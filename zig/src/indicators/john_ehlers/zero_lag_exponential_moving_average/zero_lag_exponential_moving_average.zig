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

/// Enumerates the outputs of the zero-lag exponential moving average indicator.
pub const ZeroLagExponentialMovingAverageOutput = enum(u8) {
    value = 1,
};

/// Parameters to create an instance of the Zero-lag EMA.
pub const ZeroLagExponentialMovingAverageParams = struct {
    /// Smoothing factor (alpha) of the EMA. Must be in (0, 1]. Default is 0.25.
    smoothing_factor: f64 = 0.25,
    /// Gain factor used to estimate the velocity. Default is 0.5.
    velocity_gain_factor: f64 = 0.5,
    /// Length of the momentum used to estimate velocity. Must be >= 1. Default is 3.
    velocity_momentum_length: i32 = 3,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Ehler's Zero-lag Exponential Moving Average (ZEMA).
///
/// ZEMA = alpha*(Price + gainFactor*(Price - Price[momentumLength ago])) + (1 - alpha)*ZEMA[previous]
///
/// The indicator is not primed during the first VelocityMomentumLength updates.
pub const ZeroLagExponentialMovingAverage = struct {
    line: LineIndicator,
    allocator: std.mem.Allocator,
    alpha: f64,
    one_min_alpha: f64,
    gain_factor: f64,
    momentum_length: usize,
    momentum_window: []f64,
    count: usize,
    value: f64,
    primed: bool,
    mnemonic_buf: [128]u8,
    mnemonic_len: usize,
    description_buf: [200]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: ZeroLagExponentialMovingAverageParams) !ZeroLagExponentialMovingAverage {
        const sf = params.smoothing_factor;
        if (sf <= 0 or sf > 1) {
            return error.InvalidSmoothingFactor;
        }

        const ml = params.velocity_momentum_length;
        if (ml < 1) {
            return error.InvalidMomentumLength;
        }

        const ml_usize: usize = @intCast(ml);

        // Allocate momentum window of size ml + 1.
        const window = try allocator.alloc(f64, ml_usize + 1);
        @memset(window, 0.0);

        const gf = params.velocity_gain_factor;

        // Build mnemonic.
        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(
            &triple_buf,
            params.bar_component orelse bar_component.BarComponent.close,
            params.quote_component orelse quote_component.default_quote_component,
            params.trade_component orelse trade_component.default_trade_component,
        );

        var mnemonic_buf: [128]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "zema({d}, {d}, {d}{s})", .{ sf, gf, ml, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [200]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Zero-lag Exponential Moving Average {s}", .{mnemonic_slice}) catch
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
            .allocator = allocator,
            .alpha = sf,
            .one_min_alpha = 1.0 - sf,
            .gain_factor = gf,
            .momentum_length = ml_usize,
            .momentum_window = window,
            .count = 0,
            .value = math.nan(f64),
            .primed = false,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *ZeroLagExponentialMovingAverage) void {
        self.allocator.free(self.momentum_window);
    }

    pub fn fixSlices(self: *ZeroLagExponentialMovingAverage) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    /// Update the indicator given the next sample.
    pub fn update(self: *ZeroLagExponentialMovingAverage, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        if (self.primed) {
            // Shift momentum window left by 1.
            std.mem.copyForwards(f64, self.momentum_window[0..self.momentum_length], self.momentum_window[1 .. self.momentum_length + 1]);
            self.momentum_window[self.momentum_length] = sample;
            self.value = self.calculate(sample);
            return self.value;
        }

        self.momentum_window[self.count] = sample;
        self.count += 1;

        if (self.count <= self.momentum_length) {
            self.value = sample;
            return math.nan(f64);
        }

        // count == momentumLength + 1: prime the indicator.
        self.value = self.calculate(sample);
        self.primed = true;
        return self.value;
    }

    fn calculate(self: *ZeroLagExponentialMovingAverage, sample: f64) f64 {
        const momentum = sample - self.momentum_window[0];
        return self.alpha * (sample + self.gain_factor * momentum) + self.one_min_alpha * self.value;
    }

    pub fn isPrimed(self: *const ZeroLagExponentialMovingAverage) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const ZeroLagExponentialMovingAverage, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .zero_lag_exponential_moving_average,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *ZeroLagExponentialMovingAverage, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *ZeroLagExponentialMovingAverage, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *ZeroLagExponentialMovingAverage, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *ZeroLagExponentialMovingAverage, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *ZeroLagExponentialMovingAverage) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(ZeroLagExponentialMovingAverage);
};

// --- Tests ---
const testing = std.testing;

test "ZEMA isPrimed" {
    var z = try ZeroLagExponentialMovingAverage.init(testing.allocator, .{});
    defer z.deinit();
    z.fixSlices();

    try testing.expect(!z.isPrimed());

    // First 3 updates (momentumLength=3) should not prime.
    _ = z.update(100);
    try testing.expect(!z.isPrimed());
    _ = z.update(100);
    try testing.expect(!z.isPrimed());
    _ = z.update(100);
    try testing.expect(!z.isPrimed());

    // 4th update should prime.
    _ = z.update(100);
    try testing.expect(z.isPrimed());
}

test "ZEMA update NaN passthrough" {
    var z = try ZeroLagExponentialMovingAverage.init(testing.allocator, .{});
    defer z.deinit();
    z.fixSlices();

    try testing.expect(math.isNan(z.update(math.nan(f64))));
    try testing.expect(!z.isPrimed());
}

test "ZEMA update constant" {
    const value: f64 = 42.0;

    var z = try ZeroLagExponentialMovingAverage.init(testing.allocator, .{});
    defer z.deinit();
    z.fixSlices();

    // First 3 updates should return NaN.
    for (0..3) |_| {
        try testing.expect(math.isNan(z.update(value)));
    }

    // 4th update primes.
    const act = z.update(value);
    try testing.expect(@abs(act - value) <= 1e-10);

    // Further updates with same constant should stay at value.
    for (0..10) |_| {
        const v = z.update(value);
        try testing.expect(@abs(v - value) <= 1e-10);
    }
}

test "ZEMA metadata" {
    var z = try ZeroLagExponentialMovingAverage.init(testing.allocator, .{});
    defer z.deinit();
    z.fixSlices();
    var meta: Metadata = undefined;
    z.getMetadata(&meta);

    try testing.expectEqual(Identifier.zero_lag_exponential_moving_average, meta.identifier);
    try testing.expectEqualStrings("zema(0.25, 0.5, 3)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
    try testing.expectEqualStrings(
        "Zero-lag Exponential Moving Average zema(0.25, 0.5, 3)",
        meta.outputs_buf[0].description,
    );
}

test "ZEMA constructor validation" {
    // Valid default.
    var z = try ZeroLagExponentialMovingAverage.init(testing.allocator, .{});
    z.deinit();

    // Smoothing factor out of range.
    try testing.expectError(error.InvalidSmoothingFactor, ZeroLagExponentialMovingAverage.init(testing.allocator, .{ .smoothing_factor = 0 }));
    try testing.expectError(error.InvalidSmoothingFactor, ZeroLagExponentialMovingAverage.init(testing.allocator, .{ .smoothing_factor = -0.1 }));
    try testing.expectError(error.InvalidSmoothingFactor, ZeroLagExponentialMovingAverage.init(testing.allocator, .{ .smoothing_factor = 1.1 }));

    // sf = 1 should be valid.
    var z2 = try ZeroLagExponentialMovingAverage.init(testing.allocator, .{ .smoothing_factor = 1 });
    z2.deinit();

    // Momentum length out of range.
    try testing.expectError(error.InvalidMomentumLength, ZeroLagExponentialMovingAverage.init(testing.allocator, .{ .velocity_momentum_length = 0 }));
    try testing.expectError(error.InvalidMomentumLength, ZeroLagExponentialMovingAverage.init(testing.allocator, .{ .velocity_momentum_length = -1 }));
}

test "ZEMA updateBar" {
    var z = try ZeroLagExponentialMovingAverage.init(testing.allocator, .{});
    defer z.deinit();
    z.fixSlices();

    const bar1 = Bar{ .time = 1000, .open = 91, .high = 100, .low = 90, .close = 95, .volume = 1000 };
    const out1 = z.updateBar(&bar1);
    try testing.expect(math.isNan(out1.slice()[0].scalar.value));

    // Prime (need 4 updates for ml=3).
    const bar2 = Bar{ .time = 2000, .open = 92, .high = 101, .low = 91, .close = 96, .volume = 1000 };
    _ = z.updateBar(&bar2);
    const bar3 = Bar{ .time = 3000, .open = 93, .high = 102, .low = 92, .close = 97, .volume = 1000 };
    _ = z.updateBar(&bar3);
    const bar4 = Bar{ .time = 4000, .open = 94, .high = 103, .low = 93, .close = 98, .volume = 1000 };
    const out4 = z.updateBar(&bar4);
    try testing.expect(!math.isNan(out4.slice()[0].scalar.value));
}

test "ZEMA custom bar component mnemonic" {
    var z = try ZeroLagExponentialMovingAverage.init(testing.allocator, .{ .bar_component = bar_component.BarComponent.open });
    defer z.deinit();
    z.fixSlices();
    var meta: Metadata = undefined;
    z.getMetadata(&meta);
    try testing.expectEqualStrings("zema(0.25, 0.5, 3, o)", meta.mnemonic);
}
