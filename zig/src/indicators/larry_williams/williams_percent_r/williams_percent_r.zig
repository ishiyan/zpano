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

/// Enumerates the outputs of the Williams %R indicator.
pub const WilliamsPercentROutput = enum(u8) {
    /// The scalar value of the Williams %R.
    value = 1,
};

/// Parameters to create an instance of the Williams %R indicator.
pub const WilliamsPercentRParams = struct {
    /// The number of time periods. Must be >= 2. Default is 14.
    length: usize = 14,
};

/// Larry Williams' Williams %R momentum indicator.
///
/// Williams %R reflects the level of the closing price relative to the
/// highest high over a lookback period. The oscillation ranges from 0 to -100;
/// readings from 0 to -20 are considered overbought, readings from -80 to -100
/// are considered oversold.
///
/// The value is calculated as:
///
///   %R = -100 * (HighestHigh - Close) / (HighestHigh - LowestLow)
///
/// where HighestHigh and LowestLow are computed over the last `length` bars.
/// If HighestHigh equals LowestLow, the value is 0.
///
/// The indicator requires bar data (high, low, close). For scalar, quote, and
/// trade updates, the single value is used as a substitute for all three.
pub const WilliamsPercentR = struct {
    length: usize,
    high_buf: []f64,
    low_buf: []f64,
    buffer_index: usize,
    count: usize,
    value: f64,
    primed: bool,
    allocator: std.mem.Allocator,

    pub const Error = error{
        InvalidLength,
        OutOfMemory,
    };

    pub fn init(allocator: std.mem.Allocator, params: WilliamsPercentRParams) Error!WilliamsPercentR {
        const length = params.length;
        if (length < 2) return error.InvalidLength;

        const high_buf = allocator.alloc(f64, length) catch return error.OutOfMemory;
        errdefer allocator.free(high_buf);
        const low_buf = allocator.alloc(f64, length) catch return error.OutOfMemory;

        return WilliamsPercentR{
            .length = length,
            .high_buf = high_buf,
            .low_buf = low_buf,
            .buffer_index = 0,
            .count = 0,
            .value = math.nan(f64),
            .primed = false,
            .allocator = allocator,
        };
    }

    pub fn deinit(self: *WilliamsPercentR) void {
        self.allocator.free(self.high_buf);
        self.allocator.free(self.low_buf);
    }

    pub fn fixSlices(self: *WilliamsPercentR) void {
        _ = self;
    }

    /// Core update given close, high, low values.
    pub fn update(self: *WilliamsPercentR, close: f64, high: f64, low: f64) f64 {
        if (math.isNan(close) or math.isNan(high) or math.isNan(low)) {
            return math.nan(f64);
        }

        self.high_buf[self.buffer_index] = high;
        self.low_buf[self.buffer_index] = low;

        self.buffer_index = (self.buffer_index + 1) % self.length;
        self.count += 1;

        if (self.count < self.length) {
            return self.value;
        }

        // Find highest high and lowest low in the window.
        var max_high = self.high_buf[0];
        var min_low = self.low_buf[0];
        for (self.high_buf[1..], self.low_buf[1..]) |h, l| {
            if (h > max_high) max_high = h;
            if (l < min_low) min_low = l;
        }

        const diff = max_high - min_low;
        if (@abs(diff) < math.floatMin(f64)) {
            self.value = 0;
        } else {
            self.value = -100.0 * (max_high - close) / diff;
        }

        if (!self.primed) {
            self.primed = true;
        }

        return self.value;
    }

    /// Update using a single sample value as substitute for high, low, close.
    pub fn updateSample(self: *WilliamsPercentR, sample: f64) f64 {
        return self.update(sample, sample, sample);
    }

    pub fn isPrimed(self: *const WilliamsPercentR) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const WilliamsPercentR, out: *Metadata) void {
        _ = self;
        const mnemonic = "willr";
        const description = "Williams %R";

        build_metadata_mod.buildMetadata(out, Identifier.williams_percent_r, mnemonic, description, &.{
            .{ .mnemonic = mnemonic, .description = description },
        });
    }

    fn makeOutput(self: *const WilliamsPercentR, time: i64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = Scalar{ .time = time, .value = self.value } });
        return out;
    }

    pub fn updateScalar(self: *WilliamsPercentR, sample: *const Scalar) OutputArray {
        _ = self.update(sample.value, sample.value, sample.value);
        return self.makeOutput(sample.time);
    }

    pub fn updateBar(self: *WilliamsPercentR, sample: *const Bar) OutputArray {
        _ = self.update(sample.close, sample.high, sample.low);
        return self.makeOutput(sample.time);
    }

    pub fn updateQuote(self: *WilliamsPercentR, sample: *const Quote) OutputArray {
        const mid = (sample.bid_price + sample.ask_price) / 2.0;
        _ = self.update(mid, mid, mid);
        return self.makeOutput(sample.time);
    }

    pub fn updateTrade(self: *WilliamsPercentR, sample: *const Trade) OutputArray {
        _ = self.update(sample.price, sample.price, sample.price);
        return self.makeOutput(sample.time);
    }

    // --- Indicator interface ---

    pub fn indicator(self: *WilliamsPercentR) indicator_mod.Indicator {
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
        const self: *const WilliamsPercentR = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *const anyopaque, out: *Metadata) void {
        const self: *const WilliamsPercentR = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *WilliamsPercentR = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *WilliamsPercentR = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *WilliamsPercentR = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *WilliamsPercentR = @ptrCast(@alignCast(ptr));
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

test "WilliamsPercentR update period=14" {
    const tolerance = 1e-6;
    const allocator = testing.allocator;

    var w = try WilliamsPercentR.init(allocator, .{ .length = 14 });
    defer w.deinit();

    for (0..testdata.test_close.len) |i| {
        const act = w.update(testdata.test_close[i], testdata.test_high[i], testdata.test_low[i]);
        const exp = testdata.test_expected_14[i];

        if (math.isNan(exp)) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(!math.isNan(act));
            try testing.expect(almostEqual(act, exp, tolerance));
        }
    }
}

test "WilliamsPercentR update period=2" {
    const tolerance = 1e-6;
    const allocator = testing.allocator;

    var w = try WilliamsPercentR.init(allocator, .{ .length = 2 });
    defer w.deinit();

    for (0..testdata.test_close.len) |i| {
        const act = w.update(testdata.test_close[i], testdata.test_high[i], testdata.test_low[i]);
        const exp = testdata.test_expected_2[i];

        if (math.isNan(exp)) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(!math.isNan(act));
            try testing.expect(almostEqual(act, exp, tolerance));
        }
    }
}

test "WilliamsPercentR NaN passthrough" {
    const allocator = testing.allocator;
    var w = try WilliamsPercentR.init(allocator, .{ .length = 14 });
    defer w.deinit();

    try testing.expect(math.isNan(w.update(math.nan(f64), 1, 1)));
    try testing.expect(math.isNan(w.update(1, math.nan(f64), 1)));
    try testing.expect(math.isNan(w.update(1, 1, math.nan(f64))));
}

test "WilliamsPercentR isPrimed" {
    const allocator = testing.allocator;
    var w = try WilliamsPercentR.init(allocator, .{ .length = 14 });
    defer w.deinit();

    try testing.expect(!w.isPrimed());

    for (0..13) |i| {
        _ = w.update(testdata.test_close[i], testdata.test_high[i], testdata.test_low[i]);
        try testing.expect(!w.isPrimed());
    }

    _ = w.update(testdata.test_close[13], testdata.test_high[13], testdata.test_low[13]);
    try testing.expect(w.isPrimed());
}

test "WilliamsPercentR updateSample" {
    const allocator = testing.allocator;
    var w = try WilliamsPercentR.init(allocator, .{ .length = 14 });
    defer w.deinit();

    for (0..13) |_| {
        const v = w.updateSample(9.0);
        try testing.expect(math.isNan(v));
    }

    const v = w.updateSample(9.0);
    try testing.expectEqual(@as(f64, 0.0), v);
}

test "WilliamsPercentR metadata" {
    const allocator = testing.allocator;
    var w = try WilliamsPercentR.init(allocator, .{ .length = 14 });
    defer w.deinit();

    var meta: Metadata = undefined;
    w.getMetadata(&meta);

    try testing.expectEqual(Identifier.williams_percent_r, meta.identifier);
    try testing.expectEqualStrings("willr", meta.mnemonic);
    try testing.expectEqualStrings("Williams %R", meta.description);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
}

test "WilliamsPercentR updateBar" {
    const allocator = testing.allocator;
    var w = try WilliamsPercentR.init(allocator, .{ .length = 14 });
    defer w.deinit();

    for (0..14) |i| {
        _ = w.update(testdata.test_close[i], testdata.test_high[i], testdata.test_low[i]);
    }

    const bar = Bar{ .time = 42, .open = 0, .high = testdata.test_high[14], .low = testdata.test_low[14], .close = testdata.test_close[14], .volume = 0 };
    const out = w.updateBar(&bar);
    try testing.expectEqual(@as(usize, 1), out.len);
}

test "WilliamsPercentR invalid length" {
    const allocator = testing.allocator;
    const result = WilliamsPercentR.init(allocator, .{ .length = 1 });
    try testing.expectError(error.InvalidLength, result);
}
