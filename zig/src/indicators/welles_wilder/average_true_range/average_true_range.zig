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

const true_range_mod = @import("../true_range/true_range.zig");
const TrueRange = true_range_mod.TrueRange;

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the Average True Range indicator.
pub const AverageTrueRangeOutput = enum(u8) {
    /// The scalar value of the Average True Range.
    value = 1,
};

/// Welles Wilder's Average True Range (ATR) indicator.
///
/// ATR averages True Range (TR) values over the specified length using the Wilder method:
///   - multiply the previous value by (length - 1)
///   - add the current TR value
///   - divide by length
///
/// The initial ATR value is a simple average of the first `length` TR values.
/// The indicator is not primed during the first `length` updates.
pub const AverageTrueRange = struct {
    length: i32,
    last_index: i32,
    stage: i32,
    window_count: i32,
    window: ?[]f64,
    window_sum: f64,
    value: f64,
    primed: bool,
    true_range: TrueRange,
    allocator: std.mem.Allocator,

    const mnemonic_str = "atr";
    const description_str = "Average True Range";

    pub const Error = error{
        InvalidLength,
        OutOfMemory,
    };

    pub fn init(allocator: std.mem.Allocator, params: struct { length: i32 = 14 }) Error!AverageTrueRange {
        if (params.length < 1) return Error.InvalidLength;

        const last_index = params.length - 1;
        const window: ?[]f64 = if (last_index > 0)
            allocator.alloc(f64, @intCast(params.length)) catch return Error.OutOfMemory
        else
            null;

        if (window) |w| @memset(w, 0.0);

        return .{
            .length = params.length,
            .last_index = last_index,
            .stage = 0,
            .window_count = 0,
            .window = window,
            .window_sum = 0,
            .value = math.nan(f64),
            .primed = false,
            .true_range = TrueRange.init(),
            .allocator = allocator,
        };
    }

    pub fn deinit(self: *AverageTrueRange) void {
        if (self.window) |w| self.allocator.free(w);
    }

    pub fn fixSlices(_: *AverageTrueRange) void {}

    /// Update given close, high, low values.
    pub fn update(self: *AverageTrueRange, close: f64, high: f64, low: f64) f64 {
        if (math.isNan(close) or math.isNan(high) or math.isNan(low)) return math.nan(f64);

        const tr_value = self.true_range.update(close, high, low);

        if (self.last_index == 0) {
            self.value = tr_value;
            if (self.stage == 0) {
                self.stage += 1;
            } else if (self.stage == 1) {
                self.stage += 1;
                self.primed = true;
            }
            return self.value;
        }

        if (self.stage > 1) {
            // Wilder smoothing method.
            self.value *= @as(f64, @floatFromInt(self.last_index));
            self.value += tr_value;
            self.value /= @as(f64, @floatFromInt(self.length));
            return self.value;
        }

        if (self.stage == 1) {
            self.window_sum += tr_value;
            self.window.?[@intCast(self.window_count)] = tr_value;
            self.window_count += 1;

            if (self.window_count == self.length) {
                self.stage += 1;
                self.primed = true;
                self.value = self.window_sum / @as(f64, @floatFromInt(self.length));
            }

            if (self.primed) return self.value;
            return math.nan(f64);
        }

        // The very first sample is used by the True Range.
        self.stage += 1;
        return math.nan(f64);
    }

    /// Update using a single sample value as substitute for high, low, close.
    pub fn updateSample(self: *AverageTrueRange, sample: f64) f64 {
        return self.update(sample, sample, sample);
    }

    pub fn isPrimed(self: *const AverageTrueRange) bool {
        return self.primed;
    }

    pub fn getMetadata(_: *const AverageTrueRange, out: *Metadata) void {
        build_metadata_mod.buildMetadata(out, Identifier.average_true_range, mnemonic_str, description_str, &.{
            .{ .mnemonic = mnemonic_str, .description = description_str },
        });
    }

    fn makeOutput(self: *const AverageTrueRange, time: i64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = Scalar{ .time = time, .value = self.value } });
        return out;
    }

    pub fn updateScalar(self: *AverageTrueRange, sample: *const Scalar) OutputArray {
        _ = self.update(sample.value, sample.value, sample.value);
        return self.makeOutput(sample.time);
    }

    pub fn updateBar(self: *AverageTrueRange, sample: *const Bar) OutputArray {
        _ = self.update(sample.close, sample.high, sample.low);
        return self.makeOutput(sample.time);
    }

    pub fn updateQuote(self: *AverageTrueRange, sample: *const Quote) OutputArray {
        const mid = (sample.bid_price + sample.ask_price) / 2.0;
        _ = self.update(mid, mid, mid);
        return self.makeOutput(sample.time);
    }

    pub fn updateTrade(self: *AverageTrueRange, sample: *const Trade) OutputArray {
        _ = self.update(sample.price, sample.price, sample.price);
        return self.makeOutput(sample.time);
    }

    // --- Indicator interface ---

    pub fn indicator(self: *AverageTrueRange) indicator_mod.Indicator {
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
        const self: *const AverageTrueRange = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *const anyopaque, out: *Metadata) void {
        const self: *const AverageTrueRange = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *AverageTrueRange = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *AverageTrueRange = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *AverageTrueRange = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *AverageTrueRange = @ptrCast(@alignCast(ptr));
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
test "AverageTrueRange update length=14" {
    const tolerance = 1e-12;
    var atr = try AverageTrueRange.init(testing.allocator, .{ .length = 14 });
    defer atr.deinit();

    for (0..testdata.test_input_close.len) |i| {
        const act = atr.update(testdata.test_input_close[i], testdata.test_input_high[i], testdata.test_input_low[i]);
        const exp = testdata.test_expected_atr[i];

        if (math.isNan(exp)) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(!math.isNan(act));
            try testing.expect(almostEqual(act, exp, tolerance));
        }
    }
}

test "AverageTrueRange isPrimed length=5" {
    var atr = try AverageTrueRange.init(testing.allocator, .{ .length = 5 });
    defer atr.deinit();

    try testing.expect(!atr.isPrimed());

    for (0..5) |i| {
        _ = atr.update(testdata.test_input_close[i], testdata.test_input_high[i], testdata.test_input_low[i]);
        try testing.expect(!atr.isPrimed());
    }

    for (5..10) |i| {
        _ = atr.update(testdata.test_input_close[i], testdata.test_input_high[i], testdata.test_input_low[i]);
        try testing.expect(atr.isPrimed());
    }
}

test "AverageTrueRange constructor validation" {
    try testing.expectError(error.InvalidLength, AverageTrueRange.init(testing.allocator, .{ .length = 0 }));
    try testing.expectError(error.InvalidLength, AverageTrueRange.init(testing.allocator, .{ .length = -8 }));

    var atr = try AverageTrueRange.init(testing.allocator, .{ .length = 14 });
    defer atr.deinit();
    try testing.expect(!atr.isPrimed());
}

test "AverageTrueRange metadata" {
    var atr = try AverageTrueRange.init(testing.allocator, .{ .length = 14 });
    defer atr.deinit();
    var meta: Metadata = undefined;
    atr.getMetadata(&meta);

    try testing.expectEqual(Identifier.average_true_range, meta.identifier);
    try testing.expectEqualStrings("atr", meta.mnemonic);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
}

test "AverageTrueRange NaN passthrough" {
    var atr = try AverageTrueRange.init(testing.allocator, .{ .length = 14 });
    defer atr.deinit();

    try testing.expect(math.isNan(atr.update(math.nan(f64), 1, 1)));
    try testing.expect(math.isNan(atr.update(1, math.nan(f64), 1)));
    try testing.expect(math.isNan(atr.update(1, 1, math.nan(f64))));
    try testing.expect(math.isNan(atr.updateSample(math.nan(f64))));
}

test "AverageTrueRange updateBar" {
    var atr = try AverageTrueRange.init(testing.allocator, .{ .length = 14 });
    defer atr.deinit();

    // Prime with 14 bars.
    for (0..14) |i| {
        _ = atr.update(testdata.test_input_close[i], testdata.test_input_high[i], testdata.test_input_low[i]);
    }

    const bar = Bar{
        .time = 1000,
        .open = 91,
        .high = testdata.test_input_high[14],
        .low = testdata.test_input_low[14],
        .close = testdata.test_input_close[14],
        .volume = 1000,
    };
    const out = atr.updateBar(&bar);
    try testing.expect(!math.isNan(out.slice()[0].scalar.value));
}
