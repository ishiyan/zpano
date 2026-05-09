const std = @import("std");
const math = std.math;


const entities = @import("entities");
const Bar = entities.Bar;
const Quote = entities.Quote;
const Trade = entities.Trade;
const Scalar = entities.Scalar;
const quote_component = entities.quote_component;
const trade_component = entities.trade_component;
const indicator_mod = @import("../../core/indicator.zig");
const line_indicator_mod = @import("../../core/line_indicator.zig");
const build_metadata_mod = @import("../../core/build_metadata.zig");
const identifier_mod = @import("../../core/identifier.zig");
const metadata_mod = @import("../../core/metadata.zig");

const sma_mod = @import("../../common/simple_moving_average/simple_moving_average.zig");
const ema_mod = @import("../../common/exponential_moving_average/exponential_moving_average.zig");

const OutputArray = indicator_mod.OutputArray;
const LineIndicator = line_indicator_mod.LineIndicator;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the Advance-Decline Oscillator.
pub const AdvanceDeclineOscillatorOutput = enum(u8) {
    /// The scalar value of the oscillator.
    value = 1,
};

/// Moving average type for the ADOSC.
pub const MovingAverageType = enum {
    sma,
    ema,
};

/// Parameters for creating an AdvanceDeclineOscillator.
pub const AdvanceDeclineOscillatorParams = struct {
    fast_length: u32 = 3,
    slow_length: u32 = 10,
    moving_average_type: MovingAverageType = .sma,
    first_is_average: bool = false,
};

/// Marc Chaikin's Advance-Decline (A/D) Oscillator (Chaikin Oscillator).
///
/// The Chaikin Oscillator is the difference between a fast and slow moving average
/// of the Accumulation/Distribution Line.
///
/// CLV = ((Close - Low) - (High - Close)) / (High - Low)
/// AD  = AD_previous + CLV × Volume
/// ADOSC = FastMA(AD) - SlowMA(AD)
///
/// When High equals Low, the A/D value is unchanged (no division by zero).
pub const AdvanceDeclineOscillator = struct {
    line: LineIndicator,
    ad: f64,
    fast_ma: MaUnion,
    slow_ma: MaUnion,
    value: f64,
    primed: bool,
    allocator: std.mem.Allocator,
    mnemonic_buf: [96]u8,
    mnemonic_len: usize,
    description_buf: [160]u8,
    description_len: usize,

    const MaUnion = union(enum) {
        sma: sma_mod.SimpleMovingAverage,
        ema: ema_mod.ExponentialMovingAverage,

        fn update(self: *MaUnion, sample: f64) f64 {
            return switch (self.*) {
                .sma => |*s| s.update(sample),
                .ema => |*e| e.update(sample),
            };
        }

        fn isPrimed(self: *const MaUnion) bool {
            return switch (self.*) {
                .sma => |*s| s.isPrimed(),
                .ema => |*e| e.isPrimed(),
            };
        }

        fn deinit(self: *MaUnion) void {
            switch (self.*) {
                .sma => |*s| s.deinit(),
                .ema => {},
            }
        }
    };

    pub fn init(allocator: std.mem.Allocator, params: AdvanceDeclineOscillatorParams) !AdvanceDeclineOscillator {
        if (params.fast_length < 2) {
            return error.InvalidFastLength;
        }
        if (params.slow_length < 2) {
            return error.InvalidSlowLength;
        }

        var fast_ma: MaUnion = undefined;
        var slow_ma: MaUnion = undefined;
        var ma_label: []const u8 = undefined;

        switch (params.moving_average_type) {
            .ema => {
                ma_label = "EMA";
                var fast_ema = try ema_mod.ExponentialMovingAverage.initLength(.{
                    .length = params.fast_length,
                    .first_is_average = params.first_is_average,
                });
                fast_ema.fixSlices();
                var slow_ema = try ema_mod.ExponentialMovingAverage.initLength(.{
                    .length = params.slow_length,
                    .first_is_average = params.first_is_average,
                });
                slow_ema.fixSlices();
                fast_ma = .{ .ema = fast_ema };
                slow_ma = .{ .ema = slow_ema };
            },
            .sma => {
                ma_label = "SMA";
                var fast_sma = try sma_mod.SimpleMovingAverage.init(allocator, .{
                    .length = params.fast_length,
                });
                fast_sma.fixSlices();
                var slow_sma = try sma_mod.SimpleMovingAverage.init(allocator, .{
                    .length = params.slow_length,
                });
                slow_sma.fixSlices();
                fast_ma = .{ .sma = fast_sma };
                slow_ma = .{ .sma = slow_sma };
            },
        }

        var mnemonic_buf: [96]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "adosc({s}{d}/{s}{d})", .{
            ma_label, params.fast_length, ma_label, params.slow_length,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [160]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Chaikin Advance-Decline Oscillator {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                null,
                null,
                null,
            ),
            .ad = 0,
            .fast_ma = fast_ma,
            .slow_ma = slow_ma,
            .value = math.nan(f64),
            .primed = false,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *AdvanceDeclineOscillator) void {
        self.fast_ma.deinit();
        self.slow_ma.deinit();
    }

    pub fn fixSlices(self: *AdvanceDeclineOscillator) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    /// Core update logic. For scalar/quote/trade, H=L=C so AD is unchanged,
    /// but the unchanged AD value is still fed to the MAs.
    pub fn update(self: *AdvanceDeclineOscillator, sample: f64) f64 {
        if (math.isNan(sample)) {
            return math.nan(f64);
        }
        return self.updateHlcv(sample, sample, sample, 1);
    }

    /// Updates the indicator with the given high, low, close, and volume values.
    pub fn updateHlcv(self: *AdvanceDeclineOscillator, high: f64, low: f64, close: f64, volume: f64) f64 {
        if (math.isNan(high) or math.isNan(low) or math.isNan(close) or math.isNan(volume)) {
            return math.nan(f64);
        }

        // Compute cumulative AD.
        const temp = high - low;
        if (temp > 0) {
            self.ad += ((close - low) - (high - close)) / temp * volume;
        }

        // Feed AD to both MAs.
        const fast = self.fast_ma.update(self.ad);
        const slow = self.slow_ma.update(self.ad);
        self.primed = self.fast_ma.isPrimed() and self.slow_ma.isPrimed();

        if (math.isNan(fast) or math.isNan(slow)) {
            self.value = math.nan(f64);
            return self.value;
        }

        self.value = fast - slow;
        return self.value;
    }

    pub fn isPrimed(self: *const AdvanceDeclineOscillator) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const AdvanceDeclineOscillator, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .advance_decline_oscillator,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *AdvanceDeclineOscillator, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    /// Shadows LineIndicator.updateBar to extract HLCV directly from the bar.
    pub fn updateBar(self: *AdvanceDeclineOscillator, sample: *const Bar) OutputArray {
        const value = self.updateHlcv(sample.high, sample.low, sample.close, sample.volume);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *AdvanceDeclineOscillator, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *AdvanceDeclineOscillator, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *AdvanceDeclineOscillator) indicator_mod.Indicator {
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

    fn vtableIsPrimed(ptr: *anyopaque) bool {
        const self: *AdvanceDeclineOscillator = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const AdvanceDeclineOscillator = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *AdvanceDeclineOscillator = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *AdvanceDeclineOscillator = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *AdvanceDeclineOscillator = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *AdvanceDeclineOscillator = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");


fn roundTo(v: f64, comptime digits: comptime_int) f64 {
    const p = comptime blk: {
        var result: f64 = 1.0;
        for (0..digits) |_| result *= 10.0;
        break :blk result;
    };
    return @round(v * p) / p;
}

// High test data, 252 entries.
// Low test data, 252 entries.
// Close test data, 252 entries.
// Volume test data, 252 entries.
// Expected EMA ADOSC output, 252 entries.
// Expected SMA ADOSC output, 252 entries. First 9 are NaN, then 243 valid.
test "AdvanceDeclineOscillator EMA" {
    const allocator = testing.allocator;

    const highs = testdata.testHighs();
    const lows = testdata.testLows();
    const closes = testdata.testCloses();
    const volumes = testdata.testVolumes();
    const expected = testdata.testExpectedEMA();

    var adosc = try AdvanceDeclineOscillator.init(allocator, .{
        .fast_length = 3,
        .slow_length = 10,
        .moving_average_type = .ema,
    });
    defer adosc.deinit();
    adosc.fixSlices();

    // EMA with length 10 has lookback = 9. First 9 values are NaN.
    for (0..252) |i| {
        const v = adosc.updateHlcv(highs[i], lows[i], closes[i], volumes[i]);

        if (i < 9) {
            try testing.expect(math.isNan(v));
            try testing.expect(!adosc.isPrimed());
            continue;
        }

        try testing.expect(!math.isNan(v));
        try testing.expect(adosc.isPrimed());

        const got = roundTo(v, 2);
        const exp = roundTo(expected[i], 2);
        try testing.expectEqual(exp, got);
    }
}

test "AdvanceDeclineOscillator SMA" {
    const allocator = testing.allocator;

    const highs = testdata.testHighs();
    const lows = testdata.testLows();
    const closes = testdata.testCloses();
    const volumes = testdata.testVolumes();
    const expected = testdata.testExpectedSMA();

    var adosc = try AdvanceDeclineOscillator.init(allocator, .{
        .fast_length = 3,
        .slow_length = 10,
        .moving_average_type = .sma,
    });
    defer adosc.deinit();
    adosc.fixSlices();

    for (0..252) |i| {
        const v = adosc.updateHlcv(highs[i], lows[i], closes[i], volumes[i]);

        if (i < 9) {
            try testing.expect(math.isNan(v));
            try testing.expect(!adosc.isPrimed());
            continue;
        }

        try testing.expect(!math.isNan(v));
        try testing.expect(adosc.isPrimed());

        const got = roundTo(v, 2);
        const exp = roundTo(expected[i], 2);
        try testing.expectEqual(exp, got);
    }
}

test "AdvanceDeclineOscillator updateBar" {
    const allocator = testing.allocator;

    const highs = testdata.testHighs();
    const lows = testdata.testLows();
    const closes = testdata.testCloses();
    const volumes = testdata.testVolumes();
    const expected = testdata.testExpectedEMA();

    var adosc = try AdvanceDeclineOscillator.init(allocator, .{
        .fast_length = 3,
        .slow_length = 10,
        .moving_average_type = .ema,
    });
    defer adosc.deinit();
    adosc.fixSlices();

    for (0..15) |i| {
        const bar = Bar{
            .time = 0,
            .open = highs[i],
            .high = highs[i],
            .low = lows[i],
            .close = closes[i],
            .volume = volumes[i],
        };

        const output = adosc.updateBar(&bar);
        const scalar_val = output.slice()[0].scalar.value;

        if (i < 9) {
            try testing.expect(math.isNan(scalar_val));
            continue;
        }

        const got = roundTo(scalar_val, 2);
        const exp = roundTo(expected[i], 2);
        try testing.expectEqual(exp, got);
    }
}

test "AdvanceDeclineOscillator NaN" {
    const allocator = testing.allocator;

    var adosc = try AdvanceDeclineOscillator.init(allocator, .{
        .fast_length = 3,
        .slow_length = 10,
        .moving_average_type = .ema,
    });
    defer adosc.deinit();
    adosc.fixSlices();

    try testing.expect(math.isNan(adosc.update(math.nan(f64))));
    try testing.expect(math.isNan(adosc.updateHlcv(math.nan(f64), 1, 2, 3)));
    try testing.expect(math.isNan(adosc.updateHlcv(1, math.nan(f64), 2, 3)));
    try testing.expect(math.isNan(adosc.updateHlcv(1, 2, math.nan(f64), 3)));
    try testing.expect(math.isNan(adosc.updateHlcv(1, 2, 3, math.nan(f64))));
}

test "AdvanceDeclineOscillator not primed initially" {
    const allocator = testing.allocator;

    var adosc = try AdvanceDeclineOscillator.init(allocator, .{
        .fast_length = 3,
        .slow_length = 10,
        .moving_average_type = .ema,
    });
    defer adosc.deinit();
    adosc.fixSlices();

    try testing.expect(!adosc.isPrimed());
    try testing.expect(math.isNan(adosc.update(math.nan(f64))));
    try testing.expect(!adosc.isPrimed());
}

test "AdvanceDeclineOscillator metadata" {
    const allocator = testing.allocator;

    var adosc = try AdvanceDeclineOscillator.init(allocator, .{
        .fast_length = 3,
        .slow_length = 10,
        .moving_average_type = .ema,
    });
    defer adosc.deinit();
    adosc.fixSlices();

    var meta: Metadata = undefined;
    adosc.getMetadata(&meta);

    try testing.expectEqual(Identifier.advance_decline_oscillator, meta.identifier);
    try testing.expectEqualStrings("adosc(EMA3/EMA10)", meta.mnemonic);
    try testing.expectEqualStrings("Chaikin Advance-Decline Oscillator adosc(EMA3/EMA10)", meta.description);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
}

test "AdvanceDeclineOscillator invalid params" {
    const allocator = testing.allocator;

    try testing.expectError(error.InvalidFastLength, AdvanceDeclineOscillator.init(allocator, .{ .fast_length = 1 }));
    try testing.expectError(error.InvalidSlowLength, AdvanceDeclineOscillator.init(allocator, .{ .slow_length = 1 }));
}
