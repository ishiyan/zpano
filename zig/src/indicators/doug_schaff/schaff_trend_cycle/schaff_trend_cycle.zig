const std = @import("std");
const math = std.math;

const entities = @import("entities");
const Bar = entities.Bar;
const Quote = entities.Quote;
const Trade = entities.Trade;
const Scalar = entities.Scalar;
const bar_component = entities.bar_component;
const quote_component = entities.quote_component;
const trade_component = entities.trade_component;
const indicator_mod = @import("../../core/indicator.zig");
const build_metadata_mod = @import("../../core/build_metadata.zig");
const component_triple_mnemonic_mod = @import("../../core/component_triple_mnemonic.zig");
const identifier_mod = @import("../../core/identifier.zig");
const metadata_mod = @import("../../core/metadata.zig");

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the Schaff Trend Cycle indicator.
pub const SchaffTrendCycleOutput = enum(u8) {
    /// STC oscillator value (range [0, 100]).
    stc = 1,
    /// Gated MACD line (XMAC) value.
    macd = 2,
    /// First smoothed %D stage (PF) value.
    pf = 3,
};

/// Parameters to create a Schaff Trend Cycle indicator.
pub const SchaffTrendCycleParams = struct {
    fast: usize = 23,
    slow: usize = 50,
    tclen: usize = 10,
    factor: f64 = 0.5,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Stateful streaming EMA: alpha = 2/(period+1), seeds e0 = x0.
///
/// Inlined verbatim from the Blau exponential moving average so the indicator is a
/// standalone porting unit. Do NOT change its numerics.
const Ema = struct {
    alpha: f64,
    prev: f64 = 0.0,
    primed: bool = false,

    fn init(period: usize) Ema {
        return .{ .alpha = 2.0 / (@as(f64, @floatFromInt(period)) + 1.0) };
    }

    fn update(self: *Ema, x: f64) f64 {
        if (!self.primed) {
            self.prev = x;
            self.primed = true;
            return self.prev;
        }
        self.prev = self.alpha * x + (1.0 - self.alpha) * self.prev;
        return self.prev;
    }
};

/// A fixed-capacity ring buffer of the last n values, providing min/max.
const Window = struct {
    data: []f64,
    pos: usize = 0,
    count: usize = 0,

    fn push(self: *Window, v: f64) void {
        self.data[self.pos] = v;
        self.pos = (self.pos + 1) % self.data.len;
        if (self.count < self.data.len) {
            self.count += 1;
        }
    }

    fn minMax(self: *const Window) struct { min: f64, max: f64 } {
        var min_val = self.data[0];
        var max_val = self.data[0];
        var i: usize = 1;
        while (i < self.count) : (i += 1) {
            const v = self.data[i];
            if (v < min_val) min_val = v;
            if (v > max_val) max_val = v;
        }
        return .{ .min = min_val, .max = max_val };
    }
};

/// Schaff Trend Cycle (STC) by Doug Schaff.
///
/// STC runs a MACD line through two cascaded stochastics, each followed by an
/// EMA-style smoothing, producing a cyclical oscillator bounded to [0, 100].
///
/// The indicator produces three outputs:
///   - STC: the oscillator, range [0, 100], NaN during warm-up (bars 0..slow);
///   - MACD: the gated MACD line XMAC (0.0 pre-gate), exposed for stage testing;
///   - PF: the first smoothed %D (0.0 pre-gate), exposed for stage testing.
pub const SchaffTrendCycle = struct {
    ema_fast: Ema,
    ema_slow: Ema,

    slow: usize,
    tclen: usize,
    factor: f64,

    bar: i64,

    macd_win: Window,
    pf_win: Window,

    frac1: f64,
    frac2: f64,
    pf: f64,
    pff: f64,

    primed: bool,

    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,

    allocator: std.mem.Allocator,
    mnemonic_buf: [128]u8,
    mnemonic_len: usize,
    description_buf: [192]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: SchaffTrendCycleParams) !SchaffTrendCycle {
        const fast = params.fast;
        const slow = params.slow;
        const tclen = params.tclen;
        const factor = params.factor;

        if (fast < 1) return error.InvalidFast;
        if (slow < 1) return error.InvalidSlow;
        if (tclen < 1) return error.InvalidTclen;
        if (factor <= 0.0 or factor > 1.0) return error.InvalidFactor;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        const macd_data = try allocator.alloc(f64, tclen);
        errdefer allocator.free(macd_data);
        const pf_data = try allocator.alloc(f64, tclen);
        errdefer allocator.free(pf_data);

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [128]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "stc({d},{d},{d},{d:.2}{s})", .{
            fast,
            slow,
            tclen,
            factor,
            triple,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [192]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Schaff Trend Cycle {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        return .{
            .ema_fast = Ema.init(fast),
            .ema_slow = Ema.init(slow),
            .slow = slow,
            .tclen = tclen,
            .factor = factor,
            .bar = -1,
            .macd_win = .{ .data = macd_data },
            .pf_win = .{ .data = pf_data },
            .frac1 = 0.0,
            .frac2 = 0.0,
            .pf = 0.0,
            .pff = 0.0,
            .primed = false,
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *SchaffTrendCycle) void {
        self.allocator.free(self.macd_win.data);
        self.allocator.free(self.pf_win.data);
    }

    pub fn fixSlices(self: *SchaffTrendCycle) void {
        _ = self;
        // STC doesn't use LineIndicator; mnemonic/description are read from buffers
        // directly and the windows point to heap memory, so no slice fixup is needed.
    }

    /// Returns stc, macd, pf.
    pub fn updateValues(self: *SchaffTrendCycle, sample: f64) struct { stc: f64, macd: f64, pf: f64 } {
        self.bar += 1;
        const k = self.bar;

        // Price EMAs always advance (they accumulate over the full history).
        const ema_fast = self.ema_fast.update(sample);
        const ema_slow = self.ema_slow.update(sample);

        // GATE: XMAC is only assigned while barindex > slow.
        const gate_open = k > @as(i64, @intCast(self.slow));
        const macd: f64 = if (gate_open) ema_fast - ema_slow else 0.0;
        self.macd_win.push(macd);

        if (!gate_open) {
            self.pf_win.push(self.pf);
            return .{ .stc = math.nan(f64), .macd = macd, .pf = self.pf };
        }

        // 1st stochastic of the MACD over tclen (guard on the range).
        const mm1 = self.macd_win.minMax();
        const rng1 = mm1.max - mm1.min;
        if (rng1 > 0.0) {
            self.frac1 = ((macd - mm1.min) / rng1) * 100.0;
        }

        // 1st smoothing: PF = EMA(Frac1, alpha=factor), seed 0.
        self.pf += self.factor * (self.frac1 - self.pf);
        self.pf_win.push(self.pf);

        // 2nd stochastic of PF over tclen.
        const mm2 = self.pf_win.minMax();
        const rng2 = mm2.max - mm2.min;
        if (rng2 > 0.0) {
            self.frac2 = ((self.pf - mm2.min) / rng2) * 100.0;
        }

        // 2nd smoothing: STC = PFF = EMA(Frac2, alpha=factor), seed 0.
        self.pff += self.factor * (self.frac2 - self.pff);
        self.primed = true;

        return .{ .stc = self.pff, .macd = macd, .pf = self.pf };
    }

    pub fn isPrimed(self: *const SchaffTrendCycle) bool {
        return self.primed;
    }

    fn mnemonic(self: *const SchaffTrendCycle) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const SchaffTrendCycle) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    pub fn getMetadata(self: *const SchaffTrendCycle, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();

        var stc_mn_buf: [160]u8 = undefined;
        const stc_mn = std.fmt.bufPrint(&stc_mn_buf, "{s} stc", .{mn}) catch mn;
        var macd_mn_buf: [160]u8 = undefined;
        const macd_mn = std.fmt.bufPrint(&macd_mn_buf, "{s} macd", .{mn}) catch mn;
        var pf_mn_buf: [160]u8 = undefined;
        const pf_mn = std.fmt.bufPrint(&pf_mn_buf, "{s} pf", .{mn}) catch mn;

        var stc_desc_buf: [256]u8 = undefined;
        const stc_desc = std.fmt.bufPrint(&stc_desc_buf, "{s} STC", .{desc}) catch desc;
        var macd_desc_buf: [256]u8 = undefined;
        const macd_desc = std.fmt.bufPrint(&macd_desc_buf, "{s} MACD", .{desc}) catch desc;
        var pf_desc_buf: [256]u8 = undefined;
        const pf_desc = std.fmt.bufPrint(&pf_desc_buf, "{s} PF", .{desc}) catch desc;

        build_metadata_mod.buildMetadata(
            out,
            .schaff_trend_cycle,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = stc_mn, .description = stc_desc },
                .{ .mnemonic = macd_mn, .description = macd_desc },
                .{ .mnemonic = pf_mn, .description = pf_desc },
            },
        );
    }

    pub fn updateScalar(self: *SchaffTrendCycle, sample: *const Scalar) OutputArray {
        const result = self.updateValues(sample.value);
        return makeOutput(sample.time, result.stc, result.macd, result.pf);
    }

    pub fn updateBar(self: *SchaffTrendCycle, sample: *const Bar) OutputArray {
        const v = self.bar_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    pub fn updateQuote(self: *SchaffTrendCycle, sample: *const Quote) OutputArray {
        const v = self.quote_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    pub fn updateTrade(self: *SchaffTrendCycle, sample: *const Trade) OutputArray {
        const v = self.trade_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    fn makeOutput(time: i64, stc_v: f64, macd_v: f64, pf_v: f64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = stc_v } });
        out.append(.{ .scalar = .{ .time = time, .value = macd_v } });
        out.append(.{ .scalar = .{ .time = time, .value = pf_v } });
        return out;
    }

    pub fn indicator(self: *SchaffTrendCycle) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.VTable{
        .isPrimed = vtableIsPrimed,
        .metadata = vtableMetadata,
        .updateScalar = vtableUpdateScalar,
        .updateBar = vtableUpdateBar,
        .updateQuote = vtableUpdateQuote,
        .updateTrade = vtableUpdateTrade,
    };

    fn vtableIsPrimed(ptr: *anyopaque) bool {
        const self: *SchaffTrendCycle = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const SchaffTrendCycle = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *SchaffTrendCycle = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *SchaffTrendCycle = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *SchaffTrendCycle = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *SchaffTrendCycle = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidFast,
        InvalidSlow,
        InvalidTclen,
        InvalidFactor,
        MnemonicTooLong,
        OutOfMemory,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");

const Combo = struct {
    name: []const u8,
    fast: usize,
    slow: usize,
    tclen: usize,
    factor: f64,
    stc: [252]f64,
    macd: ?[252]f64,
    pf: ?[252]f64,
};

fn checkVal(exp: f64, act: f64, tolerance: f64) !void {
    if (math.isNan(exp)) {
        try testing.expect(math.isNan(act));
        return;
    }
    try testing.expect(@abs(act - exp) <= tolerance);
}

test "STC reference data all combos" {
    const allocator = testing.allocator;
    const tolerance = 1e-9;
    const input = testdata.testInput();

    const combos = [_]Combo{
        .{ .name = "F23_S50_T10_C50", .fast = 23, .slow = 50, .tclen = 10, .factor = 0.5, .stc = testdata.expectedStcF23_S50_T10_C50(), .macd = testdata.expectedMacdF23_S50_T10_C50(), .pf = testdata.expectedPfF23_S50_T10_C50() },
        .{ .name = "F12_S26_T10_C50", .fast = 12, .slow = 26, .tclen = 10, .factor = 0.5, .stc = testdata.expectedStcF12_S26_T10_C50(), .macd = testdata.expectedMacdF12_S26_T10_C50(), .pf = testdata.expectedPfF12_S26_T10_C50() },
        .{ .name = "F5_S10_T5_C50", .fast = 5, .slow = 10, .tclen = 5, .factor = 0.5, .stc = testdata.expectedStcF5_S10_T5_C50(), .macd = testdata.expectedMacdF5_S10_T5_C50(), .pf = testdata.expectedPfF5_S10_T5_C50() },
        .{ .name = "F3_S7_T3_C50", .fast = 3, .slow = 7, .tclen = 3, .factor = 0.5, .stc = testdata.expectedStcF3_S7_T3_C50(), .macd = null, .pf = null },
        .{ .name = "F8_S21_T10_C50", .fast = 8, .slow = 21, .tclen = 10, .factor = 0.5, .stc = testdata.expectedStcF8_S21_T10_C50(), .macd = null, .pf = null },
        .{ .name = "F10_S30_T10_C50", .fast = 10, .slow = 30, .tclen = 10, .factor = 0.5, .stc = testdata.expectedStcF10_S30_T10_C50(), .macd = null, .pf = null },
        .{ .name = "F15_S40_T14_C50", .fast = 15, .slow = 40, .tclen = 14, .factor = 0.5, .stc = testdata.expectedStcF15_S40_T14_C50(), .macd = null, .pf = null },
        .{ .name = "F6_S13_T8_C60", .fast = 6, .slow = 13, .tclen = 8, .factor = 0.6, .stc = testdata.expectedStcF6_S13_T8_C60(), .macd = null, .pf = null },
        .{ .name = "F23_S50_T23_C50", .fast = 23, .slow = 50, .tclen = 23, .factor = 0.5, .stc = testdata.expectedStcF23_S50_T23_C50(), .macd = null, .pf = null },
        .{ .name = "F23_S50_T5_C50", .fast = 23, .slow = 50, .tclen = 5, .factor = 0.5, .stc = testdata.expectedStcF23_S50_T5_C50(), .macd = null, .pf = null },
        .{ .name = "F12_S26_T10_C25", .fast = 12, .slow = 26, .tclen = 10, .factor = 0.25, .stc = testdata.expectedStcF12_S26_T10_C25(), .macd = null, .pf = null },
        .{ .name = "F12_S26_T10_C80", .fast = 12, .slow = 26, .tclen = 10, .factor = 0.8, .stc = testdata.expectedStcF12_S26_T10_C80(), .macd = null, .pf = null },
        .{ .name = "F12_S26_T10_C100", .fast = 12, .slow = 26, .tclen = 10, .factor = 1.0, .stc = testdata.expectedStcF12_S26_T10_C100(), .macd = null, .pf = null },
        .{ .name = "F20_S40_T10_C50", .fast = 20, .slow = 40, .tclen = 10, .factor = 0.5, .stc = testdata.expectedStcF20_S40_T10_C50(), .macd = null, .pf = null },
    };

    for (combos) |combo| {
        var ind = try SchaffTrendCycle.init(allocator, .{
            .fast = combo.fast,
            .slow = combo.slow,
            .tclen = combo.tclen,
            .factor = combo.factor,
        });
        defer ind.deinit();

        for (0..252) |i| {
            const result = ind.updateValues(input[i]);
            try checkVal(combo.stc[i], result.stc, tolerance);
            if (combo.macd) |m| try checkVal(m[i], result.macd, tolerance);
            if (combo.pf) |p| try checkVal(p[i], result.pf, tolerance);
        }
    }
}

test "STC metadata default" {
    const allocator = testing.allocator;

    var ind = try SchaffTrendCycle.init(allocator, .{});
    defer ind.deinit();

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqual(Identifier.schaff_trend_cycle, meta.identifier);
    try testing.expectEqualStrings("stc(23,50,10,0.50)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 3), meta.outputs_len);
    try testing.expectEqual(@as(u8, 1), meta.outputs_buf[0].kind);
    try testing.expectEqual(@as(u8, 2), meta.outputs_buf[1].kind);
    try testing.expectEqual(@as(u8, 3), meta.outputs_buf[2].kind);
}

test "STC custom mnemonic" {
    const allocator = testing.allocator;

    var ind = try SchaffTrendCycle.init(allocator, .{ .fast = 12, .slow = 26, .tclen = 10, .factor = 0.25 });
    defer ind.deinit();

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqualStrings("stc(12,26,10,0.25)", meta.mnemonic);
}

test "STC invalid params" {
    const allocator = testing.allocator;

    const r1 = SchaffTrendCycle.init(allocator, .{ .fast = 0 });
    try testing.expect(if (r1) |_| false else |_| true);

    const r2 = SchaffTrendCycle.init(allocator, .{ .slow = 0 });
    try testing.expect(if (r2) |_| false else |_| true);

    const r3 = SchaffTrendCycle.init(allocator, .{ .tclen = 0 });
    try testing.expect(if (r3) |_| false else |_| true);

    const r4 = SchaffTrendCycle.init(allocator, .{ .factor = 0.0 });
    try testing.expect(if (r4) |_| false else |_| true);

    const r5 = SchaffTrendCycle.init(allocator, .{ .factor = 1.5 });
    try testing.expect(if (r5) |_| false else |_| true);
}

test "STC entity update ordering" {
    const allocator = testing.allocator;
    const tolerance = 1e-9;
    const input = testdata.testInput();
    const exp_stc = testdata.expectedStcF23_S50_T10_C50();
    const exp_macd = testdata.expectedMacdF23_S50_T10_C50();
    const exp_pf = testdata.expectedPfF23_S50_T10_C50();

    var ind = try SchaffTrendCycle.init(allocator, .{});
    defer ind.deinit();

    var last_out: OutputArray = undefined;
    for (0..252) |i| {
        const scalar = Scalar{ .time = 0, .value = input[i] };
        last_out = ind.updateScalar(&scalar);
    }
    const items = last_out.slice();

    try checkVal(exp_stc[251], items[0].scalar.value, tolerance);
    try checkVal(exp_macd[251], items[1].scalar.value, tolerance);
    try checkVal(exp_pf[251], items[2].scalar.value, tolerance);
}
