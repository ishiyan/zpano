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

/// Enumerates the outputs of the True Strength Index indicator.
pub const TrueStrengthIndexOutput = enum(u8) {
    /// True Strength Index oscillator value (range [-100, +100]).
    tsi = 1,
    /// Signal-line value: the ul-period EMA of the oscillator.
    signal = 2,
};

/// Parameters to create a True Strength Index indicator.
///
/// The field names q, r, s, u and ul are the canonical symbols from William
/// Blau's Momentum, Direction, and Divergence (Wiley, 1995), chapter 2.
pub const TrueStrengthIndexParams = struct {
    q: usize = 2,
    r: usize = 20,
    s: usize = 5,
    u: usize = 3,
    ul: usize = 3,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Stateful streaming EMA: alpha = 2/(period+1), seeds e0 = x0.
///
/// Inlined verbatim from the Blau exponential moving average so the indicator is a
/// standalone porting unit. Do NOT change its numerics.
///
/// period == 1 -> alpha == 1 -> pure passthrough (output == input).
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

/// True Strength Index (TSI) by William Blau.
///
/// A double-/triple-smoothed momentum oscillator bounded to [-100, +100], paired
/// with an EMA signal line (the Ergodic form, Blau ch.1.4):
///
///   tsi_k    = 100 * TEMA(mtm, r, s, u) / TEMA(|mtm|, r, s, u)   (the oscillator)
///   signal_k = EMA(tsi, ul)_k                                    (ul-period EMA)
///
/// where mtm_k = C_k - C_(k-(q-1)) and TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u).
///
/// The indicator produces two outputs:
///   - TSI: the oscillator, range [-100, +100], NaN during warm-up (bars 0..q-2);
///   - Signal: the ul-period EMA of the oscillator (Blau's Ergodic signal line).
pub const TrueStrengthIndex = struct {
    q: usize,

    history: []f64,
    history_len: usize,

    num_r: Ema,
    num_s: Ema,
    num_u: Ema,
    den_r: Ema,
    den_s: Ema,
    den_u: Ema,

    signal_ema: Ema,

    primed: bool,

    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,

    allocator: std.mem.Allocator,
    mnemonic_buf: [128]u8,
    mnemonic_len: usize,
    description_buf: [192]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: TrueStrengthIndexParams) !TrueStrengthIndex {
        const q = params.q;
        const r = params.r;
        const s = params.s;
        const u = params.u;
        const ul = params.ul;

        if (q < 1) return error.InvalidQ;
        if (r < 1) return error.InvalidR;
        if (s < 1) return error.InvalidS;
        if (u < 1) return error.InvalidU;
        if (ul < 1) return error.InvalidUl;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        const history = try allocator.alloc(f64, q);
        errdefer allocator.free(history);

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [128]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "tsi({d},{d},{d},{d}{s})", .{
            q,
            r,
            s,
            u,
            triple,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [192]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "True Strength Index {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        return .{
            .q = q,
            .history = history,
            .history_len = 0,
            .num_r = Ema.init(r),
            .num_s = Ema.init(s),
            .num_u = Ema.init(u),
            .den_r = Ema.init(r),
            .den_s = Ema.init(s),
            .den_u = Ema.init(u),
            .signal_ema = Ema.init(ul),
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

    pub fn deinit(self: *TrueStrengthIndex) void {
        self.allocator.free(self.history);
    }

    pub fn fixSlices(self: *TrueStrengthIndex) void {
        _ = self;
        // TSI doesn't use LineIndicator; mnemonic/description are read from buffers
        // directly and the history points to heap memory, so no slice fixup is needed.
    }

    /// Returns tsi, signal.
    pub fn updateValues(self: *TrueStrengthIndex, sample: f64) struct { tsi: f64, signal: f64 } {
        // Maintain a rolling window of the last q prices; the leftmost element is
        // C_(k-(q-1)).
        if (self.history_len < self.q) {
            self.history[self.history_len] = sample;
            self.history_len += 1;
        } else {
            var i: usize = 1;
            while (i < self.q) : (i += 1) {
                self.history[i - 1] = self.history[i];
            }
            self.history[self.q - 1] = sample;
        }

        // Momentum needs a price from q-1 bars ago, available only once the window
        // holds q prices. Before then neither output is defined and the signal EMA
        // is NOT advanced.
        if (self.history_len < self.q) {
            return .{ .tsi = math.nan(f64), .signal = math.nan(f64) };
        }

        // mtm_k = C_k - C_(k-(q-1)); the leftmost history element is C_(k-(q-1)).
        const mtm = sample - self.history[0];
        const abs_mtm = @abs(mtm);

        // Numerator cascade: TEMA(mtm, r, s, u).
        const n = self.num_u.update(self.num_s.update(self.num_r.update(mtm)));
        // Denominator cascade: TEMA(|mtm|, r, s, u).
        const d = self.den_u.update(self.den_s.update(self.den_r.update(abs_mtm)));

        // Division guard (Blau_TSI.mq5): denominator 0 -> oscillator 0.0.
        const tsi: f64 = if (d == 0.0) 0.0 else 100.0 * n / d;

        // Signal line = EMA(tsi, ul); seeds here on the first finite oscillator.
        const signal = self.signal_ema.update(tsi);
        self.primed = true;

        return .{ .tsi = tsi, .signal = signal };
    }

    pub fn isPrimed(self: *const TrueStrengthIndex) bool {
        return self.primed;
    }

    fn mnemonic(self: *const TrueStrengthIndex) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const TrueStrengthIndex) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    pub fn getMetadata(self: *const TrueStrengthIndex, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();

        var tsi_mn_buf: [160]u8 = undefined;
        const tsi_mn = std.fmt.bufPrint(&tsi_mn_buf, "{s} tsi", .{mn}) catch mn;
        var signal_mn_buf: [160]u8 = undefined;
        const signal_mn = std.fmt.bufPrint(&signal_mn_buf, "{s} signal", .{mn}) catch mn;

        var tsi_desc_buf: [256]u8 = undefined;
        const tsi_desc = std.fmt.bufPrint(&tsi_desc_buf, "{s} TSI", .{desc}) catch desc;
        var signal_desc_buf: [256]u8 = undefined;
        const signal_desc = std.fmt.bufPrint(&signal_desc_buf, "{s} signal", .{desc}) catch desc;

        build_metadata_mod.buildMetadata(
            out,
            .true_strength_index,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = tsi_mn, .description = tsi_desc },
                .{ .mnemonic = signal_mn, .description = signal_desc },
            },
        );
    }

    pub fn updateScalar(self: *TrueStrengthIndex, sample: *const Scalar) OutputArray {
        const result = self.updateValues(sample.value);
        return makeOutput(sample.time, result.tsi, result.signal);
    }

    pub fn updateBar(self: *TrueStrengthIndex, sample: *const Bar) OutputArray {
        const v = self.bar_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    pub fn updateQuote(self: *TrueStrengthIndex, sample: *const Quote) OutputArray {
        const v = self.quote_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    pub fn updateTrade(self: *TrueStrengthIndex, sample: *const Trade) OutputArray {
        const v = self.trade_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    fn makeOutput(time: i64, tsi_v: f64, signal_v: f64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = tsi_v } });
        out.append(.{ .scalar = .{ .time = time, .value = signal_v } });
        return out;
    }

    pub fn indicator(self: *TrueStrengthIndex) indicator_mod.Indicator {
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
        const self: *TrueStrengthIndex = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const TrueStrengthIndex = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *TrueStrengthIndex = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *TrueStrengthIndex = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *TrueStrengthIndex = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *TrueStrengthIndex = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidQ,
        InvalidR,
        InvalidS,
        InvalidU,
        InvalidUl,
        MnemonicTooLong,
        OutOfMemory,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");

// signal_ul is the signal-line EMA period used for every expected signal array.
const signal_ul = 3;

const Combo = struct {
    name: []const u8,
    q: usize,
    r: usize,
    s: usize,
    u: usize,
    tsi: [252]f64,
    signal: [252]f64,
};

fn checkVal(exp: f64, act: f64, tolerance: f64) !void {
    if (math.isNan(exp)) {
        try testing.expect(math.isNan(act));
        return;
    }
    try testing.expect(@abs(act - exp) <= tolerance);
}

test "TSI reference data all combos" {
    const allocator = testing.allocator;
    const tolerance = 1e-9;
    const input = testdata.testInput();

    const combos = [_]Combo{
        .{ .name = "Q2_R20_S5_U3", .q = 2, .r = 20, .s = 5, .u = 3, .tsi = testdata.expectedQ2_R20_S5_U3(), .signal = testdata.expectedQ2_R20_S5_U3_SIG_UL3() },
        .{ .name = "Q2_R25_S13_U1", .q = 2, .r = 25, .s = 13, .u = 1, .tsi = testdata.expectedQ2_R25_S13_U1(), .signal = testdata.expectedQ2_R25_S13_U1_SIG_UL3() },
        .{ .name = "Q2_R20_S5_U1", .q = 2, .r = 20, .s = 5, .u = 1, .tsi = testdata.expectedQ2_R20_S5_U1(), .signal = testdata.expectedQ2_R20_S5_U1_SIG_UL3() },
        .{ .name = "Q2_R32_S5_U1", .q = 2, .r = 32, .s = 5, .u = 1, .tsi = testdata.expectedQ2_R32_S5_U1(), .signal = testdata.expectedQ2_R32_S5_U1_SIG_UL3() },
        .{ .name = "Q2_R13_S13_U1", .q = 2, .r = 13, .s = 13, .u = 1, .tsi = testdata.expectedQ2_R13_S13_U1(), .signal = testdata.expectedQ2_R13_S13_U1_SIG_UL3() },
        .{ .name = "Q2_R20_S40_U1", .q = 2, .r = 20, .s = 40, .u = 1, .tsi = testdata.expectedQ2_R20_S40_U1(), .signal = testdata.expectedQ2_R20_S40_U1_SIG_UL3() },
        .{ .name = "Q2_R40_S20_U1", .q = 2, .r = 40, .s = 20, .u = 1, .tsi = testdata.expectedQ2_R40_S20_U1(), .signal = testdata.expectedQ2_R40_S20_U1_SIG_UL3() },
        .{ .name = "Q2_R64_S64_U1", .q = 2, .r = 64, .s = 64, .u = 1, .tsi = testdata.expectedQ2_R64_S64_U1(), .signal = testdata.expectedQ2_R64_S64_U1_SIG_UL3() },
        .{ .name = "Q2_R100_S5_U1", .q = 2, .r = 100, .s = 5, .u = 1, .tsi = testdata.expectedQ2_R100_S5_U1(), .signal = testdata.expectedQ2_R100_S5_U1_SIG_UL3() },
        .{ .name = "Q2_R1_S1_U1", .q = 2, .r = 1, .s = 1, .u = 1, .tsi = testdata.expectedQ2_R1_S1_U1(), .signal = testdata.expectedQ2_R1_S1_U1_SIG_UL3() },
        .{ .name = "Q2_R1_S5_U3", .q = 2, .r = 1, .s = 5, .u = 3, .tsi = testdata.expectedQ2_R1_S5_U3(), .signal = testdata.expectedQ2_R1_S5_U3_SIG_UL3() },
        .{ .name = "Q2_R20_S1_U1", .q = 2, .r = 20, .s = 1, .u = 1, .tsi = testdata.expectedQ2_R20_S1_U1(), .signal = testdata.expectedQ2_R20_S1_U1_SIG_UL3() },
        .{ .name = "Q2_R5_S5_U5", .q = 2, .r = 5, .s = 5, .u = 5, .tsi = testdata.expectedQ2_R5_S5_U5(), .signal = testdata.expectedQ2_R5_S5_U5_SIG_UL3() },
        .{ .name = "Q3_R20_S5_U3", .q = 3, .r = 20, .s = 5, .u = 3, .tsi = testdata.expectedQ3_R20_S5_U3(), .signal = testdata.expectedQ3_R20_S5_U3_SIG_UL3() },
        .{ .name = "Q5_R20_S5_U3", .q = 5, .r = 20, .s = 5, .u = 3, .tsi = testdata.expectedQ5_R20_S5_U3(), .signal = testdata.expectedQ5_R20_S5_U3_SIG_UL3() },
        .{ .name = "Q10_R20_S5_U1", .q = 10, .r = 20, .s = 5, .u = 1, .tsi = testdata.expectedQ10_R20_S5_U1(), .signal = testdata.expectedQ10_R20_S5_U1_SIG_UL3() },
        .{ .name = "Q2_R9_S3_U1", .q = 2, .r = 9, .s = 3, .u = 1, .tsi = testdata.expectedQ2_R9_S3_U1(), .signal = testdata.expectedQ2_R9_S3_U1_SIG_UL3() },
        .{ .name = "Q2_R7_S4_U2", .q = 2, .r = 7, .s = 4, .u = 2, .tsi = testdata.expectedQ2_R7_S4_U2(), .signal = testdata.expectedQ2_R7_S4_U2_SIG_UL3() },
    };

    for (combos) |combo| {
        var ind = try TrueStrengthIndex.init(allocator, .{
            .q = combo.q,
            .r = combo.r,
            .s = combo.s,
            .u = combo.u,
            .ul = signal_ul,
        });
        defer ind.deinit();

        for (0..252) |i| {
            const result = ind.updateValues(input[i]);
            try checkVal(combo.tsi[i], result.tsi, tolerance);
            try checkVal(combo.signal[i], result.signal, tolerance);
        }
    }
}

test "TSI metadata default" {
    const allocator = testing.allocator;

    var ind = try TrueStrengthIndex.init(allocator, .{});
    defer ind.deinit();

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqual(Identifier.true_strength_index, meta.identifier);
    try testing.expectEqualStrings("tsi(2,20,5,3)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 2), meta.outputs_len);
    try testing.expectEqual(@as(u8, 1), meta.outputs_buf[0].kind);
    try testing.expectEqual(@as(u8, 2), meta.outputs_buf[1].kind);
}

test "TSI custom mnemonic excludes ul" {
    const allocator = testing.allocator;

    var ind = try TrueStrengthIndex.init(allocator, .{ .q = 2, .r = 25, .s = 13, .u = 1, .ul = 7 });
    defer ind.deinit();

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqualStrings("tsi(2,25,13,1)", meta.mnemonic);
}

test "TSI invalid params" {
    const allocator = testing.allocator;

    const r1 = TrueStrengthIndex.init(allocator, .{ .q = 0 });
    try testing.expect(if (r1) |_| false else |_| true);

    const r2 = TrueStrengthIndex.init(allocator, .{ .r = 0 });
    try testing.expect(if (r2) |_| false else |_| true);

    const r3 = TrueStrengthIndex.init(allocator, .{ .s = 0 });
    try testing.expect(if (r3) |_| false else |_| true);

    const r4 = TrueStrengthIndex.init(allocator, .{ .u = 0 });
    try testing.expect(if (r4) |_| false else |_| true);

    const r5 = TrueStrengthIndex.init(allocator, .{ .ul = 0 });
    try testing.expect(if (r5) |_| false else |_| true);
}

test "TSI entity update ordering" {
    const allocator = testing.allocator;
    const tolerance = 1e-9;
    const input = testdata.testInput();
    const exp_tsi = testdata.expectedQ2_R20_S5_U3();
    const exp_signal = testdata.expectedQ2_R20_S5_U3_SIG_UL3();

    var ind = try TrueStrengthIndex.init(allocator, .{});
    defer ind.deinit();

    var last_out: OutputArray = undefined;
    for (0..252) |i| {
        const scalar = Scalar{ .time = 0, .value = input[i] };
        last_out = ind.updateScalar(&scalar);
    }
    const items = last_out.slice();

    try checkVal(exp_tsi[251], items[0].scalar.value, tolerance);
    try checkVal(exp_signal[251], items[1].scalar.value, tolerance);
}
