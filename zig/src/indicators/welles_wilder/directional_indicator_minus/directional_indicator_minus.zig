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

const atr_mod = @import("../average_true_range/average_true_range.zig");
const AverageTrueRange = atr_mod.AverageTrueRange;
const dmm_mod = @import("../directional_movement_minus/directional_movement_minus.zig");
const DirectionalMovementMinus = dmm_mod.DirectionalMovementMinus;

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

const epsilon = 1e-8;

/// Enumerates the outputs of the Directional Indicator Minus indicator.
pub const DirectionalIndicatorMinusOutput = enum(u8) {
    /// The scalar value of the directional indicator minus (-DI).
    value = 1,
    /// The scalar value of the directional movement minus (-DM).
    directional_movement_minus = 2,
    /// The scalar value of the average true range (ATR).
    average_true_range = 3,
    /// The scalar value of the true range (TR).
    true_range = 4,
};

/// Welles Wilder's Directional Indicator Minus (-DI).
///
/// The directional indicator minus measures the percentage of the average true range
/// that is attributable to downward movement. It is calculated as:
///   -DI = 100 * -DM(n) / (ATR * length)
///
/// where -DM(n) is the Wilder-smoothed directional movement minus and ATR is the
/// average true range over the same length.
pub const DirectionalIndicatorMinus = struct {
    length: i32,
    value: f64,
    average_true_range: AverageTrueRange,
    directional_movement_minus: DirectionalMovementMinus,

    const mnemonic_str = "-di";
    const description_str = "Directional Indicator Minus";

    pub const Error = error{
        InvalidLength,
        OutOfMemory,
    };

    pub fn init(allocator: std.mem.Allocator, params: struct { length: i32 = 14 }) Error!DirectionalIndicatorMinus {
        if (params.length < 1) return Error.InvalidLength;

        const atr = atr_mod.AverageTrueRange.init(allocator, .{ .length = params.length }) catch |e| switch (e) {
            error.InvalidLength => return Error.InvalidLength,
            error.OutOfMemory => return Error.OutOfMemory,
        };

        const dmm = DirectionalMovementMinus.init(.{ .length = @intCast(params.length) }) catch return Error.InvalidLength;

        return .{
            .length = params.length,
            .value = math.nan(f64),
            .average_true_range = atr,
            .directional_movement_minus = dmm,
        };
    }

    pub fn deinit(self: *DirectionalIndicatorMinus) void {
        self.average_true_range.deinit();
    }

    pub fn fixSlices(_: *DirectionalIndicatorMinus) void {}

    /// Update given close, high, low values.
    pub fn update(self: *DirectionalIndicatorMinus, close: f64, high: f64, low: f64) f64 {
        if (math.isNan(close) or math.isNan(high) or math.isNan(low)) return math.nan(f64);

        const atr_value = self.average_true_range.update(close, high, low);
        const dmm_value = self.directional_movement_minus.update(high, low);

        if (self.average_true_range.isPrimed() and self.directional_movement_minus.isPrimed()) {
            const atr_scaled = atr_value * @as(f64, @floatFromInt(self.length));

            if (@abs(atr_scaled) < epsilon) {
                self.value = 0;
            } else {
                self.value = 100.0 * dmm_value / atr_scaled;
            }

            return self.value;
        }

        return math.nan(f64);
    }

    /// Update using a single sample value as substitute for high, low, close.
    pub fn updateSample(self: *DirectionalIndicatorMinus, sample: f64) f64 {
        return self.update(sample, sample, sample);
    }

    pub fn isPrimed(self: *const DirectionalIndicatorMinus) bool {
        return self.average_true_range.isPrimed() and self.directional_movement_minus.isPrimed();
    }

    pub fn getMetadata(_: *const DirectionalIndicatorMinus, out: *Metadata) void {
        build_metadata_mod.buildMetadata(out, Identifier.directional_indicator_minus, mnemonic_str, description_str, &.{
            .{ .mnemonic = mnemonic_str, .description = description_str },
            .{ .mnemonic = "-dm", .description = "Directional Movement Minus" },
            .{ .mnemonic = "atr", .description = "Average True Range" },
            .{ .mnemonic = "tr", .description = "True Range" },
        });
    }

    fn makeOutput(self: *const DirectionalIndicatorMinus, time: i64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = Scalar{ .time = time, .value = self.value } });
        return out;
    }

    pub fn updateScalar(self: *DirectionalIndicatorMinus, sample: *const Scalar) OutputArray {
        _ = self.update(sample.value, sample.value, sample.value);
        return self.makeOutput(sample.time);
    }

    pub fn updateBar(self: *DirectionalIndicatorMinus, sample: *const Bar) OutputArray {
        _ = self.update(sample.close, sample.high, sample.low);
        return self.makeOutput(sample.time);
    }

    pub fn updateQuote(self: *DirectionalIndicatorMinus, sample: *const Quote) OutputArray {
        const mid = (sample.bid_price + sample.ask_price) / 2.0;
        _ = self.update(mid, mid, mid);
        return self.makeOutput(sample.time);
    }

    pub fn updateTrade(self: *DirectionalIndicatorMinus, sample: *const Trade) OutputArray {
        _ = self.update(sample.price, sample.price, sample.price);
        return self.makeOutput(sample.time);
    }

    // --- Indicator interface ---

    pub fn indicator(self: *DirectionalIndicatorMinus) indicator_mod.Indicator {
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
        const self: *const DirectionalIndicatorMinus = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *const anyopaque, out: *Metadata) void {
        const self: *const DirectionalIndicatorMinus = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *DirectionalIndicatorMinus = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *DirectionalIndicatorMinus = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *DirectionalIndicatorMinus = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *DirectionalIndicatorMinus = @ptrCast(@alignCast(ptr));
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

// TA-Lib test data (252 entries).
// Expected -DI14 (length=14), 252 entries.
test "DirectionalIndicatorMinus update length=14" {
    const tolerance = 1e-8;
    var dim = try DirectionalIndicatorMinus.init(testing.allocator, .{ .length = 14 });
    defer dim.deinit();

    for (0..testdata.test_input_close.len) |i| {
        const act = dim.update(testdata.test_input_close[i], testdata.test_input_high[i], testdata.test_input_low[i]);
        const exp = testdata.test_expected_di14[i];

        if (math.isNan(exp)) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(!math.isNan(act));
            try testing.expect(almostEqual(act, exp, tolerance));
        }
    }
}

test "DirectionalIndicatorMinus isPrimed length=14" {
    var dim = try DirectionalIndicatorMinus.init(testing.allocator, .{ .length = 14 });
    defer dim.deinit();

    for (0..14) |i| {
        _ = dim.update(testdata.test_input_close[i], testdata.test_input_high[i], testdata.test_input_low[i]);
        try testing.expect(!dim.isPrimed());
    }

    _ = dim.update(testdata.test_input_close[14], testdata.test_input_high[14], testdata.test_input_low[14]);
    try testing.expect(dim.isPrimed());
}

test "DirectionalIndicatorMinus constructor validation" {
    try testing.expectError(error.InvalidLength, DirectionalIndicatorMinus.init(testing.allocator, .{ .length = 0 }));
    try testing.expectError(error.InvalidLength, DirectionalIndicatorMinus.init(testing.allocator, .{ .length = -8 }));

    var dim = try DirectionalIndicatorMinus.init(testing.allocator, .{ .length = 14 });
    defer dim.deinit();
    try testing.expect(!dim.isPrimed());
}

test "DirectionalIndicatorMinus NaN passthrough" {
    var dim = try DirectionalIndicatorMinus.init(testing.allocator, .{ .length = 14 });
    defer dim.deinit();

    try testing.expect(math.isNan(dim.update(math.nan(f64), 1, 1)));
    try testing.expect(math.isNan(dim.update(1, math.nan(f64), 1)));
    try testing.expect(math.isNan(dim.update(1, 1, math.nan(f64))));
    try testing.expect(math.isNan(dim.updateSample(math.nan(f64))));
}

test "DirectionalIndicatorMinus metadata" {
    var dim = try DirectionalIndicatorMinus.init(testing.allocator, .{ .length = 14 });
    defer dim.deinit();
    var meta: Metadata = undefined;
    dim.getMetadata(&meta);

    try testing.expectEqual(Identifier.directional_indicator_minus, meta.identifier);
    try testing.expectEqualStrings("-di", meta.mnemonic);
    try testing.expectEqual(@as(usize, 4), meta.outputs_len);
}

test "DirectionalIndicatorMinus updateBar" {
    var dim = try DirectionalIndicatorMinus.init(testing.allocator, .{ .length = 14 });
    defer dim.deinit();

    for (0..14) |i| {
        _ = dim.update(testdata.test_input_close[i], testdata.test_input_high[i], testdata.test_input_low[i]);
    }

    const bar = Bar{
        .time = 1000,
        .open = 91,
        .high = testdata.test_input_high[14],
        .low = testdata.test_input_low[14],
        .close = testdata.test_input_close[14],
        .volume = 1000,
    };
    const out = dim.updateBar(&bar);
    try testing.expect(!math.isNan(out.slice()[0].scalar.value));
}
