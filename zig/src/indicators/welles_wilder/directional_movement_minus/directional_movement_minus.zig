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

/// Enumerates the outputs of the Directional Movement Minus indicator.
pub const DirectionalMovementMinusOutput = enum(u8) {
    /// The scalar value of -DM.
    value = 1,
};

/// Parameters for the Directional Movement Minus indicator.
pub const DirectionalMovementMinusParams = struct {
    /// The smoothing length. Must be >= 1. Default is 14.
    length: usize = 14,
};

/// Welles Wilder's Directional Movement Minus (-DM) indicator.
///
/// UpMove = today's high − yesterday's high
/// DownMove = yesterday's low − today's low
/// if DownMove > UpMove and DownMove > 0, then -DM = DownMove, else -DM = 0
///
/// When length > 1, Wilder's smoothing is applied:
///   -DM(n) = previous -DM(n) − previous -DM(n)/n + today's -DM(1)
pub const DirectionalMovementMinus = struct {
    length: usize,
    no_smoothing: bool,
    count: usize,
    previous_high: f64,
    previous_low: f64,
    value: f64,
    accumulator: f64,
    primed: bool,

    pub const Error = error{InvalidLength};

    pub fn init(params: DirectionalMovementMinusParams) Error!DirectionalMovementMinus {
        if (params.length < 1) return error.InvalidLength;

        return .{
            .length = params.length,
            .no_smoothing = params.length == 1,
            .count = 0,
            .previous_high = 0,
            .previous_low = 0,
            .value = math.nan(f64),
            .accumulator = 0,
            .primed = false,
        };
    }

    pub fn deinit(_: *DirectionalMovementMinus) void {}
    pub fn fixSlices(_: *DirectionalMovementMinus) void {}

    /// Core update given high and low values.
    pub fn update(self: *DirectionalMovementMinus, high_in: f64, low_in: f64) f64 {
        if (math.isNan(high_in) or math.isNan(low_in)) {
            return math.nan(f64);
        }

        var high = high_in;
        var low = low_in;
        if (high < low) {
            const tmp = high;
            high = low;
            low = tmp;
        }

        if (self.no_smoothing) {
            if (self.primed) {
                const delta_plus = high - self.previous_high;
                const delta_minus = self.previous_low - low;
                if (delta_minus > 0 and delta_minus > delta_plus) {
                    self.value = delta_minus;
                } else {
                    self.value = 0;
                }
            } else {
                if (self.count > 0) {
                    const delta_plus = high - self.previous_high;
                    const delta_minus = self.previous_low - low;
                    if (delta_minus > 0 and delta_minus > delta_plus) {
                        self.value = delta_minus;
                    } else {
                        self.value = 0;
                    }
                    self.primed = true;
                }
                self.count += 1;
            }
        } else {
            const n: f64 = @floatFromInt(self.length);
            if (self.primed) {
                const delta_plus = high - self.previous_high;
                const delta_minus = self.previous_low - low;
                if (delta_minus > 0 and delta_minus > delta_plus) {
                    self.accumulator += -self.accumulator / n + delta_minus;
                } else {
                    self.accumulator += -self.accumulator / n;
                }
                self.value = self.accumulator;
            } else {
                if (self.count > 0 and self.length >= self.count) {
                    const delta_plus = high - self.previous_high;
                    const delta_minus = self.previous_low - low;
                    if (self.length > self.count) {
                        if (delta_minus > 0 and delta_minus > delta_plus) {
                            self.accumulator += delta_minus;
                        }
                    } else {
                        if (delta_minus > 0 and delta_minus > delta_plus) {
                            self.accumulator += -self.accumulator / n + delta_minus;
                        } else {
                            self.accumulator += -self.accumulator / n;
                        }
                        self.value = self.accumulator;
                        self.primed = true;
                    }
                }
                self.count += 1;
            }
        }

        self.previous_low = low;
        self.previous_high = high;

        return self.value;
    }

    /// Update using a single sample value as substitute for high and low.
    pub fn updateSample(self: *DirectionalMovementMinus, sample: f64) f64 {
        return self.update(sample, sample);
    }

    pub fn isPrimed(self: *const DirectionalMovementMinus) bool {
        return self.primed;
    }

    pub fn getMetadata(_: *const DirectionalMovementMinus, out: *Metadata) void {
        const mnemonic = "-dm";
        const description = "Directional Movement Minus";
        build_metadata_mod.buildMetadata(out, Identifier.directional_movement_minus, mnemonic, description, &.{
            .{ .mnemonic = mnemonic, .description = description },
        });
    }

    fn makeOutput(self: *const DirectionalMovementMinus, time: i64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = Scalar{ .time = time, .value = self.value } });
        return out;
    }

    pub fn updateScalar(self: *DirectionalMovementMinus, sample: *const Scalar) OutputArray {
        _ = self.update(sample.value, sample.value);
        return self.makeOutput(sample.time);
    }

    pub fn updateBar(self: *DirectionalMovementMinus, sample: *const Bar) OutputArray {
        _ = self.update(sample.high, sample.low);
        return self.makeOutput(sample.time);
    }

    pub fn updateQuote(self: *DirectionalMovementMinus, sample: *const Quote) OutputArray {
        const mid = (sample.bid_price + sample.ask_price) / 2.0;
        _ = self.update(mid, mid);
        return self.makeOutput(sample.time);
    }

    pub fn updateTrade(self: *DirectionalMovementMinus, sample: *const Trade) OutputArray {
        _ = self.update(sample.price, sample.price);
        return self.makeOutput(sample.time);
    }

    // --- Indicator interface ---

    pub fn indicator(self: *DirectionalMovementMinus) indicator_mod.Indicator {
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
        const self: *const DirectionalMovementMinus = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *const anyopaque, out: *Metadata) void {
        const self: *const DirectionalMovementMinus = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *DirectionalMovementMinus = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *DirectionalMovementMinus = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *DirectionalMovementMinus = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *DirectionalMovementMinus = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }
};

// ============================================================================
// Tests
// ============================================================================

const testing = std.testing;
const testdata = @import("testdata.zig");


fn almostEqual(a: f64, b: f64, tolerance: f64) bool {
    if (math.isNan(a) and math.isNan(b)) return true;
    if (math.isNan(a) or math.isNan(b)) return false;
    return @abs(a - b) <= tolerance;
}

test "DirectionalMovementMinus update length=14" {
    const tolerance = 1e-8;
    var dmm = try DirectionalMovementMinus.init(.{ .length = 14 });

    for (0..testdata.test_high.len) |i| {
        const act = dmm.update(testdata.test_high[i], testdata.test_low[i]);
        const exp = testdata.test_expected_dmm14[i];

        if (math.isNan(exp)) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(!math.isNan(act));
            try testing.expect(almostEqual(act, exp, tolerance));
        }
    }
}

test "DirectionalMovementMinus update length=1" {
    const tolerance = 1e-8;
    var dmm = try DirectionalMovementMinus.init(.{ .length = 1 });

    for (0..testdata.test_high.len) |i| {
        const act = dmm.update(testdata.test_high[i], testdata.test_low[i]);
        const exp = testdata.test_expected_dmm1[i];

        if (math.isNan(exp)) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(!math.isNan(act));
            try testing.expect(almostEqual(act, exp, tolerance));
        }
    }
}

test "DirectionalMovementMinus constructor validation" {
    const result = DirectionalMovementMinus.init(.{ .length = 0 });
    try testing.expectError(error.InvalidLength, result);
}

test "DirectionalMovementMinus isPrimed length=1" {
    var dmm = try DirectionalMovementMinus.init(.{ .length = 1 });

    try testing.expect(!dmm.isPrimed());

    _ = dmm.update(testdata.test_high[0], testdata.test_low[0]);
    try testing.expect(!dmm.isPrimed());

    _ = dmm.update(testdata.test_high[1], testdata.test_low[1]);
    try testing.expect(dmm.isPrimed());
}

test "DirectionalMovementMinus isPrimed length=14" {
    var dmm = try DirectionalMovementMinus.init(.{ .length = 14 });

    for (0..14) |i| {
        _ = dmm.update(testdata.test_high[i], testdata.test_low[i]);
        try testing.expect(!dmm.isPrimed());
    }

    _ = dmm.update(testdata.test_high[14], testdata.test_low[14]);
    try testing.expect(dmm.isPrimed());
}

test "DirectionalMovementMinus NaN passthrough" {
    var dmm = try DirectionalMovementMinus.init(.{ .length = 14 });

    try testing.expect(math.isNan(dmm.update(math.nan(f64), 1)));
    try testing.expect(math.isNan(dmm.update(1, math.nan(f64))));
    try testing.expect(math.isNan(dmm.update(math.nan(f64), math.nan(f64))));
    try testing.expect(math.isNan(dmm.updateSample(math.nan(f64))));
}

test "DirectionalMovementMinus high/low swap" {
    var dmm1 = try DirectionalMovementMinus.init(.{ .length = 1 });
    var dmm2 = try DirectionalMovementMinus.init(.{ .length = 1 });

    _ = dmm1.update(10, 5);
    _ = dmm2.update(5, 10);

    const v1 = dmm1.update(12, 6);
    const v2 = dmm2.update(6, 12);

    try testing.expectEqual(v1, v2);
}

test "DirectionalMovementMinus metadata" {
    var dmm = try DirectionalMovementMinus.init(.{ .length = 14 });
    var meta: Metadata = undefined;
    dmm.getMetadata(&meta);

    try testing.expectEqual(Identifier.directional_movement_minus, meta.identifier);
    try testing.expectEqualStrings("-dm", meta.mnemonic);
    try testing.expectEqualStrings("Directional Movement Minus", meta.description);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
}

test "DirectionalMovementMinus updateBar" {
    var dmm = try DirectionalMovementMinus.init(.{ .length = 14 });

    for (0..14) |i| {
        _ = dmm.update(testdata.test_high[i], testdata.test_low[i]);
    }

    const bar = Bar{ .time = 42, .open = 0, .high = testdata.test_high[14], .low = testdata.test_low[14], .close = 0, .volume = 0 };
    const out = dmm.updateBar(&bar);
    try testing.expectEqual(@as(usize, 1), out.len);
}
