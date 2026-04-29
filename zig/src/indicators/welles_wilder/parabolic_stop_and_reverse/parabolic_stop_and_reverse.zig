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
const test_highs = [_]f64{
    93.25,  94.94,  96.375,  96.19,   96,      94.72,  95,     93.72,   92.47,   92.75,
    96.25,  99.625, 99.125,  92.75,   91.315,  93.25,  93.405, 90.655,  91.97,   92.25,
    90.345, 88.5,   88.25,   85.5,    84.44,   84.75,  84.44,  89.405,  88.125,  89.125,
    87.155, 87.25,  87.375,  88.97,   90,      89.845, 86.97,  85.94,   84.75,   85.47,
    84.47,  88.5,   89.47,   90,      92.44,   91.44,  92.97,  91.72,   91.155,  91.75,
    90,     88.875, 89,      85.25,   83.815,  85.25,  86.625, 87.94,   89.375,  90.625,
    90.75,  88.845, 91.97,   93.375,  93.815,  94.03,  94.03,  91.815,  92,      91.94,
    89.75,  88.75,  86.155,  84.875,  85.94,   99.375, 103.28, 105.375, 107.625, 105.25,
    104.5,  105.5,  106.125, 107.94,  106.25,  107,    108.75, 110.94,  110.94,  114.22,
    123,    121.75, 119.815, 120.315, 119.375, 118.19, 116.69, 115.345, 113,     118.315,
    116.87, 116.75, 113.87,  114.62,  115.31,  116,    121.69, 119.87,  120.87,  116.75,
    116.5,  116,    118.31,  121.5,   122,     121.44, 125.75, 127.75,  124.19,  124.44,
    125.75, 124.69, 125.31,  132,     131.31,  132.25, 133.88, 133.5,   135.5,   137.44,
    138.69, 139.19, 138.5,   138.13,  137.5,   138.88, 132.13, 129.75,  128.5,   125.44,
    125.12, 126.5,  128.69,  126.62,  126.69,  126,    123.12, 121.87,  124,     127,
    124.44, 122.5,  123.75,  123.81,  124.5,   127.87, 128.56, 129.63,  124.87,  124.37,
    124.87, 123.62, 124.06,  125.87,  125.19,  125.62, 126,    128.5,   126.75,  129.75,
    132.69, 133.94, 136.5,   137.69,  135.56,  133.56, 135,    132.38,  131.44,  130.88,
    129.63, 127.25, 127.81,  125,     126.81,  124.75, 122.81, 122.25,  121.06,  120,
    123.25, 122.75, 119.19,  115.06,  116.69,  114.87, 110.87, 107.25,  108.87,  109,
    108.5,  113.06, 93,      94.62,   95.12,   96,     95.56,  95.31,   99,      98.81,
    96.81,  95.94,  94.44,   92.94,   93.94,   95.5,   97.06,  97.5,    96.25,   96.37,
    95,     94.87,  98.25,   105.12,  108.44,  109.87, 105,    106,     104.94,  104.5,
    104.44, 106.31, 112.87,  116.5,   119.19,  121,    122.12, 111.94,  112.75,  110.19,
    107.94, 109.69, 111.06,  110.44,  110.12,  110.31, 110.44, 110,     110.75,  110.5,
    110.5,  109.5,
};

const test_lows = [_]f64{
    90.75,  91.405, 94.25,   93.5,   92.815,  93.5,   92,      89.75,   89.44,  90.625,
    92.75,  96.315, 96.03,   88.815, 86.75,   90.94,  88.905,  88.78,   89.25,  89.75,
    87.5,   86.53,  84.625,  82.28,  81.565,  80.875, 81.25,   84.065,  85.595, 85.97,
    84.405, 85.095, 85.5,    85.53,  87.875,  86.565, 84.655,  83.25,   82.565, 83.44,
    82.53,  85.065, 86.875,  88.53,  89.28,   90.125, 90.75,   89,      88.565, 90.095,
    89,     86.47,  84,      83.315, 82,      83.25,  84.75,   85.28,   87.19,  88.44,
    88.25,  87.345, 89.28,   91.095, 89.53,   91.155, 92,      90.53,   89.97,  88.815,
    86.75,  85.065, 82.03,   81.5,   82.565,  96.345, 96.47,   101.155, 104.25, 101.75,
    101.72, 101.72, 103.155, 105.69, 103.655, 104,    105.53,  108.53,  108.75, 107.75,
    117,    118,    116,     118.5,  116.53,  116.25, 114.595, 110.875, 110.5,  110.72,
    112.62, 114.19, 111.19,  109.44, 111.56,  112.44, 117.5,   116.06,  116.56, 113.31,
    112.56, 114,    114.75,  118.87, 119,     119.75, 122.62,  123,     121.75, 121.56,
    123.12, 122.19, 122.75,  124.37, 128,     129.5,  130.81,  130.63,  132.13, 133.88,
    135.38, 135.75, 136.19,  134.5,  135.38,  133.69, 126.06,  126.87,  123.5,  122.62,
    122.75, 123.56, 125.81,  124.62, 124.37,  121.81, 118.19,  118.06,  117.56, 121,
    121.12, 118.94, 119.81,  121,    122,     124.5,  126.56,  123.5,   121.25, 121.06,
    122.31, 121,    120.87,  122.06, 122.75,  122.69, 122.87,  125.5,   124.25, 128,
    128.38, 130.69, 131.63,  134.38, 132,     131.94, 131.94,  129.56,  123.75, 126,
    126.25, 124.37, 121.44,  120.44, 121.37,  121.69, 120,     119.62,  115.5,  116.75,
    119.06, 119.06, 115.06,  111.06, 113.12,  110,    105,     104.69,  103.87, 104.69,
    105.44, 107,    89,      92.5,   92.12,   94.62,  92.81,   94.25,   96.25,  96.37,
    93.69,  93.5,   90,      90.19,  90.5,    92.12,  94.12,   94.87,   93,     93.87,
    93,     92.62,  93.56,   98.37,  104.44,  106,    101.81,  104.12,  103.37, 102.12,
    102.25, 103.37, 107.94,  112.5,  115.44,  115.5,  112.25,  107.56,  106.56, 106.87,
    104.5,  105.75, 108.62,  107.75, 108.06,  108,    108.19,  108.12,  109.06, 108.75,
    108.56, 106.62,
};

// Expected SAREXT output for 252-bar dataset with default parameters.
const test_expected = [_]f64{
    math.nan(f64),   90.7500000000,   90.8338000000,   91.0554480000,   91.2682300800,
    91.4725008768,   91.6686008417,   -96.3750000000,  -96.2425000000,  -95.9704000000,
    89.4400000000,   89.5762000000,   89.9781520000,   -99.6250000000,  -99.4088000000,
    -98.9024480000,  -98.4163500800,  -97.9496960768,  -97.5017082337,  -97.0716399044,
    -96.6587743082,  -96.2624233359,  -95.6784779357,  -94.7941997009,  -93.5427797308,
    -92.1054461631,  -90.5331837003,  80.8750000000,   81.0456000000,   81.2127880000,
    81.3766322400,   81.5371995952,   81.6945556033,   81.8487644912,   81.9998892014,
    82.3198936333,   82.6270978880,   82.9220139725,   -90.0000000000,  -89.8513000000,
    -89.7055740000,  -89.4185510400,  82.5300000000,   82.6688000000,   82.9620480000,
    83.5307251200,   84.0652816128,   84.7776590838,   85.4330463571,   86.0360026485,
    86.5907224366,   -92.9700000000,  -92.8400000000,  -92.4864000000,  -91.9361160000,
    -91.1412267200,  -90.4099285824,  -89.7371342958,  82.0000000000,   82.1475000000,
    82.4866000000,   82.9824040000,   83.4484597600,   84.1301829792,   85.0546646813,
    86.1059049195,   87.2152782308,   88.1693392785,   88.9898317795,   -94.0300000000,
    -93.9257000000,  -93.6386720000,  -93.1242516800,  -92.2367115456,  -91.1630403910,
    81.5000000000,   81.8575000000,   82.7144000000,   84.0740360000,   85.9581131200,
    87.6914640704,   89.2861469448,   90.7532551892,   92.1029947741,   93.6866952966,
    95.1120257670,   96.3948231903,   97.8774444074,   99.7062021904,   101.2789338837,
    103.3495044623,  106.8865936591,  109.7870068005,  112.1653455764,  114.1155833726,
    115.7147783656,  -123.0000000000, -122.8319000000, -122.3536240000, -121.6424065600,
    -120.9738621664, -120.3454304364, -119.7547046102, -119.1994223336, -118.4186685469,
    -117.7003750632, 109.4400000000,  109.6850000000,  109.9251000000,  110.1603980000,
    110.3909900400,  110.6169702392,  110.8384308344,  111.0554622177,  111.2681529734,
    111.6974268544,  112.1095297803,  112.9279579934,  114.1137213540,  115.2046236457,
    116.2082537540,  117.1315934537,  117.9810659774,  118.7625806992,  120.0863226293,
    121.2776903663,  122.5943675224,  124.1743560693,  125.5331462196,  127.1278428244,
    128.9840311160,  130.9252248928,  132.5781799143,  133.9005439314,  134.5000000000,
    -139.1900000000, -139.0800000000, -138.8800000000, -138.3672000000, -137.4751680000,
    -136.2867545600, -135.1934141952, -134.1875410596, -133.2621377748, -132.4107667528,
    -131.6275054126, -130.6457548713, -129.1510642868, -127.5983152866, 117.5600000000,
    117.5600000000,  117.7488000000,  117.9338240000,  118.1151475200,  118.2928445696,
    118.4669876782,  118.8431081711,  119.4261216808,  120.2424319463,  120.9934373906,
    121.0600000000,  -129.6300000000, -129.4574000000, -129.1139040000, -128.7841478400,
    -128.4675819264, -128.1636786493, 120.8700000000,  121.0226000000,  121.1721480000,
    121.5152620800,  122.1857463552,  123.1260866468,  124.4634779821,  126.0506606243,
    127.4473813493,  128.6764955874,  -137.6900000000, -137.5274000000, -136.9763040000,
    -136.4472518400, -135.9393617664, -135.4517872957, -134.6110800580, -133.4773936534,
    -132.4344021611, -131.4748499882, -130.3273649894, -129.0424811907, -127.1465338240,
    -125.5160190886, -124.1137764162, -123.2500000000, -122.7500000000, -120.6458000000,
    -118.9203560000, -117.1362848000, -114.8700000000, -112.8340000000, -111.0412000000,
    -109.6069600000, 103.8700000000,  -113.0600000000, -113.0600000000, -112.5788000000,
    -112.1072240000, -111.6450795200, -111.1921779296, -110.7483343710, -110.3133676836,
    -109.8871003299, -109.4693583233, -109.0599711569, -108.6587717337, -108.2655962990,
    -107.8802843731, -107.5026786856, -107.1326251119, -106.7699726096, -106.4145731575,
    -106.0662816943, -105.7249560604, -105.3904569392, 89.0000000000,   89.3224000000,
    90.0871040000,   91.2740777600,   92.3898330944,   93.4386431087,   94.4245245222,
    95.3512530509,   96.2223778678,   97.0412351958,   98.3075363801,   100.1267827421,
    102.4143688130,  105.0163571792,  -122.1200000000, -122.1200000000, -121.4976000000,
    -120.9000960000, -119.9160902400, -118.9911248256, -118.1216573361, -117.3043578959,
    -116.5360964221, -115.8139306368, -115.1350947986, -114.4969891107, -113.8971697641,
    -113.3333395782, -112.8033392035,
};

test "ParabolicStopAndReverse 252-bar update" {
    const tol = 1e-6;
    var sar = try ParabolicStopAndReverse.init(.{});

    for (0..test_highs.len) |i| {
        const result = sar.updateHL(test_highs[i], test_lows[i]);
        const exp = test_expected[i];

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

    _ = sar.updateHL(test_highs[0], test_lows[0]);
    const result = sar.updateHL(test_highs[1], test_lows[1]);
    // Forced long: result should be positive.
    try testing.expect(result > 0);
}

test "ParabolicStopAndReverse forced start short" {
    var sar = try ParabolicStopAndReverse.init(.{ .start_value = -100.0 });

    _ = sar.updateHL(test_highs[0], test_lows[0]);
    const result = sar.updateHL(test_highs[1], test_lows[1]);
    // Forced short: result should be negative.
    try testing.expect(result < 0);
}
