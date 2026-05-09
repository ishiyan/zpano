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

/// Enumerates the outputs of the Ultimate Oscillator.
pub const UltimateOscillatorOutput = enum(u8) {
    /// The scalar value of the ultimate oscillator.
    value = 1,
};

/// Parameters to create an instance of the Ultimate Oscillator.
pub const UltimateOscillatorParams = struct {
    /// First (shortest) period. Default 7. Must be >= 2. 0 means use default.
    length1: usize = 7,
    /// Second (medium) period. Default 14. Must be >= 2. 0 means use default.
    length2: usize = 14,
    /// Third (longest) period. Default 28. Must be >= 2. 0 means use default.
    length3: usize = 28,
};

const weight1: f64 = 4.0;
const weight2: f64 = 2.0;
const weight3: f64 = 1.0;
const total_weight: f64 = weight1 + weight2 + weight3;

/// Sorts three values ascending.
fn sortThree(a: usize, b_in: usize, c_in: usize) struct { usize, usize, usize } {
    var a_v = a;
    var b_v = b_in;
    var c_v = c_in;
    if (a_v > b_v) {
        const tmp = a_v;
        a_v = b_v;
        b_v = tmp;
    }
    if (b_v > c_v) {
        const tmp = b_v;
        b_v = c_v;
        c_v = tmp;
    }
    if (a_v > b_v) {
        const tmp = a_v;
        a_v = b_v;
        b_v = tmp;
    }
    return .{ a_v, b_v, c_v };
}

/// Larry Williams' Ultimate Oscillator.
///
/// Combines three different time periods into a single oscillator that measures
/// buying pressure relative to true range. The three periods are weighted 4:2:1.
///
/// The indicator requires bar data (high, low, close). For scalar, quote, and
/// trade updates, the single value is used as a substitute for all three.
pub const UltimateOscillator = struct {
    // Sorted periods: p1 <= p2 <= p3.
    p1: usize,
    p2: usize,
    p3: usize,

    previous_close: f64,

    // Circular buffers for buying pressure and true range (size = p3).
    bp_buffer: []f64,
    tr_buffer: []f64,
    buffer_index: usize,

    // Running sums for each period window.
    bp_sum1: f64,
    bp_sum2: f64,
    bp_sum3: f64,
    tr_sum1: f64,
    tr_sum2: f64,
    tr_sum3: f64,

    // Count of values received (excluding first bar).
    count: usize,
    primed: bool,

    allocator: std.mem.Allocator,
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,

    pub const Error = error{
        InvalidLength1,
        InvalidLength2,
        InvalidLength3,
        OutOfMemory,
        MnemonicTooLong,
    };

    pub fn init(allocator: std.mem.Allocator, params: UltimateOscillatorParams) Error!UltimateOscillator {
        const l1 = if (params.length1 == 0) 7 else params.length1;
        const l2 = if (params.length2 == 0) 14 else params.length2;
        const l3 = if (params.length3 == 0) 28 else params.length3;

        if (l1 < 2) return error.InvalidLength1;
        if (l2 < 2) return error.InvalidLength2;
        if (l3 < 2) return error.InvalidLength3;

        const sorted = sortThree(l1, l2, l3);
        const s1 = sorted[0];
        const s2 = sorted[1];
        const s3 = sorted[2];

        var mnemonic_buf: [64]u8 = undefined;
        const mn_slice = std.fmt.bufPrint(&mnemonic_buf, "ultosc({d}, {d}, {d})", .{ l1, l2, l3 }) catch return error.MnemonicTooLong;
        const mnemonic_len = mn_slice.len;

        const bp_buffer = allocator.alloc(f64, s3) catch return error.OutOfMemory;
        errdefer allocator.free(bp_buffer);
        const tr_buffer = allocator.alloc(f64, s3) catch return error.OutOfMemory;

        @memset(bp_buffer, 0);
        @memset(tr_buffer, 0);

        return UltimateOscillator{
            .p1 = s1,
            .p2 = s2,
            .p3 = s3,
            .previous_close = math.nan(f64),
            .bp_buffer = bp_buffer,
            .tr_buffer = tr_buffer,
            .buffer_index = 0,
            .bp_sum1 = 0,
            .bp_sum2 = 0,
            .bp_sum3 = 0,
            .tr_sum1 = 0,
            .tr_sum2 = 0,
            .tr_sum3 = 0,
            .count = 0,
            .primed = false,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
        };
    }

    pub fn deinit(self: *UltimateOscillator) void {
        self.allocator.free(self.bp_buffer);
        self.allocator.free(self.tr_buffer);
    }

    pub fn fixSlices(self: *UltimateOscillator) void {
        _ = self;
    }

    /// Core update given close, high, low values.
    pub fn update(self: *UltimateOscillator, close: f64, high: f64, low: f64) f64 {
        if (math.isNan(close) or math.isNan(high) or math.isNan(low)) {
            return math.nan(f64);
        }

        // First bar: just store close, return NaN.
        if (math.isNan(self.previous_close)) {
            self.previous_close = close;
            return math.nan(f64);
        }

        // Calculate buying pressure and true range.
        const true_low = @min(low, self.previous_close);
        const bp = close - true_low;

        var tr = high - low;
        const d1 = @abs(self.previous_close - high);
        if (d1 > tr) tr = d1;
        const d2 = @abs(self.previous_close - low);
        if (d2 > tr) tr = d2;

        self.previous_close = close;
        self.count += 1;

        // Remove trailing values BEFORE storing new value.
        if (self.count > self.p1) {
            const old_index = (self.buffer_index + self.p3 - self.p1) % self.p3;
            self.bp_sum1 -= self.bp_buffer[old_index];
            self.tr_sum1 -= self.tr_buffer[old_index];
        }

        if (self.count > self.p2) {
            const old_index = (self.buffer_index + self.p3 - self.p2) % self.p3;
            self.bp_sum2 -= self.bp_buffer[old_index];
            self.tr_sum2 -= self.tr_buffer[old_index];
        }

        if (self.count > self.p3) {
            const old_index = (self.buffer_index + self.p3 - self.p3) % self.p3;
            self.bp_sum3 -= self.bp_buffer[old_index];
            self.tr_sum3 -= self.tr_buffer[old_index];
        }

        // Add to running sums.
        self.bp_sum1 += bp;
        self.bp_sum2 += bp;
        self.bp_sum3 += bp;
        self.tr_sum1 += tr;
        self.tr_sum2 += tr;
        self.tr_sum3 += tr;

        // Store in circular buffer.
        self.bp_buffer[self.buffer_index] = bp;
        self.tr_buffer[self.buffer_index] = tr;

        // Advance buffer index.
        self.buffer_index = (self.buffer_index + 1) % self.p3;

        // Need at least p3 values to produce output.
        if (self.count < self.p3) {
            return math.nan(f64);
        }

        self.primed = true;

        // Calculate output.
        var output: f64 = 0;

        if (self.tr_sum1 != 0) {
            output += weight1 * (self.bp_sum1 / self.tr_sum1);
        }
        if (self.tr_sum2 != 0) {
            output += weight2 * (self.bp_sum2 / self.tr_sum2);
        }
        if (self.tr_sum3 != 0) {
            output += weight3 * (self.bp_sum3 / self.tr_sum3);
        }

        return 100.0 * (output / total_weight);
    }

    pub fn isPrimed(self: *const UltimateOscillator) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const UltimateOscillator, out: *Metadata) void {
        const mnemonic = self.mnemonic_buf[0..self.mnemonic_len];

        var description_buf: [128]u8 = undefined;
        const description = std.fmt.bufPrint(&description_buf, "Ultimate Oscillator {s}", .{mnemonic}) catch "Ultimate Oscillator";

        build_metadata_mod.buildMetadata(out, Identifier.ultimate_oscillator, mnemonic, description, &.{
            .{ .mnemonic = mnemonic, .description = description },
        });
    }

    fn makeOutput(self: *const UltimateOscillator, time: i64, value: f64) OutputArray {
        _ = self;
        var out = OutputArray{};
        out.append(.{ .scalar = Scalar{ .time = time, .value = value } });
        return out;
    }

    pub fn updateScalar(self: *UltimateOscillator, sample: *const Scalar) OutputArray {
        const v = self.update(sample.value, sample.value, sample.value);
        return self.makeOutput(sample.time, v);
    }

    pub fn updateBar(self: *UltimateOscillator, sample: *const Bar) OutputArray {
        const v = self.update(sample.close, sample.high, sample.low);
        return self.makeOutput(sample.time, v);
    }

    pub fn updateQuote(self: *UltimateOscillator, sample: *const Quote) OutputArray {
        const mid = (sample.bid_price + sample.ask_price) / 2.0;
        const v = self.update(mid, mid, mid);
        return self.makeOutput(sample.time, v);
    }

    pub fn updateTrade(self: *UltimateOscillator, sample: *const Trade) OutputArray {
        const v = self.update(sample.price, sample.price, sample.price);
        return self.makeOutput(sample.time, v);
    }

    // --- Indicator interface ---

    pub fn indicator(self: *UltimateOscillator) indicator_mod.Indicator {
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
        const self: *const UltimateOscillator = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *const anyopaque, out: *Metadata) void {
        const self: *const UltimateOscillator = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *UltimateOscillator = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *UltimateOscillator = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *UltimateOscillator = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *UltimateOscillator = @ptrCast(@alignCast(ptr));
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

test "UltimateOscillator update default 7-14-28" {
    const tolerance = 1e-4;
    const allocator = testing.allocator;

    var ind = try UltimateOscillator.init(allocator, .{});
    defer ind.deinit();

    for (0..testdata.test_high.len) |i| {
        const result = ind.update(testdata.test_close[i], testdata.test_high[i], testdata.test_low[i]);
        const expected = testdata.test_expected[i];

        if (math.isNan(expected)) {
            try testing.expect(math.isNan(result));
        } else {
            try testing.expect(!math.isNan(result));
            try testing.expect(almostEqual(result, expected, tolerance));
        }
    }
}

test "UltimateOscillator isPrimed default" {
    const allocator = testing.allocator;

    var ind = try UltimateOscillator.init(allocator, .{});
    defer ind.deinit();

    try testing.expect(!ind.isPrimed());

    // Need 28 values (first bar sets prev close, then 28 more = index 28 primes).
    for (0..28) |i| {
        _ = ind.update(testdata.test_close[i], testdata.test_high[i], testdata.test_low[i]);
        try testing.expect(!ind.isPrimed());
    }

    _ = ind.update(testdata.test_close[28], testdata.test_high[28], testdata.test_low[28]);
    try testing.expect(ind.isPrimed());
}

test "UltimateOscillator isPrimed custom 2-3-4" {
    const allocator = testing.allocator;

    var ind = try UltimateOscillator.init(allocator, .{ .length1 = 2, .length2 = 3, .length3 = 4 });
    defer ind.deinit();

    for (0..4) |i| {
        _ = ind.update(testdata.test_close[i], testdata.test_high[i], testdata.test_low[i]);
        try testing.expect(!ind.isPrimed());
    }

    _ = ind.update(testdata.test_close[4], testdata.test_high[4], testdata.test_low[4]);
    try testing.expect(ind.isPrimed());
}

test "UltimateOscillator NaN passthrough" {
    const allocator = testing.allocator;

    var ind = try UltimateOscillator.init(allocator, .{});
    defer ind.deinit();

    try testing.expect(math.isNan(ind.update(math.nan(f64), 100, 90)));
    try testing.expect(math.isNan(ind.update(95, math.nan(f64), 90)));
    try testing.expect(math.isNan(ind.update(95, 100, math.nan(f64))));
}

test "UltimateOscillator metadata" {
    const allocator = testing.allocator;

    var ind = try UltimateOscillator.init(allocator, .{});
    defer ind.deinit();

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqual(Identifier.ultimate_oscillator, meta.identifier);
    try testing.expectEqualStrings("ultosc(7, 14, 28)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
}

test "UltimateOscillator invalid params" {
    const allocator = testing.allocator;

    try testing.expectError(error.InvalidLength1, UltimateOscillator.init(allocator, .{ .length1 = 1 }));
    try testing.expectError(error.InvalidLength2, UltimateOscillator.init(allocator, .{ .length2 = 1 }));
    try testing.expectError(error.InvalidLength3, UltimateOscillator.init(allocator, .{ .length3 = 1 }));
}

test "UltimateOscillator spot checks" {
    const tolerance = 1e-4;
    const allocator = testing.allocator;

    var ind = try UltimateOscillator.init(allocator, .{});
    defer ind.deinit();

    var results: [252]f64 = undefined;
    for (0..testdata.test_high.len) |i| {
        results[i] = ind.update(testdata.test_close[i], testdata.test_high[i], testdata.test_low[i]);
    }

    try testing.expect(almostEqual(results[28], 47.1713, tolerance));
    try testing.expect(almostEqual(results[29], 46.2802, tolerance));
    try testing.expect(almostEqual(results[251], 40.0854, tolerance));
}

test "UltimateOscillator updateBar" {
    const allocator = testing.allocator;

    var ind = try UltimateOscillator.init(allocator, .{});
    defer ind.deinit();

    const bar = Bar{ .time = 42, .open = 0, .high = 100, .low = 90, .close = 95, .volume = 0 };
    const out = ind.updateBar(&bar);
    try testing.expectEqual(@as(usize, 1), out.len);
}
