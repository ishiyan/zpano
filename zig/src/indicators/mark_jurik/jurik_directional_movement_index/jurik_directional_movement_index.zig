const std = @import("std");
const math = std.math;

const entities = @import("entities");
const Bar = entities.Bar;
const Quote = entities.Quote;
const Trade = entities.Trade;
const Scalar = entities.Scalar;
const indicator_mod = @import("../../core/indicator.zig");
const build_metadata_mod = @import("../../core/build_metadata.zig");
const metadata_mod = @import("../../core/metadata.zig");
const line_indicator_mod = @import("../../core/line_indicator.zig");

const jma_mod = @import("../jurik_moving_average/jurik_moving_average.zig");
const JurikMovingAverage = jma_mod.JurikMovingAverage;

const OutputArray = indicator_mod.OutputArray;
const LineIndicator = line_indicator_mod.LineIndicator;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the Jurik Directional Movement Index indicator.
pub const JurikDirectionalMovementIndexOutput = enum(u8) {
    /// Bipolar: 100*(Plus-Minus)/(Plus+Minus).
    bipolar = 1,
    /// Plus: JMA(upward) / JMA(TrueRange).
    plus = 2,
    /// Minus: JMA(downward) / JMA(TrueRange).
    minus = 3,
};

/// Parameters for the Jurik Directional Movement Index.
pub const JurikDirectionalMovementIndexParams = struct {
    /// Smoothing length for the internal JMA instances. Must be >= 1.
    length: u32 = 14,
};

/// Jurik Directional Movement Index (DMX), see http://jurikres.com/.
///
/// Produces three output lines: Bipolar, Plus, Minus.
/// Internal JMA instances use phase=-100.
pub const JurikDirectionalMovementIndex = struct {
    primed: bool,
    bar: u32,
    prev_high: f64,
    prev_low: f64,
    prev_close: f64,
    jma_plus: JurikMovingAverage,
    jma_minus: JurikMovingAverage,
    jma_denom: JurikMovingAverage,
    bipolar_val: f64,
    plus_val: f64,
    minus_val: f64,

    mnemonic_buf: [96]u8,
    mnemonic_len: usize,
    desc_buf: [128]u8,
    desc_len: usize,

    pub fn init(params: JurikDirectionalMovementIndexParams) !JurikDirectionalMovementIndex {
        const length = params.length;
        if (length < 1) return error.InvalidLength;

        const jma_params = jma_mod.JurikMovingAverageParams{
            .length = length,
            .phase = -100,
        };

        var mnemonic_buf: [96]u8 = undefined;
        const mnemonic = std.fmt.bufPrint(&mnemonic_buf, "dmx({d})", .{length}) catch unreachable;
        const mnemonic_len = mnemonic.len;

        var desc_buf: [128]u8 = undefined;
        const desc = std.fmt.bufPrint(&desc_buf, "Jurik directional movement index dmx({d})", .{length}) catch unreachable;
        const desc_len = desc.len;

        return .{
            .primed = false,
            .bar = 0,
            .prev_high = math.nan(f64),
            .prev_low = math.nan(f64),
            .prev_close = math.nan(f64),
            .jma_plus = JurikMovingAverage.init(jma_params) catch return error.InvalidLength,
            .jma_minus = JurikMovingAverage.init(jma_params) catch return error.InvalidLength,
            .jma_denom = JurikMovingAverage.init(jma_params) catch return error.InvalidLength,
            .bipolar_val = math.nan(f64),
            .plus_val = math.nan(f64),
            .minus_val = math.nan(f64),
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .desc_buf = desc_buf,
            .desc_len = desc_len,
        };
    }

    pub fn fixSlices(self: *JurikDirectionalMovementIndex) void {
        // JMA instances also have fixSlices - call them.
        self.jma_plus.fixSlices();
        self.jma_minus.fixSlices();
        self.jma_denom.fixSlices();
    }

    /// Update with high, low, close values. Returns [3]f64 = { bipolar, plus, minus }.
    pub fn updateHLC(self: *JurikDirectionalMovementIndex, high: f64, low: f64, close: f64) [3]f64 {
        const warmup: u32 = 41;
        const epsilon: f64 = 0.00001;
        const hundred: f64 = 100.0;

        self.bar += 1;

        var true_range: f64 = 0;
        var upward: f64 = 0;
        var downward: f64 = 0;

        if (self.bar >= 2) {
            const v1 = hundred * (high - self.prev_high);
            const v2 = hundred * (self.prev_low - low);

            if (v1 > v2 and v1 > 0) {
                upward = v1;
            }

            if (v2 > v1 and v2 > 0) {
                downward = v2;
            }
        }

        if (self.bar >= 3) {
            const m1 = @abs(high - low);
            const m2 = @abs(high - self.prev_close);
            const m3 = @abs(low - self.prev_close);
            true_range = @max(@max(m1, m2), m3);
        }

        self.prev_high = high;
        self.prev_low = low;
        self.prev_close = close;

        const numer_plus = self.jma_plus.update(upward);
        const numer_minus = self.jma_minus.update(downward);
        const denom = self.jma_denom.update(true_range);

        if (self.bar <= warmup) {
            self.bipolar_val = math.nan(f64);
            self.plus_val = math.nan(f64);
            self.minus_val = math.nan(f64);
            return .{ math.nan(f64), math.nan(f64), math.nan(f64) };
        }

        self.primed = true;

        if (denom > epsilon) {
            self.plus_val = numer_plus / denom;
        } else {
            self.plus_val = 0;
        }

        if (denom > epsilon) {
            self.minus_val = numer_minus / denom;
        } else {
            self.minus_val = 0;
        }

        const sum = self.plus_val + self.minus_val;
        if (sum > epsilon) {
            self.bipolar_val = hundred * (self.plus_val - self.minus_val) / sum;
        } else {
            self.bipolar_val = 0;
        }

        return .{ self.bipolar_val, self.plus_val, self.minus_val };
    }

    pub fn isPrimed(self: *const JurikDirectionalMovementIndex) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const JurikDirectionalMovementIndex, out: *Metadata) void {
        const mn = self.mnemonic_buf[0..self.mnemonic_len];
        const desc = self.desc_buf[0..self.desc_len];

        // Build mnemonic/description strings for each output.
        var bp_mn_buf: [128]u8 = undefined;
        const bp_mn = std.fmt.bufPrint(&bp_mn_buf, "{s}:bipolar", .{mn}) catch unreachable;
        var bp_desc_buf: [128]u8 = undefined;
        const bp_desc = std.fmt.bufPrint(&bp_desc_buf, "{s} bipolar", .{desc}) catch unreachable;

        var pl_mn_buf: [128]u8 = undefined;
        const pl_mn = std.fmt.bufPrint(&pl_mn_buf, "{s}:plus", .{mn}) catch unreachable;
        var pl_desc_buf: [128]u8 = undefined;
        const pl_desc = std.fmt.bufPrint(&pl_desc_buf, "{s} plus", .{desc}) catch unreachable;

        var mi_mn_buf: [128]u8 = undefined;
        const mi_mn = std.fmt.bufPrint(&mi_mn_buf, "{s}:minus", .{mn}) catch unreachable;
        var mi_desc_buf: [128]u8 = undefined;
        const mi_desc = std.fmt.bufPrint(&mi_desc_buf, "{s} minus", .{desc}) catch unreachable;

        build_metadata_mod.buildMetadata(
            out,
            .jurik_directional_movement_index,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = bp_mn, .description = bp_desc },
                .{ .mnemonic = pl_mn, .description = pl_desc },
                .{ .mnemonic = mi_mn, .description = mi_desc },
            },
        );
    }

    pub fn updateScalar(self: *JurikDirectionalMovementIndex, sample: *const Scalar) OutputArray {
        const v = sample.value;
        return self.updateEntity(sample.time, v, v, v);
    }

    pub fn updateBar(self: *JurikDirectionalMovementIndex, sample: *const Bar) OutputArray {
        return self.updateEntity(sample.time, sample.high, sample.low, sample.close);
    }

    pub fn updateQuote(self: *JurikDirectionalMovementIndex, sample: *const Quote) OutputArray {
        return self.updateEntity(sample.time, sample.ask_price, sample.bid_price, (sample.ask_price + sample.bid_price) / 2.0);
    }

    pub fn updateTrade(self: *JurikDirectionalMovementIndex, sample: *const Trade) OutputArray {
        const v = sample.price;
        return self.updateEntity(sample.time, v, v, v);
    }

    fn updateEntity(self: *JurikDirectionalMovementIndex, time: i64, high: f64, low: f64, close: f64) OutputArray {
        const result = self.updateHLC(high, low, close);
        var output = OutputArray{};
        output.append(.{ .scalar = .{ .time = time, .value = result[0] } });
        output.append(.{ .scalar = .{ .time = time, .value = result[1] } });
        output.append(.{ .scalar = .{ .time = time, .value = result[2] } });
        return output;
    }

    pub fn indicator(self: *JurikDirectionalMovementIndex) indicator_mod.Indicator {
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
        const self: *JurikDirectionalMovementIndex = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const JurikDirectionalMovementIndex = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *JurikDirectionalMovementIndex = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *JurikDirectionalMovementIndex = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *JurikDirectionalMovementIndex = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *JurikDirectionalMovementIndex = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const Error = error{
        InvalidLength,
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

fn runDmxTest(length: u32, expected_bipolar: [252]f64, expected_plus: [252]f64, comptime has_minus: bool, expected_minus: if (has_minus) [252]f64 else [1]f64) !void {
    var dmx = JurikDirectionalMovementIndex.init(.{ .length = length }) catch unreachable;
    dmx.fixSlices();
    const close_data = testdata.testInputClose();
    const high_data = testdata.testInputHigh();
    const low_data = testdata.testInputLow();
    const eps = 1e-10;

    for (0..252) |i| {
        const result = dmx.updateHLC(high_data[i], low_data[i], close_data[i]);
        const bipolar = result[0];
        const plus = result[1];
        const minus = result[2];

        // First 41 bars (indices 0-40) are warmup.
        if (i <= 40) {
            try testing.expect(math.isNan(bipolar));
            continue;
        }

        if (!almostEqual(bipolar, expected_bipolar[i], eps)) {
            std.debug.print("FAIL [{d}] len={d}: bipolar expected {d}, got {d}, diff {d}\n", .{ i, length, expected_bipolar[i], bipolar, @abs(bipolar - expected_bipolar[i]) });
            return error.TestUnexpectedResult;
        }

        if (!almostEqual(plus, expected_plus[i], eps)) {
            std.debug.print("FAIL [{d}] len={d}: plus expected {d}, got {d}, diff {d}\n", .{ i, length, expected_plus[i], plus, @abs(plus - expected_plus[i]) });
            return error.TestUnexpectedResult;
        }

        if (has_minus) {
            if (!almostEqual(minus, expected_minus[i], eps)) {
                std.debug.print("FAIL [{d}] len={d}: minus expected {d}, got {d}, diff {d}\n", .{ i, length, expected_minus[i], minus, @abs(minus - expected_minus[i]) });
                return error.TestUnexpectedResult;
            }
        }
    }
}

test "dmx length 2" {
    try runDmxTest(2, testdata.dmxBipolarLen2(), testdata.dmxPlusLen2(), true, testdata.dmxMinusLen2());
}
test "dmx length 3" {
    try runDmxTest(3, testdata.dmxBipolarLen3(), testdata.dmxPlusLen3(), true, testdata.dmxMinusLen3());
}
test "dmx length 4" {
    try runDmxTest(4, testdata.dmxBipolarLen4(), testdata.dmxPlusLen4(), true, testdata.dmxMinusLen4());
}
test "dmx length 5" {
    try runDmxTest(5, testdata.dmxBipolarLen5(), testdata.dmxPlusLen5(), true, testdata.dmxMinusLen5());
}
test "dmx length 6" {
    try runDmxTest(6, testdata.dmxBipolarLen6(), testdata.dmxPlusLen6(), true, testdata.dmxMinusLen6());
}
test "dmx length 7" {
    try runDmxTest(7, testdata.dmxBipolarLen7(), testdata.dmxPlusLen7(), true, testdata.dmxMinusLen7());
}
test "dmx length 8" {
    try runDmxTest(8, testdata.dmxBipolarLen8(), testdata.dmxPlusLen8(), true, testdata.dmxMinusLen8());
}
test "dmx length 9" {
    try runDmxTest(9, testdata.dmxBipolarLen9(), testdata.dmxPlusLen9(), true, testdata.dmxMinusLen9());
}
test "dmx length 10" {
    try runDmxTest(10, testdata.dmxBipolarLen10(), testdata.dmxPlusLen10(), true, testdata.dmxMinusLen10());
}
test "dmx length 11" {
    try runDmxTest(11, testdata.dmxBipolarLen11(), testdata.dmxPlusLen11(), true, testdata.dmxMinusLen11());
}
test "dmx length 12" {
    try runDmxTest(12, testdata.dmxBipolarLen12(), testdata.dmxPlusLen12(), true, testdata.dmxMinusLen12());
}
test "dmx length 13" {
    try runDmxTest(13, testdata.dmxBipolarLen13(), testdata.dmxPlusLen13(), true, testdata.dmxMinusLen13());
}
test "dmx length 14" {
    // dmxMinusLen14 is [1]f64 (intentionally empty) — skip minus check.
    try runDmxTest(14, testdata.dmxBipolarLen14(), testdata.dmxPlusLen14(), false, testdata.dmxMinusLen14());
}
test "dmx length 15" {
    try runDmxTest(15, testdata.dmxBipolarLen15(), testdata.dmxPlusLen15(), true, testdata.dmxMinusLen15());
}
test "dmx length 16" {
    try runDmxTest(16, testdata.dmxBipolarLen16(), testdata.dmxPlusLen16(), true, testdata.dmxMinusLen16());
}
test "dmx length 17" {
    try runDmxTest(17, testdata.dmxBipolarLen17(), testdata.dmxPlusLen17(), true, testdata.dmxMinusLen17());
}
test "dmx length 18" {
    try runDmxTest(18, testdata.dmxBipolarLen18(), testdata.dmxPlusLen18(), true, testdata.dmxMinusLen18());
}
test "dmx length 19" {
    try runDmxTest(19, testdata.dmxBipolarLen19(), testdata.dmxPlusLen19(), true, testdata.dmxMinusLen19());
}
test "dmx length 20" {
    try runDmxTest(20, testdata.dmxBipolarLen20(), testdata.dmxPlusLen20(), true, testdata.dmxMinusLen20());
}
