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

/// Enumerates the outputs of the zero-lag error-correcting exponential moving average indicator.
pub const ZeroLagErrorCorrectingExponentialMovingAverageOutput = enum(u8) {
    value = 1,
};

/// Parameters to create an instance of the Zero-lag Error-Correcting EMA.
pub const ZeroLagErrorCorrectingExponentialMovingAverageParams = struct {
    /// Smoothing factor (alpha) of the EMA. Must be in (0, 1]. Default is 0.095.
    smoothing_factor: f64 = 0.095,
    /// Range [-g, g] for finding the best gain factor. Must be positive. Default is 5.
    gain_limit: f64 = 5,
    /// Iteration step for finding the best gain factor. Must be positive. Default is 0.1.
    gain_step: f64 = 0.1,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Ehler's adaptive zero-lag error-correcting exponential moving average (ZECEMA).
///
/// The algorithm iterates over gain values in [-gainLimit, gainLimit] with the given
/// gainStep to find the gain that minimizes the error between the sample and the
/// error-corrected EMA value.
///
/// The indicator is not primed during the first two updates; it primes on the third.
pub const ZeroLagErrorCorrectingExponentialMovingAverage = struct {
    line: LineIndicator,
    alpha: f64,
    one_min_alpha: f64,
    gain_limit: f64,
    gain_step: f64,
    count: usize,
    value: f64,
    ema_value: f64,
    primed: bool,
    mnemonic_buf: [128]u8,
    mnemonic_len: usize,
    description_buf: [200]u8,
    description_len: usize,

    pub fn init(params: ZeroLagErrorCorrectingExponentialMovingAverageParams) !ZeroLagErrorCorrectingExponentialMovingAverage {
        const sf = params.smoothing_factor;
        if (sf <= 0 or sf > 1) {
            return error.InvalidSmoothingFactor;
        }

        const gl = params.gain_limit;
        if (gl <= 0) {
            return error.InvalidGainLimit;
        }

        const gs = params.gain_step;
        if (gs <= 0) {
            return error.InvalidGainStep;
        }

        // Build mnemonic.
        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(
            &triple_buf,
            params.bar_component orelse bar_component.BarComponent.close,
            params.quote_component orelse quote_component.default_quote_component,
            params.trade_component orelse trade_component.default_trade_component,
        );

        var mnemonic_buf: [128]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "zecema({d}, {d}, {d}{s})", .{ sf, gl, gs, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [200]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Zero-lag Error-Correcting Exponential Moving Average {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                params.bar_component orelse bar_component.BarComponent.close,
                params.quote_component orelse quote_component.default_quote_component,
                params.trade_component orelse trade_component.default_trade_component,
            ),
            .alpha = sf,
            .one_min_alpha = 1.0 - sf,
            .gain_limit = gl,
            .gain_step = gs,
            .count = 0,
            .value = math.nan(f64),
            .ema_value = math.nan(f64),
            .primed = false,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn fixSlices(self: *ZeroLagErrorCorrectingExponentialMovingAverage) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    /// Update the indicator given the next sample.
    pub fn update(self: *ZeroLagErrorCorrectingExponentialMovingAverage, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        if (self.primed) {
            self.value = self.calculate(sample);
            return self.value;
        }

        self.count += 1;

        if (self.count == 1) {
            self.ema_value = sample;
            return math.nan(f64);
        }

        if (self.count == 2) {
            self.ema_value = self.calculateEma(sample);
            self.value = self.ema_value;
            return math.nan(f64);
        }

        // count == 3: prime the indicator.
        self.value = self.calculate(sample);
        self.primed = true;
        return self.value;
    }

    fn calculateEma(self: *ZeroLagErrorCorrectingExponentialMovingAverage, sample: f64) f64 {
        return self.alpha * sample + self.one_min_alpha * self.ema_value;
    }

    fn calculate(self: *ZeroLagErrorCorrectingExponentialMovingAverage, sample: f64) f64 {
        self.ema_value = self.calculateEma(sample);

        var least_error: f64 = math.floatMax(f64);
        var best_ec: f64 = 0.0;

        var gain: f64 = -self.gain_limit;
        while (gain <= self.gain_limit) : (gain += self.gain_step) {
            const ec = self.alpha * (self.ema_value + gain * (sample - self.value)) + self.one_min_alpha * self.value;
            const err = @abs(sample - ec);

            if (least_error > err) {
                least_error = err;
                best_ec = ec;
            }
        }

        return best_ec;
    }

    pub fn isPrimed(self: *const ZeroLagErrorCorrectingExponentialMovingAverage) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const ZeroLagErrorCorrectingExponentialMovingAverage, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .zero_lag_error_correcting_exponential_moving_average,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *ZeroLagErrorCorrectingExponentialMovingAverage, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *ZeroLagErrorCorrectingExponentialMovingAverage, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *ZeroLagErrorCorrectingExponentialMovingAverage, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *ZeroLagErrorCorrectingExponentialMovingAverage, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *ZeroLagErrorCorrectingExponentialMovingAverage) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(ZeroLagErrorCorrectingExponentialMovingAverage);
};

// --- Tests ---
const testing = std.testing;

test "ZECEMA isPrimed" {
    var z = try ZeroLagErrorCorrectingExponentialMovingAverage.init(.{});
    z.fixSlices();

    try testing.expect(!z.isPrimed());

    _ = z.update(100);
    try testing.expect(!z.isPrimed());

    _ = z.update(100);
    try testing.expect(!z.isPrimed());

    _ = z.update(100);
    try testing.expect(z.isPrimed());
}

test "ZECEMA update NaN passthrough" {
    var z = try ZeroLagErrorCorrectingExponentialMovingAverage.init(.{});
    z.fixSlices();

    try testing.expect(math.isNan(z.update(math.nan(f64))));
    try testing.expect(!z.isPrimed());
}

test "ZECEMA update constant" {
    const value: f64 = 42.0;

    var z = try ZeroLagErrorCorrectingExponentialMovingAverage.init(.{});
    z.fixSlices();

    // First 2 updates should return NaN.
    try testing.expect(math.isNan(z.update(value)));
    try testing.expect(math.isNan(z.update(value)));

    // 3rd update primes.
    const act = z.update(value);
    try testing.expect(!math.isNan(act));
    try testing.expect(@abs(act - value) <= 1e-6);

    // Further updates with same constant should stay close.
    for (0..10) |_| {
        const v = z.update(value);
        try testing.expect(@abs(v - value) <= 1e-6);
    }
}

test "ZECEMA metadata" {
    var z = try ZeroLagErrorCorrectingExponentialMovingAverage.init(.{});
    z.fixSlices();
    var meta: Metadata = undefined;
    z.getMetadata(&meta);

    try testing.expectEqual(Identifier.zero_lag_error_correcting_exponential_moving_average, meta.identifier);
    try testing.expectEqualStrings("zecema(0.095, 5, 0.1)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
    try testing.expectEqualStrings(
        "Zero-lag Error-Correcting Exponential Moving Average zecema(0.095, 5, 0.1)",
        meta.outputs_buf[0].description,
    );
}

test "ZECEMA constructor validation" {
    // Valid default.
    _ = try ZeroLagErrorCorrectingExponentialMovingAverage.init(.{});

    // Smoothing factor out of range.
    try testing.expectError(error.InvalidSmoothingFactor, ZeroLagErrorCorrectingExponentialMovingAverage.init(.{ .smoothing_factor = 0 }));
    try testing.expectError(error.InvalidSmoothingFactor, ZeroLagErrorCorrectingExponentialMovingAverage.init(.{ .smoothing_factor = -0.1 }));
    try testing.expectError(error.InvalidSmoothingFactor, ZeroLagErrorCorrectingExponentialMovingAverage.init(.{ .smoothing_factor = 1.1 }));

    // sf = 1 should be valid.
    _ = try ZeroLagErrorCorrectingExponentialMovingAverage.init(.{ .smoothing_factor = 1 });

    // Gain limit out of range.
    try testing.expectError(error.InvalidGainLimit, ZeroLagErrorCorrectingExponentialMovingAverage.init(.{ .gain_limit = 0 }));
    try testing.expectError(error.InvalidGainLimit, ZeroLagErrorCorrectingExponentialMovingAverage.init(.{ .gain_limit = -1 }));

    // Gain step out of range.
    try testing.expectError(error.InvalidGainStep, ZeroLagErrorCorrectingExponentialMovingAverage.init(.{ .gain_step = 0 }));
    try testing.expectError(error.InvalidGainStep, ZeroLagErrorCorrectingExponentialMovingAverage.init(.{ .gain_step = -0.1 }));
}

test "ZECEMA updateBar" {
    var z = try ZeroLagErrorCorrectingExponentialMovingAverage.init(.{});
    z.fixSlices();

    const bar1 = Bar{ .time = 1000, .open = 91, .high = 100, .low = 90, .close = 95, .volume = 1000 };
    const out1 = z.updateBar(&bar1);
    try testing.expect(math.isNan(out1.slice()[0].scalar.value));

    // Prime.
    const bar2 = Bar{ .time = 2000, .open = 92, .high = 101, .low = 91, .close = 96, .volume = 1000 };
    _ = z.updateBar(&bar2);
    const bar3 = Bar{ .time = 3000, .open = 93, .high = 102, .low = 92, .close = 97, .volume = 1000 };
    const out3 = z.updateBar(&bar3);
    try testing.expect(!math.isNan(out3.slice()[0].scalar.value));
}

test "ZECEMA custom bar component mnemonic" {
    var z = try ZeroLagErrorCorrectingExponentialMovingAverage.init(.{ .bar_component = bar_component.BarComponent.open });
    z.fixSlices();
    var meta: Metadata = undefined;
    z.getMetadata(&meta);
    try testing.expectEqualStrings("zecema(0.095, 5, 0.1, o)", meta.mnemonic);
}
