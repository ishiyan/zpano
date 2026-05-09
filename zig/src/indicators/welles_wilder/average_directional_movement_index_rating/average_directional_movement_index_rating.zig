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

const adx_mod = @import("../average_directional_movement_index/average_directional_movement_index.zig");
const AverageDirectionalMovementIndex = adx_mod.AverageDirectionalMovementIndex;

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the Average Directional Movement Index Rating indicator.
pub const AverageDirectionalMovementIndexRatingOutput = enum(u8) {
    /// The scalar value of the average directional movement index rating (ADXR).
    value = 1,
    /// The scalar value of the average directional movement index (ADX).
    average_directional_movement_index = 2,
    /// The scalar value of the directional movement index (DX).
    directional_movement_index = 3,
    /// The scalar value of the directional indicator plus (+DI).
    directional_indicator_plus = 4,
    /// The scalar value of the directional indicator minus (-DI).
    directional_indicator_minus = 5,
    /// The scalar value of the directional movement plus (+DM).
    directional_movement_plus = 6,
    /// The scalar value of the directional movement minus (-DM).
    directional_movement_minus = 7,
    /// The scalar value of the average true range (ATR).
    average_true_range = 8,
    /// The scalar value of the true range (TR).
    true_range = 9,
};

/// Welles Wilder's Average Directional Movement Index Rating (ADXR).
///
/// The average directional movement index rating averages the current ADX value with
/// the ADX value from (length - 1) periods ago. It is calculated as:
///
///   ADXR = (ADX[current] + ADX[current - (length - 1)]) / 2
///
/// The indicator requires close, high, and low values.
pub const AverageDirectionalMovementIndexRating = struct {
    length: i32,
    buffer_size: usize,
    buffer: []f64,
    buffer_index: usize,
    buffer_count: usize,
    primed: bool,
    value: f64,
    average_directional_movement_index: AverageDirectionalMovementIndex,
    allocator: std.mem.Allocator,

    const mnemonic_str = "adxr";
    const description_str = "Average Directional Movement Index Rating";

    pub const Error = error{
        InvalidLength,
        OutOfMemory,
    };

    pub fn init(allocator: std.mem.Allocator, params: struct { length: i32 = 14 }) Error!AverageDirectionalMovementIndexRating {
        if (params.length < 1) return Error.InvalidLength;

        const adx = AverageDirectionalMovementIndex.init(allocator, .{ .length = params.length }) catch |e| switch (e) {
            error.InvalidLength => return Error.InvalidLength,
            error.OutOfMemory => return Error.OutOfMemory,
        };

        const buffer_size: usize = @intCast(params.length);
        const buffer = allocator.alloc(f64, buffer_size) catch return Error.OutOfMemory;

        return .{
            .length = params.length,
            .buffer_size = buffer_size,
            .buffer = buffer,
            .buffer_index = 0,
            .buffer_count = 0,
            .primed = false,
            .value = math.nan(f64),
            .average_directional_movement_index = adx,
            .allocator = allocator,
        };
    }

    pub fn deinit(self: *AverageDirectionalMovementIndexRating) void {
        self.allocator.free(self.buffer);
        self.average_directional_movement_index.deinit();
    }

    pub fn fixSlices(_: *AverageDirectionalMovementIndexRating) void {}

    /// Update given close, high, low values.
    pub fn update(self: *AverageDirectionalMovementIndexRating, close: f64, high: f64, low: f64) f64 {
        if (math.isNan(close) or math.isNan(high) or math.isNan(low)) return math.nan(f64);

        const adx_value = self.average_directional_movement_index.update(close, high, low);

        if (!self.average_directional_movement_index.isPrimed()) {
            return math.nan(f64);
        }

        // Store ADX value in circular buffer.
        self.buffer[self.buffer_index] = adx_value;
        self.buffer_index = (self.buffer_index + 1) % self.buffer_size;
        self.buffer_count += 1;

        if (self.buffer_count < self.buffer_size) {
            return math.nan(f64);
        }

        // The oldest value in the buffer is at buffer_index (since we just advanced it).
        const old_adx = self.buffer[self.buffer_index % self.buffer_size];
        self.value = (adx_value + old_adx) / 2.0;
        self.primed = true;

        return self.value;
    }

    /// Update using a single sample value as substitute for high, low, close.
    pub fn updateSample(self: *AverageDirectionalMovementIndexRating, sample: f64) f64 {
        return self.update(sample, sample, sample);
    }

    pub fn isPrimed(self: *const AverageDirectionalMovementIndexRating) bool {
        return self.primed;
    }

    pub fn getMetadata(_: *const AverageDirectionalMovementIndexRating, out: *Metadata) void {
        build_metadata_mod.buildMetadata(out, Identifier.average_directional_movement_index_rating, mnemonic_str, description_str, &.{
            .{ .mnemonic = mnemonic_str, .description = description_str },
            .{ .mnemonic = "adx", .description = "Average Directional Movement Index" },
            .{ .mnemonic = "dx", .description = "Directional Movement Index" },
            .{ .mnemonic = "+di", .description = "Directional Indicator Plus" },
            .{ .mnemonic = "-di", .description = "Directional Indicator Minus" },
            .{ .mnemonic = "+dm", .description = "Directional Movement Plus" },
            .{ .mnemonic = "-dm", .description = "Directional Movement Minus" },
            .{ .mnemonic = "atr", .description = "Average True Range" },
            .{ .mnemonic = "tr", .description = "True Range" },
        });
    }

    fn makeOutput(self: *const AverageDirectionalMovementIndexRating, time: i64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = Scalar{ .time = time, .value = self.value } });
        return out;
    }

    pub fn updateScalar(self: *AverageDirectionalMovementIndexRating, sample: *const Scalar) OutputArray {
        _ = self.update(sample.value, sample.value, sample.value);
        return self.makeOutput(sample.time);
    }

    pub fn updateBar(self: *AverageDirectionalMovementIndexRating, sample: *const Bar) OutputArray {
        _ = self.update(sample.close, sample.high, sample.low);
        return self.makeOutput(sample.time);
    }

    pub fn updateQuote(self: *AverageDirectionalMovementIndexRating, sample: *const Quote) OutputArray {
        const mid = (sample.bid_price + sample.ask_price) / 2.0;
        _ = self.update(mid, mid, mid);
        return self.makeOutput(sample.time);
    }

    pub fn updateTrade(self: *AverageDirectionalMovementIndexRating, sample: *const Trade) OutputArray {
        _ = self.update(sample.price, sample.price, sample.price);
        return self.makeOutput(sample.time);
    }

    // --- Indicator interface ---

    pub fn indicator(self: *AverageDirectionalMovementIndexRating) indicator_mod.Indicator {
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
        const self: *const AverageDirectionalMovementIndexRating = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *const anyopaque, out: *Metadata) void {
        const self: *const AverageDirectionalMovementIndexRating = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *AverageDirectionalMovementIndexRating = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *AverageDirectionalMovementIndexRating = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *AverageDirectionalMovementIndexRating = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *AverageDirectionalMovementIndexRating = @ptrCast(@alignCast(ptr));
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

// TA-Lib test data (252 entries) — same as ADX test data.
// Expected ADXR14 (length=14), 252 entries. 40 NaN (indices 0-39), values from index 40 onward.
test "AverageDirectionalMovementIndexRating update length=14" {
    const tolerance = 1e-8;
    var adxr = try AverageDirectionalMovementIndexRating.init(testing.allocator, .{ .length = 14 });
    defer adxr.deinit();

    for (0..testdata.test_input_close.len) |i| {
        const act = adxr.update(testdata.test_input_close[i], testdata.test_input_high[i], testdata.test_input_low[i]);
        const exp = testdata.test_expected_adxr14[i];

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

test "AverageDirectionalMovementIndexRating isPrimed length=14" {
    var adxr = try AverageDirectionalMovementIndexRating.init(testing.allocator, .{ .length = 14 });
    defer adxr.deinit();

    // ADX primes at index 27. ADXR needs (length)=14 ADX values in buffer,
    // so ADXR primes at index 40.
    for (0..40) |i| {
        _ = adxr.update(testdata.test_input_close[i], testdata.test_input_high[i], testdata.test_input_low[i]);
        try testing.expect(!adxr.isPrimed());
    }

    _ = adxr.update(testdata.test_input_close[40], testdata.test_input_high[40], testdata.test_input_low[40]);
    try testing.expect(adxr.isPrimed());
}

test "AverageDirectionalMovementIndexRating constructor validation" {
    try testing.expectError(error.InvalidLength, AverageDirectionalMovementIndexRating.init(testing.allocator, .{ .length = 0 }));
    try testing.expectError(error.InvalidLength, AverageDirectionalMovementIndexRating.init(testing.allocator, .{ .length = -8 }));

    var adxr = try AverageDirectionalMovementIndexRating.init(testing.allocator, .{ .length = 14 });
    defer adxr.deinit();
    try testing.expect(!adxr.isPrimed());
}

test "AverageDirectionalMovementIndexRating NaN passthrough" {
    var adxr = try AverageDirectionalMovementIndexRating.init(testing.allocator, .{ .length = 14 });
    defer adxr.deinit();

    try testing.expect(math.isNan(adxr.update(math.nan(f64), 1, 1)));
    try testing.expect(math.isNan(adxr.update(1, math.nan(f64), 1)));
    try testing.expect(math.isNan(adxr.update(1, 1, math.nan(f64))));
    try testing.expect(math.isNan(adxr.updateSample(math.nan(f64))));
}

test "AverageDirectionalMovementIndexRating metadata" {
    var adxr = try AverageDirectionalMovementIndexRating.init(testing.allocator, .{ .length = 14 });
    defer adxr.deinit();
    var meta: Metadata = undefined;
    adxr.getMetadata(&meta);

    try testing.expectEqual(Identifier.average_directional_movement_index_rating, meta.identifier);
    try testing.expectEqualStrings("adxr", meta.mnemonic);
    try testing.expectEqual(@as(usize, 9), meta.outputs_len);
}

test "AverageDirectionalMovementIndexRating updateBar" {
    var adxr = try AverageDirectionalMovementIndexRating.init(testing.allocator, .{ .length = 14 });
    defer adxr.deinit();

    for (0..40) |i| {
        _ = adxr.update(testdata.test_input_close[i], testdata.test_input_high[i], testdata.test_input_low[i]);
    }

    const bar = Bar{
        .time = 1000,
        .open = 91,
        .high = testdata.test_input_high[40],
        .low = testdata.test_input_low[40],
        .close = testdata.test_input_close[40],
        .volume = 1000,
    };
    const out = adxr.updateBar(&bar);
    try testing.expect(!math.isNan(out.slice()[0].scalar.value));
}
