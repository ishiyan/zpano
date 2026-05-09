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

/// Enumerates the outputs of the Aroon indicator.
pub const AroonOutput = enum(u8) {
    /// The Aroon Up line.
    up = 1,
    /// The Aroon Down line.
    down = 2,
    /// The Aroon Oscillator (Up - Down).
    osc = 3,
};

/// Parameters to create an instance of the Aroon indicator.
pub const AroonParams = struct {
    /// The lookback period. Must be >= 2. Default is 14.
    length: usize = 14,
};

/// Tushar Chande's Aroon indicator.
///
/// Measures the number of periods since the highest high and lowest low
/// within a lookback window. Produces three outputs:
///   - Up: 100 * (Length - periods since highest high) / Length
///   - Down: 100 * (Length - periods since lowest low) / Length
///   - Osc: Up - Down
///
/// The indicator requires bar data (high, low). For scalar, quote, and
/// trade updates, the single value substitutes for both.
pub const Aroon = struct {
    length: usize,
    factor: f64,
    high_buf: []f64,
    low_buf: []f64,
    buffer_index: usize,
    count: usize,
    highest_index: usize,
    lowest_index: usize,
    up: f64,
    down: f64,
    osc: f64,
    primed: bool,
    allocator: std.mem.Allocator,
    mnemonic_buf: [32]u8,
    mnemonic_len: usize,

    pub const Error = error{
        InvalidLength,
        OutOfMemory,
    };

    pub fn init(allocator: std.mem.Allocator, params: AroonParams) Error!Aroon {
        const length = params.length;
        if (length < 2) return error.InvalidLength;

        const window_size = length + 1;
        const high_buf = allocator.alloc(f64, window_size) catch return error.OutOfMemory;
        errdefer allocator.free(high_buf);
        const low_buf = allocator.alloc(f64, window_size) catch return error.OutOfMemory;

        var mnemonic_buf: [32]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "aroon({d})", .{length}) catch
            return error.InvalidLength;

        return Aroon{
            .length = length,
            .factor = 100.0 / @as(f64, @floatFromInt(length)),
            .high_buf = high_buf,
            .low_buf = low_buf,
            .buffer_index = 0,
            .count = 0,
            .highest_index = 0,
            .lowest_index = 0,
            .up = math.nan(f64),
            .down = math.nan(f64),
            .osc = math.nan(f64),
            .primed = false,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_slice.len,
        };
    }

    pub fn deinit(self: *Aroon) void {
        self.allocator.free(self.high_buf);
        self.allocator.free(self.low_buf);
    }

    pub fn fixSlices(self: *Aroon) void {
        _ = self;
    }

    /// Core update given high and low values.
    /// Returns [up, down, osc].
    pub fn updateHighLow(self: *Aroon, high: f64, low: f64) [3]f64 {
        if (math.isNan(high) or math.isNan(low)) {
            return .{ math.nan(f64), math.nan(f64), math.nan(f64) };
        }

        const window_size = self.length + 1;
        const today = self.count;

        // Store in circular buffer.
        const pos = self.buffer_index;
        self.high_buf[pos] = high;
        self.low_buf[pos] = low;
        self.buffer_index = (self.buffer_index + 1) % window_size;
        self.count += 1;

        // Need at least length+1 bars (indices 0..length).
        if (self.count < window_size) {
            return .{ self.up, self.down, self.osc };
        }

        const trailing_index = today - self.length;

        if (self.count == window_size) {
            // First time: scan entire window.
            self.highest_index = trailing_index;
            self.lowest_index = trailing_index;

            var i = trailing_index + 1;
            while (i <= today) : (i += 1) {
                const buf_pos = i % window_size;
                if (self.high_buf[buf_pos] >= self.high_buf[self.highest_index % window_size]) {
                    self.highest_index = i;
                }
                if (self.low_buf[buf_pos] <= self.low_buf[self.lowest_index % window_size]) {
                    self.lowest_index = i;
                }
            }
        } else {
            // Subsequent: optimized update.
            if (self.highest_index < trailing_index) {
                self.highest_index = trailing_index;
                var i = trailing_index + 1;
                while (i <= today) : (i += 1) {
                    const buf_pos = i % window_size;
                    if (self.high_buf[buf_pos] >= self.high_buf[self.highest_index % window_size]) {
                        self.highest_index = i;
                    }
                }
            } else if (high >= self.high_buf[self.highest_index % window_size]) {
                self.highest_index = today;
            }

            if (self.lowest_index < trailing_index) {
                self.lowest_index = trailing_index;
                var i = trailing_index + 1;
                while (i <= today) : (i += 1) {
                    const buf_pos = i % window_size;
                    if (self.low_buf[buf_pos] <= self.low_buf[self.lowest_index % window_size]) {
                        self.lowest_index = i;
                    }
                }
            } else if (low <= self.low_buf[self.lowest_index % window_size]) {
                self.lowest_index = today;
            }
        }

        self.up = self.factor * @as(f64, @floatFromInt(self.length - (today - self.highest_index)));
        self.down = self.factor * @as(f64, @floatFromInt(self.length - (today - self.lowest_index)));
        self.osc = self.up - self.down;

        if (!self.primed) {
            self.primed = true;
        }

        return .{ self.up, self.down, self.osc };
    }

    pub fn isPrimed(self: *const Aroon) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const Aroon, out: *Metadata) void {
        const mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        const desc_prefix = "Aroon ";

        var desc_buf: [64]u8 = undefined;
        const desc = std.fmt.bufPrint(&desc_buf, "{s}{s}", .{ desc_prefix, mnemonic }) catch mnemonic;

        var up_mnemonic_buf: [48]u8 = undefined;
        const up_mnemonic = std.fmt.bufPrint(&up_mnemonic_buf, "{s} up", .{mnemonic}) catch mnemonic;
        var down_mnemonic_buf: [48]u8 = undefined;
        const down_mnemonic = std.fmt.bufPrint(&down_mnemonic_buf, "{s} down", .{mnemonic}) catch mnemonic;
        var osc_mnemonic_buf: [48]u8 = undefined;
        const osc_mnemonic = std.fmt.bufPrint(&osc_mnemonic_buf, "{s} osc", .{mnemonic}) catch mnemonic;

        var up_desc_buf: [64]u8 = undefined;
        const up_desc = std.fmt.bufPrint(&up_desc_buf, "{s} Up", .{desc}) catch desc;
        var down_desc_buf: [64]u8 = undefined;
        const down_desc = std.fmt.bufPrint(&down_desc_buf, "{s} Down", .{desc}) catch desc;
        var osc_desc_buf: [64]u8 = undefined;
        const osc_desc = std.fmt.bufPrint(&osc_desc_buf, "{s} Oscillator", .{desc}) catch desc;

        build_metadata_mod.buildMetadata(out, Identifier.aroon, mnemonic, desc, &.{
            .{ .mnemonic = up_mnemonic, .description = up_desc },
            .{ .mnemonic = down_mnemonic, .description = down_desc },
            .{ .mnemonic = osc_mnemonic, .description = osc_desc },
        });
    }

    fn makeOutput(self: *const Aroon, time: i64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = Scalar{ .time = time, .value = self.up } });
        out.append(.{ .scalar = Scalar{ .time = time, .value = self.down } });
        out.append(.{ .scalar = Scalar{ .time = time, .value = self.osc } });
        return out;
    }

    pub fn updateScalar(self: *Aroon, sample: *const Scalar) OutputArray {
        _ = self.updateHighLow(sample.value, sample.value);
        return self.makeOutput(sample.time);
    }

    pub fn updateBar(self: *Aroon, sample: *const Bar) OutputArray {
        _ = self.updateHighLow(sample.high, sample.low);
        return self.makeOutput(sample.time);
    }

    pub fn updateQuote(self: *Aroon, sample: *const Quote) OutputArray {
        const mid = (sample.bid_price + sample.ask_price) / 2.0;
        _ = self.updateHighLow(mid, mid);
        return self.makeOutput(sample.time);
    }

    pub fn updateTrade(self: *Aroon, sample: *const Trade) OutputArray {
        _ = self.updateHighLow(sample.price, sample.price);
        return self.makeOutput(sample.time);
    }

    // --- Indicator interface ---

    pub fn indicator(self: *Aroon) indicator_mod.Indicator {
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

    fn vtableIsPrimed(ptr: *const anyopaque) bool {
        const self: *const Aroon = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *const anyopaque, out: *Metadata) void {
        const self: *const Aroon = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *Aroon = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *Aroon = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *Aroon = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *Aroon = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }
};

// --- Tests ---

const tolerance: f64 = 1e-6;

fn almostEqual(a: f64, b: f64, tol: f64) bool {
    if (math.isNan(a) and math.isNan(b)) return true;
    if (math.isNan(a) or math.isNan(b)) return false;
    return @abs(a - b) <= tol;
}

const testdata = @import("testdata.zig");
const ExpectedRow = testdata.ExpectedRow;

test "Aroon length=14 full data" {
    const allocator = std.testing.allocator;
    var ind = try Aroon.init(allocator, .{ .length = 14 });
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateHighLow(testdata.test_high[i], testdata.test_low[i]);
        const exp = testdata.test_expected[i];

        if (math.isNan(exp.up)) {
            try std.testing.expect(math.isNan(result[0]));
            continue;
        }

        try std.testing.expect(almostEqual(result[0], exp.up, tolerance));
        try std.testing.expect(almostEqual(result[1], exp.down, tolerance));
        try std.testing.expect(almostEqual(result[2], exp.osc, tolerance));
    }
}

test "Aroon isPrimed" {
    const allocator = std.testing.allocator;
    var ind = try Aroon.init(allocator, .{ .length = 14 });
    defer ind.deinit();

    try std.testing.expect(!ind.isPrimed());

    for (0..14) |i| {
        _ = ind.updateHighLow(testdata.test_high[i], testdata.test_low[i]);
        try std.testing.expect(!ind.isPrimed());
    }

    _ = ind.updateHighLow(testdata.test_high[14], testdata.test_low[14]);
    try std.testing.expect(ind.isPrimed());
}

test "Aroon NaN input" {
    const allocator = std.testing.allocator;
    var ind = try Aroon.init(allocator, .{ .length = 14 });
    defer ind.deinit();

    const result = ind.updateHighLow(math.nan(f64), 1.0);
    try std.testing.expect(math.isNan(result[0]));
    try std.testing.expect(math.isNan(result[1]));
    try std.testing.expect(math.isNan(result[2]));
}

test "Aroon invalid params" {
    const allocator = std.testing.allocator;
    try std.testing.expectError(error.InvalidLength, Aroon.init(allocator, .{ .length = 1 }));
    try std.testing.expectError(error.InvalidLength, Aroon.init(allocator, .{ .length = 0 }));
}

test "Aroon metadata" {
    const allocator = std.testing.allocator;
    var ind = try Aroon.init(allocator, .{ .length = 14 });
    defer ind.deinit();

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try std.testing.expectEqual(Identifier.aroon, meta.identifier);
    try std.testing.expectEqualStrings("aroon(14)", meta.mnemonic);
    try std.testing.expectEqual(@as(usize, 3), meta.outputs_len);
}

test "Aroon updateBar" {
    const allocator = std.testing.allocator;
    var ind = try Aroon.init(allocator, .{ .length = 14 });
    defer ind.deinit();

    // Feed 14 bars (not primed yet).
    for (0..14) |i| {
        const bar = Bar{ .time = 0, .open = 0, .high = testdata.test_high[i], .low = testdata.test_low[i], .close = 0, .volume = 0 };
        const out = ind.updateBar(&bar);
        try std.testing.expect(math.isNan(out.slice()[0].scalar.value));
    }

    // Index 14: first valid.
    const bar14 = Bar{ .time = 0, .open = 0, .high = testdata.test_high[14], .low = testdata.test_low[14], .close = 0, .volume = 0 };
    const out = ind.updateBar(&bar14);
    try std.testing.expect(almostEqual(out.slice()[0].scalar.value, testdata.test_expected[14].up, tolerance));
    try std.testing.expect(almostEqual(out.slice()[1].scalar.value, testdata.test_expected[14].down, tolerance));
    try std.testing.expect(almostEqual(out.slice()[2].scalar.value, testdata.test_expected[14].osc, tolerance));
}
