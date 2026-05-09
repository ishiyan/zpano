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
const dmp_mod = @import("../directional_movement_plus/directional_movement_plus.zig");
const DirectionalMovementPlus = dmp_mod.DirectionalMovementPlus;

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

const epsilon = 1e-8;

/// Enumerates the outputs of the Directional Indicator Plus indicator.
pub const DirectionalIndicatorPlusOutput = enum(u8) {
    /// The scalar value of the directional indicator plus (+DI).
    value = 1,
    /// The scalar value of the directional movement plus (+DM).
    directional_movement_plus = 2,
    /// The scalar value of the average true range (ATR).
    average_true_range = 3,
    /// The scalar value of the true range (TR).
    true_range = 4,
};

/// Welles Wilder's Directional Indicator Plus (+DI).
///
/// The directional indicator plus measures the percentage of the average true range
/// that is attributable to upward movement. It is calculated as:
///   +DI = 100 * +DM(n) / (ATR * length)
///
/// where +DM(n) is the Wilder-smoothed directional movement plus and ATR is the
/// average true range over the same length.
pub const DirectionalIndicatorPlus = struct {
    length: i32,
    value: f64,
    average_true_range: AverageTrueRange,
    directional_movement_plus: DirectionalMovementPlus,

    const mnemonic_str = "+di";
    const description_str = "Directional Indicator Plus";

    pub const Error = error{
        InvalidLength,
        OutOfMemory,
    };

    pub fn init(allocator: std.mem.Allocator, params: struct { length: i32 = 14 }) Error!DirectionalIndicatorPlus {
        if (params.length < 1) return Error.InvalidLength;

        const atr = atr_mod.AverageTrueRange.init(allocator, .{ .length = params.length }) catch |e| switch (e) {
            error.InvalidLength => return Error.InvalidLength,
            error.OutOfMemory => return Error.OutOfMemory,
        };

        const dmp = DirectionalMovementPlus.init(.{ .length = @intCast(params.length) }) catch return Error.InvalidLength;

        return .{
            .length = params.length,
            .value = math.nan(f64),
            .average_true_range = atr,
            .directional_movement_plus = dmp,
        };
    }

    pub fn deinit(self: *DirectionalIndicatorPlus) void {
        self.average_true_range.deinit();
    }

    pub fn fixSlices(_: *DirectionalIndicatorPlus) void {}

    /// Update given close, high, low values.
    pub fn update(self: *DirectionalIndicatorPlus, close: f64, high: f64, low: f64) f64 {
        if (math.isNan(close) or math.isNan(high) or math.isNan(low)) return math.nan(f64);

        const atr_value = self.average_true_range.update(close, high, low);
        const dmp_value = self.directional_movement_plus.update(high, low);

        if (self.average_true_range.isPrimed() and self.directional_movement_plus.isPrimed()) {
            const atr_scaled = atr_value * @as(f64, @floatFromInt(self.length));

            if (@abs(atr_scaled) < epsilon) {
                self.value = 0;
            } else {
                self.value = 100.0 * dmp_value / atr_scaled;
            }

            return self.value;
        }

        return math.nan(f64);
    }

    /// Update using a single sample value as substitute for high, low, close.
    pub fn updateSample(self: *DirectionalIndicatorPlus, sample: f64) f64 {
        return self.update(sample, sample, sample);
    }

    pub fn isPrimed(self: *const DirectionalIndicatorPlus) bool {
        return self.average_true_range.isPrimed() and self.directional_movement_plus.isPrimed();
    }

    pub fn getMetadata(_: *const DirectionalIndicatorPlus, out: *Metadata) void {
        build_metadata_mod.buildMetadata(out, Identifier.directional_indicator_plus, mnemonic_str, description_str, &.{
            .{ .mnemonic = mnemonic_str, .description = description_str },
            .{ .mnemonic = "+dm", .description = "Directional Movement Plus" },
            .{ .mnemonic = "atr", .description = "Average True Range" },
            .{ .mnemonic = "tr", .description = "True Range" },
        });
    }

    fn makeOutput(self: *const DirectionalIndicatorPlus, time: i64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = Scalar{ .time = time, .value = self.value } });
        return out;
    }

    pub fn updateScalar(self: *DirectionalIndicatorPlus, sample: *const Scalar) OutputArray {
        _ = self.update(sample.value, sample.value, sample.value);
        return self.makeOutput(sample.time);
    }

    pub fn updateBar(self: *DirectionalIndicatorPlus, sample: *const Bar) OutputArray {
        _ = self.update(sample.close, sample.high, sample.low);
        return self.makeOutput(sample.time);
    }

    pub fn updateQuote(self: *DirectionalIndicatorPlus, sample: *const Quote) OutputArray {
        const mid = (sample.bid_price + sample.ask_price) / 2.0;
        _ = self.update(mid, mid, mid);
        return self.makeOutput(sample.time);
    }

    pub fn updateTrade(self: *DirectionalIndicatorPlus, sample: *const Trade) OutputArray {
        _ = self.update(sample.price, sample.price, sample.price);
        return self.makeOutput(sample.time);
    }

    // --- Indicator interface ---

    pub fn indicator(self: *DirectionalIndicatorPlus) indicator_mod.Indicator {
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
        const self: *const DirectionalIndicatorPlus = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *const anyopaque, out: *Metadata) void {
        const self: *const DirectionalIndicatorPlus = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *DirectionalIndicatorPlus = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *DirectionalIndicatorPlus = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *DirectionalIndicatorPlus = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *DirectionalIndicatorPlus = @ptrCast(@alignCast(ptr));
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
// Expected +DI14 (length=14), 252 entries.
test "DirectionalIndicatorPlus update length=14" {
    const tolerance = 1e-8;
    var dip = try DirectionalIndicatorPlus.init(testing.allocator, .{ .length = 14 });
    defer dip.deinit();

    for (0..testdata.test_input_close.len) |i| {
        const act = dip.update(testdata.test_input_close[i], testdata.test_input_high[i], testdata.test_input_low[i]);
        const exp = testdata.test_expected_di14[i];

        if (math.isNan(exp)) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(!math.isNan(act));
            try testing.expect(almostEqual(act, exp, tolerance));
        }
    }
}

test "DirectionalIndicatorPlus isPrimed length=14" {
    var dip = try DirectionalIndicatorPlus.init(testing.allocator, .{ .length = 14 });
    defer dip.deinit();

    for (0..14) |i| {
        _ = dip.update(testdata.test_input_close[i], testdata.test_input_high[i], testdata.test_input_low[i]);
        try testing.expect(!dip.isPrimed());
    }

    _ = dip.update(testdata.test_input_close[14], testdata.test_input_high[14], testdata.test_input_low[14]);
    try testing.expect(dip.isPrimed());
}

test "DirectionalIndicatorPlus constructor validation" {
    try testing.expectError(error.InvalidLength, DirectionalIndicatorPlus.init(testing.allocator, .{ .length = 0 }));
    try testing.expectError(error.InvalidLength, DirectionalIndicatorPlus.init(testing.allocator, .{ .length = -8 }));

    var dip = try DirectionalIndicatorPlus.init(testing.allocator, .{ .length = 14 });
    defer dip.deinit();
    try testing.expect(!dip.isPrimed());
}

test "DirectionalIndicatorPlus NaN passthrough" {
    var dip = try DirectionalIndicatorPlus.init(testing.allocator, .{ .length = 14 });
    defer dip.deinit();

    try testing.expect(math.isNan(dip.update(math.nan(f64), 1, 1)));
    try testing.expect(math.isNan(dip.update(1, math.nan(f64), 1)));
    try testing.expect(math.isNan(dip.update(1, 1, math.nan(f64))));
    try testing.expect(math.isNan(dip.updateSample(math.nan(f64))));
}

test "DirectionalIndicatorPlus metadata" {
    var dip = try DirectionalIndicatorPlus.init(testing.allocator, .{ .length = 14 });
    defer dip.deinit();
    var meta: Metadata = undefined;
    dip.getMetadata(&meta);

    try testing.expectEqual(Identifier.directional_indicator_plus, meta.identifier);
    try testing.expectEqualStrings("+di", meta.mnemonic);
    try testing.expectEqual(@as(usize, 4), meta.outputs_len);
}

test "DirectionalIndicatorPlus updateBar" {
    var dip = try DirectionalIndicatorPlus.init(testing.allocator, .{ .length = 14 });
    defer dip.deinit();

    for (0..14) |i| {
        _ = dip.update(testdata.test_input_close[i], testdata.test_input_high[i], testdata.test_input_low[i]);
    }

    const bar = Bar{
        .time = 1000,
        .open = 91,
        .high = testdata.test_input_high[14],
        .low = testdata.test_input_low[14],
        .close = testdata.test_input_close[14],
        .volume = 1000,
    };
    const out = dip.updateBar(&bar);
    try testing.expect(!math.isNan(out.slice()[0].scalar.value));
}
