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

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the Normalized Average True Range indicator.
pub const NormalizedAverageTrueRangeOutput = enum(u8) {
    /// The scalar value of the normalized average true range.
    value = 1,
};

/// Welles Wilder's Normalized Average True Range (NATR) indicator.
///
/// NATR is calculated as (ATR / close) * 100, where ATR is the Average True Range.
/// If close == 0, the result is 0 (not division by zero).
/// The indicator is not primed during the first length updates.
pub const NormalizedAverageTrueRange = struct {
    length: i32,
    value: f64,
    primed: bool,
    average_true_range: AverageTrueRange,

    const mnemonic_str = "natr";
    const description_str = "Normalized Average True Range";

    pub const Error = error{
        InvalidLength,
        OutOfMemory,
    };

    pub fn init(allocator: std.mem.Allocator, params: struct { length: i32 = 14 }) Error!NormalizedAverageTrueRange {
        if (params.length < 1) return Error.InvalidLength;

        const atr = atr_mod.AverageTrueRange.init(allocator, .{ .length = params.length }) catch |err| switch (err) {
            error.InvalidLength => return Error.InvalidLength,
            error.OutOfMemory => return Error.OutOfMemory,
        };

        return .{
            .length = params.length,
            .value = math.nan(f64),
            .primed = false,
            .average_true_range = atr,
        };
    }

    pub fn deinit(self: *NormalizedAverageTrueRange) void {
        self.average_true_range.deinit();
    }

    pub fn fixSlices(_: *NormalizedAverageTrueRange) void {}

    /// Update given close, high, low values.
    pub fn update(self: *NormalizedAverageTrueRange, close: f64, high: f64, low: f64) f64 {
        if (math.isNan(close) or math.isNan(high) or math.isNan(low)) return math.nan(f64);

        const atr_value = self.average_true_range.update(close, high, low);

        if (self.average_true_range.isPrimed()) {
            self.primed = true;

            if (close == 0) {
                self.value = 0;
            } else {
                self.value = (atr_value / close) * 100.0;
            }
        }

        if (self.primed) {
            return self.value;
        }

        return math.nan(f64);
    }

    /// Update using a single sample value as substitute for high, low, close.
    pub fn updateSample(self: *NormalizedAverageTrueRange, sample: f64) f64 {
        return self.update(sample, sample, sample);
    }

    pub fn isPrimed(self: *const NormalizedAverageTrueRange) bool {
        return self.primed;
    }

    pub fn getMetadata(_: *const NormalizedAverageTrueRange, out: *Metadata) void {
        build_metadata_mod.buildMetadata(out, Identifier.normalized_average_true_range, mnemonic_str, description_str, &.{
            .{ .mnemonic = mnemonic_str, .description = description_str },
        });
    }

    fn makeOutput(self: *const NormalizedAverageTrueRange, time: i64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = Scalar{ .time = time, .value = self.value } });
        return out;
    }

    pub fn updateScalar(self: *NormalizedAverageTrueRange, sample: *const Scalar) OutputArray {
        _ = self.update(sample.value, sample.value, sample.value);
        return self.makeOutput(sample.time);
    }

    pub fn updateBar(self: *NormalizedAverageTrueRange, sample: *const Bar) OutputArray {
        _ = self.update(sample.close, sample.high, sample.low);
        return self.makeOutput(sample.time);
    }

    pub fn updateQuote(self: *NormalizedAverageTrueRange, sample: *const Quote) OutputArray {
        const mid = (sample.bid_price + sample.ask_price) / 2.0;
        _ = self.update(mid, mid, mid);
        return self.makeOutput(sample.time);
    }

    pub fn updateTrade(self: *NormalizedAverageTrueRange, sample: *const Trade) OutputArray {
        _ = self.update(sample.price, sample.price, sample.price);
        return self.makeOutput(sample.time);
    }

    // --- Indicator interface ---

    pub fn indicator(self: *NormalizedAverageTrueRange) indicator_mod.Indicator {
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
        const self: *const NormalizedAverageTrueRange = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *const anyopaque, out: *Metadata) void {
        const self: *const NormalizedAverageTrueRange = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *NormalizedAverageTrueRange = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *NormalizedAverageTrueRange = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *NormalizedAverageTrueRange = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *NormalizedAverageTrueRange = @ptrCast(@alignCast(ptr));
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
test "NormalizedAverageTrueRange update length=14" {
    const tolerance = 1e-11;
    var natr = try NormalizedAverageTrueRange.init(testing.allocator, .{ .length = 14 });
    defer natr.deinit();

    for (0..testdata.test_input_close.len) |i| {
        const act = natr.update(testdata.test_input_close[i], testdata.test_input_high[i], testdata.test_input_low[i]);
        const exp = testdata.test_expected_natr[i];

        if (math.isNan(exp)) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(!math.isNan(act));
            try testing.expect(almostEqual(act, exp, tolerance));
        }
    }
}

test "NormalizedAverageTrueRange update length=1" {
    const tolerance = 1e-11;
    var natr = try NormalizedAverageTrueRange.init(testing.allocator, .{ .length = 1 });
    defer natr.deinit();

    for (0..testdata.test_input_close.len) |i| {
        const act = natr.update(testdata.test_input_close[i], testdata.test_input_high[i], testdata.test_input_low[i]);
        const exp = testdata.test_expected_natr1[i];

        if (math.isNan(exp)) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(!math.isNan(act));
            try testing.expect(almostEqual(act, exp, tolerance));
        }
    }
}

test "NormalizedAverageTrueRange isPrimed length=5" {
    var natr = try NormalizedAverageTrueRange.init(testing.allocator, .{ .length = 5 });
    defer natr.deinit();

    try testing.expect(!natr.isPrimed());

    for (0..5) |i| {
        _ = natr.update(testdata.test_input_close[i], testdata.test_input_high[i], testdata.test_input_low[i]);
        try testing.expect(!natr.isPrimed());
    }

    for (5..10) |i| {
        _ = natr.update(testdata.test_input_close[i], testdata.test_input_high[i], testdata.test_input_low[i]);
        try testing.expect(natr.isPrimed());
    }
}

test "NormalizedAverageTrueRange constructor validation" {
    try testing.expectError(error.InvalidLength, NormalizedAverageTrueRange.init(testing.allocator, .{ .length = 0 }));
    try testing.expectError(error.InvalidLength, NormalizedAverageTrueRange.init(testing.allocator, .{ .length = -8 }));

    var natr = try NormalizedAverageTrueRange.init(testing.allocator, .{ .length = 14 });
    defer natr.deinit();
    try testing.expect(!natr.isPrimed());
}

test "NormalizedAverageTrueRange close=0" {
    var natr = try NormalizedAverageTrueRange.init(testing.allocator, .{ .length = 14 });
    defer natr.deinit();

    // Prime the indicator.
    for (0..15) |i| {
        _ = natr.update(testdata.test_input_close[i], testdata.test_input_high[i], testdata.test_input_low[i]);
    }

    // close=0 should return 0, not panic or NaN.
    const result = natr.update(0, 3.3, 2.2);
    try testing.expect(result == 0);
}

test "NormalizedAverageTrueRange metadata" {
    var natr = try NormalizedAverageTrueRange.init(testing.allocator, .{ .length = 14 });
    defer natr.deinit();
    var meta: Metadata = undefined;
    natr.getMetadata(&meta);

    try testing.expectEqual(Identifier.normalized_average_true_range, meta.identifier);
    try testing.expectEqualStrings("natr", meta.mnemonic);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
}

test "NormalizedAverageTrueRange NaN passthrough" {
    var natr = try NormalizedAverageTrueRange.init(testing.allocator, .{ .length = 14 });
    defer natr.deinit();

    try testing.expect(math.isNan(natr.update(math.nan(f64), 1, 1)));
    try testing.expect(math.isNan(natr.update(1, math.nan(f64), 1)));
    try testing.expect(math.isNan(natr.update(1, 1, math.nan(f64))));
    try testing.expect(math.isNan(natr.updateSample(math.nan(f64))));
}

test "NormalizedAverageTrueRange updateBar" {
    var natr = try NormalizedAverageTrueRange.init(testing.allocator, .{ .length = 14 });
    defer natr.deinit();

    for (0..14) |i| {
        _ = natr.update(testdata.test_input_close[i], testdata.test_input_high[i], testdata.test_input_low[i]);
    }

    const bar = Bar{
        .time = 1000,
        .open = 91,
        .high = testdata.test_input_high[14],
        .low = testdata.test_input_low[14],
        .close = testdata.test_input_close[14],
        .volume = 1000,
    };
    const out = natr.updateBar(&bar);
    try testing.expect(!math.isNan(out.slice()[0].scalar.value));
}
