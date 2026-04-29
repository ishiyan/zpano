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

const sma_mod = @import("../../common/simple_moving_average/simple_moving_average.zig");
const ema_mod = @import("../../common/exponential_moving_average/exponential_moving_average.zig");

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the Stochastic Oscillator.
pub const StochasticOutput = enum(u8) {
    /// The Fast-K line (raw stochastic).
    fast_k = 1,
    /// The Slow-K line (smoothed Fast-K).
    slow_k = 2,
    /// The Slow-D line (smoothed Slow-K).
    slow_d = 3,
};

/// Specifies the type of moving average used for smoothing.
pub const MovingAverageType = enum(u8) {
    sma = 0,
    ema = 1,
};

/// Parameters to create a Stochastic Oscillator.
pub const StochasticParams = struct {
    /// Lookback period for raw %K. Must be >= 1. Default 5.
    fast_k_length: usize = 5,
    /// Smoothing period for Slow-K. Must be >= 1. Default 3.
    slow_k_length: usize = 3,
    /// Smoothing period for Slow-D. Must be >= 1. Default 3.
    slow_d_length: usize = 3,
    /// MA type for Slow-K smoothing.
    slow_k_ma_type: MovingAverageType = .sma,
    /// MA type for Slow-D smoothing.
    slow_d_ma_type: MovingAverageType = .sma,
    /// When true, EMA is seeded with average of first period values.
    first_is_average: bool = false,
};

/// A line smoother: either SMA, EMA, or passthrough (length 1).
const Smoother = union(enum) {
    sma: sma_mod.SimpleMovingAverage,
    ema: ema_mod.ExponentialMovingAverage,
    passthrough: void,

    fn update(self: *Smoother, v: f64) f64 {
        return switch (self.*) {
            .sma => |*s| s.update(v),
            .ema => |*e| e.update(v),
            .passthrough => v,
        };
    }

    fn isPrimed(self: *const Smoother) bool {
        return switch (self.*) {
            .sma => |*s| s.isPrimed(),
            .ema => |*e| e.isPrimed(),
            .passthrough => true,
        };
    }

    fn deinit(self: *Smoother) void {
        switch (self.*) {
            .sma => |*s| s.deinit(),
            .ema, .passthrough => {},
        }
    }
};

/// George Lane's Stochastic Oscillator.
///
/// Produces three outputs: Fast-K, Slow-K, and Slow-D.
/// Requires bar data (high, low, close). For scalar/quote/trade updates,
/// the single value substitutes for all three.
pub const Stochastic = struct {
    fast_k_length: usize,
    high_buf: []f64,
    low_buf: []f64,
    buffer_index: usize,
    count: usize,

    slow_k_ma: Smoother,
    slow_d_ma: Smoother,

    fast_k: f64,
    slow_k: f64,
    slow_d: f64,
    primed: bool,

    allocator: std.mem.Allocator,
    mnemonic_buf: [96]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub const Error = error{
        InvalidFastKLength,
        InvalidSlowKLength,
        InvalidSlowDLength,
        OutOfMemory,
        MnemonicTooLong,
    };

    pub fn init(allocator: std.mem.Allocator, params: StochasticParams) Error!Stochastic {
        if (params.fast_k_length < 1) return error.InvalidFastKLength;
        if (params.slow_k_length < 1) return error.InvalidSlowKLength;
        if (params.slow_d_length < 1) return error.InvalidSlowDLength;

        var slow_k_ma = try createSmoother(allocator, params.slow_k_ma_type, params.slow_k_length, params.first_is_average);
        errdefer slow_k_ma.deinit();

        var slow_d_ma = try createSmoother(allocator, params.slow_d_ma_type, params.slow_d_length, params.first_is_average);
        errdefer slow_d_ma.deinit();

        const slow_k_label = maLabel(params.slow_k_ma_type, params.slow_k_length);
        const slow_d_label = maLabel(params.slow_d_ma_type, params.slow_d_length);

        var mnemonic_buf: [96]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "stoch({d}/{s}{d}/{s}{d})", .{
            params.fast_k_length, slow_k_label, params.slow_k_length, slow_d_label, params.slow_d_length,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Stochastic Oscillator {s}", .{
            mnemonic_buf[0..mnemonic_len],
        }) catch return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const high_buf = allocator.alloc(f64, params.fast_k_length) catch return error.OutOfMemory;
        errdefer allocator.free(high_buf);
        const low_buf = allocator.alloc(f64, params.fast_k_length) catch return error.OutOfMemory;

        return Stochastic{
            .fast_k_length = params.fast_k_length,
            .high_buf = high_buf,
            .low_buf = low_buf,
            .buffer_index = 0,
            .count = 0,
            .slow_k_ma = slow_k_ma,
            .slow_d_ma = slow_d_ma,
            .fast_k = math.nan(f64),
            .slow_k = math.nan(f64),
            .slow_d = math.nan(f64),
            .primed = false,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *Stochastic) void {
        self.allocator.free(self.high_buf);
        self.allocator.free(self.low_buf);
        self.slow_k_ma.deinit();
        self.slow_d_ma.deinit();
    }

    pub fn fixSlices(self: *Stochastic) void {
        // No-op: mnemonic/description live in embedded buffers, no slice fields to fix.
        _ = self;
    }

    /// Core update given close, high, low. Returns (fastK, slowK, slowD).
    pub fn update(self: *Stochastic, close: f64, high: f64, low: f64) struct { fast_k: f64, slow_k: f64, slow_d: f64 } {
        if (math.isNan(close) or math.isNan(high) or math.isNan(low)) {
            return .{ .fast_k = math.nan(f64), .slow_k = math.nan(f64), .slow_d = math.nan(f64) };
        }

        self.high_buf[self.buffer_index] = high;
        self.low_buf[self.buffer_index] = low;
        self.buffer_index = (self.buffer_index + 1) % self.fast_k_length;
        self.count += 1;

        if (self.count < self.fast_k_length) {
            return .{ .fast_k = self.fast_k, .slow_k = self.slow_k, .slow_d = self.slow_d };
        }

        // Find highest high and lowest low in the window.
        var hh = self.high_buf[0];
        var ll = self.low_buf[0];
        for (self.high_buf[1..], self.low_buf[1..]) |h, l| {
            if (h > hh) hh = h;
            if (l < ll) ll = l;
        }

        // Calculate Fast-K.
        const diff = hh - ll;
        if (diff > 0) {
            self.fast_k = 100.0 * (close - ll) / diff;
        } else {
            self.fast_k = 0;
        }

        // Feed Fast-K to Slow-K smoother.
        self.slow_k = self.slow_k_ma.update(self.fast_k);

        // Feed Slow-K to Slow-D smoother (only when Slow-K MA is primed).
        if (self.slow_k_ma.isPrimed()) {
            self.slow_d = self.slow_d_ma.update(self.slow_k);

            if (!self.primed and self.slow_d_ma.isPrimed()) {
                self.primed = true;
            }
        }

        return .{ .fast_k = self.fast_k, .slow_k = self.slow_k, .slow_d = self.slow_d };
    }

    pub fn isPrimed(self: *const Stochastic) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const Stochastic, out: *Metadata) void {
        const mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        const description = self.description_buf[0..self.description_len];

        // Build output mnemonic/description strings.
        var fk_mn_buf: [128]u8 = undefined;
        const fk_mn = std.fmt.bufPrint(&fk_mn_buf, "{s} fastK", .{mnemonic}) catch mnemonic;
        var sk_mn_buf: [128]u8 = undefined;
        const sk_mn = std.fmt.bufPrint(&sk_mn_buf, "{s} slowK", .{mnemonic}) catch mnemonic;
        var sd_mn_buf: [128]u8 = undefined;
        const sd_mn = std.fmt.bufPrint(&sd_mn_buf, "{s} slowD", .{mnemonic}) catch mnemonic;

        var fk_desc_buf: [160]u8 = undefined;
        const fk_desc = std.fmt.bufPrint(&fk_desc_buf, "{s} Fast-K", .{description}) catch description;
        var sk_desc_buf: [160]u8 = undefined;
        const sk_desc = std.fmt.bufPrint(&sk_desc_buf, "{s} Slow-K", .{description}) catch description;
        var sd_desc_buf: [160]u8 = undefined;
        const sd_desc = std.fmt.bufPrint(&sd_desc_buf, "{s} Slow-D", .{description}) catch description;

        build_metadata_mod.buildMetadata(out, Identifier.stochastic, mnemonic, description, &.{
            .{ .mnemonic = fk_mn, .description = fk_desc },
            .{ .mnemonic = sk_mn, .description = sk_desc },
            .{ .mnemonic = sd_mn, .description = sd_desc },
        });
    }

    fn makeOutput(self: *const Stochastic, time: i64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = Scalar{ .time = time, .value = self.fast_k } });
        out.append(.{ .scalar = Scalar{ .time = time, .value = self.slow_k } });
        out.append(.{ .scalar = Scalar{ .time = time, .value = self.slow_d } });
        return out;
    }

    pub fn updateScalar(self: *Stochastic, sample: *const Scalar) OutputArray {
        _ = self.update(sample.value, sample.value, sample.value);
        return self.makeOutput(sample.time);
    }

    pub fn updateBar(self: *Stochastic, sample: *const Bar) OutputArray {
        _ = self.update(sample.close, sample.high, sample.low);
        return self.makeOutput(sample.time);
    }

    pub fn updateQuote(self: *Stochastic, sample: *const Quote) OutputArray {
        const mid = (sample.bid_price + sample.ask_price) / 2.0;
        _ = self.update(mid, mid, mid);
        return self.makeOutput(sample.time);
    }

    pub fn updateTrade(self: *Stochastic, sample: *const Trade) OutputArray {
        _ = self.update(sample.price, sample.price, sample.price);
        return self.makeOutput(sample.time);
    }

    // --- Indicator interface ---

    pub fn indicator(self: *Stochastic) indicator_mod.Indicator {
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
        const self: *const Stochastic = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *const anyopaque, out: *Metadata) void {
        const self: *const Stochastic = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *Stochastic = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *Stochastic = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *Stochastic = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *Stochastic = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }
};

fn createSmoother(allocator: std.mem.Allocator, ma_type: MovingAverageType, length: usize, first_is_average: bool) Stochastic.Error!Smoother {
    if (length < 2) {
        return Smoother{ .passthrough = {} };
    }

    switch (ma_type) {
        .ema => {
            const ema = ema_mod.ExponentialMovingAverage.initLength(.{
                .length = length,
                .first_is_average = first_is_average,
            }) catch return error.InvalidSlowKLength;
            return Smoother{ .ema = ema };
        },
        .sma => {
            const sma = sma_mod.SimpleMovingAverage.init(allocator, .{
                .length = length,
            }) catch return error.OutOfMemory;
            return Smoother{ .sma = sma };
        },
    }
}

fn maLabel(ma_type: MovingAverageType, length: usize) []const u8 {
    if (length < 2) return "SMA";
    return switch (ma_type) {
        .ema => "EMA",
        .sma => "SMA",
    };
}

// =============================================================================
// Tests
// =============================================================================

fn testInputHigh() [252]f64 {
    return .{
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
}

fn testInputLow() [252]f64 {
    return .{
        90.75,  91.405, 94.25,   93.5,   92.815, 93.5,   92.0,    89.75,   89.44,  90.625,
        92.75,  96.315, 96.03,   88.815, 86.75,  90.94,  88.905,  88.78,   89.25,  89.75,
        87.5,   86.53,  84.625,  82.28,  81.565, 80.875, 81.25,   84.065,  85.595, 85.97,
        84.405, 85.095, 85.5,    85.53,  87.875, 86.565, 84.655,  83.25,   82.565, 83.44,
        82.53,  85.065, 86.875,  88.53,  89.28,  90.125, 90.75,   89.0,    88.565, 90.095,
        89.0,   86.47,  84.0,    83.315, 82.0,   83.25,  84.75,   85.28,   87.19,  88.44,
        88.25,  87.345, 89.28,   91.095, 89.53,  91.155, 92.0,    90.53,   89.97,  88.815,
        86.75,  85.065, 82.03,   81.5,   82.565, 96.345, 96.47,   101.155, 104.25, 101.75,
        101.72, 101.72, 103.155, 105.69, 106.25, 104.0,  105.53,  108.53,  108.75, 107.75,
        117.0,  118.0,  116.0,   118.5,  116.53, 116.25, 114.595, 110.875, 110.5,  110.72,
        112.62, 114.19, 111.19,  109.44, 111.56, 112.44, 117.5,   116.06,  116.56, 113.31,
        112.56, 114.0,  114.75,  118.87, 119.0,  119.75, 122.62,  123.0,   121.75, 121.56,
        123.12, 122.19, 122.75,  124.37, 128.0,  129.5,  130.81,  130.63,  132.13, 133.88,
        135.38, 135.75, 136.19,  134.5,  135.38, 133.69, 126.06,  126.87,  123.5,  122.62,
        122.75, 123.56, 125.81,  124.62, 124.37, 121.81, 118.19,  118.06,  117.56, 121.0,
        121.12, 118.94, 119.81,  121.0,  122.0,  124.5,  126.56,  123.5,   121.25, 121.06,
        122.31, 121.0,  120.87,  122.06, 122.75, 122.69, 122.87,  125.5,   124.25, 128.0,
        128.38, 130.69, 131.63,  134.38, 132.0,  131.94, 131.94,  129.56,  123.75, 126.0,
        126.25, 124.37, 121.44,  120.44, 121.37, 121.69, 120.0,   119.62,  115.5,  116.75,
        119.06, 119.06, 115.06,  111.06, 113.12, 110.0,  105.0,   104.69,  103.87, 104.69,
        105.44, 107.0,  89.0,    92.5,   92.12,  94.62,  92.81,   94.25,   96.25,  96.37,
        93.69,  93.5,   90.0,    90.19,  90.5,   92.12,  94.12,   94.87,   93.0,   93.87,
        93.0,   92.62,  93.56,   98.37,  104.44, 106.0,  101.81,  104.12,  103.37, 102.12,
        102.25, 103.37, 107.94,  112.5,  115.44, 115.5,  112.25,  107.56,  106.56, 106.87,
        104.5,  105.75, 108.62,  107.75, 108.06, 108.0,  108.19,  108.12,  109.06, 108.75,
        108.56, 106.62,
    };
}

fn testInputClose() [252]f64 {
    return .{
        91.5,    94.815,  94.375,  95.095, 93.78,   94.625,  92.53,   92.75,   90.315,  92.47,
        96.125,  97.25,   98.5,    89.875, 91.0,    92.815,  89.155,  89.345,  91.625,  89.875,
        88.375,  87.625,  84.78,   83.0,   83.5,    81.375,  84.44,   89.25,   86.375,  86.25,
        85.25,   87.125,  85.815,  88.97,  88.47,   86.875,  86.815,  84.875,  84.19,   83.875,
        83.375,  85.5,    89.19,   89.44,  91.095,  90.75,   91.44,   89.0,    91.0,    90.5,
        89.03,   88.815,  84.28,   83.5,   82.69,   84.75,   85.655,  86.19,   88.94,   89.28,
        88.625,  88.5,    91.97,   91.5,   93.25,   93.5,    93.155,  91.72,   90.0,    89.69,
        88.875,  85.19,   83.375,  84.875, 85.94,   97.25,   99.875,  104.94,  106.0,   102.5,
        102.405, 104.595, 106.125, 106.0,  106.065, 104.625, 108.625, 109.315, 110.5,   112.75,
        123.0,   119.625, 118.75,  119.25, 117.94,  116.44,  115.19,  111.875, 110.595, 118.125,
        116.0,   116.0,   112.0,   113.75, 112.94,  116.0,   120.5,   116.62,  117.0,   115.25,
        114.31,  115.5,   115.87,  120.69, 120.19,  120.75,  124.75,  123.37,  122.94,  122.56,
        123.12,  122.56,  124.62,  129.25, 131.0,   132.25,  131.0,   132.81,  134.0,   137.38,
        137.81,  137.88,  137.25,  136.31, 136.25,  134.63,  128.25,  129.0,   123.87,  124.81,
        123.0,   126.25,  128.38,  125.37, 125.69,  122.25,  119.37,  118.5,   123.19,  123.5,
        122.19,  119.31,  123.31,  121.12, 123.37,  127.37,  128.5,   123.87,  122.94,  121.75,
        124.44,  122.0,   122.37,  122.94, 124.0,   123.19,  124.56,  127.25,  125.87,  128.86,
        132.0,   130.75,  134.75,  135.0,  132.38,  133.31,  131.94,  130.0,   125.37,  130.13,
        127.12,  125.19,  122.0,   125.0,  123.0,   123.5,   120.06,  121.0,   117.75,  119.87,
        122.0,   119.19,  116.37,  113.5,  114.25,  110.0,   105.06,  107.0,   107.87,  107.0,
        107.12,  107.0,   91.0,    93.94,  93.87,   95.5,    93.0,    94.94,   98.25,   96.75,
        94.81,   94.37,   91.56,   90.25,  93.94,   93.62,   97.0,    95.0,    95.87,   94.06,
        94.62,   93.75,   98.0,    103.94, 107.87,  106.06,  104.5,   105.0,   104.19,  103.06,
        103.42,  105.27,  111.87,  116.0,  116.62,  118.28,  113.37,  109.0,   109.7,   109.25,
        107.0,   109.19,  110.0,   109.2,  110.12,  108.0,   108.62,  109.75,  109.81,  109.0,
        108.75,  107.87,
    };
}

test "stochastic 5/SMA3/SMA4 single value" {
    const allocator = std.testing.allocator;
    var ind = try Stochastic.init(allocator, .{
        .fast_k_length = 5,
        .slow_k_length = 3,
        .slow_d_length = 4,
    });
    defer ind.deinit();

    const high = testInputHigh();
    const low = testInputLow();
    const close = testInputClose();

    const tolerance = 1e-2;

    // Feed first 9 bars (indices 0..8).
    for (0..9) |i| {
        _ = ind.update(close[i], high[i], low[i]);
    }

    // Index 9: first primed value.
    const r = ind.update(close[9], high[9], low[9]);

    try std.testing.expect(@abs(r.slow_k - 38.139) < tolerance);
    try std.testing.expect(@abs(r.slow_d - 36.725) < tolerance);
    try std.testing.expect(ind.isPrimed());
}

test "stochastic 5/SMA3/SMA3 first value" {
    const allocator = std.testing.allocator;
    var ind = try Stochastic.init(allocator, .{
        .fast_k_length = 5,
        .slow_k_length = 3,
        .slow_d_length = 3,
    });
    defer ind.deinit();

    const high = testInputHigh();
    const low = testInputLow();
    const close = testInputClose();

    const tolerance = 1e-2;

    for (0..8) |i| {
        _ = ind.update(close[i], high[i], low[i]);
    }

    const r = ind.update(close[8], high[8], low[8]);

    try std.testing.expect(@abs(r.slow_k - 24.0128) < tolerance);
    try std.testing.expect(@abs(r.slow_d - 36.254) < tolerance);
    try std.testing.expect(ind.isPrimed());
}

test "stochastic 5/SMA3/SMA3 last value" {
    const allocator = std.testing.allocator;
    var ind = try Stochastic.init(allocator, .{
        .fast_k_length = 5,
        .slow_k_length = 3,
        .slow_d_length = 3,
    });
    defer ind.deinit();

    const high = testInputHigh();
    const low = testInputLow();
    const close = testInputClose();

    const tolerance = 1e-2;

    var r: @TypeOf(ind.update(0, 0, 0)) = undefined;
    for (0..252) |i| {
        r = ind.update(close[i], high[i], low[i]);
    }

    try std.testing.expect(@abs(r.slow_k - 30.194) < tolerance);
    try std.testing.expect(@abs(r.slow_d - 43.69) < tolerance);
}

test "stochastic 5/SMA3/SMA4 last value" {
    const allocator = std.testing.allocator;
    var ind = try Stochastic.init(allocator, .{
        .fast_k_length = 5,
        .slow_k_length = 3,
        .slow_d_length = 4,
    });
    defer ind.deinit();

    const high = testInputHigh();
    const low = testInputLow();
    const close = testInputClose();

    const tolerance = 1e-2;

    var r: @TypeOf(ind.update(0, 0, 0)) = undefined;
    for (0..252) |i| {
        r = ind.update(close[i], high[i], low[i]);
    }

    try std.testing.expect(@abs(r.slow_k - 30.194) < tolerance);
    try std.testing.expect(@abs(r.slow_d - 46.641) < tolerance);
}

test "stochastic is primed" {
    const allocator = std.testing.allocator;
    var ind = try Stochastic.init(allocator, .{
        .fast_k_length = 5,
        .slow_k_length = 3,
        .slow_d_length = 3,
    });
    defer ind.deinit();

    const high = testInputHigh();
    const low = testInputLow();
    const close = testInputClose();

    try std.testing.expect(!ind.isPrimed());

    for (0..8) |i| {
        _ = ind.update(close[i], high[i], low[i]);
        try std.testing.expect(!ind.isPrimed());
    }

    _ = ind.update(close[8], high[8], low[8]);
    try std.testing.expect(ind.isPrimed());
}

test "stochastic nan" {
    const allocator = std.testing.allocator;
    var ind = try Stochastic.init(allocator, .{
        .fast_k_length = 5,
        .slow_k_length = 3,
        .slow_d_length = 3,
    });
    defer ind.deinit();

    const r = ind.update(math.nan(f64), 1.0, 1.0);
    try std.testing.expect(math.isNan(r.fast_k));
    try std.testing.expect(math.isNan(r.slow_k));
    try std.testing.expect(math.isNan(r.slow_d));
}

test "stochastic metadata" {
    const allocator = std.testing.allocator;
    var ind = try Stochastic.init(allocator, .{
        .fast_k_length = 5,
        .slow_k_length = 3,
        .slow_d_length = 3,
    });
    defer ind.deinit();

    var m: Metadata = undefined;
    ind.getMetadata(&m);

    try std.testing.expectEqual(Identifier.stochastic, m.identifier);
    try std.testing.expectEqualStrings("stoch(5/SMA3/SMA3)", m.mnemonic);
    try std.testing.expectEqual(@as(usize, 3), m.outputs_len);

    const outs = m.outputs_buf[0..m.outputs_len];
    try std.testing.expectEqual(@as(i32, 1), outs[0].kind);
    try std.testing.expectEqual(@as(i32, 2), outs[1].kind);
    try std.testing.expectEqual(@as(i32, 3), outs[2].kind);
}

test "stochastic update bar" {
    const allocator = std.testing.allocator;
    var ind = try Stochastic.init(allocator, .{
        .fast_k_length = 5,
        .slow_k_length = 3,
        .slow_d_length = 3,
    });
    defer ind.deinit();

    const high = testInputHigh();
    const low = testInputLow();
    const close = testInputClose();

    const tolerance = 1e-2;

    for (0..8) |i| {
        const bar = Bar{
            .time = 0,
            .open = 0,
            .high = high[i],
            .low = low[i],
            .close = close[i],
            .volume = 0,
        };
        const out = ind.updateBar(&bar);
        const items = out.slice();
        const slow_d_val = items[2].scalar.value;
        try std.testing.expect(math.isNan(slow_d_val));
    }

    const bar = Bar{
        .time = 0,
        .open = 0,
        .high = high[8],
        .low = low[8],
        .close = close[8],
        .volume = 0,
    };
    const out = ind.updateBar(&bar);
    const items2 = out.slice();
    const slow_k_val = items2[1].scalar.value;
    const slow_d_val = items2[2].scalar.value;

    try std.testing.expect(@abs(slow_k_val - 24.0128) < tolerance);
    try std.testing.expect(@abs(slow_d_val - 36.254) < tolerance);
}

test "stochastic invalid params" {
    const allocator = std.testing.allocator;

    // fastK too small
    if (Stochastic.init(allocator, .{ .fast_k_length = 0, .slow_k_length = 3, .slow_d_length = 3 })) |_| {
        return error.ExpectedError;
    } else |_| {}

    // slowK too small
    if (Stochastic.init(allocator, .{ .fast_k_length = 5, .slow_k_length = 0, .slow_d_length = 3 })) |_| {
        return error.ExpectedError;
    } else |_| {}

    // slowD too small
    if (Stochastic.init(allocator, .{ .fast_k_length = 5, .slow_k_length = 3, .slow_d_length = 0 })) |_| {
        return error.ExpectedError;
    } else |_| {}
}
