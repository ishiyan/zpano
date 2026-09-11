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

/// Enumerates the outputs of the MACD Index indicator.
pub const MacdIndexOutput = enum(u8) {
    /// The MACD Index line value, in raw price units (unbounded).
    macdi = 1,
    /// Signal-line value: the ul-period EMA of the index.
    signal = 2,
};

/// Parameters to create a MACD Index indicator.
///
/// The field names r, s, u and ul are the canonical symbols from William Blau's
/// Momentum, Direction, and Divergence (Wiley, 1995), chapter 5.
pub const MacdIndexParams = struct {
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

/// MACD Index (MACD_I) by William Blau.
///
/// Blau's MACD line is the difference of two EMAs of the close, optionally
/// smoothed by a third EMA, paired with an EMA signal line (the Ergodic form,
/// Blau ch. 5):
///
///   macd_k   = EMA(close, s)_k - EMA(close, r)_k          (MACD line; s fast, r slow)
///   macdi_k  = EMA(macd, u)_k                             (the MACD_I line)
///   signal_k = EMA(macdi, ul)_k                           (ul-period EMA)
///
/// with the fast period s strictly shorter than the slow period r (s < r).
/// Setting u=1 recovers the book's pure two-EMA MACD line. Blau notes the MACD
/// and the MDI are both double-smoothed momentum indicators with nearly
/// interchangeable shapes (within a scale factor).
///
/// The index is NOT normalized: there is no 100 * TEMA/TEMA ratio and no fixed
/// range, so the output is in the same price units as the input and may take any
/// sign or magnitude. Because there is no division there is also no division guard.
///
/// The indicator produces two outputs:
///   - MACDI: the index line, in raw price units, finite from bar 0;
///   - Signal: the ul-period EMA of the index (Blau's Ergodic signal line).
///
/// Priming: both price EMAs are defined from bar 0, so macd_0 = 0 and the u
/// smoothing and signal EMAs seed on that 0 -- there is no NaN warm-up region
/// and bar 0 is exactly 0.0.
pub const MacdIndex = struct {
    ema_fast: Ema,
    ema_slow: Ema,
    smooth_u: Ema,

    signal_ema: Ema,

    primed: bool,

    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,

    mnemonic_buf: [128]u8,
    mnemonic_len: usize,
    description_buf: [192]u8,
    description_len: usize,

    pub fn init(params: MacdIndexParams) !MacdIndex {
        const r = params.r;
        const s = params.s;
        const u = params.u;
        const ul = params.ul;

        if (r < 1) return error.InvalidR;
        if (s < 1) return error.InvalidS;
        if (u < 1) return error.InvalidU;
        if (ul < 1) return error.InvalidUl;
        if (s >= r) return error.InvalidS;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [128]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "macdi({d},{d},{d}{s})", .{
            r,
            s,
            u,
            triple,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [192]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "MACD Index {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        return .{
            .ema_fast = Ema.init(s),
            .ema_slow = Ema.init(r),
            .smooth_u = Ema.init(u),
            .signal_ema = Ema.init(ul),
            .primed = false,
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn fixSlices(self: *MacdIndex) void {
        _ = self;
        // MacdIndex doesn't use LineIndicator; mnemonic/description are read from
        // owned buffers directly and the indicator owns no heap memory, so no
        // slice fixup is needed.
    }

    /// Returns macdi, signal.
    pub fn updateValues(self: *MacdIndex, sample: f64) struct { macdi: f64, signal: f64 } {
        // MACD line = fast EMA - slow EMA. Both seed at bar 0 (to close_0), so
        // the line is 0.0 on bar 0 and defined on every bar thereafter.
        const macd = self.ema_fast.update(sample) - self.ema_slow.update(sample);

        // Smooth the MACD line: EMA(macd, u). No normalization, no guard.
        const macdi = self.smooth_u.update(macd);

        // Signal line = EMA(macdi, ul); seeds here on the bar-0 index value.
        const signal = self.signal_ema.update(macdi);
        self.primed = true;

        return .{ .macdi = macdi, .signal = signal };
    }

    pub fn isPrimed(self: *const MacdIndex) bool {
        return self.primed;
    }

    fn mnemonic(self: *const MacdIndex) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const MacdIndex) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    pub fn getMetadata(self: *const MacdIndex, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();

        var macdi_mn_buf: [160]u8 = undefined;
        const macdi_mn = std.fmt.bufPrint(&macdi_mn_buf, "{s} macdi", .{mn}) catch mn;
        var signal_mn_buf: [160]u8 = undefined;
        const signal_mn = std.fmt.bufPrint(&signal_mn_buf, "{s} signal", .{mn}) catch mn;

        var macdi_desc_buf: [256]u8 = undefined;
        const macdi_desc = std.fmt.bufPrint(&macdi_desc_buf, "{s} MACDI", .{desc}) catch desc;
        var signal_desc_buf: [256]u8 = undefined;
        const signal_desc = std.fmt.bufPrint(&signal_desc_buf, "{s} signal", .{desc}) catch desc;

        build_metadata_mod.buildMetadata(
            out,
            .macd_index,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = macdi_mn, .description = macdi_desc },
                .{ .mnemonic = signal_mn, .description = signal_desc },
            },
        );
    }

    pub fn updateScalar(self: *MacdIndex, sample: *const Scalar) OutputArray {
        const result = self.updateValues(sample.value);
        return makeOutput(sample.time, result.macdi, result.signal);
    }

    pub fn updateBar(self: *MacdIndex, sample: *const Bar) OutputArray {
        const v = self.bar_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    pub fn updateQuote(self: *MacdIndex, sample: *const Quote) OutputArray {
        const v = self.quote_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    pub fn updateTrade(self: *MacdIndex, sample: *const Trade) OutputArray {
        const v = self.trade_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    fn makeOutput(time: i64, macdi_v: f64, signal_v: f64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = macdi_v } });
        out.append(.{ .scalar = .{ .time = time, .value = signal_v } });
        return out;
    }

    pub fn indicator(self: *MacdIndex) indicator_mod.Indicator {
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
        const self: *MacdIndex = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const MacdIndex = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *MacdIndex = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *MacdIndex = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *MacdIndex = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *MacdIndex = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidR,
        InvalidS,
        InvalidU,
        InvalidUl,
        MnemonicTooLong,
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
    r: usize,
    s: usize,
    u: usize,
    macdi: [252]f64,
    signal: [252]f64,
};

fn checkVal(exp: f64, act: f64, tolerance: f64) !void {
    try testing.expect(@abs(act - exp) <= tolerance);
}

test "MACDI reference data all combos" {
    const tolerance = 1e-9;
    const input = testdata.testInput();

    const combos = [_]Combo{
        .{ .name = "R20_S5_U3", .r = 20, .s = 5, .u = 3, .macdi = testdata.expectedR20_S5_U3(), .signal = testdata.expectedR20_S5_U3_SIG_UL3() },
        .{ .name = "R20_S5_U1", .r = 20, .s = 5, .u = 1, .macdi = testdata.expectedR20_S5_U1(), .signal = testdata.expectedR20_S5_U1_SIG_UL3() },
        .{ .name = "R26_S12_U3", .r = 26, .s = 12, .u = 3, .macdi = testdata.expectedR26_S12_U3(), .signal = testdata.expectedR26_S12_U3_SIG_UL3() },
        .{ .name = "R26_S12_U1", .r = 26, .s = 12, .u = 1, .macdi = testdata.expectedR26_S12_U1(), .signal = testdata.expectedR26_S12_U1_SIG_UL3() },
        .{ .name = "R35_S5_U3", .r = 35, .s = 5, .u = 3, .macdi = testdata.expectedR35_S5_U3(), .signal = testdata.expectedR35_S5_U3_SIG_UL3() },
        .{ .name = "R10_S3_U5", .r = 10, .s = 3, .u = 5, .macdi = testdata.expectedR10_S3_U5(), .signal = testdata.expectedR10_S3_U5_SIG_UL3() },
        .{ .name = "R32_S12_U5", .r = 32, .s = 12, .u = 5, .macdi = testdata.expectedR32_S12_U5(), .signal = testdata.expectedR32_S12_U5_SIG_UL3() },
        .{ .name = "R17_S8_U1", .r = 17, .s = 8, .u = 1, .macdi = testdata.expectedR17_S8_U1(), .signal = testdata.expectedR17_S8_U1_SIG_UL3() },
        .{ .name = "R20_S10_U3", .r = 20, .s = 10, .u = 3, .macdi = testdata.expectedR20_S10_U3(), .signal = testdata.expectedR20_S10_U3_SIG_UL3() },
        .{ .name = "R8_S4_U2", .r = 8, .s = 4, .u = 2, .macdi = testdata.expectedR8_S4_U2(), .signal = testdata.expectedR8_S4_U2_SIG_UL3() },
        .{ .name = "R30_S15_U1", .r = 30, .s = 15, .u = 1, .macdi = testdata.expectedR30_S15_U1(), .signal = testdata.expectedR30_S15_U1_SIG_UL3() },
        .{ .name = "R3_S2_U3", .r = 3, .s = 2, .u = 3, .macdi = testdata.expectedR3_S2_U3(), .signal = testdata.expectedR3_S2_U3_SIG_UL3() },
        .{ .name = "R50_S12_U1", .r = 50, .s = 12, .u = 1, .macdi = testdata.expectedR50_S12_U1(), .signal = testdata.expectedR50_S12_U1_SIG_UL3() },
        .{ .name = "R19_S6_U3", .r = 19, .s = 6, .u = 3, .macdi = testdata.expectedR19_S6_U3(), .signal = testdata.expectedR19_S6_U3_SIG_UL3() },
        .{ .name = "R20_S5_U5", .r = 20, .s = 5, .u = 5, .macdi = testdata.expectedR20_S5_U5(), .signal = testdata.expectedR20_S5_U5_SIG_UL3() },
        .{ .name = "R60_S30_U10", .r = 60, .s = 30, .u = 10, .macdi = testdata.expectedR60_S30_U10(), .signal = testdata.expectedR60_S30_U10_SIG_UL3() },
    };

    for (combos) |combo| {
        var ind = try MacdIndex.init(.{
            .r = combo.r,
            .s = combo.s,
            .u = combo.u,
            .ul = signal_ul,
        });

        for (0..252) |i| {
            const result = ind.updateValues(input[i]);
            try checkVal(combo.macdi[i], result.macdi, tolerance);
            try checkVal(combo.signal[i], result.signal, tolerance);
        }
    }
}

test "MACDI has no warm-up region" {
    const input = testdata.testInput();

    var ind = try MacdIndex.init(.{});

    const first = ind.updateValues(input[0]);
    try testing.expectEqual(@as(f64, 0.0), first.macdi);
    try testing.expectEqual(@as(f64, 0.0), first.signal);

    for (1..252) |i| {
        const result = ind.updateValues(input[i]);
        try testing.expect(!math.isNan(result.macdi));
        try testing.expect(!math.isNan(result.signal));
    }
}

test "MACDI passthrough smoothing equals fast EMA minus slow EMA" {
    const tolerance = 1e-12;

    var ind = try MacdIndex.init(.{ .r = 2, .s = 1, .u = 1, .ul = 1 });

    // Bar 0 seeds both price EMAs, so the MACD line is exactly 0.
    const first = ind.updateValues(10.0);
    try checkVal(0.0, first.macdi, tolerance);
    try checkVal(0.0, first.signal, tolerance);

    // EMA(1) is a passthrough -> fast = 12. EMA(2) = (2/3)*12 + (1/3)*10 = 11.3333.
    const second = ind.updateValues(12.0);
    try checkVal(12.0 - (2.0 / 3.0 * 12.0 + 1.0 / 3.0 * 10.0), second.macdi, tolerance);
    try checkVal(second.macdi, second.signal, tolerance);
}

test "MACDI signal passthrough when ul = 1" {
    const input = testdata.testInput();

    var ind = try MacdIndex.init(.{ .r = 20, .s = 5, .u = 3, .ul = 1 });

    for (0..252) |i| {
        const result = ind.updateValues(input[i]);
        try testing.expectEqual(result.macdi, result.signal);
    }
}

test "MACDI is primed after the first update" {
    const input = testdata.testInput();

    var ind = try MacdIndex.init(.{});

    try testing.expect(!ind.isPrimed());
    _ = ind.updateValues(input[0]);
    try testing.expect(ind.isPrimed());
}

test "MACDI metadata default" {
    var ind = try MacdIndex.init(.{});

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqual(Identifier.macd_index, meta.identifier);
    try testing.expectEqualStrings("macdi(20,5,3)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 2), meta.outputs_len);
    try testing.expectEqual(@as(u8, 1), meta.outputs_buf[0].kind);
    try testing.expectEqual(@as(u8, 2), meta.outputs_buf[1].kind);
}

test "MACDI custom mnemonic excludes ul" {
    var ind = try MacdIndex.init(.{ .r = 26, .s = 12, .u = 9, .ul = 7 });

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqualStrings("macdi(26,12,9)", meta.mnemonic);
}

test "MACDI invalid params" {
    const r1 = MacdIndex.init(.{ .r = 0 });
    try testing.expect(if (r1) |_| false else |_| true);

    const r2 = MacdIndex.init(.{ .s = 0 });
    try testing.expect(if (r2) |_| false else |_| true);

    const r3 = MacdIndex.init(.{ .u = 0 });
    try testing.expect(if (r3) |_| false else |_| true);

    const r4 = MacdIndex.init(.{ .ul = 0 });
    try testing.expect(if (r4) |_| false else |_| true);

    const r5 = MacdIndex.init(.{ .r = 5, .s = 5 });
    try testing.expect(if (r5) |_| false else |_| true);

    const r6 = MacdIndex.init(.{ .r = 5, .s = 6 });
    try testing.expect(if (r6) |_| false else |_| true);
}

test "MACDI entity update ordering" {
    const tolerance = 1e-9;
    const input = testdata.testInput();
    const exp_macdi = testdata.expectedR20_S5_U3();
    const exp_signal = testdata.expectedR20_S5_U3_SIG_UL3();

    var ind = try MacdIndex.init(.{});

    var last_out: OutputArray = undefined;
    for (0..252) |i| {
        const scalar = Scalar{ .time = 0, .value = input[i] };
        last_out = ind.updateScalar(&scalar);
    }
    const items = last_out.slice();

    try checkVal(exp_macdi[251], items[0].scalar.value, tolerance);
    try checkVal(exp_signal[251], items[1].scalar.value, tolerance);
}
