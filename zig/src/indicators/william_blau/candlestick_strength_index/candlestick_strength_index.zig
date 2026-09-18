const std = @import("std");
const math = std.math;

const entities = @import("entities");
const Bar = entities.Bar;
const Quote = entities.Quote;
const Trade = entities.Trade;
const Scalar = entities.Scalar;
const indicator_mod = @import("../../core/indicator.zig");
const build_metadata_mod = @import("../../core/build_metadata.zig");
const identifier_mod = @import("../../core/identifier.zig");
const metadata_mod = @import("../../core/metadata.zig");

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the Candlestick Strength Index indicator.
pub const CandlestickStrengthIndexOutput = enum(u8) {
    /// Candlestick Strength Index oscillator value (range [-100, +100]).
    csi = 1,
    /// Signal-line value: the ul-period EMA of the oscillator.
    signal = 2,
};

/// Parameters to create a Candlestick Strength Index indicator.
///
/// The field names r, s, u and ul are the canonical symbols from William Blau's
/// Momentum, Direction, and Divergence (Wiley, 1995), chapter 6.
///
/// The indicator consumes the open, high, low and close prices of a bar, so it
/// has no configurable price-component fields.
pub const CandlestickStrengthIndexParams = struct {
    r: usize = 20,
    s: usize = 5,
    u: usize = 3,
    ul: usize = 3,
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

/// Candlestick Strength Index (CSI) by William Blau, known in the book as the
/// CandleStick Indicator.
///
/// A double-/triple-smoothed candle-body-vs-range oscillator bounded to [-100, +100],
/// paired with an EMA signal line (the Ergodic form, Blau ch.6.4):
///
///   csi_k    = 100 * TEMA(close-open, r, s, u) / TEMA(high-low, r, s, u)   (the oscillator)
///   signal_k = EMA(csi, ul)_k                                              (ul-period EMA)
///
/// where the two intra-bar quantities are the signed candle body co_k = close_k - open_k
/// and the bar range hl_k = high_k - low_k >= 0, and TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u).
///
/// It is the range-normalized sibling of the Candlestick Momentum Index (CMI): both share
/// the signed numerator TEMA(close-open), but the CSI divides by the smoothed range while
/// the CMI divides by the smoothed absolute body. Because every bar has |close-open| <= high-low,
/// the ratio is bounded to [-100, +100]. The inputs are the open, high, low and close prices.
///
/// The indicator produces two outputs:
///   - CSI: the oscillator, range [-100, +100];
///   - Signal: the ul-period EMA of the oscillator (Blau's Ergodic signal line).
///
/// Priming convention (book / EasyLanguage): each EMA stage seeds on its first
/// received value. Both intra-bar series are defined from bar 0, so there is no NaN
/// warm-up region. Division guard: denominator <= 0 -> oscillator 0.0.
pub const CandlestickStrengthIndex = struct {
    num_r: Ema,
    num_s: Ema,
    num_u: Ema,
    den_r: Ema,
    den_s: Ema,
    den_u: Ema,

    signal_ema: Ema,

    primed: bool,

    mnemonic_buf: [128]u8,
    mnemonic_len: usize,
    description_buf: [192]u8,
    description_len: usize,

    pub fn init(params: CandlestickStrengthIndexParams) !CandlestickStrengthIndex {
        const r = params.r;
        const s = params.s;
        const u = params.u;
        const ul = params.ul;

        if (r < 1) return error.InvalidR;
        if (s < 1) return error.InvalidS;
        if (u < 1) return error.InvalidU;
        if (ul < 1) return error.InvalidUl;

        var mnemonic_buf: [128]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "csi({d},{d},{d},{d})", .{
            r,
            s,
            u,
            ul,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [192]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Candlestick Strength Index {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        return .{
            .num_r = Ema.init(r),
            .num_s = Ema.init(s),
            .num_u = Ema.init(u),
            .den_r = Ema.init(r),
            .den_s = Ema.init(s),
            .den_u = Ema.init(u),
            .signal_ema = Ema.init(ul),
            .primed = false,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn fixSlices(self: *CandlestickStrengthIndex) void {
        _ = self;
        // CSI doesn't use LineIndicator; mnemonic/description are read from the
        // owned buffers directly, so no slice fixup is needed.
    }

    /// Updates the indicator given the next bar's open, high, low and close values.
    /// Returns csi, signal.
    pub fn updateValues(
        self: *CandlestickStrengthIndex,
        open: f64,
        high: f64,
        low: f64,
        close: f64,
    ) struct { csi: f64, signal: f64 } {
        // Two intra-bar quantities: the signed candle body and the (non-negative) range.
        const co = close - open;
        const hl = high - low;

        // Numerator cascade: TEMA(close-open, r, s, u).
        const n = self.num_u.update(self.num_s.update(self.num_r.update(co)));
        // Denominator cascade: TEMA(high-low, r, s, u).
        const d = self.den_u.update(self.den_s.update(self.den_r.update(hl)));

        // Division guard: zero range so far -> oscillator 0.0.
        const csi: f64 = if (d > 0.0) 100.0 * n / d else 0.0;

        // Signal line = EMA(csi, ul); seeds on bar 0's oscillator value.
        const signal = self.signal_ema.update(csi);
        self.primed = true;

        return .{ .csi = csi, .signal = signal };
    }

    pub fn isPrimed(self: *const CandlestickStrengthIndex) bool {
        return self.primed;
    }

    fn mnemonic(self: *const CandlestickStrengthIndex) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const CandlestickStrengthIndex) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    pub fn getMetadata(self: *const CandlestickStrengthIndex, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();

        var csi_mn_buf: [160]u8 = undefined;
        const csi_mn = std.fmt.bufPrint(&csi_mn_buf, "{s} csi", .{mn}) catch mn;
        var signal_mn_buf: [160]u8 = undefined;
        const signal_mn = std.fmt.bufPrint(&signal_mn_buf, "{s} signal", .{mn}) catch mn;

        var csi_desc_buf: [256]u8 = undefined;
        const csi_desc = std.fmt.bufPrint(&csi_desc_buf, "{s} CSI", .{desc}) catch desc;
        var signal_desc_buf: [256]u8 = undefined;
        const signal_desc = std.fmt.bufPrint(&signal_desc_buf, "{s} signal", .{desc}) catch desc;

        build_metadata_mod.buildMetadata(
            out,
            .candlestick_strength_index,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = csi_mn, .description = csi_desc },
                .{ .mnemonic = signal_mn, .description = signal_desc },
            },
        );
    }

    /// Updates the indicator given the next scalar sample.
    ///
    /// A scalar carries a single value, so the candle body and the bar range are both zero.
    pub fn updateScalar(self: *CandlestickStrengthIndex, sample: *const Scalar) OutputArray {
        return self.updateEntity(sample.time, sample.value, sample.value, sample.value, sample.value);
    }

    pub fn updateBar(self: *CandlestickStrengthIndex, sample: *const Bar) OutputArray {
        return self.updateEntity(sample.time, sample.open, sample.high, sample.low, sample.close);
    }

    /// Updates the indicator given the next quote sample.
    ///
    /// A quote maps the bid to the open and the low, and the ask to the close and the high,
    /// so the candle body and the bar range are both the spread.
    pub fn updateQuote(self: *CandlestickStrengthIndex, sample: *const Quote) OutputArray {
        return self.updateEntity(
            sample.time,
            sample.bid_price,
            sample.ask_price,
            sample.bid_price,
            sample.ask_price,
        );
    }

    /// Updates the indicator given the next trade sample.
    ///
    /// A trade carries a single price, so the candle body and the bar range are both zero.
    pub fn updateTrade(self: *CandlestickStrengthIndex, sample: *const Trade) OutputArray {
        return self.updateEntity(sample.time, sample.price, sample.price, sample.price, sample.price);
    }

    fn updateEntity(
        self: *CandlestickStrengthIndex,
        time: i64,
        open: f64,
        high: f64,
        low: f64,
        close: f64,
    ) OutputArray {
        const result = self.updateValues(open, high, low, close);

        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = result.csi } });
        out.append(.{ .scalar = .{ .time = time, .value = result.signal } });
        return out;
    }

    pub fn indicator(self: *CandlestickStrengthIndex) indicator_mod.Indicator {
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
        const self: *CandlestickStrengthIndex = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const CandlestickStrengthIndex = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *CandlestickStrengthIndex = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *CandlestickStrengthIndex = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *CandlestickStrengthIndex = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *CandlestickStrengthIndex = @ptrCast(@alignCast(ptr));
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
    csi: [252]f64,
    signal: [252]f64,
};

fn checkVal(exp: f64, act: f64, tolerance: f64) !void {
    if (math.isNan(exp)) {
        try testing.expect(math.isNan(act));
        return;
    }
    try testing.expect(@abs(act - exp) <= tolerance);
}

test "CSI reference data all combos" {
    const tolerance = 1e-9;
    const input = testdata.testInput();
    const open = testdata.testOpen();
    const high = testdata.testHigh();
    const low = testdata.testLow();

    const combos = [_]Combo{
        .{ .name = "R20_S5_U3", .r = 20, .s = 5, .u = 3, .csi = testdata.expectedR20_S5_U3(), .signal = testdata.expectedR20_S5_U3_SIG_UL3() },
        .{ .name = "R32_S32_U1", .r = 32, .s = 32, .u = 1, .csi = testdata.expectedR32_S32_U1(), .signal = testdata.expectedR32_S32_U1_SIG_UL3() },
        .{ .name = "R1_S1_U1", .r = 1, .s = 1, .u = 1, .csi = testdata.expectedR1_S1_U1(), .signal = testdata.expectedR1_S1_U1_SIG_UL3() },
        .{ .name = "R25_S13_U1", .r = 25, .s = 13, .u = 1, .csi = testdata.expectedR25_S13_U1(), .signal = testdata.expectedR25_S13_U1_SIG_UL3() },
        .{ .name = "R13_S13_U1", .r = 13, .s = 13, .u = 1, .csi = testdata.expectedR13_S13_U1(), .signal = testdata.expectedR13_S13_U1_SIG_UL3() },
        .{ .name = "R5_S5_U5", .r = 5, .s = 5, .u = 5, .csi = testdata.expectedR5_S5_U5(), .signal = testdata.expectedR5_S5_U5_SIG_UL3() },
        .{ .name = "R9_S3_U1", .r = 9, .s = 3, .u = 1, .csi = testdata.expectedR9_S3_U1(), .signal = testdata.expectedR9_S3_U1_SIG_UL3() },
        .{ .name = "R64_S64_U1", .r = 64, .s = 64, .u = 1, .csi = testdata.expectedR64_S64_U1(), .signal = testdata.expectedR64_S64_U1_SIG_UL3() },
        .{ .name = "R32_S32_U3", .r = 32, .s = 32, .u = 3, .csi = testdata.expectedR32_S32_U3(), .signal = testdata.expectedR32_S32_U3_SIG_UL3() },
        .{ .name = "R40_S20_U1", .r = 40, .s = 20, .u = 1, .csi = testdata.expectedR40_S20_U1(), .signal = testdata.expectedR40_S20_U1_SIG_UL3() },
        .{ .name = "R2_S2_U2", .r = 2, .s = 2, .u = 2, .csi = testdata.expectedR2_S2_U2(), .signal = testdata.expectedR2_S2_U2_SIG_UL3() },
        .{ .name = "R7_S4_U2", .r = 7, .s = 4, .u = 2, .csi = testdata.expectedR7_S4_U2(), .signal = testdata.expectedR7_S4_U2_SIG_UL3() },
        .{ .name = "R12_S12_U12", .r = 12, .s = 12, .u = 12, .csi = testdata.expectedR12_S12_U12(), .signal = testdata.expectedR12_S12_U12_SIG_UL3() },
        .{ .name = "R3_S10_U10", .r = 3, .s = 10, .u = 10, .csi = testdata.expectedR3_S10_U10(), .signal = testdata.expectedR3_S10_U10_SIG_UL3() },
        .{ .name = "R50_S1_U1", .r = 50, .s = 1, .u = 1, .csi = testdata.expectedR50_S1_U1(), .signal = testdata.expectedR50_S1_U1_SIG_UL3() },
        .{ .name = "R32_S5_U3", .r = 32, .s = 5, .u = 3, .csi = testdata.expectedR32_S5_U3(), .signal = testdata.expectedR32_S5_U3_SIG_UL3() },
    };

    for (combos) |combo| {
        var ind = try CandlestickStrengthIndex.init(.{
            .r = combo.r,
            .s = combo.s,
            .u = combo.u,
            .ul = signal_ul,
        });

        for (0..252) |i| {
            const result = ind.updateValues(open[i], high[i], low[i], input[i]);
            try checkVal(combo.csi[i], result.csi, tolerance);
            try checkVal(combo.signal[i], result.signal, tolerance);
        }
    }
}

test "CSI passthrough reduces to the body over the range" {
    const tolerance = 1e-9;

    var ind = try CandlestickStrengthIndex.init(.{ .r = 1, .s = 1, .u = 1, .ul = 1 });

    const up = ind.updateValues(10.0, 12.0, 10.0, 12.0);
    try checkVal(100.0, up.csi, tolerance);
    try checkVal(100.0, up.signal, tolerance);

    const down = ind.updateValues(12.0, 12.0, 10.0, 10.0);
    try checkVal(-100.0, down.csi, tolerance);
    try checkVal(-100.0, down.signal, tolerance);

    const flat = ind.updateValues(11.0, 11.0, 11.0, 11.0);
    try checkVal(0.0, flat.csi, tolerance);
}

test "CSI is primed after the first bar" {
    const input = testdata.testInput();
    const open = testdata.testOpen();
    const high = testdata.testHigh();
    const low = testdata.testLow();

    var ind = try CandlestickStrengthIndex.init(.{});
    try testing.expect(!ind.isPrimed());

    _ = ind.updateValues(open[0], high[0], low[0], input[0]);
    try testing.expect(ind.isPrimed());
}

test "CSI metadata default" {
    var ind = try CandlestickStrengthIndex.init(.{});

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqual(Identifier.candlestick_strength_index, meta.identifier);
    try testing.expectEqualStrings("csi(20,5,3,3)", meta.mnemonic);
    try testing.expectEqualStrings("Candlestick Strength Index csi(20,5,3,3)", meta.description);
    try testing.expectEqual(@as(usize, 2), meta.outputs_len);
    try testing.expectEqual(@as(u8, 1), meta.outputs_buf[0].kind);
    try testing.expectEqual(@as(u8, 2), meta.outputs_buf[1].kind);
}

test "CSI custom mnemonic" {
    var ind = try CandlestickStrengthIndex.init(.{ .r = 25, .s = 13, .u = 1, .ul = 7 });

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqualStrings("csi(25,13,1,7)", meta.mnemonic);
}

test "CSI invalid params" {
    const r1 = CandlestickStrengthIndex.init(.{ .r = 0 });
    try testing.expect(if (r1) |_| false else |_| true);

    const r2 = CandlestickStrengthIndex.init(.{ .s = 0 });
    try testing.expect(if (r2) |_| false else |_| true);

    const r3 = CandlestickStrengthIndex.init(.{ .u = 0 });
    try testing.expect(if (r3) |_| false else |_| true);

    const r4 = CandlestickStrengthIndex.init(.{ .ul = 0 });
    try testing.expect(if (r4) |_| false else |_| true);
}

test "CSI bar update ordering" {
    const tolerance = 1e-9;
    const input = testdata.testInput();
    const open = testdata.testOpen();
    const high = testdata.testHigh();
    const low = testdata.testLow();
    const exp_csi = testdata.expectedR20_S5_U3();
    const exp_signal = testdata.expectedR20_S5_U3_SIG_UL3();

    var ind = try CandlestickStrengthIndex.init(.{});

    var last_out: OutputArray = undefined;
    for (0..252) |i| {
        const bar = Bar{
            .time = 0,
            .open = open[i],
            .high = high[i],
            .low = low[i],
            .close = input[i],
            .volume = 0.0,
        };
        last_out = ind.updateBar(&bar);
    }
    const items = last_out.slice();

    try checkVal(exp_csi[251], items[0].scalar.value, tolerance);
    try checkVal(exp_signal[251], items[1].scalar.value, tolerance);
}

test "CSI quote maps bid to open and low, ask to close and high" {
    const tolerance = 1e-9;

    var ind = try CandlestickStrengthIndex.init(.{ .r = 1, .s = 1, .u = 1, .ul = 1 });

    const quote = Quote{ .time = 0, .bid_price = 10.0, .ask_price = 12.0, .bid_size = 1.0, .ask_size = 1.0 };
    const items = ind.updateQuote(&quote).slice();

    try checkVal(100.0, items[0].scalar.value, tolerance);
    try checkVal(100.0, items[1].scalar.value, tolerance);
}

test "CSI scalar and trade have a zero candle body and a zero range" {
    const tolerance = 1e-9;

    var scalar_ind = try CandlestickStrengthIndex.init(.{ .r = 1, .s = 1, .u = 1, .ul = 1 });
    const scalar = Scalar{ .time = 0, .value = 10.0 };
    const scalar_items = scalar_ind.updateScalar(&scalar).slice();
    try checkVal(0.0, scalar_items[0].scalar.value, tolerance);
    try checkVal(0.0, scalar_items[1].scalar.value, tolerance);

    var trade_ind = try CandlestickStrengthIndex.init(.{ .r = 1, .s = 1, .u = 1, .ul = 1 });
    const trade = Trade{ .time = 0, .price = 10.0, .volume = 1.0 };
    const trade_items = trade_ind.updateTrade(&trade).slice();
    try checkVal(0.0, trade_items[0].scalar.value, tolerance);
    try checkVal(0.0, trade_items[1].scalar.value, tolerance);
}
