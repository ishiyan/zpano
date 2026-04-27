const std = @import("std");
const math = std.math;

const Bar = @import("bar").Bar;
const Quote = @import("quote").Quote;
const Trade = @import("trade").Trade;
const Scalar = @import("scalar").Scalar;
const bar_component = @import("bar_component");
const quote_component = @import("quote_component");
const trade_component = @import("trade_component");

const indicator_mod = @import("../../core/indicator.zig");
const line_indicator_mod = @import("../../core/line_indicator.zig");
const build_metadata_mod = @import("../../core/build_metadata.zig");
const component_triple_mnemonic_mod = @import("../../core/component_triple_mnemonic.zig");
const identifier_mod = @import("../../core/identifier.zig");
const metadata_mod = @import("../../core/metadata.zig");

const OutputArray = indicator_mod.OutputArray;
const LineIndicator = line_indicator_mod.LineIndicator;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the exponential moving average indicator.
pub const ExponentialMovingAverageOutput = enum(u8) {
    value = 1,
};

/// Parameters to create an instance of the EMA based on length.
pub const ExponentialMovingAverageLengthParams = struct {
    length: usize,
    first_is_average: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Parameters to create an instance of the EMA based on smoothing factor.
pub const ExponentialMovingAverageSmoothingFactorParams = struct {
    smoothing_factor: f64,
    first_is_average: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the exponential moving average (EMA).
///
///   EMAᵢ = αPᵢ + (1-α)EMAᵢ₋₁ = EMAᵢ₋₁ + α(Pᵢ - EMAᵢ₋₁), 0 < α ≤ 1.
///
/// The indicator is not primed during the first ℓ−1 updates.
pub const ExponentialMovingAverage = struct {
    line: LineIndicator,
    value: f64,
    sum: f64,
    smoothing_factor: f64,
    length: usize,
    count: usize,
    first_is_average: bool,
    primed: bool,
    mnemonic_buf: [96]u8,
    mnemonic_len: usize,
    description_buf: [160]u8,
    description_len: usize,

    const epsilon: f64 = 0.00000001;

    /// Create EMA from length.
    pub fn initLength(params: ExponentialMovingAverageLengthParams) !ExponentialMovingAverage {
        if (params.length < 1) {
            return error.InvalidLength;
        }
        const alpha = 2.0 / @as(f64, @floatFromInt(1 + params.length));
        return initInternal(params.length, alpha, false, params.first_is_average, params.bar_component, params.quote_component, params.trade_component);
    }

    /// Create EMA from smoothing factor.
    pub fn initSmoothingFactor(params: ExponentialMovingAverageSmoothingFactorParams) !ExponentialMovingAverage {
        var alpha = params.smoothing_factor;
        if (alpha < 0.0 or alpha > 1.0) {
            return error.InvalidSmoothingFactor;
        }
        if (alpha < epsilon) {
            alpha = epsilon;
        }
        const length: usize = @intFromFloat(@round(2.0 / alpha) - 1.0);
        return initInternal(length, alpha, true, params.first_is_average, params.bar_component, params.quote_component, params.trade_component);
    }

    fn initInternal(
        length: usize,
        alpha: f64,
        is_alpha_mode: bool,
        first_is_average: bool,
        bc_opt: ?bar_component.BarComponent,
        qc_opt: ?quote_component.QuoteComponent,
        tc_opt: ?trade_component.TradeComponent,
    ) !ExponentialMovingAverage {
        const bc = bc_opt orelse bar_component.default_bar_component;
        const qc = qc_opt orelse quote_component.default_quote_component;
        const tc = tc_opt orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [96]u8 = undefined;
        var mnemonic_slice: []u8 = undefined;
        if (is_alpha_mode) {
            mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "ema({d}, {d:.8}{s})", .{ length, alpha, triple }) catch
                return error.MnemonicTooLong;
        } else {
            mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "ema({d}{s})", .{ length, triple }) catch
                return error.MnemonicTooLong;
        }
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [160]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Exponential moving average {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                bc_opt,
                qc_opt,
                tc_opt,
            ),
            .value = 0.0,
            .sum = 0.0,
            .smoothing_factor = alpha,
            .length = length,
            .count = 0,
            .first_is_average = first_is_average,
            .primed = false,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn fixSlices(self: *ExponentialMovingAverage) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    pub fn update(self: *ExponentialMovingAverage, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        const temp = sample;

        if (self.primed) {
            self.value += (temp - self.value) * self.smoothing_factor;
        } else {
            self.count += 1;
            if (self.first_is_average) {
                self.sum += temp;
                if (self.count < self.length) {
                    return math.nan(f64);
                }
                self.value = self.sum / @as(f64, @floatFromInt(self.length));
            } else {
                if (self.count == 1) {
                    self.value = temp;
                } else {
                    self.value += (temp - self.value) * self.smoothing_factor;
                }
                if (self.count < self.length) {
                    return math.nan(f64);
                }
            }
            self.primed = true;
        }

        return self.value;
    }

    pub fn isPrimed(self: *const ExponentialMovingAverage) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const ExponentialMovingAverage) Metadata {
        return build_metadata_mod.buildMetadata(
            .exponential_moving_average,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *ExponentialMovingAverage, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *ExponentialMovingAverage, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *ExponentialMovingAverage, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *ExponentialMovingAverage, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *ExponentialMovingAverage) indicator_mod.Indicator {
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
        const self: *ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque) Metadata {
        const self: *const ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.getMetadata();
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidLength,
        InvalidSmoothingFactor,
        MnemonicTooLong,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;

fn testInput() [252]f64 {
    return .{
        91.500000,  94.815000,  94.375000,  95.095000,  93.780000,  94.625000,  92.530000,  92.750000,  90.315000,  92.470000,
        96.125000,  97.250000,  98.500000,  89.875000,  91.000000,  92.815000,  89.155000,  89.345000,  91.625000,  89.875000,
        88.375000,  87.625000,  84.780000,  83.000000,  83.500000,  81.375000,  84.440000,  89.250000,  86.375000,  86.250000,
        85.250000,  87.125000,  85.815000,  88.970000,  88.470000,  86.875000,  86.815000,  84.875000,  84.190000,  83.875000,
        83.375000,  85.500000,  89.190000,  89.440000,  91.095000,  90.750000,  91.440000,  89.000000,  91.000000,  90.500000,
        89.030000,  88.815000,  84.280000,  83.500000,  82.690000,  84.750000,  85.655000,  86.190000,  88.940000,  89.280000,
        88.625000,  88.500000,  91.970000,  91.500000,  93.250000,  93.500000,  93.155000,  91.720000,  90.000000,  89.690000,
        88.875000,  85.190000,  83.375000,  84.875000,  85.940000,  97.250000,  99.875000,  104.940000, 106.000000, 102.500000,
        102.405000, 104.595000, 106.125000, 106.000000, 106.065000, 104.625000, 108.625000, 109.315000, 110.500000, 112.750000,
        123.000000, 119.625000, 118.750000, 119.250000, 117.940000, 116.440000, 115.190000, 111.875000, 110.595000, 118.125000,
        116.000000, 116.000000, 112.000000, 113.750000, 112.940000, 116.000000, 120.500000, 116.620000, 117.000000, 115.250000,
        114.310000, 115.500000, 115.870000, 120.690000, 120.190000, 120.750000, 124.750000, 123.370000, 122.940000, 122.560000,
        123.120000, 122.560000, 124.620000, 129.250000, 131.000000, 132.250000, 131.000000, 132.810000, 134.000000, 137.380000,
        137.810000, 137.880000, 137.250000, 136.310000, 136.250000, 134.630000, 128.250000, 129.000000, 123.870000, 124.810000,
        123.000000, 126.250000, 128.380000, 125.370000, 125.690000, 122.250000, 119.370000, 118.500000, 123.190000, 123.500000,
        122.190000, 119.310000, 123.310000, 121.120000, 123.370000, 127.370000, 128.500000, 123.870000, 122.940000, 121.750000,
        124.440000, 122.000000, 122.370000, 122.940000, 124.000000, 123.190000, 124.560000, 127.250000, 125.870000, 128.860000,
        132.000000, 130.750000, 134.750000, 135.000000, 132.380000, 133.310000, 131.940000, 130.000000, 125.370000, 130.130000,
        127.120000, 125.190000, 122.000000, 125.000000, 123.000000, 123.500000, 120.060000, 121.000000, 117.750000, 119.870000,
        122.000000, 119.190000, 116.370000, 113.500000, 114.250000, 110.000000, 105.060000, 107.000000, 107.870000, 107.000000,
        107.120000, 107.000000, 91.000000,  93.940000,  93.870000,  95.500000,  93.000000,  94.940000,  98.250000,  96.750000,
        94.810000,  94.370000,  91.560000,  90.250000,  93.940000,  93.620000,  97.000000,  95.000000,  95.870000,  94.060000,
        94.620000,  93.750000,  98.000000,  103.940000, 107.870000, 106.060000, 104.500000, 105.000000, 104.190000, 103.060000,
        103.420000, 105.270000, 111.870000, 116.000000, 116.620000, 118.280000, 113.370000, 109.000000, 109.700000, 109.250000,
        107.000000, 109.190000, 110.000000, 109.200000, 110.120000, 108.000000, 108.620000, 109.750000, 109.810000, 109.000000,
        108.750000, 107.870000,
    };
}

fn createEmaLength(length: usize, first_is_average: bool) !ExponentialMovingAverage {
    var ema = try ExponentialMovingAverage.initLength(.{
        .length = length,
        .first_is_average = first_is_average,
    });
    ema.fixSlices();
    return ema;
}

fn createEmaAlpha(alpha: f64, first_is_average: bool) !ExponentialMovingAverage {
    var ema = try ExponentialMovingAverage.initSmoothingFactor(.{
        .smoothing_factor = alpha,
        .first_is_average = first_is_average,
    });
    ema.fixSlices();
    return ema;
}

test "ema update length 2 firstIsAverage true" {
    const input = testInput();
    var ema = try createEmaLength(2, true);

    // Index 0: NaN
    try testing.expect(math.isNan(ema.update(input[0])));
    // Index 1: 93.15
    try testing.expect(@abs(93.15 - ema.update(input[1])) < 1e-2);
    // Index 2: 93.96
    try testing.expect(@abs(93.96 - ema.update(input[2])) < 1e-2);

    for (3..252) |i| {
        _ = ema.update(input[i]);
    }
    // last not checked here — we test via full run
    try testing.expect(math.isNan(ema.update(math.nan(f64))));
}

test "ema update length 10 firstIsAverage true" {
    const input = testInput();
    var ema = try createEmaLength(10, true);

    for (0..9) |i| {
        try testing.expect(math.isNan(ema.update(input[i])));
    }

    var results: [252]f64 = undefined;
    var ema2 = try createEmaLength(10, true);
    for (0..252) |i| {
        results[i] = ema2.update(input[i]);
    }

    try testing.expect(@abs(93.22 - results[9]) < 1e-2);
    try testing.expect(@abs(93.75 - results[10]) < 1e-2);
    try testing.expect(@abs(86.46 - results[29]) < 1e-2);
    try testing.expect(@abs(108.97 - results[251]) < 1e-2);
    try testing.expect(math.isNan(ema2.update(math.nan(f64))));
}

test "ema update length 2 firstIsAverage false (Metastock)" {
    const input = testInput();
    var ema = try createEmaLength(2, false);

    // Index 0: NaN (count < length for length=2)
    try testing.expect(math.isNan(ema.update(input[0])));
    // Index 1: 93.71
    try testing.expect(@abs(93.71 - ema.update(input[1])) < 1e-2);
    // Index 2: 94.15
    try testing.expect(@abs(94.15 - ema.update(input[2])) < 1e-2);

    for (3..252) |i| {
        _ = ema.update(input[i]);
    }
    try testing.expect(math.isNan(ema.update(math.nan(f64))));
}

test "ema update length 10 firstIsAverage false (Metastock)" {
    const input = testInput();
    var ema = try createEmaLength(10, false);

    for (0..9) |i| {
        try testing.expect(math.isNan(ema.update(input[i])));
    }

    var results: [252]f64 = undefined;
    var ema2 = try createEmaLength(10, false);
    for (0..252) |i| {
        results[i] = ema2.update(input[i]);
    }

    try testing.expect(@abs(92.60 - results[9]) < 1e-2);
    try testing.expect(@abs(93.24 - results[10]) < 1e-2);
    try testing.expect(@abs(93.97 - results[11]) < 1e-2);
    try testing.expect(@abs(86.23 - results[30]) < 1e-2);
    try testing.expect(@abs(108.97 - results[251]) < 1e-2);
    try testing.expect(math.isNan(ema2.update(math.nan(f64))));
}

test "ema is primed length 10" {
    const input = testInput();
    var ema = try createEmaLength(10, true);

    try testing.expect(!ema.isPrimed());
    for (0..9) |i| {
        _ = ema.update(input[i]);
        try testing.expect(!ema.isPrimed());
    }
    _ = ema.update(input[9]);
    try testing.expect(ema.isPrimed());
}

test "ema metadata length" {
    var ema = try createEmaLength(10, true);
    const m = ema.getMetadata();

    try testing.expectEqual(Identifier.exponential_moving_average, m.identifier);
    try testing.expectEqualStrings("ema(10)", m.mnemonic);
    try testing.expectEqualStrings("Exponential moving average ema(10)", m.description);
}

test "ema metadata alpha" {
    const alpha: f64 = 2.0 / 11.0;
    var ema = try createEmaAlpha(alpha, false);
    const m = ema.getMetadata();

    try testing.expectEqual(Identifier.exponential_moving_average, m.identifier);
    try testing.expectEqualStrings("ema(10, 0.18181818)", m.mnemonic);
    try testing.expectEqualStrings("Exponential moving average ema(10, 0.18181818)", m.description);
}

test "ema update entity" {
    const alpha: f64 = 2.0 / 3.0;
    const inp: f64 = 3.0;
    const exp: f64 = alpha * inp;
    const time: i64 = 1617235200;

    // scalar
    {
        var ema = try createEmaLength(2, false);
        _ = ema.update(0.0);
        _ = ema.update(0.0);
        const out = ema.updateScalar(&.{ .time = time, .value = inp });
        try testing.expectEqual(@as(usize, 1), out.len);
        const s = out.slice()[0].scalar;
        try testing.expectEqual(time, s.time);
        try testing.expectEqual(exp, s.value);
    }

    // bar
    {
        var ema = try createEmaLength(2, false);
        _ = ema.update(0.0);
        _ = ema.update(0.0);
        const bar = Bar{ .time = time, .open = 0, .high = 0, .low = 0, .close = inp, .volume = 0 };
        const out = ema.updateBar(&bar);
        const s = out.slice()[0].scalar;
        try testing.expectEqual(exp, s.value);
    }
}

test "ema init invalid" {
    // length = 0
    const result = ExponentialMovingAverage.initLength(.{ .length = 0 });
    try testing.expectError(error.InvalidLength, result);

    // alpha < 0
    const result2 = ExponentialMovingAverage.initSmoothingFactor(.{ .smoothing_factor = -1.0 });
    try testing.expectError(error.InvalidSmoothingFactor, result2);

    // alpha > 1
    const result3 = ExponentialMovingAverage.initSmoothingFactor(.{ .smoothing_factor = 2.0 });
    try testing.expectError(error.InvalidSmoothingFactor, result3);
}

test "ema mnemonic with bar component" {
    var ema = try ExponentialMovingAverage.initLength(.{
        .length = 10,
        .bar_component = .median,
    });
    ema.fixSlices();
    try testing.expectEqualStrings("ema(10, hl/2)", ema.line.mnemonic);
    try testing.expectEqualStrings("Exponential moving average ema(10, hl/2)", ema.line.description);
}

test "ema alpha with quote component" {
    var ema = try ExponentialMovingAverage.initSmoothingFactor(.{
        .smoothing_factor = 2.0 / 11.0,
        .bar_component = null,
        .quote_component = .bid,
    });
    ema.fixSlices();
    try testing.expectEqualStrings("ema(10, 0.18181818, b)", ema.line.mnemonic);
}
