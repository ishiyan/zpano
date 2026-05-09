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

const testdata = @import("testdata.zig");

test "stochastic 5/SMA3/SMA4 single value" {
    const allocator = std.testing.allocator;
    var ind = try Stochastic.init(allocator, .{
        .fast_k_length = 5,
        .slow_k_length = 3,
        .slow_d_length = 4,
    });
    defer ind.deinit();

    const high = testdata.testInputHigh();
    const low = testdata.testInputLow();
    const close = testdata.testInputClose();

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

    const high = testdata.testInputHigh();
    const low = testdata.testInputLow();
    const close = testdata.testInputClose();

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

    const high = testdata.testInputHigh();
    const low = testdata.testInputLow();
    const close = testdata.testInputClose();

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

    const high = testdata.testInputHigh();
    const low = testdata.testInputLow();
    const close = testdata.testInputClose();

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

    const high = testdata.testInputHigh();
    const low = testdata.testInputLow();
    const close = testdata.testInputClose();

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

    const high = testdata.testInputHigh();
    const low = testdata.testInputLow();
    const close = testdata.testInputClose();

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
