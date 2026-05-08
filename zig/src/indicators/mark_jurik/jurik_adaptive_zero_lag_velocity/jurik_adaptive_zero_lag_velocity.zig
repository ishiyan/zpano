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
const line_indicator_mod = @import("../../core/line_indicator.zig");
const build_metadata_mod = @import("../../core/build_metadata.zig");
const component_triple_mnemonic_mod = @import("../../core/component_triple_mnemonic.zig");
const metadata_mod = @import("../../core/metadata.zig");

const OutputArray = indicator_mod.OutputArray;
const LineIndicator = line_indicator_mod.LineIndicator;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the Jurik Adaptive Zero Lag Velocity indicator.
pub const JurikAdaptiveZeroLagVelocityOutput = enum(u8) {
    /// The velocity value.
    value = 1,
};

/// Parameters for the Jurik Adaptive Zero Lag Velocity.
pub const JurikAdaptiveZeroLagVelocityParams = struct {
    /// Minimum adaptive length. Must be >= 2.
    lo_length: u32 = 5,
    /// Maximum adaptive length. Must be >= lo_length.
    hi_length: u32 = 30,
    /// Sensitivity for adaptive depth. Default 1.0.
    sensitivity: f64 = 1.0,
    /// Period for the velocity smoother. Must be > 0. Default 3.0.
    period: f64 = 3.0,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

const MAX_BARS = 1024;
const SMOOTH_BUFFER_SIZE = 1001;

/// Adaptive smoother (Stage 2) for JAVEL.
const VelSmooth = struct {
    jrc03: f64,
    jrc06: u32,
    jrc07: u32,
    ema_factor: f64,
    damping: f64,
    eps2: f64,
    buffer: [SMOOTH_BUFFER_SIZE]f64,
    head: u32,
    length: u32,
    bar_count: u32,
    velocity: f64,
    position: f64,
    smoothed_mad: f64,
    initialized: bool,

    fn init(period: f64) VelSmooth {
        const eps2: f64 = 0.0001;
        const jrc03 = @min(500.0, @max(eps2, period));
        const jrc06: u32 = @intFromFloat(@max(31.0, @ceil(2.0 * period)));
        const jrc07: u32 = @intFromFloat(@min(30.0, @ceil(period)));
        const ema_factor = 1.0 - @exp(-@log(4.0) / (period / 2.0));
        const damping = 0.86 - 0.55 / @sqrt(jrc03);

        return .{
            .jrc03 = jrc03,
            .jrc06 = jrc06,
            .jrc07 = jrc07,
            .ema_factor = ema_factor,
            .damping = damping,
            .eps2 = eps2,
            .buffer = [_]f64{0} ** SMOOTH_BUFFER_SIZE,
            .head = 0,
            .length = 0,
            .bar_count = 0,
            .velocity = 0,
            .position = 0,
            .smoothed_mad = 0,
            .initialized = false,
        };
    }

    fn update(self: *VelSmooth, value: f64) f64 {
        self.bar_count += 1;

        // Store in circular buffer.
        const old_index = self.head % SMOOTH_BUFFER_SIZE;
        self.buffer[old_index] = value;
        self.head += 1;

        if (self.length < self.jrc06) {
            self.length += 1;
        }

        const length = self.length;

        // First bar: initialize position.
        if (length < 2) {
            if (!self.initialized) {
                self.position = value;
                self.initialized = true;
            }
            return self.position;
        }

        if (!self.initialized) {
            self.position = value;
            self.initialized = true;
        }

        // Linear regression over buffer.
        var sum_values: f64 = 0;
        var sum_weighted: f64 = 0;

        for (0..length) |k| {
            var idx: i32 = @as(i32, @intCast(self.head)) - @as(i32, @intCast(length)) + @as(i32, @intCast(k));
            if (idx < 0) idx += SMOOTH_BUFFER_SIZE;
            const uidx: usize = @intCast(@mod(idx, SMOOTH_BUFFER_SIZE));

            const bv = self.buffer[uidx];
            sum_values += bv;
            sum_weighted += bv * @as(f64, @floatFromInt(k));
        }

        const fl: f64 = @floatFromInt(length);
        const midpoint = (fl - 1.0) / 2.0;
        const sum_x_sq = fl * (fl - 1.0) * (2.0 * fl - 1.0) / 6.0;
        const regression_denom = sum_x_sq - fl * midpoint * midpoint;

        var regression_slope: f64 = 0;
        if (@abs(regression_denom) >= self.eps2) {
            regression_slope = (sum_weighted - midpoint * sum_values) / regression_denom;
        }

        const intercept = sum_values / fl - regression_slope * midpoint;

        // Compute MAD from regression residuals.
        var sum_abs_dev: f64 = 0;

        for (0..length) |k| {
            var idx: i32 = @as(i32, @intCast(self.head)) - @as(i32, @intCast(length)) + @as(i32, @intCast(k));
            if (idx < 0) idx += SMOOTH_BUFFER_SIZE;
            const uidx: usize = @intCast(@mod(idx, SMOOTH_BUFFER_SIZE));

            const predicted = intercept + regression_slope * @as(f64, @floatFromInt(k));
            sum_abs_dev += @abs(self.buffer[uidx] - predicted);
        }

        var raw_mad = sum_abs_dev / fl;
        const scale = 1.2 * std.math.pow(f64, @as(f64, @floatFromInt(self.jrc06)) / fl, 0.25);
        raw_mad *= scale;

        // Smooth MAD with EMA.
        if (self.bar_count <= self.jrc07 + 1) {
            self.smoothed_mad = raw_mad;
        } else {
            self.smoothed_mad += self.ema_factor * (raw_mad - self.smoothed_mad);
        }

        // Adaptive velocity/position dynamics.
        const prediction_error = value - self.position;

        var response_factor: f64 = undefined;
        if (self.smoothed_mad * self.jrc03 < self.eps2) {
            response_factor = 1.0;
        } else {
            response_factor = 1.0 - @exp(-@abs(prediction_error) / (self.smoothed_mad * self.jrc03));
        }

        self.velocity = response_factor * prediction_error + self.velocity * self.damping;
        self.position += self.velocity;

        return self.position;
    }
};

/// Jurik Adaptive Zero Lag Velocity (JAVEL) indicator.
pub const JurikAdaptiveZeroLagVelocity = struct {
    line: LineIndicator,
    primed: bool,
    lo_length: u32,
    hi_length: u32,
    sensitivity: f64,
    eps: f64,

    prices: [MAX_BARS]f64,
    value1: [MAX_BARS]f64,
    bar_count: u32,
    smooth: VelSmooth,

    mnemonic_buf: [128]u8,
    mnemonic_len: usize,

    pub fn init(params: JurikAdaptiveZeroLagVelocityParams) !JurikAdaptiveZeroLagVelocity {
        const lo_length = params.lo_length;
        const hi_length = params.hi_length;
        const sensitivity = params.sensitivity;
        const period = params.period;

        if (lo_length < 2) return error.InvalidLoLength;
        if (hi_length < lo_length) return error.InvalidHiLength;
        if (period <= 0) return error.InvalidPeriod;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [128]u8 = undefined;
        const mnemonic = std.fmt.bufPrint(&mnemonic_buf, "javel({d}, {d}, {d:.2}, {d:.2}{s})", .{
            lo_length, hi_length, sensitivity, period, triple,
        }) catch unreachable;
        const mnemonic_len = mnemonic.len;

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                "Jurik adaptive zero lag velocity ",
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .primed = false,
            .lo_length = lo_length,
            .hi_length = hi_length,
            .sensitivity = sensitivity,
            .eps = 0.001,
            .prices = [_]f64{0} ** MAX_BARS,
            .value1 = [_]f64{0} ** MAX_BARS,
            .bar_count = 0,
            .smooth = VelSmooth.init(period),
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
        };
    }

    pub fn fixSlices(self: *JurikAdaptiveZeroLagVelocity) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn computeAdaptiveDepth(self: *JurikAdaptiveZeroLagVelocity, bar: u32) f64 {
        var long_window = bar;
        if (long_window > 99) long_window = 99;
        long_window += 1;

        var short_window = bar;
        if (short_window > 9) short_window = 9;
        short_window += 1;

        var avg1: f64 = 0;
        for (0..long_window) |i| {
            avg1 += self.value1[bar + 1 - long_window + @as(u32, @intCast(i))];
        }
        avg1 /= @as(f64, @floatFromInt(long_window));

        var avg2: f64 = 0;
        for (0..short_window) |i| {
            avg2 += self.value1[bar + 1 - short_window + @as(u32, @intCast(i))];
        }
        avg2 /= @as(f64, @floatFromInt(short_window));

        const value2 = self.sensitivity * @log((self.eps + avg1) / (self.eps + avg2));
        const value3 = value2 / (1.0 + @abs(value2));

        return @as(f64, @floatFromInt(self.lo_length)) +
            @as(f64, @floatFromInt(self.hi_length - self.lo_length)) * (1.0 + value3) / 2.0;
    }

    fn computeWLSSlope(self: *JurikAdaptiveZeroLagVelocity, bar: u32, depth: u32) f64 {
        const n: f64 = @floatFromInt(depth + 1);
        const s1 = n * (n + 1.0) / 2.0;
        const s2 = s1 * (2.0 * n + 1.0) / 3.0;
        const denom = s1 * s1 * s1 - s2 * s2;

        var sum_xw: f64 = 0;
        var sum_xw2: f64 = 0;

        for (0..depth + 1) |i| {
            const w = n - @as(f64, @floatFromInt(i));
            const p = self.prices[bar - @as(u32, @intCast(i))];
            sum_xw += p * w;
            sum_xw2 += p * w * w;
        }

        return (sum_xw2 * s1 - sum_xw * s2) / denom;
    }

    pub fn update(self: *JurikAdaptiveZeroLagVelocity, sample: f64) f64 {
        if (math.isNan(sample)) return sample;

        const bar = self.bar_count;
        self.bar_count += 1;

        self.prices[bar] = sample;

        // Compute value1 (abs diff).
        if (bar == 0) {
            self.value1[bar] = 0.0;
        } else {
            self.value1[bar] = @abs(sample - self.prices[bar - 1]);
        }

        // Compute adaptive depth.
        const adaptive_depth = self.computeAdaptiveDepth(bar);
        const depth: u32 = @intFromFloat(@ceil(adaptive_depth));

        // Check if we have enough prices for WLS.
        if (bar < depth) {
            return math.nan(f64);
        }

        // Stage 1: WLS slope.
        const slope = self.computeWLSSlope(bar, depth);

        // Stage 2: adaptive smoother.
        const result = self.smooth.update(slope);

        if (!self.primed) {
            self.primed = true;
        }

        return result;
    }

    pub fn isPrimed(self: *const JurikAdaptiveZeroLagVelocity) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const JurikAdaptiveZeroLagVelocity, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .jurik_adaptive_zero_lag_velocity,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *JurikAdaptiveZeroLagVelocity, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *JurikAdaptiveZeroLagVelocity, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *JurikAdaptiveZeroLagVelocity, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *JurikAdaptiveZeroLagVelocity, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *JurikAdaptiveZeroLagVelocity) indicator_mod.Indicator {
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
        const self: *JurikAdaptiveZeroLagVelocity = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const JurikAdaptiveZeroLagVelocity = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *JurikAdaptiveZeroLagVelocity = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *JurikAdaptiveZeroLagVelocity = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *JurikAdaptiveZeroLagVelocity = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *JurikAdaptiveZeroLagVelocity = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const Error = error{
        InvalidLoLength,
        InvalidHiLength,
        InvalidPeriod,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) <= epsilon;
}

fn runJavelTest(lo_length: u32, hi_length: u32, sensitivity: f64, period: f64, expected: [252]f64) !void {
    var ind = JurikAdaptiveZeroLagVelocity.init(.{
        .lo_length = lo_length,
        .hi_length = hi_length,
        .sensitivity = sensitivity,
        .period = period,
    }) catch unreachable;
    ind.fixSlices();
    const input = testdata.testInput();
    const eps = 1e-13;

    for (0..252) |i| {
        const v = ind.update(input[i]);
        if (math.isNan(expected[i])) {
            try testing.expect(math.isNan(v));
        } else {
            try testing.expect(!math.isNan(v));
            if (!almostEqual(v, expected[i], eps)) {
                std.debug.print("FAIL [{d}] lo={d} hi={d} sens={d:.2} per={d:.2}: expected {d}, got {d}, diff {d}\n", .{ i, lo_length, hi_length, sensitivity, period, expected[i], v, @abs(v - expected[i]) });
                return error.TestUnexpectedResult;
            }
        }
    }

    try testing.expect(math.isNan(ind.update(math.nan(f64))));
}

test "javel lo=2 hi=15" {
    try runJavelTest(2, 15, 1.0, 3.0, testdata.expectedLo2Hi15());
}
test "javel lo=2 hi=30" {
    try runJavelTest(2, 30, 1.0, 3.0, testdata.expectedLo2Hi30());
}
test "javel lo=2 hi=60" {
    try runJavelTest(2, 60, 1.0, 3.0, testdata.expectedLo2Hi60());
}
test "javel lo=5 hi=15" {
    try runJavelTest(5, 15, 1.0, 3.0, testdata.expectedLo5Hi15());
}
test "javel lo=5 hi=30" {
    try runJavelTest(5, 30, 1.0, 3.0, testdata.expectedLo5Hi30());
}
test "javel lo=5 hi=60" {
    try runJavelTest(5, 60, 1.0, 3.0, testdata.expectedLo5Hi60());
}
test "javel lo=10 hi=15" {
    try runJavelTest(10, 15, 1.0, 3.0, testdata.expectedLo10Hi15());
}
test "javel lo=10 hi=30" {
    try runJavelTest(10, 30, 1.0, 3.0, testdata.expectedLo10Hi30());
}
test "javel lo=10 hi=60" {
    try runJavelTest(10, 60, 1.0, 3.0, testdata.expectedLo10Hi60());
}
test "javel sens=0.5" {
    try runJavelTest(5, 30, 0.5, 3.0, testdata.expectedSens05());
}
test "javel sens=2.5" {
    try runJavelTest(5, 30, 2.5, 3.0, testdata.expectedSens25());
}
test "javel sens=5.0" {
    try runJavelTest(5, 30, 5.0, 3.0, testdata.expectedSens50());
}
test "javel period=1.5" {
    try runJavelTest(5, 30, 1.0, 1.5, testdata.expectedPeriod15());
}
test "javel period=10.0" {
    try runJavelTest(5, 30, 1.0, 10.0, testdata.expectedPeriod100());
}
test "javel period=30.0" {
    try runJavelTest(5, 30, 1.0, 30.0, testdata.expectedPeriod300());
}
