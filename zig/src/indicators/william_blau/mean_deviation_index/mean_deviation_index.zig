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

/// Enumerates the outputs of the Mean Deviation Index indicator.
pub const MeanDeviationIndexOutput = enum(u8) {
    /// The Mean Deviation Index line value, in raw price units (unbounded).
    mdi = 1,
    /// Signal-line value: the ul-period EMA of the index.
    signal = 2,
};

/// Parameters to create a Mean Deviation Index indicator.
///
/// The field names r, s, u and ul are the canonical symbols from William Blau's
/// Momentum, Direction, and Divergence (Wiley, 1995), chapter 5.
pub const MeanDeviationIndexParams = struct {
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

/// Mean Deviation Index (MDI) by William Blau.
///
/// A detrended, double-/triple-smoothed momentum line in raw price units, paired
/// with an EMA signal line (the Ergodic form, Blau ch. 5):
///
///   md_k     = price_k - EMA(price, r)_k                  (deviation from trend)
///   mdi_k    = EMA(EMA(md, s), u)_k                       (the MDI line)
///   signal_k = EMA(mdi, ul)_k                             (ul-period EMA)
///
/// The price series is detrended by subtracting its own r-period EMA, then the
/// deviation is smoothed by an s-period EMA and an optional u-period EMA. Blau
/// notes the MDI approximates the MACD when r is long and s is short.
///
/// The index is NOT normalized: there is no 100 * TEMA/TEMA ratio and no fixed
/// range, so the output is in the same price units as the input and may take any
/// sign or magnitude. Because there is no division there is also no division guard.
///
/// The indicator produces two outputs:
///   - MDI: the index line, in raw price units, finite from bar 0;
///   - Signal: the ul-period EMA of the index (Blau's Ergodic signal line).
///
/// Priming: the detrending EMA is defined from bar 0, so md_0 = 0 and both
/// smoothing EMAs seed on that 0 -- there is no NaN warm-up region and bar 0 is
/// exactly 0.0. Degenerate case: r=1 makes the detrend a passthrough, so the
/// index is identically 0.0.
pub const MeanDeviationIndex = struct {
    trend: Ema,
    smooth_s: Ema,
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

    pub fn init(params: MeanDeviationIndexParams) !MeanDeviationIndex {
        const r = params.r;
        const s = params.s;
        const u = params.u;
        const ul = params.ul;

        if (r < 1) return error.InvalidR;
        if (s < 1) return error.InvalidS;
        if (u < 1) return error.InvalidU;
        if (ul < 1) return error.InvalidUl;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [128]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "mdi({d},{d},{d}{s})", .{
            r,
            s,
            u,
            triple,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [192]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Mean Deviation Index {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        return .{
            .trend = Ema.init(r),
            .smooth_s = Ema.init(s),
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

    pub fn fixSlices(self: *MeanDeviationIndex) void {
        _ = self;
        // MDI doesn't use LineIndicator; mnemonic/description are read from owned
        // buffers directly and the indicator owns no heap memory, so no slice
        // fixup is needed.
    }

    /// Returns mdi, signal.
    pub fn updateValues(self: *MeanDeviationIndex, sample: f64) struct { mdi: f64, signal: f64 } {
        // Mean deviation: the price minus its own r-period EMA trend. The
        // baseline EMA seeds on bar 0, so the bar-0 deviation is exactly 0.
        const md = sample - self.trend.update(sample);

        // Smooth the deviation: EMA(EMA(md, s), u). No normalization, no guard.
        const mdi = self.smooth_u.update(self.smooth_s.update(md));

        // Signal line = EMA(mdi, ul); seeds here on the bar-0 index value.
        const signal = self.signal_ema.update(mdi);
        self.primed = true;

        return .{ .mdi = mdi, .signal = signal };
    }

    pub fn isPrimed(self: *const MeanDeviationIndex) bool {
        return self.primed;
    }

    fn mnemonic(self: *const MeanDeviationIndex) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const MeanDeviationIndex) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    pub fn getMetadata(self: *const MeanDeviationIndex, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();

        var mdi_mn_buf: [160]u8 = undefined;
        const mdi_mn = std.fmt.bufPrint(&mdi_mn_buf, "{s} mdi", .{mn}) catch mn;
        var signal_mn_buf: [160]u8 = undefined;
        const signal_mn = std.fmt.bufPrint(&signal_mn_buf, "{s} signal", .{mn}) catch mn;

        var mdi_desc_buf: [256]u8 = undefined;
        const mdi_desc = std.fmt.bufPrint(&mdi_desc_buf, "{s} MDI", .{desc}) catch desc;
        var signal_desc_buf: [256]u8 = undefined;
        const signal_desc = std.fmt.bufPrint(&signal_desc_buf, "{s} signal", .{desc}) catch desc;

        build_metadata_mod.buildMetadata(
            out,
            .mean_deviation_index,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = mdi_mn, .description = mdi_desc },
                .{ .mnemonic = signal_mn, .description = signal_desc },
            },
        );
    }

    pub fn updateScalar(self: *MeanDeviationIndex, sample: *const Scalar) OutputArray {
        const result = self.updateValues(sample.value);
        return makeOutput(sample.time, result.mdi, result.signal);
    }

    pub fn updateBar(self: *MeanDeviationIndex, sample: *const Bar) OutputArray {
        const v = self.bar_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    pub fn updateQuote(self: *MeanDeviationIndex, sample: *const Quote) OutputArray {
        const v = self.quote_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    pub fn updateTrade(self: *MeanDeviationIndex, sample: *const Trade) OutputArray {
        const v = self.trade_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    fn makeOutput(time: i64, mdi_v: f64, signal_v: f64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = mdi_v } });
        out.append(.{ .scalar = .{ .time = time, .value = signal_v } });
        return out;
    }

    pub fn indicator(self: *MeanDeviationIndex) indicator_mod.Indicator {
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
        const self: *MeanDeviationIndex = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const MeanDeviationIndex = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *MeanDeviationIndex = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *MeanDeviationIndex = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *MeanDeviationIndex = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *MeanDeviationIndex = @ptrCast(@alignCast(ptr));
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
    mdi: [252]f64,
    signal: [252]f64,
};

fn checkVal(exp: f64, act: f64, tolerance: f64) !void {
    try testing.expect(@abs(act - exp) <= tolerance);
}

test "MDI reference data all combos" {
    const tolerance = 1e-9;
    const input = testdata.testInput();

    const combos = [_]Combo{
        .{ .name = "R20_S5_U3", .r = 20, .s = 5, .u = 3, .mdi = testdata.expectedR20_S5_U3(), .signal = testdata.expectedR20_S5_U3_SIG_UL3() },
        .{ .name = "R20_S5_U1", .r = 20, .s = 5, .u = 1, .mdi = testdata.expectedR20_S5_U1(), .signal = testdata.expectedR20_S5_U1_SIG_UL3() },
        .{ .name = "R1_S5_U3", .r = 1, .s = 5, .u = 3, .mdi = testdata.expectedR1_S5_U3(), .signal = testdata.expectedR1_S5_U3_SIG_UL3() },
        .{ .name = "R40_S5_U3", .r = 40, .s = 5, .u = 3, .mdi = testdata.expectedR40_S5_U3(), .signal = testdata.expectedR40_S5_U3_SIG_UL3() },
        .{ .name = "R10_S5_U3", .r = 10, .s = 5, .u = 3, .mdi = testdata.expectedR10_S5_U3(), .signal = testdata.expectedR10_S5_U3_SIG_UL3() },
        .{ .name = "R5_S5_U5", .r = 5, .s = 5, .u = 5, .mdi = testdata.expectedR5_S5_U5(), .signal = testdata.expectedR5_S5_U5_SIG_UL3() },
        .{ .name = "R20_S9_U1", .r = 20, .s = 9, .u = 1, .mdi = testdata.expectedR20_S9_U1(), .signal = testdata.expectedR20_S9_U1_SIG_UL3() },
        .{ .name = "R26_S12_U9", .r = 26, .s = 12, .u = 9, .mdi = testdata.expectedR26_S12_U9(), .signal = testdata.expectedR26_S12_U9_SIG_UL3() },
        .{ .name = "R50_S13_U1", .r = 50, .s = 13, .u = 1, .mdi = testdata.expectedR50_S13_U1(), .signal = testdata.expectedR50_S13_U1_SIG_UL3() },
        .{ .name = "R30_S5_U3", .r = 30, .s = 5, .u = 3, .mdi = testdata.expectedR30_S5_U3(), .signal = testdata.expectedR30_S5_U3_SIG_UL3() },
        .{ .name = "R3_S3_U3", .r = 3, .s = 3, .u = 3, .mdi = testdata.expectedR3_S3_U3(), .signal = testdata.expectedR3_S3_U3_SIG_UL3() },
        .{ .name = "R7_S4_U2", .r = 7, .s = 4, .u = 2, .mdi = testdata.expectedR7_S4_U2(), .signal = testdata.expectedR7_S4_U2_SIG_UL3() },
        .{ .name = "R2_S5_U3", .r = 2, .s = 5, .u = 3, .mdi = testdata.expectedR2_S5_U3(), .signal = testdata.expectedR2_S5_U3_SIG_UL3() },
        .{ .name = "R20_S1_U1", .r = 20, .s = 1, .u = 1, .mdi = testdata.expectedR20_S1_U1(), .signal = testdata.expectedR20_S1_U1_SIG_UL3() },
        .{ .name = "R20_S20_U5", .r = 20, .s = 20, .u = 5, .mdi = testdata.expectedR20_S20_U5(), .signal = testdata.expectedR20_S20_U5_SIG_UL3() },
        .{ .name = "R60_S30_U10", .r = 60, .s = 30, .u = 10, .mdi = testdata.expectedR60_S30_U10(), .signal = testdata.expectedR60_S30_U10_SIG_UL3() },
    };

    for (combos) |combo| {
        var ind = try MeanDeviationIndex.init(.{
            .r = combo.r,
            .s = combo.s,
            .u = combo.u,
            .ul = signal_ul,
        });

        for (0..252) |i| {
            const result = ind.updateValues(input[i]);
            try checkVal(combo.mdi[i], result.mdi, tolerance);
            try checkVal(combo.signal[i], result.signal, tolerance);
        }
    }
}

test "MDI has no warm-up region" {
    const input = testdata.testInput();

    var ind = try MeanDeviationIndex.init(.{});

    const first = ind.updateValues(input[0]);
    try testing.expectEqual(@as(f64, 0.0), first.mdi);
    try testing.expectEqual(@as(f64, 0.0), first.signal);

    for (1..252) |i| {
        const result = ind.updateValues(input[i]);
        try testing.expect(!math.isNan(result.mdi));
        try testing.expect(!math.isNan(result.signal));
    }
}

test "MDI degenerate r = 1 is identically zero" {
    const input = testdata.testInput();

    var ind = try MeanDeviationIndex.init(.{ .r = 1, .s = 5, .u = 3, .ul = 3 });

    for (0..252) |i| {
        const result = ind.updateValues(input[i]);
        try testing.expectEqual(@as(f64, 0.0), result.mdi);
        try testing.expectEqual(@as(f64, 0.0), result.signal);
    }
}

test "MDI passthrough smoothing equals price minus EMA" {
    const tolerance = 1e-12;

    var ind = try MeanDeviationIndex.init(.{ .r = 2, .s = 1, .u = 1, .ul = 1 });

    // Bar 0 seeds the baseline, so the deviation is exactly 0.
    const first = ind.updateValues(10.0);
    try checkVal(0.0, first.mdi, tolerance);
    try checkVal(0.0, first.signal, tolerance);

    // EMA(2) = (2/3)*13 + (1/3)*10 = 12, so the deviation is 13-12 = 1.
    const second = ind.updateValues(13.0);
    try checkVal(1.0, second.mdi, tolerance);
    try checkVal(1.0, second.signal, tolerance);
}

test "MDI signal passthrough when ul = 1" {
    const input = testdata.testInput();

    var ind = try MeanDeviationIndex.init(.{ .r = 20, .s = 5, .u = 3, .ul = 1 });

    for (0..252) |i| {
        const result = ind.updateValues(input[i]);
        try testing.expectEqual(result.mdi, result.signal);
    }
}

test "MDI is primed after the first update" {
    const input = testdata.testInput();

    var ind = try MeanDeviationIndex.init(.{});

    try testing.expect(!ind.isPrimed());
    _ = ind.updateValues(input[0]);
    try testing.expect(ind.isPrimed());
}

test "MDI metadata default" {
    var ind = try MeanDeviationIndex.init(.{});

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqual(Identifier.mean_deviation_index, meta.identifier);
    try testing.expectEqualStrings("mdi(20,5,3)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 2), meta.outputs_len);
    try testing.expectEqual(@as(u8, 1), meta.outputs_buf[0].kind);
    try testing.expectEqual(@as(u8, 2), meta.outputs_buf[1].kind);
}

test "MDI custom mnemonic excludes ul" {
    var ind = try MeanDeviationIndex.init(.{ .r = 26, .s = 12, .u = 9, .ul = 7 });

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqualStrings("mdi(26,12,9)", meta.mnemonic);
}

test "MDI invalid params" {
    const r1 = MeanDeviationIndex.init(.{ .r = 0 });
    try testing.expect(if (r1) |_| false else |_| true);

    const r2 = MeanDeviationIndex.init(.{ .s = 0 });
    try testing.expect(if (r2) |_| false else |_| true);

    const r3 = MeanDeviationIndex.init(.{ .u = 0 });
    try testing.expect(if (r3) |_| false else |_| true);

    const r4 = MeanDeviationIndex.init(.{ .ul = 0 });
    try testing.expect(if (r4) |_| false else |_| true);
}

test "MDI entity update ordering" {
    const tolerance = 1e-9;
    const input = testdata.testInput();
    const exp_mdi = testdata.expectedR20_S5_U3();
    const exp_signal = testdata.expectedR20_S5_U3_SIG_UL3();

    var ind = try MeanDeviationIndex.init(.{});

    var last_out: OutputArray = undefined;
    for (0..252) |i| {
        const scalar = Scalar{ .time = 0, .value = input[i] };
        last_out = ind.updateScalar(&scalar);
    }
    const items = last_out.slice();

    try checkVal(exp_mdi[251], items[0].scalar.value, tolerance);
    try checkVal(exp_signal[251], items[1].scalar.value, tolerance);
}
