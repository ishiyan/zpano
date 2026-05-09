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

const dx_mod = @import("../directional_movement_index/directional_movement_index.zig");
const DirectionalMovementIndex = dx_mod.DirectionalMovementIndex;

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the Average Directional Movement Index indicator.
pub const AverageDirectionalMovementIndexOutput = enum(u8) {
    /// The scalar value of the average directional movement index (ADX).
    value = 1,
    /// The scalar value of the directional movement index (DX).
    directional_movement_index = 2,
    /// The scalar value of the directional indicator plus (+DI).
    directional_indicator_plus = 3,
    /// The scalar value of the directional indicator minus (-DI).
    directional_indicator_minus = 4,
    /// The scalar value of the directional movement plus (+DM).
    directional_movement_plus = 5,
    /// The scalar value of the directional movement minus (-DM).
    directional_movement_minus = 6,
    /// The scalar value of the average true range (ATR).
    average_true_range = 7,
    /// The scalar value of the true range (TR).
    true_range = 8,
};

/// Welles Wilder's Average Directional Movement Index (ADX).
///
/// The average directional movement index smooths the directional movement index (DX)
/// using Wilder's smoothing technique. It is calculated as:
///
///   Initial ADX = SMA of first `length` DX values
///   Subsequent ADX = (previousADX * (length-1) + DX) / length
///
/// The indicator requires close, high, and low values.
pub const AverageDirectionalMovementIndex = struct {
    length: i32,
    length_f: f64,
    length_minus_one: f64,
    count: i32,
    sum: f64,
    primed: bool,
    value: f64,
    directional_movement_index: DirectionalMovementIndex,

    const mnemonic_str = "adx";
    const description_str = "Average Directional Movement Index";

    pub const Error = error{
        InvalidLength,
        OutOfMemory,
    };

    pub fn init(allocator: std.mem.Allocator, params: struct { length: i32 = 14 }) Error!AverageDirectionalMovementIndex {
        if (params.length < 1) return Error.InvalidLength;

        const dx = DirectionalMovementIndex.init(allocator, .{ .length = params.length }) catch |e| switch (e) {
            error.InvalidLength => return Error.InvalidLength,
            error.OutOfMemory => return Error.OutOfMemory,
        };

        return .{
            .length = params.length,
            .length_f = @as(f64, @floatFromInt(params.length)),
            .length_minus_one = @as(f64, @floatFromInt(params.length - 1)),
            .count = 0,
            .sum = 0,
            .primed = false,
            .value = math.nan(f64),
            .directional_movement_index = dx,
        };
    }

    pub fn deinit(self: *AverageDirectionalMovementIndex) void {
        self.directional_movement_index.deinit();
    }

    pub fn fixSlices(_: *AverageDirectionalMovementIndex) void {}

    /// Update given close, high, low values.
    pub fn update(self: *AverageDirectionalMovementIndex, close: f64, high: f64, low: f64) f64 {
        if (math.isNan(close) or math.isNan(high) or math.isNan(low)) return math.nan(f64);

        const dx_value = self.directional_movement_index.update(close, high, low);

        if (!self.directional_movement_index.isPrimed()) {
            return math.nan(f64);
        }

        if (self.primed) {
            self.value = (self.value * self.length_minus_one + dx_value) / self.length_f;
            return self.value;
        }

        self.count += 1;
        self.sum += dx_value;

        if (self.count == self.length) {
            self.value = self.sum / self.length_f;
            self.primed = true;
            return self.value;
        }

        return math.nan(f64);
    }

    /// Update using a single sample value as substitute for high, low, close.
    pub fn updateSample(self: *AverageDirectionalMovementIndex, sample: f64) f64 {
        return self.update(sample, sample, sample);
    }

    pub fn isPrimed(self: *const AverageDirectionalMovementIndex) bool {
        return self.primed;
    }

    pub fn getMetadata(_: *const AverageDirectionalMovementIndex, out: *Metadata) void {
        build_metadata_mod.buildMetadata(out, Identifier.average_directional_movement_index, mnemonic_str, description_str, &.{
            .{ .mnemonic = mnemonic_str, .description = description_str },
            .{ .mnemonic = "dx", .description = "Directional Movement Index" },
            .{ .mnemonic = "+di", .description = "Directional Indicator Plus" },
            .{ .mnemonic = "-di", .description = "Directional Indicator Minus" },
            .{ .mnemonic = "+dm", .description = "Directional Movement Plus" },
            .{ .mnemonic = "-dm", .description = "Directional Movement Minus" },
            .{ .mnemonic = "atr", .description = "Average True Range" },
            .{ .mnemonic = "tr", .description = "True Range" },
        });
    }

    fn makeOutput(self: *const AverageDirectionalMovementIndex, time: i64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = Scalar{ .time = time, .value = self.value } });
        return out;
    }

    pub fn updateScalar(self: *AverageDirectionalMovementIndex, sample: *const Scalar) OutputArray {
        _ = self.update(sample.value, sample.value, sample.value);
        return self.makeOutput(sample.time);
    }

    pub fn updateBar(self: *AverageDirectionalMovementIndex, sample: *const Bar) OutputArray {
        _ = self.update(sample.close, sample.high, sample.low);
        return self.makeOutput(sample.time);
    }

    pub fn updateQuote(self: *AverageDirectionalMovementIndex, sample: *const Quote) OutputArray {
        const mid = (sample.bid_price + sample.ask_price) / 2.0;
        _ = self.update(mid, mid, mid);
        return self.makeOutput(sample.time);
    }

    pub fn updateTrade(self: *AverageDirectionalMovementIndex, sample: *const Trade) OutputArray {
        _ = self.update(sample.price, sample.price, sample.price);
        return self.makeOutput(sample.time);
    }

    // --- Indicator interface ---

    pub fn indicator(self: *AverageDirectionalMovementIndex) indicator_mod.Indicator {
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
        const self: *const AverageDirectionalMovementIndex = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *const anyopaque, out: *Metadata) void {
        const self: *const AverageDirectionalMovementIndex = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *AverageDirectionalMovementIndex = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *AverageDirectionalMovementIndex = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *AverageDirectionalMovementIndex = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *AverageDirectionalMovementIndex = @ptrCast(@alignCast(ptr));
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
// Expected ADX14 (length=14), 252 entries. 27 NaN (indices 0-26), values from index 27 onward.
test "AverageDirectionalMovementIndex update length=14" {
    const tolerance = 1e-8;
    var adx = try AverageDirectionalMovementIndex.init(testing.allocator, .{ .length = 14 });
    defer adx.deinit();

    for (0..testdata.test_input_close.len) |i| {
        const act = adx.update(testdata.test_input_close[i], testdata.test_input_high[i], testdata.test_input_low[i]);
        const exp = testdata.test_expected_adx14[i];

        if (math.isNan(exp)) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(!math.isNan(act));
            if (!almostEqual(act, exp, tolerance)) {
                std.debug.print("[{d}] expected {d}, got {d}\n", .{ i, exp, act });
                try testing.expect(false);
            }
        }
    }
}

test "AverageDirectionalMovementIndex isPrimed length=14" {
    var adx = try AverageDirectionalMovementIndex.init(testing.allocator, .{ .length = 14 });
    defer adx.deinit();

    // ADX primes after DX primes (at index 14) + length more DX values.
    // DX primes at index 14 (after 15 updates). Then ADX needs 14 DX values: indices 14..27.
    // So ADX primes at index 27 (after 28 updates).
    for (0..27) |i| {
        _ = adx.update(testdata.test_input_close[i], testdata.test_input_high[i], testdata.test_input_low[i]);
        try testing.expect(!adx.isPrimed());
    }

    _ = adx.update(testdata.test_input_close[27], testdata.test_input_high[27], testdata.test_input_low[27]);
    try testing.expect(adx.isPrimed());
}

test "AverageDirectionalMovementIndex constructor validation" {
    try testing.expectError(error.InvalidLength, AverageDirectionalMovementIndex.init(testing.allocator, .{ .length = 0 }));
    try testing.expectError(error.InvalidLength, AverageDirectionalMovementIndex.init(testing.allocator, .{ .length = -8 }));

    var adx = try AverageDirectionalMovementIndex.init(testing.allocator, .{ .length = 14 });
    defer adx.deinit();
    try testing.expect(!adx.isPrimed());
}

test "AverageDirectionalMovementIndex NaN passthrough" {
    var adx = try AverageDirectionalMovementIndex.init(testing.allocator, .{ .length = 14 });
    defer adx.deinit();

    try testing.expect(math.isNan(adx.update(math.nan(f64), 1, 1)));
    try testing.expect(math.isNan(adx.update(1, math.nan(f64), 1)));
    try testing.expect(math.isNan(adx.update(1, 1, math.nan(f64))));
    try testing.expect(math.isNan(adx.updateSample(math.nan(f64))));
}

test "AverageDirectionalMovementIndex metadata" {
    var adx = try AverageDirectionalMovementIndex.init(testing.allocator, .{ .length = 14 });
    defer adx.deinit();
    var meta: Metadata = undefined;
    adx.getMetadata(&meta);

    try testing.expectEqual(Identifier.average_directional_movement_index, meta.identifier);
    try testing.expectEqualStrings("adx", meta.mnemonic);
    try testing.expectEqual(@as(usize, 8), meta.outputs_len);
}

test "AverageDirectionalMovementIndex updateBar" {
    var adx = try AverageDirectionalMovementIndex.init(testing.allocator, .{ .length = 14 });
    defer adx.deinit();

    for (0..27) |i| {
        _ = adx.update(testdata.test_input_close[i], testdata.test_input_high[i], testdata.test_input_low[i]);
    }

    const bar = Bar{
        .time = 1000,
        .open = 91,
        .high = testdata.test_input_high[27],
        .low = testdata.test_input_low[27],
        .close = testdata.test_input_close[27],
        .volume = 1000,
    };
    const out = adx.updateBar(&bar);
    try testing.expect(!math.isNan(out.slice()[0].scalar.value));
}
