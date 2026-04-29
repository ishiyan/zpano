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

const test_high = [252]f64{
    93.25,  94.94,  96.375,  96.19,   96.0,    94.72,  95.0,   93.72,   92.47,   92.75,
    96.25,  99.625, 99.125,  92.75,   91.315,  93.25,  93.405, 90.655,  91.97,   92.25,
    90.345, 88.5,   88.25,   85.5,    84.44,   84.75,  84.44,  89.405,  88.125,  89.125,
    87.155, 87.25,  87.375,  88.97,   90.0,    89.845, 86.97,  85.94,   84.75,   85.47,
    84.47,  88.5,   89.47,   90.0,    92.44,   91.44,  92.97,  91.72,   91.155,  91.75,
    90.0,   88.875, 89.0,    85.25,   83.815,  85.25,  86.625, 87.94,   89.375,  90.625,
    90.75,  88.845, 91.97,   93.375,  93.815,  94.03,  94.03,  91.815,  92.0,    91.94,
    89.75,  88.75,  86.155,  84.875,  85.94,   99.375, 103.28, 105.375, 107.625, 105.25,
    104.5,  105.5,  106.125, 107.94,  106.25,  107.0,  108.75, 110.94,  110.94,  114.22,
    123.0,  121.75, 119.815, 120.315, 119.375, 118.19, 116.69, 115.345, 113.0,   118.315,
    116.87, 116.75, 113.87,  114.62,  115.31,  116.0,  121.69, 119.87,  120.87,  116.75,
    116.5,  116.0,  118.31,  121.5,   122.0,   121.44, 125.75, 127.75,  124.19,  124.44,
    125.75, 124.69, 125.31,  132.0,   131.31,  132.25, 133.88, 133.5,   135.5,   137.44,
    138.69, 139.19, 138.5,   138.13,  137.5,   138.88, 132.13, 129.75,  128.5,   125.44,
    125.12, 126.5,  128.69,  126.62,  126.69,  126.0,  123.12, 121.87,  124.0,   127.0,
    124.44, 122.5,  123.75,  123.81,  124.5,   127.87, 128.56, 129.63,  124.87,  124.37,
    124.87, 123.62, 124.06,  125.87,  125.19,  125.62, 126.0,  128.5,   126.75,  129.75,
    132.69, 133.94, 136.5,   137.69,  135.56,  133.56, 135.0,  132.38,  131.44,  130.88,
    129.63, 127.25, 127.81,  125.0,   126.81,  124.75, 122.81, 122.25,  121.06,  120.0,
    123.25, 122.75, 119.19,  115.06,  116.69,  114.87, 110.87, 107.25,  108.87,  109.0,
    108.5,  113.06, 93.0,    94.62,   95.12,   96.0,   95.56,  95.31,   99.0,    98.81,
    96.81,  95.94,  94.44,   92.94,   93.94,   95.5,   97.06,  97.5,    96.25,   96.37,
    95.0,   94.87,  98.25,   105.12,  108.44,  109.87, 105.0,  106.0,   104.94,  104.5,
    104.44, 106.31, 112.87,  116.5,   119.19,  121.0,  122.12, 111.94,  112.75,  110.19,
    107.94, 109.69, 111.06,  110.44,  110.12,  110.31, 110.44, 110.0,   110.75,  110.5,
    110.5,  109.5,
};

const test_low = [252]f64{
    90.75,  91.405, 94.25,   93.5,   92.815,  93.5,   92.0,    89.75,   89.44,  90.625,
    92.75,  96.315, 96.03,   88.815, 86.75,   90.94,  88.905,  88.78,   89.25,  89.75,
    87.5,   86.53,  84.625,  82.28,  81.565,  80.875, 81.25,   84.065,  85.595, 85.97,
    84.405, 85.095, 85.5,    85.53,  87.875,  86.565, 84.655,  83.25,   82.565, 83.44,
    82.53,  85.065, 86.875,  88.53,  89.28,   90.125, 90.75,   89.0,    88.565, 90.095,
    89.0,   86.47,  84.0,    83.315, 82.0,    83.25,  84.75,   85.28,   87.19,  88.44,
    88.25,  87.345, 89.28,   91.095, 89.53,   91.155, 92.0,    90.53,   89.97,  88.815,
    86.75,  85.065, 82.03,   81.5,   82.565,  96.345, 96.47,   101.155, 104.25, 101.75,
    101.72, 101.72, 103.155, 105.69, 103.655, 104.0,  105.53,  108.53,  108.75, 107.75,
    117.0,  118.0,  116.0,   118.5,  116.53,  116.25, 114.595, 110.875, 110.5,  110.72,
    112.62, 114.19, 111.19,  109.44, 111.56,  112.44, 117.5,   116.06,  116.56, 113.31,
    112.56, 114.0,  114.75,  118.87, 119.0,   119.75, 122.62,  123.0,   121.75, 121.56,
    123.12, 122.19, 122.75,  124.37, 128.0,   129.5,  130.81,  130.63,  132.13, 133.88,
    135.38, 135.75, 136.19,  134.5,  135.38,  133.69, 126.06,  126.87,  123.5,  122.62,
    122.75, 123.56, 125.81,  124.62, 124.37,  121.81, 118.19,  118.06,  117.56, 121.0,
    121.12, 118.94, 119.81,  121.0,  122.0,   124.5,  126.56,  123.5,   121.25, 121.06,
    122.31, 121.0,  120.87,  122.06, 122.75,  122.69, 122.87,  125.5,   124.25, 128.0,
    128.38, 130.69, 131.63,  134.38, 132.0,   131.94, 131.94,  129.56,  123.75, 126.0,
    126.25, 124.37, 121.44,  120.44, 121.37,  121.69, 120.0,   119.62,  115.5,  116.75,
    119.06, 119.06, 115.06,  111.06, 113.12,  110.0,  105.0,   104.69,  103.87, 104.69,
    105.44, 107.0,  89.0,    92.5,   92.12,   94.62,  92.81,   94.25,   96.25,  96.37,
    93.69,  93.5,   90.0,    90.19,  90.5,    92.12,  94.12,   94.87,   93.0,   93.87,
    93.0,   92.62,  93.56,   98.37,  104.44,  106.0,  101.81,  104.12,  103.37, 102.12,
    102.25, 103.37, 107.94,  112.5,  115.44,  115.5,  112.25,  107.56,  106.56, 106.87,
    104.5,  105.75, 108.62,  107.75, 108.06,  108.0,  108.19,  108.12,  109.06, 108.75,
    108.56, 106.62,
};

const ExpectedRow = struct { up: f64, down: f64, osc: f64 };

const test_expected = blk: {
    const nan = math.nan(f64);
    break :blk [252]ExpectedRow{
        .{ .up = nan, .down = nan, .osc = nan }, // 0
        .{ .up = nan, .down = nan, .osc = nan }, // 1
        .{ .up = nan, .down = nan, .osc = nan }, // 2
        .{ .up = nan, .down = nan, .osc = nan }, // 3
        .{ .up = nan, .down = nan, .osc = nan }, // 4
        .{ .up = nan, .down = nan, .osc = nan }, // 5
        .{ .up = nan, .down = nan, .osc = nan }, // 6
        .{ .up = nan, .down = nan, .osc = nan }, // 7
        .{ .up = nan, .down = nan, .osc = nan }, // 8
        .{ .up = nan, .down = nan, .osc = nan }, // 9
        .{ .up = nan, .down = nan, .osc = nan }, // 10
        .{ .up = nan, .down = nan, .osc = nan }, // 11
        .{ .up = nan, .down = nan, .osc = nan }, // 12
        .{ .up = nan, .down = nan, .osc = nan }, // 13
        .{ .up = 78.571428571428600, .down = 100.000000000000000, .osc = -21.428571428571400 },
        .{ .up = 71.428571428571400, .down = 92.857142857142900, .osc = -21.428571428571400 },
        .{ .up = 64.285714285714300, .down = 85.714285714285700, .osc = -21.428571428571400 },
        .{ .up = 57.142857142857100, .down = 78.571428571428600, .osc = -21.428571428571400 },
        .{ .up = 50.000000000000000, .down = 71.428571428571400, .osc = -21.428571428571400 },
        .{ .up = 42.857142857142900, .down = 64.285714285714300, .osc = -21.428571428571400 },
        .{ .up = 35.714285714285700, .down = 57.142857142857100, .osc = -21.428571428571400 },
        .{ .up = 28.571428571428600, .down = 100.000000000000000, .osc = -71.428571428571400 },
        .{ .up = 21.428571428571400, .down = 100.000000000000000, .osc = -78.571428571428600 },
        .{ .up = 14.285714285714300, .down = 100.000000000000000, .osc = -85.714285714285700 },
        .{ .up = 7.142857142857140, .down = 100.000000000000000, .osc = -92.857142857142900 },
        .{ .up = 0.000000000000000, .down = 100.000000000000000, .osc = -100.000000000000000 },
        .{ .up = 0.000000000000000, .down = 92.857142857142900, .osc = -92.857142857142900 },
        .{ .up = 21.428571428571400, .down = 85.714285714285700, .osc = -64.285714285714300 },
        .{ .up = 14.285714285714300, .down = 78.571428571428600, .osc = -64.285714285714300 },
        .{ .up = 7.142857142857140, .down = 71.428571428571400, .osc = -64.285714285714300 },
        .{ .up = 0.000000000000000, .down = 64.285714285714300, .osc = -64.285714285714300 },
        .{ .up = 14.285714285714300, .down = 57.142857142857100, .osc = -42.857142857142900 },
        .{ .up = 7.142857142857140, .down = 50.000000000000000, .osc = -42.857142857142900 },
        .{ .up = 0.000000000000000, .down = 42.857142857142900, .osc = -42.857142857142900 },
        .{ .up = 0.000000000000000, .down = 35.714285714285700, .osc = -35.714285714285700 },
        .{ .up = 92.857142857142900, .down = 28.571428571428600, .osc = 64.285714285714300 },
        .{ .up = 85.714285714285700, .down = 21.428571428571400, .osc = 64.285714285714300 },
        .{ .up = 78.571428571428600, .down = 14.285714285714300, .osc = 64.285714285714300 },
        .{ .up = 71.428571428571400, .down = 7.142857142857140, .osc = 64.285714285714300 },
        .{ .up = 64.285714285714300, .down = 0.000000000000000, .osc = 64.285714285714300 },
        .{ .up = 57.142857142857100, .down = 0.000000000000000, .osc = 57.142857142857100 },
        .{ .up = 50.000000000000000, .down = 92.857142857142900, .osc = -42.857142857142900 },
        .{ .up = 42.857142857142900, .down = 85.714285714285700, .osc = -42.857142857142900 },
        .{ .up = 100.000000000000000, .down = 78.571428571428600, .osc = 21.428571428571400 },
        .{ .up = 100.000000000000000, .down = 71.428571428571400, .osc = 28.571428571428600 },
        .{ .up = 92.857142857142900, .down = 64.285714285714300, .osc = 28.571428571428600 },
        .{ .up = 100.000000000000000, .down = 57.142857142857100, .osc = 42.857142857142900 },
        .{ .up = 92.857142857142900, .down = 50.000000000000000, .osc = 42.857142857142900 },
        .{ .up = 85.714285714285700, .down = 42.857142857142900, .osc = 42.857142857142900 },
        .{ .up = 78.571428571428600, .down = 35.714285714285700, .osc = 42.857142857142900 },
        .{ .up = 71.428571428571400, .down = 28.571428571428600, .osc = 42.857142857142900 },
        .{ .up = 64.285714285714300, .down = 21.428571428571400, .osc = 42.857142857142900 },
        .{ .up = 57.142857142857100, .down = 14.285714285714300, .osc = 42.857142857142900 },
        .{ .up = 50.000000000000000, .down = 7.142857142857140, .osc = 42.857142857142900 },
        .{ .up = 42.857142857142900, .down = 100.000000000000000, .osc = -57.142857142857100 },
        .{ .up = 35.714285714285700, .down = 92.857142857142900, .osc = -57.142857142857100 },
        .{ .up = 28.571428571428600, .down = 85.714285714285700, .osc = -57.142857142857100 },
        .{ .up = 21.428571428571400, .down = 78.571428571428600, .osc = -57.142857142857100 },
        .{ .up = 14.285714285714300, .down = 71.428571428571400, .osc = -57.142857142857100 },
        .{ .up = 7.142857142857140, .down = 64.285714285714300, .osc = -57.142857142857200 },
        .{ .up = 0.000000000000000, .down = 57.142857142857100, .osc = -57.142857142857100 },
        .{ .up = 14.285714285714300, .down = 50.000000000000000, .osc = -35.714285714285700 },
        .{ .up = 100.000000000000000, .down = 42.857142857142900, .osc = 57.142857142857100 },
        .{ .up = 100.000000000000000, .down = 35.714285714285700, .osc = 64.285714285714300 },
        .{ .up = 100.000000000000000, .down = 28.571428571428600, .osc = 71.428571428571400 },
        .{ .up = 100.000000000000000, .down = 21.428571428571400, .osc = 78.571428571428600 },
        .{ .up = 100.000000000000000, .down = 14.285714285714300, .osc = 85.714285714285700 },
        .{ .up = 92.857142857142900, .down = 7.142857142857140, .osc = 85.714285714285700 },
        .{ .up = 85.714285714285700, .down = 0.000000000000000, .osc = 85.714285714285700 },
        .{ .up = 78.571428571428600, .down = 0.000000000000000, .osc = 78.571428571428600 },
        .{ .up = 71.428571428571400, .down = 0.000000000000000, .osc = 71.428571428571400 },
        .{ .up = 64.285714285714300, .down = 100.000000000000000, .osc = -35.714285714285700 },
        .{ .up = 57.142857142857100, .down = 100.000000000000000, .osc = -42.857142857142900 },
        .{ .up = 50.000000000000000, .down = 100.000000000000000, .osc = -50.000000000000000 },
        .{ .up = 42.857142857142900, .down = 92.857142857142900, .osc = -50.000000000000000 },
        .{ .up = 100.000000000000000, .down = 85.714285714285700, .osc = 14.285714285714300 },
        .{ .up = 100.000000000000000, .down = 78.571428571428600, .osc = 21.428571428571400 },
        .{ .up = 100.000000000000000, .down = 71.428571428571400, .osc = 28.571428571428600 },
        .{ .up = 100.000000000000000, .down = 64.285714285714300, .osc = 35.714285714285700 },
        .{ .up = 92.857142857142900, .down = 57.142857142857100, .osc = 35.714285714285700 },
        .{ .up = 85.714285714285700, .down = 50.000000000000000, .osc = 35.714285714285700 },
        .{ .up = 78.571428571428600, .down = 42.857142857142900, .osc = 35.714285714285700 },
        .{ .up = 71.428571428571400, .down = 35.714285714285700, .osc = 35.714285714285700 },
        .{ .up = 100.000000000000000, .down = 28.571428571428600, .osc = 71.428571428571400 },
        .{ .up = 92.857142857142900, .down = 21.428571428571400, .osc = 71.428571428571400 },
        .{ .up = 85.714285714285700, .down = 14.285714285714300, .osc = 71.428571428571400 },
        .{ .up = 100.000000000000000, .down = 7.142857142857140, .osc = 92.857142857142900 },
        .{ .up = 100.000000000000000, .down = 0.000000000000000, .osc = 100.000000000000000 },
        .{ .up = 100.000000000000000, .down = 0.000000000000000, .osc = 100.000000000000000 },
        .{ .up = 100.000000000000000, .down = 0.000000000000000, .osc = 100.000000000000000 },
        .{ .up = 100.000000000000000, .down = 0.000000000000000, .osc = 100.000000000000000 },
        .{ .up = 92.857142857142900, .down = 0.000000000000000, .osc = 92.857142857142900 },
        .{ .up = 85.714285714285700, .down = 21.428571428571400, .osc = 64.285714285714300 },
        .{ .up = 78.571428571428600, .down = 14.285714285714300, .osc = 64.285714285714300 },
        .{ .up = 71.428571428571400, .down = 7.142857142857140, .osc = 64.285714285714300 },
        .{ .up = 64.285714285714300, .down = 0.000000000000000, .osc = 64.285714285714300 },
        .{ .up = 57.142857142857100, .down = 0.000000000000000, .osc = 57.142857142857100 },
        .{ .up = 50.000000000000000, .down = 7.142857142857140, .osc = 42.857142857142900 },
        .{ .up = 42.857142857142900, .down = 0.000000000000000, .osc = 42.857142857142900 },
        .{ .up = 35.714285714285700, .down = 0.000000000000000, .osc = 35.714285714285700 },
        .{ .up = 28.571428571428600, .down = 0.000000000000000, .osc = 28.571428571428600 },
        .{ .up = 21.428571428571400, .down = 14.285714285714300, .osc = 7.142857142857140 },
        .{ .up = 14.285714285714300, .down = 7.142857142857140, .osc = 7.142857142857140 },
        .{ .up = 7.142857142857140, .down = 0.000000000000000, .osc = 7.142857142857140 },
        .{ .up = 0.000000000000000, .down = 92.857142857142900, .osc = -92.857142857142900 },
        .{ .up = 0.000000000000000, .down = 85.714285714285700, .osc = -85.714285714285700 },
        .{ .up = 100.000000000000000, .down = 78.571428571428600, .osc = 21.428571428571400 },
        .{ .up = 92.857142857142900, .down = 71.428571428571400, .osc = 21.428571428571400 },
        .{ .up = 85.714285714285700, .down = 64.285714285714300, .osc = 21.428571428571400 },
        .{ .up = 78.571428571428600, .down = 57.142857142857100, .osc = 21.428571428571400 },
        .{ .up = 71.428571428571400, .down = 50.000000000000000, .osc = 21.428571428571400 },
        .{ .up = 64.285714285714300, .down = 42.857142857142900, .osc = 21.428571428571400 },
        .{ .up = 57.142857142857100, .down = 35.714285714285700, .osc = 21.428571428571400 },
        .{ .up = 50.000000000000000, .down = 28.571428571428600, .osc = 21.428571428571400 },
        .{ .up = 100.000000000000000, .down = 21.428571428571400, .osc = 78.571428571428600 },
        .{ .up = 92.857142857142900, .down = 14.285714285714300, .osc = 78.571428571428600 },
        .{ .up = 100.000000000000000, .down = 7.142857142857140, .osc = 92.857142857142900 },
        .{ .up = 100.000000000000000, .down = 0.000000000000000, .osc = 100.000000000000000 },
        .{ .up = 92.857142857142900, .down = 0.000000000000000, .osc = 92.857142857142900 },
        .{ .up = 85.714285714285700, .down = 0.000000000000000, .osc = 85.714285714285700 },
        .{ .up = 78.571428571428600, .down = 28.571428571428600, .osc = 50.000000000000000 },
        .{ .up = 71.428571428571400, .down = 21.428571428571400, .osc = 50.000000000000000 },
        .{ .up = 64.285714285714300, .down = 14.285714285714300, .osc = 50.000000000000000 },
        .{ .up = 100.000000000000000, .down = 7.142857142857140, .osc = 92.857142857142900 },
        .{ .up = 92.857142857142900, .down = 0.000000000000000, .osc = 92.857142857142900 },
        .{ .up = 100.000000000000000, .down = 0.000000000000000, .osc = 100.000000000000000 },
        .{ .up = 100.000000000000000, .down = 0.000000000000000, .osc = 100.000000000000000 },
        .{ .up = 92.857142857142900, .down = 0.000000000000000, .osc = 92.857142857142900 },
        .{ .up = 100.000000000000000, .down = 0.000000000000000, .osc = 100.000000000000000 },
        .{ .up = 100.000000000000000, .down = 0.000000000000000, .osc = 100.000000000000000 },
        .{ .up = 100.000000000000000, .down = 21.428571428571400, .osc = 78.571428571428600 },
        .{ .up = 100.000000000000000, .down = 14.285714285714300, .osc = 85.714285714285700 },
        .{ .up = 92.857142857142900, .down = 7.142857142857140, .osc = 85.714285714285700 },
        .{ .up = 85.714285714285700, .down = 0.000000000000000, .osc = 85.714285714285700 },
        .{ .up = 78.571428571428600, .down = 7.142857142857140, .osc = 71.428571428571400 },
        .{ .up = 71.428571428571400, .down = 0.000000000000000, .osc = 71.428571428571400 },
        .{ .up = 64.285714285714300, .down = 0.000000000000000, .osc = 64.285714285714300 },
        .{ .up = 57.142857142857100, .down = 0.000000000000000, .osc = 57.142857142857100 },
        .{ .up = 50.000000000000000, .down = 100.000000000000000, .osc = -50.000000000000000 },
        .{ .up = 42.857142857142900, .down = 100.000000000000000, .osc = -57.142857142857100 },
        .{ .up = 35.714285714285700, .down = 92.857142857142900, .osc = -57.142857142857100 },
        .{ .up = 28.571428571428600, .down = 85.714285714285700, .osc = -57.142857142857100 },
        .{ .up = 21.428571428571400, .down = 78.571428571428600, .osc = -57.142857142857100 },
        .{ .up = 14.285714285714300, .down = 71.428571428571400, .osc = -57.142857142857100 },
        .{ .up = 7.142857142857140, .down = 64.285714285714300, .osc = -57.142857142857200 },
        .{ .up = 0.000000000000000, .down = 100.000000000000000, .osc = -100.000000000000000 },
        .{ .up = 21.428571428571400, .down = 100.000000000000000, .osc = -78.571428571428600 },
        .{ .up = 14.285714285714300, .down = 100.000000000000000, .osc = -85.714285714285700 },
        .{ .up = 7.142857142857140, .down = 100.000000000000000, .osc = -92.857142857142900 },
        .{ .up = 0.000000000000000, .down = 92.857142857142900, .osc = -92.857142857142900 },
        .{ .up = 0.000000000000000, .down = 85.714285714285700, .osc = -85.714285714285700 },
        .{ .up = 0.000000000000000, .down = 78.571428571428600, .osc = -78.571428571428600 },
        .{ .up = 28.571428571428600, .down = 71.428571428571400, .osc = -42.857142857142900 },
        .{ .up = 21.428571428571400, .down = 64.285714285714300, .osc = -42.857142857142900 },
        .{ .up = 14.285714285714300, .down = 57.142857142857100, .osc = -42.857142857142900 },
        .{ .up = 7.142857142857140, .down = 50.000000000000000, .osc = -42.857142857142900 },
        .{ .up = 0.000000000000000, .down = 42.857142857142900, .osc = -42.857142857142900 },
        .{ .up = 100.000000000000000, .down = 35.714285714285700, .osc = 64.285714285714300 },
        .{ .up = 92.857142857142900, .down = 28.571428571428600, .osc = 64.285714285714300 },
        .{ .up = 85.714285714285700, .down = 21.428571428571400, .osc = 64.285714285714300 },
        .{ .up = 78.571428571428600, .down = 14.285714285714300, .osc = 64.285714285714300 },
        .{ .up = 71.428571428571400, .down = 7.142857142857140, .osc = 64.285714285714300 },
        .{ .up = 64.285714285714300, .down = 0.000000000000000, .osc = 64.285714285714300 },
        .{ .up = 57.142857142857100, .down = 14.285714285714300, .osc = 42.857142857142900 },
        .{ .up = 50.000000000000000, .down = 7.142857142857140, .osc = 42.857142857142900 },
        .{ .up = 42.857142857142900, .down = 0.000000000000000, .osc = 42.857142857142900 },
        .{ .up = 35.714285714285700, .down = 0.000000000000000, .osc = 35.714285714285700 },
        .{ .up = 28.571428571428600, .down = 64.285714285714300, .osc = -35.714285714285700 },
        .{ .up = 21.428571428571400, .down = 57.142857142857100, .osc = -35.714285714285700 },
        .{ .up = 100.000000000000000, .down = 50.000000000000000, .osc = 50.000000000000000 },
        .{ .up = 100.000000000000000, .down = 42.857142857142900, .osc = 57.142857142857100 },
        .{ .up = 100.000000000000000, .down = 35.714285714285700, .osc = 64.285714285714300 },
        .{ .up = 100.000000000000000, .down = 28.571428571428600, .osc = 71.428571428571400 },
        .{ .up = 100.000000000000000, .down = 21.428571428571400, .osc = 78.571428571428600 },
        .{ .up = 92.857142857142900, .down = 14.285714285714300, .osc = 78.571428571428600 },
        .{ .up = 85.714285714285700, .down = 7.142857142857140, .osc = 78.571428571428600 },
        .{ .up = 78.571428571428600, .down = 0.000000000000000, .osc = 78.571428571428600 },
        .{ .up = 71.428571428571400, .down = 0.000000000000000, .osc = 71.428571428571400 },
        .{ .up = 64.285714285714300, .down = 7.142857142857140, .osc = 57.142857142857200 },
        .{ .up = 57.142857142857100, .down = 0.000000000000000, .osc = 57.142857142857100 },
        .{ .up = 50.000000000000000, .down = 0.000000000000000, .osc = 50.000000000000000 },
        .{ .up = 42.857142857142900, .down = 78.571428571428600, .osc = -35.714285714285700 },
        .{ .up = 35.714285714285700, .down = 100.000000000000000, .osc = -64.285714285714300 },
        .{ .up = 28.571428571428600, .down = 100.000000000000000, .osc = -71.428571428571400 },
        .{ .up = 21.428571428571400, .down = 92.857142857142900, .osc = -71.428571428571400 },
        .{ .up = 14.285714285714300, .down = 85.714285714285700, .osc = -71.428571428571400 },
        .{ .up = 7.142857142857140, .down = 100.000000000000000, .osc = -92.857142857142900 },
        .{ .up = 0.000000000000000, .down = 100.000000000000000, .osc = -100.000000000000000 },
        .{ .up = 0.000000000000000, .down = 100.000000000000000, .osc = -100.000000000000000 },
        .{ .up = 7.142857142857140, .down = 92.857142857142900, .osc = -85.714285714285700 },
        .{ .up = 0.000000000000000, .down = 85.714285714285700, .osc = -85.714285714285700 },
        .{ .up = 0.000000000000000, .down = 78.571428571428600, .osc = -78.571428571428600 },
        .{ .up = 0.000000000000000, .down = 100.000000000000000, .osc = -100.000000000000000 },
        .{ .up = 0.000000000000000, .down = 100.000000000000000, .osc = -100.000000000000000 },
        .{ .up = 0.000000000000000, .down = 92.857142857142900, .osc = -92.857142857142900 },
        .{ .up = 7.142857142857140, .down = 100.000000000000000, .osc = -92.857142857142900 },
        .{ .up = 0.000000000000000, .down = 100.000000000000000, .osc = -100.000000000000000 },
        .{ .up = 7.142857142857140, .down = 100.000000000000000, .osc = -92.857142857142900 },
        .{ .up = 0.000000000000000, .down = 100.000000000000000, .osc = -100.000000000000000 },
        .{ .up = 0.000000000000000, .down = 92.857142857142900, .osc = -92.857142857142900 },
        .{ .up = 28.571428571428600, .down = 85.714285714285700, .osc = -57.142857142857100 },
        .{ .up = 21.428571428571400, .down = 78.571428571428600, .osc = -57.142857142857100 },
        .{ .up = 14.285714285714300, .down = 100.000000000000000, .osc = -85.714285714285700 },
        .{ .up = 7.142857142857140, .down = 92.857142857142900, .osc = -85.714285714285700 },
        .{ .up = 0.000000000000000, .down = 85.714285714285700, .osc = -85.714285714285700 },
        .{ .up = 0.000000000000000, .down = 78.571428571428600, .osc = -78.571428571428600 },
        .{ .up = 0.000000000000000, .down = 71.428571428571400, .osc = -71.428571428571400 },
        .{ .up = 7.142857142857140, .down = 64.285714285714300, .osc = -57.142857142857200 },
        .{ .up = 0.000000000000000, .down = 57.142857142857100, .osc = -57.142857142857100 },
        .{ .up = 0.000000000000000, .down = 50.000000000000000, .osc = -50.000000000000000 },
        .{ .up = 35.714285714285700, .down = 42.857142857142900, .osc = -7.142857142857140 },
        .{ .up = 28.571428571428600, .down = 35.714285714285700, .osc = -7.142857142857150 },
        .{ .up = 21.428571428571400, .down = 28.571428571428600, .osc = -7.142857142857140 },
        .{ .up = 14.285714285714300, .down = 21.428571428571400, .osc = -7.142857142857140 },
        .{ .up = 7.142857142857140, .down = 14.285714285714300, .osc = -7.142857142857140 },
        .{ .up = 0.000000000000000, .down = 7.142857142857140, .osc = -7.142857142857140 },
        .{ .up = 42.857142857142900, .down = 0.000000000000000, .osc = 42.857142857142900 },
        .{ .up = 35.714285714285700, .down = 64.285714285714300, .osc = -28.571428571428600 },
        .{ .up = 28.571428571428600, .down = 57.142857142857100, .osc = -28.571428571428600 },
        .{ .up = 21.428571428571400, .down = 50.000000000000000, .osc = -28.571428571428600 },
        .{ .up = 14.285714285714300, .down = 42.857142857142900, .osc = -28.571428571428600 },
        .{ .up = 7.142857142857140, .down = 35.714285714285700, .osc = -28.571428571428600 },
        .{ .up = 0.000000000000000, .down = 28.571428571428600, .osc = -28.571428571428600 },
        .{ .up = 100.000000000000000, .down = 21.428571428571400, .osc = 78.571428571428600 },
        .{ .up = 100.000000000000000, .down = 14.285714285714300, .osc = 85.714285714285700 },
        .{ .up = 100.000000000000000, .down = 7.142857142857140, .osc = 92.857142857142900 },
        .{ .up = 92.857142857142900, .down = 0.000000000000000, .osc = 92.857142857142900 },
        .{ .up = 85.714285714285700, .down = 0.000000000000000, .osc = 85.714285714285700 },
        .{ .up = 78.571428571428600, .down = 0.000000000000000, .osc = 78.571428571428600 },
        .{ .up = 71.428571428571400, .down = 0.000000000000000, .osc = 71.428571428571400 },
        .{ .up = 64.285714285714300, .down = 35.714285714285700, .osc = 28.571428571428600 },
        .{ .up = 57.142857142857100, .down = 28.571428571428600, .osc = 28.571428571428600 },
        .{ .up = 100.000000000000000, .down = 21.428571428571400, .osc = 78.571428571428600 },
        .{ .up = 100.000000000000000, .down = 14.285714285714300, .osc = 85.714285714285700 },
        .{ .up = 100.000000000000000, .down = 7.142857142857140, .osc = 92.857142857142900 },
        .{ .up = 100.000000000000000, .down = 0.000000000000000, .osc = 100.000000000000000 },
        .{ .up = 100.000000000000000, .down = 0.000000000000000, .osc = 100.000000000000000 },
        .{ .up = 92.857142857142900, .down = 0.000000000000000, .osc = 92.857142857142900 },
        .{ .up = 85.714285714285700, .down = 14.285714285714300, .osc = 71.428571428571400 },
        .{ .up = 78.571428571428600, .down = 7.142857142857140, .osc = 71.428571428571400 },
        .{ .up = 71.428571428571400, .down = 0.000000000000000, .osc = 71.428571428571400 },
        .{ .up = 64.285714285714300, .down = 14.285714285714300, .osc = 50.000000000000000 },
        .{ .up = 57.142857142857100, .down = 7.142857142857140, .osc = 50.000000000000000 },
        .{ .up = 50.000000000000000, .down = 0.000000000000000, .osc = 50.000000000000000 },
        .{ .up = 42.857142857142900, .down = 0.000000000000000, .osc = 42.857142857142900 },
        .{ .up = 35.714285714285700, .down = 0.000000000000000, .osc = 35.714285714285700 },
        .{ .up = 28.571428571428600, .down = 57.142857142857100, .osc = -28.571428571428600 },
        .{ .up = 21.428571428571400, .down = 50.000000000000000, .osc = -28.571428571428600 },
        .{ .up = 14.285714285714300, .down = 42.857142857142900, .osc = -28.571428571428600 },
        .{ .up = 7.142857142857140, .down = 35.714285714285700, .osc = -28.571428571428600 },
        .{ .up = 0.000000000000000, .down = 28.571428571428600, .osc = -28.571428571428600 },
        .{ .up = 7.142857142857140, .down = 21.428571428571400, .osc = -14.285714285714300 },
    };
};

test "Aroon length=14 full data" {
    const allocator = std.testing.allocator;
    var ind = try Aroon.init(allocator, .{ .length = 14 });
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateHighLow(test_high[i], test_low[i]);
        const exp = test_expected[i];

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
        _ = ind.updateHighLow(test_high[i], test_low[i]);
        try std.testing.expect(!ind.isPrimed());
    }

    _ = ind.updateHighLow(test_high[14], test_low[14]);
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
        const bar = Bar{ .time = 0, .open = 0, .high = test_high[i], .low = test_low[i], .close = 0, .volume = 0 };
        const out = ind.updateBar(&bar);
        try std.testing.expect(math.isNan(out.slice()[0].scalar.value));
    }

    // Index 14: first valid.
    const bar14 = Bar{ .time = 0, .open = 0, .high = test_high[14], .low = test_low[14], .close = 0, .volume = 0 };
    const out = ind.updateBar(&bar14);
    try std.testing.expect(almostEqual(out.slice()[0].scalar.value, test_expected[14].up, tolerance));
    try std.testing.expect(almostEqual(out.slice()[1].scalar.value, test_expected[14].down, tolerance));
    try std.testing.expect(almostEqual(out.slice()[2].scalar.value, test_expected[14].osc, tolerance));
}
