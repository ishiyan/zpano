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

/// Enumerates the outputs of the Parabolic Stop And Reverse indicator.
pub const ParabolicStopAndReverseOutput = enum(u8) {
    /// The scalar value of the Parabolic Stop And Reverse.
    /// Positive values indicate a long position; negative values indicate a short position.
    value = 1,
};

const default_acceleration_init: f64 = 0.02;
const default_acceleration_step: f64 = 0.02;
const default_acceleration_max: f64 = 0.20;

pub const ParabolicStopAndReverseParams = struct {
    start_value: f64 = 0.0,
    offset_on_reverse: f64 = 0.0,
    acceleration_init_long: f64 = 0.0,
    acceleration_long: f64 = 0.0,
    acceleration_max_long: f64 = 0.0,
    acceleration_init_short: f64 = 0.0,
    acceleration_short: f64 = 0.0,
    acceleration_max_short: f64 = 0.0,
};

/// Welles Wilder's Parabolic Stop And Reverse (SAR) indicator.
///
/// The Parabolic SAR provides potential entry and exit points. It places dots above or below
/// the price to indicate the direction of the trend. When the dots are below the price, it
/// signals a long (upward) trend; when above, it signals a short (downward) trend.
///
/// This is the "extended" version (SAREXT) which supports separate acceleration factor
/// parameters for long and short directions, an optional start value to force the initial
/// direction, and a percent offset on reversal. The output is signed: positive values
/// indicate long positions, negative values indicate short positions.
pub const ParabolicStopAndReverse = struct {
    // Parameters (resolved from defaults).
    start_value: f64,
    offset_on_reverse: f64,
    af_init_long: f64,
    af_step_long: f64,
    af_max_long: f64,
    af_init_short: f64,
    af_step_short: f64,
    af_max_short: f64,

    // State.
    count: i32,
    is_long: bool,
    sar: f64,
    ep: f64,
    af_long: f64,
    af_short: f64,
    previous_high: f64,
    previous_low: f64,
    new_high: f64,
    new_low: f64,
    primed: bool,
    value: f64,

    const mnemonic_str = "sar()";
    const description_str = "Parabolic Stop And Reverse sar()";

    pub const Error = error{
        InvalidAcceleration,
        InvalidOffset,
    };

    pub fn init(p: ParabolicStopAndReverseParams) Error!ParabolicStopAndReverse {
        // Resolve defaults.
        var af_init_long: f64 = if (p.acceleration_init_long == 0) default_acceleration_init else p.acceleration_init_long;
        var af_step_long: f64 = if (p.acceleration_long == 0) default_acceleration_step else p.acceleration_long;
        const af_max_long: f64 = if (p.acceleration_max_long == 0) default_acceleration_max else p.acceleration_max_long;
        var af_init_short: f64 = if (p.acceleration_init_short == 0) default_acceleration_init else p.acceleration_init_short;
        var af_step_short: f64 = if (p.acceleration_short == 0) default_acceleration_step else p.acceleration_short;
        const af_max_short: f64 = if (p.acceleration_max_short == 0) default_acceleration_max else p.acceleration_max_short;

        // Validate.
        if (af_init_long < 0 or af_step_long < 0 or af_max_long < 0) {
            return Error.InvalidAcceleration;
        }
        if (af_init_short < 0 or af_step_short < 0 or af_max_short < 0) {
            return Error.InvalidAcceleration;
        }
        if (p.offset_on_reverse < 0) {
            return Error.InvalidOffset;
        }

        // Clamp: init and step cannot exceed max.
        if (af_init_long > af_max_long) af_init_long = af_max_long;
        if (af_step_long > af_max_long) af_step_long = af_max_long;
        if (af_init_short > af_max_short) af_init_short = af_max_short;
        if (af_step_short > af_max_short) af_step_short = af_max_short;

        return .{
            .start_value = p.start_value,
            .offset_on_reverse = p.offset_on_reverse,
            .af_init_long = af_init_long,
            .af_step_long = af_step_long,
            .af_max_long = af_max_long,
            .af_init_short = af_init_short,
            .af_step_short = af_step_short,
            .af_max_short = af_max_short,
            .af_long = af_init_long,
            .af_short = af_init_short,
            .count = 0,
            .is_long = false,
            .sar = 0,
            .ep = 0,
            .previous_high = 0,
            .previous_low = 0,
            .new_high = 0,
            .new_low = 0,
            .primed = false,
            .value = math.nan(f64),
        };
    }

    pub fn deinit(_: *ParabolicStopAndReverse) void {}
    pub fn fixSlices(_: *ParabolicStopAndReverse) void {}

    /// Update with a single scalar sample (high = low = sample).
    pub fn update(self: *ParabolicStopAndReverse, sample: f64) f64 {
        if (math.isNan(sample)) return math.nan(f64);
        return self.updateHL(sample, sample);
    }

    /// Update with high and low values.
    pub fn updateHL(self: *ParabolicStopAndReverse, high: f64, low: f64) f64 {
        if (math.isNan(high) or math.isNan(low)) return math.nan(f64);

        self.count += 1;

        // First bar: store high/low, no output yet.
        if (self.count == 1) {
            self.new_high = high;
            self.new_low = low;
            self.value = math.nan(f64);
            return math.nan(f64);
        }

        // Second bar: initialize SAR, EP, and direction.
        if (self.count == 2) {
            const previous_high = self.new_high;
            const previous_low = self.new_low;

            if (self.start_value == 0) {
                // Auto-detect direction using MINUS_DM logic.
                var minus_dm = previous_low - low;
                var plus_dm = high - previous_high;
                if (minus_dm < 0) minus_dm = 0;
                if (plus_dm < 0) plus_dm = 0;

                self.is_long = (minus_dm <= plus_dm);

                if (self.is_long) {
                    self.ep = high;
                    self.sar = previous_low;
                } else {
                    self.ep = low;
                    self.sar = previous_high;
                }
            } else if (self.start_value > 0) {
                self.is_long = true;
                self.ep = high;
                self.sar = self.start_value;
            } else {
                self.is_long = false;
                self.ep = low;
                self.sar = @abs(self.start_value);
            }

            self.new_high = high;
            self.new_low = low;
            self.primed = true;
            // Fall through to main loop logic below.
        }

        // Main SAR calculation (bars 2+).
        if (self.count >= 2) {
            self.previous_low = self.new_low;
            self.previous_high = self.new_high;
            self.new_low = low;
            self.new_high = high;

            if (self.count == 2) {
                // On the second call, re-assign to match TaLib algorithm behavior.
                self.previous_low = self.new_low;
                self.previous_high = self.new_high;
            }

            if (self.is_long) {
                self.value = self.updateLong();
            } else {
                self.value = self.updateShort();
            }
            return self.value;
        }

        return math.nan(f64);
    }

    fn updateLong(self: *ParabolicStopAndReverse) f64 {
        // Switch to short if the low penetrates the SAR value.
        if (self.new_low <= self.sar) {
            self.is_long = false;
            self.sar = self.ep;

            if (self.sar < self.previous_high) self.sar = self.previous_high;
            if (self.sar < self.new_high) self.sar = self.new_high;

            if (self.offset_on_reverse != 0.0) {
                self.sar += self.sar * self.offset_on_reverse;
            }

            const result = -self.sar;

            // Reset short AF and set EP.
            self.af_short = self.af_init_short;
            self.ep = self.new_low;

            // Calculate the new SAR.
            self.sar = self.sar + self.af_short * (self.ep - self.sar);

            if (self.sar < self.previous_high) self.sar = self.previous_high;
            if (self.sar < self.new_high) self.sar = self.new_high;

            return result;
        }

        // No switch — output the current SAR.
        const result = self.sar;

        // Adjust AF and EP.
        if (self.new_high > self.ep) {
            self.ep = self.new_high;
            self.af_long += self.af_step_long;
            if (self.af_long > self.af_max_long) self.af_long = self.af_max_long;
        }

        // Calculate the new SAR.
        self.sar = self.sar + self.af_long * (self.ep - self.sar);

        if (self.sar > self.previous_low) self.sar = self.previous_low;
        if (self.sar > self.new_low) self.sar = self.new_low;

        return result;
    }

    fn updateShort(self: *ParabolicStopAndReverse) f64 {
        // Switch to long if the high penetrates the SAR value.
        if (self.new_high >= self.sar) {
            self.is_long = true;
            self.sar = self.ep;

            if (self.sar > self.previous_low) self.sar = self.previous_low;
            if (self.sar > self.new_low) self.sar = self.new_low;

            if (self.offset_on_reverse != 0.0) {
                self.sar -= self.sar * self.offset_on_reverse;
            }

            const result = self.sar;

            // Reset long AF and set EP.
            self.af_long = self.af_init_long;
            self.ep = self.new_high;

            // Calculate the new SAR.
            self.sar = self.sar + self.af_long * (self.ep - self.sar);

            if (self.sar > self.previous_low) self.sar = self.previous_low;
            if (self.sar > self.new_low) self.sar = self.new_low;

            return result;
        }

        // No switch — output the negated SAR.
        const result = -self.sar;

        // Adjust AF and EP.
        if (self.new_low < self.ep) {
            self.ep = self.new_low;
            self.af_short += self.af_step_short;
            if (self.af_short > self.af_max_short) self.af_short = self.af_max_short;
        }

        // Calculate the new SAR.
        self.sar = self.sar + self.af_short * (self.ep - self.sar);

        if (self.sar < self.previous_high) self.sar = self.previous_high;
        if (self.sar < self.new_high) self.sar = self.new_high;

        return result;
    }

    pub fn isPrimed(self: *const ParabolicStopAndReverse) bool {
        return self.primed;
    }

    pub fn getMetadata(_: *const ParabolicStopAndReverse, out: *Metadata) void {
        build_metadata_mod.buildMetadata(out, Identifier.parabolic_stop_and_reverse, mnemonic_str, description_str, &.{
            .{ .mnemonic = mnemonic_str, .description = description_str },
        });
    }

    fn makeOutput(self: *const ParabolicStopAndReverse, time: i64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = Scalar{ .time = time, .value = self.value } });
        return out;
    }

    pub fn updateScalar(self: *ParabolicStopAndReverse, sample: *const Scalar) OutputArray {
        _ = self.update(sample.value);
        return self.makeOutput(sample.time);
    }

    pub fn updateBar(self: *ParabolicStopAndReverse, sample: *const Bar) OutputArray {
        _ = self.updateHL(sample.high, sample.low);
        return self.makeOutput(sample.time);
    }

    pub fn updateQuote(self: *ParabolicStopAndReverse, sample: *const Quote) OutputArray {
        const mid = (sample.bid_price + sample.ask_price) / 2.0;
        _ = self.update(mid);
        return self.makeOutput(sample.time);
    }

    pub fn updateTrade(self: *ParabolicStopAndReverse, sample: *const Trade) OutputArray {
        _ = self.update(sample.price);
        return self.makeOutput(sample.time);
    }

    // --- Indicator interface ---

    pub fn indicator(self: *ParabolicStopAndReverse) indicator_mod.Indicator {
        return indicator_mod.Indicator{
            .ptr = @ptrCast(self),
            .vtable = &.{
                .isPrimed = vtableIsPrimed,
                .metadata = vtableMetadata,
                .updateScalar = vtableUpdateScalar,
                .updateBar = vtableUpdateBar,
                .updateQuote = vtableUpdateQuote,
                .updateTrade = vtableUpdateTrade,
            },
        };
    }

    fn vtableIsPrimed(ptr: *const anyopaque) bool {
        const self: *const ParabolicStopAndReverse = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *const anyopaque, out: *Metadata) void {
        const self: *const ParabolicStopAndReverse = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *ParabolicStopAndReverse = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *ParabolicStopAndReverse = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *ParabolicStopAndReverse = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *ParabolicStopAndReverse = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }
};

// ============================================================================
// Tests
// ============================================================================

const testing = std.testing;
const testdata = @import("testdata.zig");


fn almostEqual(a: f64, b: f64, tolerance: f64) bool {
    return @abs(a - b) <= tolerance;
}

// Wilder's original SAR test data (38 bars).
const wilder_highs = [_]f64{
    51.12, 52.35, 52.1,  51.8,  52.1,  52.5,  52.8,  52.5,  53.5,  53.5,
    53.8,  54.2,  53.4,  53.5,  54.4,  55.2,  55.7,  57,    57.5,  58,
    57.7,  58,    57.5,  57,    56.7,  57.5,  56.70, 56.00, 56.20, 54.80,
    55.50, 54.70, 54.00, 52.50, 51.00, 51.50, 51.70, 53.00,
};

const wilder_lows = [_]f64{
    50.0,  51.5,  51,    50.5,  51.25, 51.7,  51.85, 51.5,  52.3,  52.5,
    53,    53.5,  52.5,  52.1,  53,    54,    55,    56,    56.5,  57,
    56.5,  57.3,  56.7,  56.3,  56.2,  56,    55.50, 55.00, 54.90, 54.00,
    54.50, 53.80, 53.00, 51.50, 50.00, 50.50, 50.20, 51.50,
};

// High test data, 252 entries.
// Expected SAREXT output for 252-bar dataset with default parameters.
test "ParabolicStopAndReverse 252-bar update" {
    const tol = 1e-6;
    var sar = try ParabolicStopAndReverse.init(.{});

    for (0..testdata.test_highs.len) |i| {
        const result = sar.updateHL(testdata.test_highs[i], testdata.test_lows[i]);
        const exp = testdata.test_expected[i];

        if (math.isNan(exp)) {
            try testing.expect(math.isNan(result));
        } else {
            try testing.expect(!math.isNan(result));
            try testing.expect(almostEqual(result, exp, tol));
        }
    }
}

test "ParabolicStopAndReverse Wilder spot checks" {
    const tol = 1e-3;
    var sar = try ParabolicStopAndReverse.init(.{});

    var results: [wilder_highs.len]f64 = undefined;
    for (0..wilder_highs.len) |i| {
        results[i] = sar.updateHL(wilder_highs[i], wilder_lows[i]);
    }

    // Wilder spot checks from test_sar.c (TA_SAR, absolute values).
    // output[0] corresponds to results[1].
    const SpotCheck = struct { out_index: usize, expected: f64 };
    const spot_checks = [_]SpotCheck{
        .{ .out_index = 0, .expected = 50.00 },
        .{ .out_index = 1, .expected = 50.047 },
        .{ .out_index = 4, .expected = 50.182 },
        .{ .out_index = 35, .expected = 52.93 },
        .{ .out_index = 36, .expected = 50.00 },
    };

    for (spot_checks) |sc| {
        const actual = @abs(results[sc.out_index + 1]); // +1 because results[0] = NaN
        try testing.expect(almostEqual(actual, sc.expected, tol));
    }
}

test "ParabolicStopAndReverse isPrimed" {
    var sar = try ParabolicStopAndReverse.init(.{});

    try testing.expect(!sar.isPrimed());

    _ = sar.updateHL(93.25, 90.75);
    try testing.expect(!sar.isPrimed());

    _ = sar.updateHL(94.94, 91.405);
    try testing.expect(sar.isPrimed());
}

test "ParabolicStopAndReverse constructor validation" {
    // Valid defaults.
    _ = try ParabolicStopAndReverse.init(.{});

    // Negative long init.
    try testing.expectError(error.InvalidAcceleration, ParabolicStopAndReverse.init(.{ .acceleration_init_long = -0.01 }));

    // Negative short step.
    try testing.expectError(error.InvalidAcceleration, ParabolicStopAndReverse.init(.{ .acceleration_short = -0.01 }));

    // Negative offset.
    try testing.expectError(error.InvalidOffset, ParabolicStopAndReverse.init(.{ .offset_on_reverse = -0.01 }));

    // Custom valid.
    _ = try ParabolicStopAndReverse.init(.{
        .acceleration_init_long = 0.01,
        .acceleration_long = 0.01,
        .acceleration_max_long = 0.10,
        .acceleration_init_short = 0.03,
        .acceleration_short = 0.03,
        .acceleration_max_short = 0.30,
    });

    // Forced start values.
    _ = try ParabolicStopAndReverse.init(.{ .start_value = 100.0 });
    _ = try ParabolicStopAndReverse.init(.{ .start_value = -100.0 });
}

test "ParabolicStopAndReverse metadata" {
    var sar = try ParabolicStopAndReverse.init(.{});
    var meta: Metadata = undefined;
    sar.getMetadata(&meta);

    try testing.expectEqual(Identifier.parabolic_stop_and_reverse, meta.identifier);
    try testing.expectEqualStrings("sar()", meta.mnemonic);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
}

test "ParabolicStopAndReverse NaN passthrough" {
    var sar = try ParabolicStopAndReverse.init(.{});

    // Prime with two bars.
    _ = sar.updateHL(93.25, 90.75);
    _ = sar.updateHL(94.94, 91.405);

    // NaN should not corrupt state.
    const nan_result = sar.updateHL(math.nan(f64), 92.0);
    try testing.expect(math.isNan(nan_result));

    // Valid data should still work.
    const result = sar.updateHL(96.375, 94.25);
    try testing.expect(!math.isNan(result));
}

test "ParabolicStopAndReverse updateBar" {
    var sar = try ParabolicStopAndReverse.init(.{});

    const bar1 = Bar{ .time = 1000, .open = 91, .high = 93.25, .low = 90.75, .close = 91.5, .volume = 1000 };
    const out1 = sar.updateBar(&bar1);
    try testing.expect(math.isNan(out1.slice()[0].scalar.value));

    const bar2 = Bar{ .time = 2000, .open = 92, .high = 94.94, .low = 91.405, .close = 94.815, .volume = 1000 };
    const out2 = sar.updateBar(&bar2);
    try testing.expect(!math.isNan(out2.slice()[0].scalar.value));
}

test "ParabolicStopAndReverse forced start long" {
    var sar = try ParabolicStopAndReverse.init(.{ .start_value = 85.0 });

    _ = sar.updateHL(testdata.test_highs[0], testdata.test_lows[0]);
    const result = sar.updateHL(testdata.test_highs[1], testdata.test_lows[1]);
    // Forced long: result should be positive.
    try testing.expect(result > 0);
}

test "ParabolicStopAndReverse forced start short" {
    var sar = try ParabolicStopAndReverse.init(.{ .start_value = -100.0 });

    _ = sar.updateHL(testdata.test_highs[0], testdata.test_lows[0]);
    const result = sar.updateHL(testdata.test_highs[1], testdata.test_lows[1]);
    // Forced short: result should be negative.
    try testing.expect(result < 0);
}
