const std = @import("std");
const math = std.math;


const entities = @import("entities");
const Bar = entities.Bar;
const Quote = entities.Quote;
const Trade = entities.Trade;
const Scalar = entities.Scalar;
const bar_component = entities.bar_component;
const quote_component = entities.quote_component;
const trade_component = entities.trade_component;
const indicator_mod = @import("../../core/indicator.zig");
const build_metadata_mod = @import("../../core/build_metadata.zig");
const component_triple_mnemonic_mod = @import("../../core/component_triple_mnemonic.zig");
const identifier_mod = @import("../../core/identifier.zig");
const metadata_mod = @import("../../core/metadata.zig");
const sma_mod = @import("../../common/simple_moving_average/simple_moving_average.zig");
const ema_mod = @import("../../common/exponential_moving_average/exponential_moving_average.zig");
const variance_mod = @import("../../common/variance/variance.zig");

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the Bollinger Bands Trend indicator.
pub const BollingerBandsTrendOutput = enum(u8) {
    /// BBTrend value.
    value = 1,
};

/// Specifies the type of moving average.
pub const MovingAverageType = enum(u8) {
    sma = 0,
    ema = 1,
};

/// Parameters to create a Bollinger Bands Trend indicator.
pub const BollingerBandsTrendParams = struct {
    fast_length: usize = 20,
    slow_length: usize = 50,
    upper_multiplier: f64 = 2.0,
    lower_multiplier: f64 = 2.0,
    is_unbiased: ?bool = null,
    moving_average_type: MovingAverageType = .sma,
    first_is_average: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

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

fn newMa(allocator: std.mem.Allocator, ma_type: MovingAverageType, length: usize, first_is_average: bool) !MaUnion {
    switch (ma_type) {
        .sma => {
            var sma = try sma_mod.SimpleMovingAverage.init(allocator, .{ .length = length });
            sma.fixSlices();
            return .{ .sma = sma };
        },
        .ema => {
            var ema = try ema_mod.ExponentialMovingAverage.initLength(.{
                .length = length,
                .first_is_average = first_is_average,
            });
            ema.fixSlices();
            return .{ .ema = ema };
        },
    }
}

/// A single Bollinger Band line (MA + Variance), used as a sub-component.
const BbLine = struct {
    ma: MaUnion,
    variance: variance_mod.Variance,
    upper_multiplier: f64,
    lower_multiplier: f64,

    fn update(self: *BbLine, sample: f64) struct { lower: f64, middle: f64, upper: f64, primed: bool } {
        const nan = math.nan(f64);

        const middle = self.ma.update(sample);
        const v = self.variance.update(sample);

        const primed = self.ma.isPrimed() and self.variance.isPrimed();

        if (math.isNan(middle) or math.isNan(v)) {
            return .{ .lower = nan, .middle = nan, .upper = nan, .primed = primed };
        }

        const stddev = @sqrt(v);
        const upper = middle + self.upper_multiplier * stddev;
        const lower = middle - self.lower_multiplier * stddev;

        return .{ .lower = lower, .middle = middle, .upper = upper, .primed = primed };
    }

    fn deinit(self: *BbLine) void {
        self.variance.deinit();
        self.ma.deinit();
    }

    pub fn fixSlices(self: *BbLine) void {
        self.variance.fixSlices();
        switch (self.ma) {
            .sma => |*s| s.fixSlices(),
            .ema => |*e| e.fixSlices(),
        }
    }
};

fn newBbLine(
    allocator: std.mem.Allocator,
    length: usize,
    upper_multiplier: f64,
    lower_multiplier: f64,
    is_unbiased: bool,
    ma_type: MovingAverageType,
    first_is_average: bool,
    bc: ?bar_component.BarComponent,
    qc: ?quote_component.QuoteComponent,
    tc: ?trade_component.TradeComponent,
) !BbLine {
    var v = try variance_mod.Variance.init(allocator, .{
        .length = length,
        .is_unbiased = is_unbiased,
        .bar_component = bc,
        .quote_component = qc,
        .trade_component = tc,
    });
    v.fixSlices();

    var ma = try newMa(allocator, ma_type, length, first_is_average);

    // Fix slices after potential move.
    v.fixSlices();
    switch (ma) {
        .sma => |*s| s.fixSlices(),
        .ema => |*e| e.fixSlices(),
    }

    return .{
        .ma = ma,
        .variance = v,
        .upper_multiplier = upper_multiplier,
        .lower_multiplier = lower_multiplier,
    };
}

/// John Bollinger's Bollinger Bands Trend indicator.
///
/// BBTrend measures the difference between the widths of fast and slow Bollinger Bands
/// relative to the fast middle band, indicating trend strength and direction.
///
/// bbtrend = (|fastLower - slowLower| - |fastUpper - slowUpper|) / fastMiddle
pub const BollingerBandsTrend = struct {
    fast_bb: BbLine,
    slow_bb: BbLine,

    value: f64,
    primed: bool,

    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,

    allocator: std.mem.Allocator,
    mnemonic_buf: [128]u8,
    mnemonic_len: usize,
    description_buf: [192]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: BollingerBandsTrendParams) !BollingerBandsTrend {
        const fast_length = params.fast_length;
        const slow_length = params.slow_length;

        if (fast_length < 2) return error.InvalidFastLength;
        if (slow_length < 2) return error.InvalidSlowLength;
        if (slow_length <= fast_length) return error.SlowMustBeGreaterThanFast;

        const upper_multiplier = params.upper_multiplier;
        const lower_multiplier = params.lower_multiplier;
        const is_unbiased = params.is_unbiased orelse true;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [128]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "bbtrend({d},{d},{d:.0},{d:.0}{s})", .{ fast_length, slow_length, upper_multiplier, lower_multiplier, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [192]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Bollinger Bands Trend {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        var fast_bb = try newBbLine(allocator, fast_length, upper_multiplier, lower_multiplier, is_unbiased, params.moving_average_type, params.first_is_average, params.bar_component, params.quote_component, params.trade_component);
        var slow_bb = try newBbLine(allocator, slow_length, upper_multiplier, lower_multiplier, is_unbiased, params.moving_average_type, params.first_is_average, params.bar_component, params.quote_component, params.trade_component);

        // Fix slices after move into struct.
        fast_bb.fixSlices();
        slow_bb.fixSlices();

        return .{
            .fast_bb = fast_bb,
            .slow_bb = slow_bb,
            .value = math.nan(f64),
            .primed = false,
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *BollingerBandsTrend) void {
        self.fast_bb.deinit();
        self.slow_bb.deinit();
    }

    pub fn fixSlices(self: *BollingerBandsTrend) void {
        self.fast_bb.fixSlices();
        self.slow_bb.fixSlices();
    }

    pub fn update(self: *BollingerBandsTrend, sample: f64) f64 {
        const nan = math.nan(f64);

        if (math.isNan(sample)) {
            return nan;
        }

        const fast = self.fast_bb.update(sample);
        const slow = self.slow_bb.update(sample);

        self.primed = fast.primed and slow.primed;

        if (!self.primed or math.isNan(fast.middle) or math.isNan(fast.lower) or math.isNan(slow.lower)) {
            self.value = nan;
            return nan;
        }

        const epsilon: f64 = 1e-10;

        const lower_diff = @abs(fast.lower - slow.lower);
        const upper_diff = @abs(fast.upper - slow.upper);

        if (@abs(fast.middle) < epsilon) {
            self.value = 0;
            return 0;
        }

        const result = (lower_diff - upper_diff) / fast.middle;
        self.value = result;

        return result;
    }

    pub fn isPrimed(self: *const BollingerBandsTrend) bool {
        return self.primed;
    }

    pub fn mnemonic(self: *const BollingerBandsTrend) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    pub fn description(self: *const BollingerBandsTrend) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    pub fn getMetadata(self: *const BollingerBandsTrend, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();

        build_metadata_mod.buildMetadata(
            out,
            .bollinger_bands_trend,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = mn, .description = desc },
            },
        );
    }

    pub fn updateScalar(self: *BollingerBandsTrend, sample: *const Scalar) OutputArray {
        const v = self.update(sample.value);
        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = sample.time, .value = v } });
        return out;
    }

    pub fn updateBar(self: *BollingerBandsTrend, sample: *const Bar) OutputArray {
        const v = self.bar_func(sample.*);
        return self.updateScalar(&.{ .time = sample.time, .value = v });
    }

    pub fn updateQuote(self: *BollingerBandsTrend, sample: *const Quote) OutputArray {
        const v = self.quote_func(sample.*);
        return self.updateScalar(&.{ .time = sample.time, .value = v });
    }

    pub fn updateTrade(self: *BollingerBandsTrend, sample: *const Trade) OutputArray {
        const v = self.trade_func(sample.*);
        return self.updateScalar(&.{ .time = sample.time, .value = v });
    }

    pub fn indicator(self: *BollingerBandsTrend) indicator_mod.Indicator {
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
        const self: *BollingerBandsTrend = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const BollingerBandsTrend = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *BollingerBandsTrend = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *BollingerBandsTrend = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *BollingerBandsTrend = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *BollingerBandsTrend = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidFastLength,
        InvalidSlowLength,
        SlowMustBeGreaterThanFast,
        MnemonicTooLong,
        OutOfMemory,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");


fn almostEqual(a: f64, b: f64, eps: f64) bool {
    return @abs(a - b) < eps;
}

fn createBBTrend(allocator: std.mem.Allocator, is_unbiased: bool) !BollingerBandsTrend {
    var bbt = try BollingerBandsTrend.init(allocator, .{
        .fast_length = 20,
        .slow_length = 50,
        .is_unbiased = is_unbiased,
    });
    bbt.fixSlices();
    return bbt;
}

test "bollinger bands trend sample stddev full data" {
    const tolerance: f64 = 1e-8;
    const closing = testdata.testClosingPrice();
    const expected = testdata.testSampleExpected();

    var bbt = try createBBTrend(testing.allocator, true);
    defer bbt.deinit();

    for (0..252) |i| {
        const v = bbt.update(closing[i]);

        if (math.isNan(expected[i])) {
            try testing.expect(math.isNan(v));
            continue;
        }

        try testing.expect(almostEqual(v, expected[i], tolerance));
    }
}

test "bollinger bands trend population stddev full data" {
    const tolerance: f64 = 1e-8;
    const closing = testdata.testClosingPrice();
    const expected = testdata.testPopulationExpected();

    var bbt = try createBBTrend(testing.allocator, false);
    defer bbt.deinit();

    for (0..252) |i| {
        const v = bbt.update(closing[i]);

        if (math.isNan(expected[i])) {
            try testing.expect(math.isNan(v));
            continue;
        }

        try testing.expect(almostEqual(v, expected[i], tolerance));
    }
}

test "bollinger bands trend is primed" {
    var bbt = try createBBTrend(testing.allocator, true);
    defer bbt.deinit();

    const closing = testdata.testClosingPrice();

    try testing.expect(!bbt.isPrimed());

    for (0..49) |i| {
        _ = bbt.update(closing[i]);
        try testing.expect(!bbt.isPrimed());
    }

    _ = bbt.update(closing[49]);
    try testing.expect(bbt.isPrimed());
}

test "bollinger bands trend nan input" {
    var bbt = try createBBTrend(testing.allocator, true);
    defer bbt.deinit();

    const v = bbt.update(math.nan(f64));
    try testing.expect(math.isNan(v));
}

test "bollinger bands trend metadata" {
    var bbt = try createBBTrend(testing.allocator, true);
    defer bbt.deinit();

    var m: Metadata = undefined;
    bbt.getMetadata(&m);

    try testing.expectEqual(Identifier.bollinger_bands_trend, m.identifier);
    try testing.expectEqual(@as(usize, 1), m.outputs_len);
    try testing.expectEqual(@as(u16, 1), m.outputs_buf[0].kind);
}

test "bollinger bands trend update scalar" {
    const tolerance: f64 = 1e-8;
    const closing = testdata.testClosingPrice();
    const expected = testdata.testSampleExpected();

    var bbt = try createBBTrend(testing.allocator, true);
    defer bbt.deinit();

    const time: i64 = 1617235200;

    // Feed first 49 — all NaN.
    for (0..49) |i| {
        const out = bbt.updateScalar(&.{ .time = time, .value = closing[i] });
        try testing.expect(math.isNan(out.slice()[0].scalar.value));
    }

    // Feed index 49 — first primed value.
    const out = bbt.updateScalar(&.{ .time = time, .value = closing[49] });
    const v = out.slice()[0].scalar.value;

    try testing.expect(almostEqual(v, expected[49], tolerance));
}

test "bollinger bands trend invalid params" {
    // fast too small
    try testing.expectError(error.InvalidFastLength, BollingerBandsTrend.init(testing.allocator, .{ .fast_length = 1, .slow_length = 50 }));
    // slow too small
    try testing.expectError(error.InvalidSlowLength, BollingerBandsTrend.init(testing.allocator, .{ .fast_length = 20, .slow_length = 1 }));
    // slow not greater than fast
    try testing.expectError(error.SlowMustBeGreaterThanFast, BollingerBandsTrend.init(testing.allocator, .{ .fast_length = 20, .slow_length = 20 }));
    // slow less than fast
    try testing.expectError(error.SlowMustBeGreaterThanFast, BollingerBandsTrend.init(testing.allocator, .{ .fast_length = 50, .slow_length = 20 }));
}
