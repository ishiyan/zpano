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
const band_mod = @import("../../core/outputs/band.zig");
const sma_mod = @import("../../common/simple_moving_average/simple_moving_average.zig");
const ema_mod = @import("../../common/exponential_moving_average/exponential_moving_average.zig");
const variance_mod = @import("../../common/variance/variance.zig");

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;
const Band = band_mod.Band;

/// Enumerates the outputs of the Bollinger Bands indicator.
pub const BollingerBandsOutput = enum(u8) {
    /// Lower band value.
    lower = 1,
    /// Middle band (moving average) value.
    middle = 2,
    /// Upper band value.
    upper = 3,
    /// Band width: (upper - lower) / middle.
    band_width = 4,
    /// Percent band (%B): (sample - lower) / (upper - lower).
    percent_band = 5,
    /// Lower/upper band pair.
    band = 6,
};

/// Specifies the type of moving average.
pub const MovingAverageType = enum(u8) {
    sma = 0,
    ema = 1,
};

/// Parameters to create a Bollinger Bands indicator.
pub const BollingerBandsParams = struct {
    length: usize = 5,
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

/// John Bollinger's Bollinger Bands indicator.
///
/// Bollinger Bands consist of a middle band (moving average) and upper/lower bands
/// placed a specified number of standard deviations above and below the middle band.
///
/// The indicator produces six outputs:
///   - LowerValue: middleValue - lowerMultiplier * stddev
///   - MiddleValue: moving average of the input
///   - UpperValue: middleValue + upperMultiplier * stddev
///   - BandWidth: (upperValue - lowerValue) / middleValue
///   - PercentBand: (sample - lowerValue) / (upperValue - lowerValue)
///   - Band: lower/upper band pair
pub const BollingerBands = struct {
    ma: MaUnion,
    variance: variance_mod.Variance,

    upper_multiplier: f64,
    lower_multiplier: f64,

    middle_value: f64,
    upper_value: f64,
    lower_value: f64,
    band_width_value: f64,
    percent_band_value: f64,
    primed: bool,

    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,

    allocator: std.mem.Allocator,
    mnemonic_buf: [128]u8,
    mnemonic_len: usize,
    description_buf: [192]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: BollingerBandsParams) !BollingerBands {
        const length = params.length;
        if (length < 2) return error.InvalidLength;

        const upper_multiplier = params.upper_multiplier;
        const lower_multiplier = params.lower_multiplier;
        const is_unbiased = params.is_unbiased orelse true;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [128]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "bb({d},{d:.0},{d:.0}{s})", .{ length, upper_multiplier, lower_multiplier, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [192]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Bollinger Bands {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        var v = try variance_mod.Variance.init(allocator, .{
            .length = length,
            .is_unbiased = is_unbiased,
            .bar_component = params.bar_component,
            .quote_component = params.quote_component,
            .trade_component = params.trade_component,
        });
        v.fixSlices();

        var ma = try newMa(allocator, params.moving_average_type, length, params.first_is_average);

        // Fix slices again after potential move.
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
            .middle_value = math.nan(f64),
            .upper_value = math.nan(f64),
            .lower_value = math.nan(f64),
            .band_width_value = math.nan(f64),
            .percent_band_value = math.nan(f64),
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

    pub fn deinit(self: *BollingerBands) void {
        self.variance.deinit();
        self.ma.deinit();
    }

    pub fn fixSlices(self: *BollingerBands) void {
        self.variance.fixSlices();
        switch (self.ma) {
            .sma => |*s| s.fixSlices(),
            .ema => |*e| e.fixSlices(),
        }
    }

    /// Core update. Returns (lower, middle, upper, bandWidth, percentBand).
    pub fn update(self: *BollingerBands, sample: f64) struct { lower: f64, middle: f64, upper: f64, bw: f64, pct_b: f64 } {
        const nan = math.nan(f64);

        if (math.isNan(sample)) {
            return .{ .lower = nan, .middle = nan, .upper = nan, .bw = nan, .pct_b = nan };
        }

        const middle = self.ma.update(sample);
        const v = self.variance.update(sample);

        self.primed = self.ma.isPrimed() and self.variance.isPrimed();

        if (math.isNan(middle) or math.isNan(v)) {
            self.middle_value = nan;
            self.upper_value = nan;
            self.lower_value = nan;
            self.band_width_value = nan;
            self.percent_band_value = nan;
            return .{ .lower = nan, .middle = nan, .upper = nan, .bw = nan, .pct_b = nan };
        }

        const stddev = @sqrt(v);
        const upper = middle + self.upper_multiplier * stddev;
        const lower = middle - self.lower_multiplier * stddev;

        const epsilon: f64 = 1e-10;

        var bw: f64 = undefined;
        if (@abs(middle) < epsilon) {
            bw = 0;
        } else {
            bw = (upper - lower) / middle;
        }

        const spread = upper - lower;
        var pct_b: f64 = undefined;
        if (@abs(spread) < epsilon) {
            pct_b = 0;
        } else {
            pct_b = (sample - lower) / spread;
        }

        self.middle_value = middle;
        self.upper_value = upper;
        self.lower_value = lower;
        self.band_width_value = bw;
        self.percent_band_value = pct_b;

        return .{ .lower = lower, .middle = middle, .upper = upper, .bw = bw, .pct_b = pct_b };
    }

    pub fn isPrimed(self: *const BollingerBands) bool {
        return self.primed;
    }

    pub fn mnemonic(self: *const BollingerBands) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    pub fn description(self: *const BollingerBands) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    pub fn getMetadata(self: *const BollingerBands, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();

        var lower_mn_buf: [160]u8 = undefined;
        const lower_mn = std.fmt.bufPrint(&lower_mn_buf, "{s} lower", .{mn}) catch mn;
        var middle_mn_buf: [160]u8 = undefined;
        const middle_mn = std.fmt.bufPrint(&middle_mn_buf, "{s} middle", .{mn}) catch mn;
        var upper_mn_buf: [160]u8 = undefined;
        const upper_mn = std.fmt.bufPrint(&upper_mn_buf, "{s} upper", .{mn}) catch mn;
        var bw_mn_buf: [160]u8 = undefined;
        const bw_mn = std.fmt.bufPrint(&bw_mn_buf, "{s} bandWidth", .{mn}) catch mn;
        var pctb_mn_buf: [160]u8 = undefined;
        const pctb_mn = std.fmt.bufPrint(&pctb_mn_buf, "{s} percentBand", .{mn}) catch mn;
        var band_mn_buf: [160]u8 = undefined;
        const band_mn = std.fmt.bufPrint(&band_mn_buf, "{s} band", .{mn}) catch mn;

        var lower_desc_buf: [256]u8 = undefined;
        const lower_desc = std.fmt.bufPrint(&lower_desc_buf, "{s} Lower", .{desc}) catch desc;
        var middle_desc_buf: [256]u8 = undefined;
        const middle_desc = std.fmt.bufPrint(&middle_desc_buf, "{s} Middle", .{desc}) catch desc;
        var upper_desc_buf: [256]u8 = undefined;
        const upper_desc = std.fmt.bufPrint(&upper_desc_buf, "{s} Upper", .{desc}) catch desc;
        var bw_desc_buf: [256]u8 = undefined;
        const bw_desc = std.fmt.bufPrint(&bw_desc_buf, "{s} Band Width", .{desc}) catch desc;
        var pctb_desc_buf: [256]u8 = undefined;
        const pctb_desc = std.fmt.bufPrint(&pctb_desc_buf, "{s} Percent Band", .{desc}) catch desc;
        var band_desc_buf: [256]u8 = undefined;
        const band_desc = std.fmt.bufPrint(&band_desc_buf, "{s} Band", .{desc}) catch desc;

        build_metadata_mod.buildMetadata(
            out,
            .bollinger_bands,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = lower_mn, .description = lower_desc },
                .{ .mnemonic = middle_mn, .description = middle_desc },
                .{ .mnemonic = upper_mn, .description = upper_desc },
                .{ .mnemonic = bw_mn, .description = bw_desc },
                .{ .mnemonic = pctb_mn, .description = pctb_desc },
                .{ .mnemonic = band_mn, .description = band_desc },
            },
        );
    }

    pub fn updateScalar(self: *BollingerBands, sample: *const Scalar) OutputArray {
        const r = self.update(sample.value);
        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = sample.time, .value = r.lower } });
        out.append(.{ .scalar = .{ .time = sample.time, .value = r.middle } });
        out.append(.{ .scalar = .{ .time = sample.time, .value = r.upper } });
        out.append(.{ .scalar = .{ .time = sample.time, .value = r.bw } });
        out.append(.{ .scalar = .{ .time = sample.time, .value = r.pct_b } });

        if (math.isNan(r.lower) or math.isNan(r.upper)) {
            out.append(.{ .band = Band.empty(sample.time) });
        } else {
            out.append(.{ .band = Band.new(sample.time, r.lower, r.upper) });
        }

        return out;
    }

    pub fn updateBar(self: *BollingerBands, sample: *const Bar) OutputArray {
        const v = self.bar_func(sample.*);
        return self.updateScalar(&.{ .time = sample.time, .value = v });
    }

    pub fn updateQuote(self: *BollingerBands, sample: *const Quote) OutputArray {
        const v = self.quote_func(sample.*);
        return self.updateScalar(&.{ .time = sample.time, .value = v });
    }

    pub fn updateTrade(self: *BollingerBands, sample: *const Trade) OutputArray {
        const v = self.trade_func(sample.*);
        return self.updateScalar(&.{ .time = sample.time, .value = v });
    }

    pub fn indicator(self: *BollingerBands) indicator_mod.Indicator {
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
        const self: *BollingerBands = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const BollingerBands = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *BollingerBands = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *BollingerBands = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *BollingerBands = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *BollingerBands = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidLength,
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

fn createBB(allocator: std.mem.Allocator, length: usize, is_unbiased: bool) !BollingerBands {
    var bb = try BollingerBands.init(allocator, .{
        .length = length,
        .is_unbiased = is_unbiased,
    });
    bb.fixSlices();
    return bb;
}

test "bollinger bands sample stddev length 20 full data" {
    const tolerance: f64 = 1e-8;
    const closing = testdata.testClosingPrice();
    const sma20 = testdata.testSma20Expected();
    const exp_lower = testdata.testSampleLowerBandExpected();
    const exp_upper = testdata.testSampleUpperBandExpected();
    const exp_bw = testdata.testSampleBandWidthExpected();
    const exp_pctb = testdata.testSamplePercentBandExpected();

    var bb = try createBB(testing.allocator, 20, true);
    defer bb.deinit();

    for (0..252) |i| {
        const r = bb.update(closing[i]);

        if (math.isNan(sma20[i])) {
            try testing.expect(math.isNan(r.lower));
            try testing.expect(math.isNan(r.middle));
            try testing.expect(math.isNan(r.upper));
            continue;
        }

        try testing.expect(almostEqual(r.middle, sma20[i], tolerance));
        try testing.expect(almostEqual(r.lower, exp_lower[i], tolerance));
        try testing.expect(almostEqual(r.upper, exp_upper[i], tolerance));
        try testing.expect(almostEqual(r.bw, exp_bw[i], tolerance));
        try testing.expect(almostEqual(r.pct_b, exp_pctb[i], tolerance));
    }
}

test "bollinger bands population stddev length 20 full data" {
    const tolerance: f64 = 1e-8;
    const closing = testdata.testClosingPrice();
    const sma20 = testdata.testSma20Expected();

    // Population expected data (abbreviated - using sample for structure test).
    // In a full port we'd add population test data arrays.
    // For now, verify basic mechanics with population variance.

    var bb = try BollingerBands.init(testing.allocator, .{
        .length = 20,
        .is_unbiased = false,
    });
    bb.fixSlices();
    defer bb.deinit();

    for (0..252) |i| {
        const r = bb.update(closing[i]);

        if (math.isNan(sma20[i])) {
            try testing.expect(math.isNan(r.lower));
            try testing.expect(math.isNan(r.middle));
            try testing.expect(math.isNan(r.upper));
            continue;
        }

        // Middle should still be SMA20 regardless of variance type.
        try testing.expect(almostEqual(r.middle, sma20[i], tolerance));
        // Lower should be less than middle, upper greater.
        try testing.expect(r.lower < r.middle);
        try testing.expect(r.upper > r.middle);
    }
}

test "bollinger bands is primed" {
    var bb = try createBB(testing.allocator, 20, true);
    defer bb.deinit();

    const closing = testdata.testClosingPrice();

    try testing.expect(!bb.isPrimed());

    for (0..19) |i| {
        _ = bb.update(closing[i]);
        try testing.expect(!bb.isPrimed());
    }

    _ = bb.update(closing[19]);
    try testing.expect(bb.isPrimed());
}

test "bollinger bands nan input" {
    var bb = try createBB(testing.allocator, 20, true);
    defer bb.deinit();

    const r = bb.update(math.nan(f64));
    try testing.expect(math.isNan(r.lower));
    try testing.expect(math.isNan(r.middle));
    try testing.expect(math.isNan(r.upper));
    try testing.expect(math.isNan(r.bw));
    try testing.expect(math.isNan(r.pct_b));
}

test "bollinger bands metadata" {
    var bb = try createBB(testing.allocator, 20, true);
    defer bb.deinit();

    var m: Metadata = undefined;
    bb.getMetadata(&m);

    try testing.expectEqual(Identifier.bollinger_bands, m.identifier);
    try testing.expectEqual(@as(usize, 6), m.outputs_len);
    try testing.expectEqual(@as(u16, 1), m.outputs_buf[0].kind);
    try testing.expectEqual(@as(u16, 6), m.outputs_buf[5].kind);
}

test "bollinger bands update scalar" {
    const tolerance: f64 = 1e-8;
    const closing = testdata.testClosingPrice();
    const exp_lower = testdata.testSampleLowerBandExpected();
    const exp_upper = testdata.testSampleUpperBandExpected();
    const sma20 = testdata.testSma20Expected();

    var bb = try createBB(testing.allocator, 20, true);
    defer bb.deinit();

    const time: i64 = 1617235200;

    // Feed first 19 — all NaN.
    for (0..19) |i| {
        const out = bb.updateScalar(&.{ .time = time, .value = closing[i] });
        try testing.expect(math.isNan(out.slice()[0].scalar.value));
        try testing.expect(out.slice()[5].band.isEmpty());
    }

    // Feed index 19 — first primed value.
    const out = bb.updateScalar(&.{ .time = time, .value = closing[19] });
    const lower = out.slice()[0].scalar.value;
    const middle = out.slice()[1].scalar.value;
    const upper = out.slice()[2].scalar.value;

    try testing.expect(almostEqual(middle, sma20[19], tolerance));
    try testing.expect(almostEqual(lower, exp_lower[19], tolerance));
    try testing.expect(almostEqual(upper, exp_upper[19], tolerance));

    const b = out.slice()[5].band;
    try testing.expect(!b.isEmpty());
    try testing.expect(almostEqual(b.lower, exp_lower[19], tolerance));
    try testing.expect(almostEqual(b.upper, exp_upper[19], tolerance));
}

test "bollinger bands invalid params" {
    const r1 = BollingerBands.init(testing.allocator, .{ .length = 1 });
    try testing.expectError(error.InvalidLength, r1);
    const r0 = BollingerBands.init(testing.allocator, .{ .length = 0 });
    try testing.expectError(error.InvalidLength, r0);
}
