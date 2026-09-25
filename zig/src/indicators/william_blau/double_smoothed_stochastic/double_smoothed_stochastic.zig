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

/// Enumerates the outputs of the Double Smoothed Stochastic indicator.
pub const DoubleSmoothedStochasticOutput = enum(u8) {
    /// Double Smoothed Stochastic oscillator value (range [0, 100]).
    dss = 1,
    /// Signal-line value: the g-period SMA of the oscillator.
    signal = 2,
};

/// Parameters to create a Double Smoothed Stochastic indicator.
///
/// The field names q, r, s and g are the canonical symbols from William Blau's
/// Momentum, Direction, and Divergence (Wiley, 1995).
///
/// The indicator consumes the high, low and close prices of a bar, so it has no
/// configurable price-component fields.
pub const DoubleSmoothedStochasticParams = struct {
    q: usize = 5,
    r: usize = 7,
    s: usize = 3,
    g: usize = 3,
};

/// Stateful streaming EMA: alpha = 2/(period+1), seeds e0 = x0.
///
/// Inlined verbatim from the Blau exponential moving average so the indicator is a
/// standalone porting unit. Do NOT change its numerics.
///
/// period == 1 -> alpha == 1 -> pure passthrough (output == input).
const Ema = struct {
    alpha: f64,
    previous: f64 = 0.0,
    primed: bool = false,

    fn init(period: usize) Ema {
        return .{ .alpha = 2.0 / (@as(f64, @floatFromInt(period)) + 1.0) };
    }

    fn update(self: *Ema, x: f64) f64 {
        if (!self.primed) {
            self.previous = x;
            self.primed = true;
            return self.previous;
        }
        self.previous = self.alpha * x + (1.0 - self.alpha) * self.previous;
        return self.previous;
    }
};

/// Stateful streaming SMA over the last period inputs.
///
/// Returns the mean of the window's current contents on every update: an expanding
/// window while fewer than period values have arrived, then a rolling period-bar
/// window. There is no NaN warm-up (finite from the first input).
///
/// period == 1 -> pure passthrough (output == input).
const Sma = struct {
    window: []f64,
    count: usize = 0,
    index: usize = 0,

    fn update(self: *Sma, x: f64) f64 {
        const period = self.window.len;
        self.window[self.index] = x;
        self.index = (self.index + 1) % period;

        if (self.count < period) {
            self.count += 1;
        }

        // Naive left-to-right sum from the oldest to the newest value (NOT a
        // compensated sum), so that every port reproduces the same values.
        // Once full, the oldest value sits at the next write position.
        const start: usize = if (self.count == period) self.index else 0;

        var sum: f64 = 0.0;
        for (0..self.count) |i| {
            sum += self.window[(start + i) % period];
        }

        return sum / @as(f64, @floatFromInt(self.count));
    }
};

/// Double Smoothed Stochastic (DSS) by William Blau.
///
/// A classic double-smoothed stochastic oscillator bounded to [0, 100], paired
/// with a short simple-moving-average signal line:
///
///   dss_k    = 100 * EMA(EMA(st, r), s) / EMA(EMA(rng, r), s)   (the oscillator)
///   signal_k = SMA(dss, g)_k                                    (g-period SMA)
///
/// where, over the last q bars, HH_k is the highest high and LL_k is the lowest low,
/// st_k = close_k - LL_k >= 0 is the raw stochastic (the close above the low), and
/// rng_k = HH_k - LL_k >= 0 is the q-bar range.
///
/// The raw stochastic and the range are smoothed separately with the same two-stage
/// EMA cascade (r then s), then divided. Because 0 <= st <= rng on every bar, the
/// ratio is bounded to [0, 100]. With q = 1 it is Blau's one-bar HLC index. It is
/// exactly the MQL5 Blau_TStochI with its third EMA period u = 1. The inputs are the
/// high, low and close prices.
///
/// The indicator produces two outputs:
///   - DSS: the oscillator, range [0, 100];
///   - Signal: the g-period SMA of the oscillator.
///
/// Priming convention (book / EasyLanguage): st and rng become valid once q bars of
/// high/low exist, i.e. at bar q-1. All four cascade stages seed there together, so
/// both outputs are NaN for bars 0..q-2 and finite from bar q-1; for q = 1 there is
/// no NaN warm-up. The signal SMA seeds on the first finite oscillator value and
/// returns the mean of the oscillator values seen so far (expanding window <= g),
/// then the full g-bar rolling mean. Division guard: denominator <= 0 -> oscillator 0.0.
pub const DoubleSmoothedStochastic = struct {
    q: usize,
    highs: []f64,
    lows: []f64,
    window_count: usize,
    window_index: usize,

    num_r: Ema,
    num_s: Ema,
    den_r: Ema,
    den_s: Ema,

    signal_sma: Sma,

    primed: bool,

    mnemonic_buf: [128]u8,
    mnemonic_len: usize,
    description_buf: [192]u8,
    description_len: usize,

    allocator: std.mem.Allocator,

    pub fn init(allocator: std.mem.Allocator, params: DoubleSmoothedStochasticParams) InitError!DoubleSmoothedStochastic {
        const q = params.q;
        const r = params.r;
        const s = params.s;
        const g = params.g;

        if (q < 1) return error.InvalidQ;
        if (r < 1) return error.InvalidR;
        if (s < 1) return error.InvalidS;
        if (g < 1) return error.InvalidG;

        var mnemonic_buf: [128]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "dss({d},{d},{d},{d})", .{
            q,
            r,
            s,
            g,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [192]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Double Smoothed Stochastic {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const highs = allocator.alloc(f64, q) catch return error.OutOfMemory;
        errdefer allocator.free(highs);
        const lows = allocator.alloc(f64, q) catch return error.OutOfMemory;
        errdefer allocator.free(lows);
        const signal_window = allocator.alloc(f64, g) catch return error.OutOfMemory;

        return .{
            .q = q,
            .highs = highs,
            .lows = lows,
            .window_count = 0,
            .window_index = 0,
            .num_r = Ema.init(r),
            .num_s = Ema.init(s),
            .den_r = Ema.init(r),
            .den_s = Ema.init(s),
            .signal_sma = .{ .window = signal_window },
            .primed = false,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
            .allocator = allocator,
        };
    }

    pub fn deinit(self: *DoubleSmoothedStochastic) void {
        self.allocator.free(self.highs);
        self.allocator.free(self.lows);
        self.allocator.free(self.signal_sma.window);
    }

    pub fn fixSlices(self: *DoubleSmoothedStochastic) void {
        _ = self;
        // DSS doesn't use LineIndicator; mnemonic/description are read from the
        // owned buffers directly, so no slice fixup is needed.
    }

    /// Updates the indicator given the next bar's high, low and close values.
    /// Returns dss, signal; both are NaN until q bars have been seen.
    pub fn updateValues(
        self: *DoubleSmoothedStochastic,
        high: f64,
        low: f64,
        close: f64,
    ) struct { dss: f64, signal: f64 } {
        self.highs[self.window_index] = high;
        self.lows[self.window_index] = low;
        self.window_index = (self.window_index + 1) % self.q;

        if (self.window_count < self.q) {
            self.window_count += 1;
        }

        // Need q bars of high/low before the stochastic is defined. Until then
        // neither output exists -- do NOT advance the EMA cascades or the SMA.
        if (self.window_count < self.q) {
            return .{ .dss = math.nan(f64), .signal = math.nan(f64) };
        }

        // Rolling extremes over the last q bars.
        var hh = self.highs[0];
        var ll = self.lows[0];
        for (1..self.q) |i| {
            hh = @max(hh, self.highs[i]);
            ll = @min(ll, self.lows[i]);
        }

        // Raw stochastic and range (both non-negative).
        const st = close - ll;
        const rng = hh - ll;

        // Numerator cascade: EMA(EMA(st, r), s).
        const n = self.num_s.update(self.num_r.update(st));
        // Denominator cascade: EMA(EMA(rng, r), s).
        const d = self.den_s.update(self.den_r.update(rng));

        // Division guard: flat window so far -> oscillator 0.0.
        const dss: f64 = if (d > 0.0) 100.0 * n / d else 0.0;

        // Signal line = SMA(dss, g); seeds on the first finite oscillator value.
        const signal = self.signal_sma.update(dss);
        self.primed = true;

        return .{ .dss = dss, .signal = signal };
    }

    pub fn isPrimed(self: *const DoubleSmoothedStochastic) bool {
        return self.primed;
    }

    fn mnemonic(self: *const DoubleSmoothedStochastic) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const DoubleSmoothedStochastic) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    pub fn getMetadata(self: *const DoubleSmoothedStochastic, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();

        var dss_mn_buf: [160]u8 = undefined;
        const dss_mn = std.fmt.bufPrint(&dss_mn_buf, "{s} dss", .{mn}) catch mn;
        var signal_mn_buf: [160]u8 = undefined;
        const signal_mn = std.fmt.bufPrint(&signal_mn_buf, "{s} signal", .{mn}) catch mn;

        var dss_desc_buf: [256]u8 = undefined;
        const dss_desc = std.fmt.bufPrint(&dss_desc_buf, "{s} DSS", .{desc}) catch desc;
        var signal_desc_buf: [256]u8 = undefined;
        const signal_desc = std.fmt.bufPrint(&signal_desc_buf, "{s} signal", .{desc}) catch desc;

        build_metadata_mod.buildMetadata(
            out,
            .double_smoothed_stochastic,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = dss_mn, .description = dss_desc },
                .{ .mnemonic = signal_mn, .description = signal_desc },
            },
        );
    }

    /// Updates the indicator given the next scalar sample.
    ///
    /// A scalar carries a single value, used as the high, the low and the close.
    pub fn updateScalar(self: *DoubleSmoothedStochastic, sample: *const Scalar) OutputArray {
        const v = sample.value;
        return self.updateEntity(sample.time, v, v, v);
    }

    pub fn updateBar(self: *DoubleSmoothedStochastic, sample: *const Bar) OutputArray {
        return self.updateEntity(sample.time, sample.high, sample.low, sample.close);
    }

    /// Updates the indicator given the next quote sample.
    ///
    /// A quote maps the mid price to the high, the low and the close.
    pub fn updateQuote(self: *DoubleSmoothedStochastic, sample: *const Quote) OutputArray {
        const v = (sample.bid_price + sample.ask_price) / 2.0;
        return self.updateEntity(sample.time, v, v, v);
    }

    /// Updates the indicator given the next trade sample.
    ///
    /// A trade carries a single price, used as the high, the low and the close.
    pub fn updateTrade(self: *DoubleSmoothedStochastic, sample: *const Trade) OutputArray {
        const v = sample.price;
        return self.updateEntity(sample.time, v, v, v);
    }

    fn updateEntity(
        self: *DoubleSmoothedStochastic,
        time: i64,
        high: f64,
        low: f64,
        close: f64,
    ) OutputArray {
        const result = self.updateValues(high, low, close);

        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = result.dss } });
        out.append(.{ .scalar = .{ .time = time, .value = result.signal } });
        return out;
    }

    pub fn indicator(self: *DoubleSmoothedStochastic) indicator_mod.Indicator {
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
        const self: *DoubleSmoothedStochastic = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const DoubleSmoothedStochastic = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *DoubleSmoothedStochastic = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *DoubleSmoothedStochastic = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *DoubleSmoothedStochastic = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *DoubleSmoothedStochastic = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidQ,
        InvalidR,
        InvalidS,
        InvalidG,
        MnemonicTooLong,
        OutOfMemory,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");

const tolerance = 1e-10;

const Combo = struct {
    name: []const u8,
    q: usize,
    r: usize,
    s: usize,
    g: usize,
    dss: [252]f64,
    signal: [252]f64,
};

fn checkVal(exp: f64, act: f64) !void {
    if (math.isNan(exp)) {
        try testing.expect(math.isNan(act));
        return;
    }
    try testing.expect(@abs(act - exp) <= tolerance);
}

test "DSS reference data all combos" {
    const input = testdata.testInput();
    const high = testdata.testHigh();
    const low = testdata.testLow();

    const combos = [_]Combo{
        .{ .name = "Q5_R7_S3_G3", .q = 5, .r = 7, .s = 3, .g = 3, .dss = testdata.expectedDsQ5_R7_S3_G3(), .signal = testdata.expectedSigQ5_R7_S3_G3() },
        .{ .name = "Q2_R3_S15_G3", .q = 2, .r = 3, .s = 15, .g = 3, .dss = testdata.expectedDsQ2_R3_S15_G3(), .signal = testdata.expectedSigQ2_R3_S15_G3() },
        .{ .name = "Q5_R20_S5_G3", .q = 5, .r = 20, .s = 5, .g = 3, .dss = testdata.expectedDsQ5_R20_S5_G3(), .signal = testdata.expectedSigQ5_R20_S5_G3() },
        .{ .name = "Q5_R7_S3_G1", .q = 5, .r = 7, .s = 3, .g = 1, .dss = testdata.expectedDsQ5_R7_S3_G1(), .signal = testdata.expectedSigQ5_R7_S3_G1() },
        .{ .name = "Q2_R3_S15_G1", .q = 2, .r = 3, .s = 15, .g = 1, .dss = testdata.expectedDsQ2_R3_S15_G1(), .signal = testdata.expectedSigQ2_R3_S15_G1() },
        .{ .name = "Q1_R1_S1_G1", .q = 1, .r = 1, .s = 1, .g = 1, .dss = testdata.expectedDsQ1_R1_S1_G1(), .signal = testdata.expectedSigQ1_R1_S1_G1() },
        .{ .name = "Q1_R5_S5_G3", .q = 1, .r = 5, .s = 5, .g = 3, .dss = testdata.expectedDsQ1_R5_S5_G3(), .signal = testdata.expectedSigQ1_R5_S5_G3() },
        .{ .name = "Q8_R5_S3_G3", .q = 8, .r = 5, .s = 3, .g = 3, .dss = testdata.expectedDsQ8_R5_S3_G3(), .signal = testdata.expectedSigQ8_R5_S3_G3() },
        .{ .name = "Q21_R13_S4_G3", .q = 21, .r = 13, .s = 4, .g = 3, .dss = testdata.expectedDsQ21_R13_S4_G3(), .signal = testdata.expectedSigQ21_R13_S4_G3() },
        .{ .name = "Q5_R1_S1_G3", .q = 5, .r = 1, .s = 1, .g = 3, .dss = testdata.expectedDsQ5_R1_S1_G3(), .signal = testdata.expectedSigQ5_R1_S1_G3() },
        .{ .name = "Q3_R10_S10_G5", .q = 3, .r = 10, .s = 10, .g = 5, .dss = testdata.expectedDsQ3_R10_S10_G5(), .signal = testdata.expectedSigQ3_R10_S10_G5() },
        .{ .name = "Q34_R5_S5_G3", .q = 34, .r = 5, .s = 5, .g = 3, .dss = testdata.expectedDsQ34_R5_S5_G3(), .signal = testdata.expectedSigQ34_R5_S5_G3() },
        .{ .name = "Q2_R3_S15_G5", .q = 2, .r = 3, .s = 15, .g = 5, .dss = testdata.expectedDsQ2_R3_S15_G5(), .signal = testdata.expectedSigQ2_R3_S15_G5() },
        .{ .name = "Q10_R7_S3_G3", .q = 10, .r = 7, .s = 3, .g = 3, .dss = testdata.expectedDsQ10_R7_S3_G3(), .signal = testdata.expectedSigQ10_R7_S3_G3() },
    };

    for (combos) |combo| {
        var ind = try DoubleSmoothedStochastic.init(testing.allocator, .{
            .q = combo.q,
            .r = combo.r,
            .s = combo.s,
            .g = combo.g,
        });
        defer ind.deinit();

        for (0..252) |i| {
            const result = ind.updateValues(high[i], low[i], input[i]);
            try checkVal(combo.dss[i], result.dss);
            try checkVal(combo.signal[i], result.signal);
        }
    }
}

test "DSS passthrough reduces to the raw one-bar HLC index" {
    var ind = try DoubleSmoothedStochastic.init(testing.allocator, .{ .q = 1, .r = 1, .s = 1, .g = 1 });
    defer ind.deinit();

    const at_high = ind.updateValues(12.0, 10.0, 12.0);
    try checkVal(100.0, at_high.dss);
    try checkVal(100.0, at_high.signal);

    const at_low = ind.updateValues(12.0, 10.0, 10.0);
    try checkVal(0.0, at_low.dss);
    try checkVal(0.0, at_low.signal);

    const at_mid = ind.updateValues(12.0, 10.0, 11.0);
    try checkVal(50.0, at_mid.dss);
    try checkVal(50.0, at_mid.signal);

    const flat = ind.updateValues(11.0, 11.0, 11.0);
    try checkVal(0.0, flat.dss);
    try checkVal(0.0, flat.signal);
}

test "DSS signal averages an expanding window, then rolls over the last g values" {
    var ind = try DoubleSmoothedStochastic.init(testing.allocator, .{ .q = 1, .r = 1, .s = 1, .g = 3 });
    defer ind.deinit();

    try checkVal(100.0, ind.updateValues(12.0, 10.0, 12.0).signal); // dss 100
    try checkVal(50.0, ind.updateValues(12.0, 10.0, 10.0).signal); // dss 0
    try checkVal(50.0, ind.updateValues(12.0, 10.0, 11.0).signal); // dss 50
    try checkVal(100.0 / 3.0, ind.updateValues(12.0, 10.0, 11.0).signal); // dss 50, 100 dropped
}

test "DSS signal equals dss when g = 1" {
    const input = testdata.testInput();
    const high = testdata.testHigh();
    const low = testdata.testLow();

    var ind = try DoubleSmoothedStochastic.init(testing.allocator, .{ .q = 5, .r = 7, .s = 3, .g = 1 });
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateValues(high[i], low[i], input[i]);
        if (math.isNan(result.dss)) {
            try testing.expect(math.isNan(result.signal));
        } else {
            try testing.expectEqual(result.dss, result.signal);
        }
    }
}

test "DSS is primed from bar q-1" {
    const input = testdata.testInput();
    const high = testdata.testHigh();
    const low = testdata.testLow();
    const q = 5;

    var ind = try DoubleSmoothedStochastic.init(testing.allocator, .{ .q = q });
    defer ind.deinit();
    try testing.expect(!ind.isPrimed());

    for (0..q - 1) |i| {
        const result = ind.updateValues(high[i], low[i], input[i]);
        try testing.expect(!ind.isPrimed());
        try testing.expect(math.isNan(result.dss));
        try testing.expect(math.isNan(result.signal));
    }

    for (q - 1..252) |i| {
        _ = ind.updateValues(high[i], low[i], input[i]);
        try testing.expect(ind.isPrimed());
    }
}

test "DSS with q = 1 is primed after the first bar" {
    const input = testdata.testInput();
    const high = testdata.testHigh();
    const low = testdata.testLow();

    var ind = try DoubleSmoothedStochastic.init(testing.allocator, .{ .q = 1 });
    defer ind.deinit();
    try testing.expect(!ind.isPrimed());

    _ = ind.updateValues(high[0], low[0], input[0]);
    try testing.expect(ind.isPrimed());
}

test "DSS metadata default" {
    var ind = try DoubleSmoothedStochastic.init(testing.allocator, .{});
    defer ind.deinit();

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqual(Identifier.double_smoothed_stochastic, meta.identifier);
    try testing.expectEqualStrings("dss(5,7,3,3)", meta.mnemonic);
    try testing.expectEqualStrings("Double Smoothed Stochastic dss(5,7,3,3)", meta.description);
    try testing.expectEqual(@as(usize, 2), meta.outputs_len);
    try testing.expectEqual(@as(u8, 1), meta.outputs_buf[0].kind);
    try testing.expectEqual(@as(u8, 2), meta.outputs_buf[1].kind);
    try testing.expectEqualStrings("dss(5,7,3,3) dss", meta.outputs_buf[0].mnemonic);
    try testing.expectEqualStrings("Double Smoothed Stochastic dss(5,7,3,3) DSS", meta.outputs_buf[0].description);
    try testing.expectEqualStrings("dss(5,7,3,3) signal", meta.outputs_buf[1].mnemonic);
    try testing.expectEqualStrings("Double Smoothed Stochastic dss(5,7,3,3) signal", meta.outputs_buf[1].description);
}

test "DSS custom mnemonic" {
    var ind = try DoubleSmoothedStochastic.init(testing.allocator, .{ .q = 2, .r = 3, .s = 15, .g = 5 });
    defer ind.deinit();

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqualStrings("dss(2,3,15,5)", meta.mnemonic);
}

test "DSS invalid params" {
    try testing.expectError(error.InvalidQ, DoubleSmoothedStochastic.init(testing.allocator, .{ .q = 0 }));
    try testing.expectError(error.InvalidR, DoubleSmoothedStochastic.init(testing.allocator, .{ .r = 0 }));
    try testing.expectError(error.InvalidS, DoubleSmoothedStochastic.init(testing.allocator, .{ .s = 0 }));
    try testing.expectError(error.InvalidG, DoubleSmoothedStochastic.init(testing.allocator, .{ .g = 0 }));
}

test "DSS bar update ordering" {
    const input = testdata.testInput();
    const high = testdata.testHigh();
    const low = testdata.testLow();
    const exp_dss = testdata.expectedDsQ5_R7_S3_G3();
    const exp_signal = testdata.expectedSigQ5_R7_S3_G3();

    var ind = try DoubleSmoothedStochastic.init(testing.allocator, .{});
    defer ind.deinit();

    var last_out: OutputArray = undefined;
    for (0..252) |i| {
        const bar = Bar{
            .time = 0,
            .open = 0.0,
            .high = high[i],
            .low = low[i],
            .close = input[i],
            .volume = 0.0,
        };
        last_out = ind.updateBar(&bar);
    }
    const items = last_out.slice();

    try testing.expectEqual(@as(usize, 2), items.len);
    try checkVal(exp_dss[251], items[0].scalar.value);
    try checkVal(exp_signal[251], items[1].scalar.value);
}

test "DSS scalar is used as high, low and close" {
    var ind = try DoubleSmoothedStochastic.init(testing.allocator, .{ .q = 2, .r = 1, .s = 1, .g = 1 });
    defer ind.deinit();

    const s0 = Scalar{ .time = 0, .value = 10.0 };
    const first = ind.updateScalar(&s0).slice();
    try testing.expect(math.isNan(first[0].scalar.value));
    try testing.expect(math.isNan(first[1].scalar.value));

    // Close at the 2-bar high.
    const s1 = Scalar{ .time = 0, .value = 12.0 };
    const items = ind.updateScalar(&s1).slice();
    try checkVal(100.0, items[0].scalar.value);
    try checkVal(100.0, items[1].scalar.value);
}

test "DSS quote mid price is used as high, low and close" {
    var ind = try DoubleSmoothedStochastic.init(testing.allocator, .{ .q = 2, .r = 1, .s = 1, .g = 1 });
    defer ind.deinit();

    const q0 = Quote{ .time = 0, .bid_price = 12.0, .ask_price = 14.0, .bid_size = 1.0, .ask_size = 1.0 };
    _ = ind.updateQuote(&q0);

    // Mid 11 is at the 2-bar low.
    const q1 = Quote{ .time = 0, .bid_price = 10.0, .ask_price = 12.0, .bid_size = 1.0, .ask_size = 1.0 };
    const items = ind.updateQuote(&q1).slice();
    try checkVal(0.0, items[0].scalar.value);
    try checkVal(0.0, items[1].scalar.value);
}

test "DSS trade price is used as high, low and close" {
    var ind = try DoubleSmoothedStochastic.init(testing.allocator, .{ .q = 2, .r = 1, .s = 1, .g = 1 });
    defer ind.deinit();

    const t0 = Trade{ .time = 0, .price = 10.0, .volume = 1.0 };
    _ = ind.updateTrade(&t0);

    const t1 = Trade{ .time = 0, .price = 12.0, .volume = 1.0 };
    const items = ind.updateTrade(&t1).slice();
    try checkVal(100.0, items[0].scalar.value);
    try checkVal(100.0, items[1].scalar.value);
}
