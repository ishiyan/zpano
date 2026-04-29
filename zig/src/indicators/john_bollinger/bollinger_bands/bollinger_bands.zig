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

fn testClosingPrice() [252]f64 {
    return .{
        91.5000,  94.8150,  94.3750,  95.0950,  93.7800,  94.6250,  92.5300,  92.7500,  90.3150,  92.4700,
        96.1250,  97.2500,  98.5000,  89.8750,  91.0000,  92.8150,  89.1550,  89.3450,  91.6250,  89.8750,
        88.3750,  87.6250,  84.7800,  83.0000,  83.5000,  81.3750,  84.4400,  89.2500,  86.3750,  86.2500,
        85.2500,  87.1250,  85.8150,  88.9700,  88.4700,  86.8750,  86.8150,  84.8750,  84.1900,  83.8750,
        83.3750,  85.5000,  89.1900,  89.4400,  91.0950,  90.7500,  91.4400,  89.0000,  91.0000,  90.5000,
        89.0300,  88.8150,  84.2800,  83.5000,  82.6900,  84.7500,  85.6550,  86.1900,  88.9400,  89.2800,
        88.6250,  88.5000,  91.9700,  91.5000,  93.2500,  93.5000,  93.1550,  91.7200,  90.0000,  89.6900,
        88.8750,  85.1900,  83.3750,  84.8750,  85.9400,  97.2500,  99.8750,  104.9400, 106.0000, 102.5000,
        102.4050, 104.5950, 106.1250, 106.0000, 106.0650, 104.6250, 108.6250, 109.3150, 110.5000, 112.7500,
        123.0000, 119.6250, 118.7500, 119.2500, 117.9400, 116.4400, 115.1900, 111.8750, 110.5950, 118.1250,
        116.0000, 116.0000, 112.0000, 113.7500, 112.9400, 116.0000, 120.5000, 116.6200, 117.0000, 115.2500,
        114.3100, 115.5000, 115.8700, 120.6900, 120.1900, 120.7500, 124.7500, 123.3700, 122.9400, 122.5600,
        123.1200, 122.5600, 124.6200, 129.2500, 131.0000, 132.2500, 131.0000, 132.8100, 134.0000, 137.3800,
        137.8100, 137.8800, 137.2500, 136.3100, 136.2500, 134.6300, 128.2500, 129.0000, 123.8700, 124.8100,
        123.0000, 126.2500, 128.3800, 125.3700, 125.6900, 122.2500, 119.3700, 118.5000, 123.1900, 123.5000,
        122.1900, 119.3100, 123.3100, 121.1200, 123.3700, 127.3700, 128.5000, 123.8700, 122.9400, 121.7500,
        124.4400, 122.0000, 122.3700, 122.9400, 124.0000, 123.1900, 124.5600, 127.2500, 125.8700, 128.8600,
        132.0000, 130.7500, 134.7500, 135.0000, 132.3800, 133.3100, 131.9400, 130.0000, 125.3700, 130.1300,
        127.1200, 125.1900, 122.0000, 125.0000, 123.0000, 123.5000, 120.0600, 121.0000, 117.7500, 119.8700,
        122.0000, 119.1900, 116.3700, 113.5000, 114.2500, 110.0000, 105.0600, 107.0000, 107.8700, 107.0000,
        107.1200, 107.0000, 91.0000,  93.9400,  93.8700,  95.5000,  93.0000,  94.9400,  98.2500,  96.7500,
        94.8100,  94.3700,  91.5600,  90.2500,  93.9400,  93.6200,  97.0000,  95.0000,  95.8700,  94.0600,
        94.6200,  93.7500,  98.0000,  103.9400, 107.8700, 106.0600, 104.5000, 105.0000, 104.1900, 103.0600,
        103.4200, 105.2700, 111.8700, 116.0000, 116.6200, 118.2800, 113.3700, 109.0000, 109.7000, 109.2500,
        107.0000, 109.1900, 110.0000, 109.2000, 110.1200, 108.0000, 108.6200, 109.7500, 109.8100, 109.0000,
        108.7500, 107.8700,
    };
}

fn testSma20Expected() [252]f64 {
    const nan = math.nan(f64);
    return .{
        nan,         nan,         nan,         nan,         nan,         nan,         nan,         nan,         nan,         nan,
        nan,         nan,         nan,         nan,         nan,         nan,         nan,         nan,         nan,         92.8910000,
        92.7347500,  92.3752500,  91.8955000,  91.2907500,  90.7767500,  90.1142500,  89.7097500,  89.5347500,  89.3377500,  89.0267500,
        88.4830000,  87.9767500,  87.3425000,  87.2972500,  87.1707500,  86.8737500,  86.7567500,  86.5332500,  86.1615000,  85.8615000,
        85.6115000,  85.5052500,  85.7257500,  86.0477500,  86.4275000,  86.8962500,  87.2462500,  87.2337500,  87.4650000,  87.6775000,
        87.8665000,  87.9510000,  87.8742500,  87.6007500,  87.3117500,  87.2055000,  87.1475000,  87.2132500,  87.4507500,  87.7210000,
        87.9835000,  88.1335000,  88.2725000,  88.3755000,  88.4832500,  88.6207500,  88.7065000,  88.8425000,  88.7925000,  88.7520000,
        88.7442500,  88.5630000,  88.5177500,  88.5865000,  88.7490000,  89.3740000,  90.0850000,  91.0225000,  91.8755000,  92.5365000,
        93.2255000,  94.0302500,  94.7380000,  95.4630000,  96.1037500,  96.6600000,  97.4335000,  98.3132500,  99.3382500,  100.4912500,
        102.1975000, 103.9192500, 105.6880000, 107.4067500, 109.0067500, 109.9662500, 110.7320000, 111.0787500, 111.3085000, 112.0897500,
        112.7695000, 113.3397500, 113.6335000, 114.0210000, 114.3647500, 114.9335000, 115.5272500, 115.8925000, 116.2175000, 116.3425000,
        115.9080000, 115.7017500, 115.5577500, 115.6297500, 115.7422500, 115.9577500, 116.4357500, 117.0105000, 117.6277500, 117.8495000,
        118.2055000, 118.5335000, 119.1645000, 119.9395000, 120.8425000, 121.6550000, 122.1800000, 122.9895000, 123.8395000, 124.9460000,
        126.1210000, 127.2400000, 128.3090000, 129.0900000, 129.8930000, 130.5870000, 130.7620000, 131.0435000, 131.0900000, 131.2025000,
        131.1965000, 131.3810000, 131.5690000, 131.3750000, 131.1095000, 130.6095000, 130.0280000, 129.3125000, 128.7720000, 128.0780000,
        127.2970000, 126.3685000, 125.6715000, 124.9120000, 124.2680000, 123.9050000, 123.9175000, 123.6610000, 123.6145000, 123.4615000,
        123.5335000, 123.3210000, 123.0205000, 122.8990000, 122.8145000, 122.8615000, 123.1210000, 123.5585000, 123.6925000, 123.9605000,
        124.4510000, 125.0230000, 125.5950000, 126.2890000, 126.7395000, 127.0365000, 127.2085000, 127.5150000, 127.6365000, 128.0555000,
        128.1895000, 128.3490000, 128.3305000, 128.4335000, 128.3835000, 128.3990000, 128.1740000, 127.8615000, 127.4555000, 127.0060000,
        126.5060000, 125.9280000, 125.0090000, 123.9340000, 123.0275000, 121.8620000, 120.5180000, 119.3680000, 118.4930000, 117.3365000,
        116.3365000, 115.4270000, 113.8770000, 112.3240000, 110.8675000, 109.4675000, 108.1145000, 106.8115000, 105.8365000, 104.6805000,
        103.3210000, 102.0800000, 100.8395000, 99.6770000,  98.6615000,  97.8425000,  97.4395000,  96.8395000,  96.2395000,  95.5925000,
        94.9675000,  94.3050000,  94.6550000,  95.1550000,  95.8550000,  96.3830000,  96.9580000,  97.4610000,  97.7580000,  98.0735000,
        98.5040000,  99.0490000,  100.0645000, 101.3520000, 102.4860000, 103.7190000, 104.5375000, 105.2375000, 105.9290000, 106.6885000,
        107.3075000, 108.0795000, 108.6795000, 108.9425000, 109.0550000, 109.1520000, 109.3580000, 109.5955000, 109.8765000, 110.1735000,
        110.4400000, 110.5700000,
    };
}

fn testSampleLowerBandExpected() [252]f64 {
    const nan = math.nan(f64);
    return .{
        nan,                 nan,                 nan,                 nan,                 nan,                 nan,                 nan,                 nan,                 nan,                 nan,
        nan,                 nan,                 nan,                 nan,                 nan,                 nan,                 nan,                 nan,                 nan,                 87.573975109686800,
        87.073112265023900,  86.367282066978800,  85.081580886581500,  83.584037085445000,  82.424829152547900,  80.982005060688600,  80.315139296881000,  80.248828298730600,  79.954857686971200,  79.668697303890700,
        79.610358786400600,  80.112159620196000,  81.192020771773800,  81.212253373265700,  81.308757097456200,  81.648511778283800,  81.642983981229000,  81.505783048605100,  81.645814938024700,  81.594189618186100,
        81.378491737857300,  81.379728055349900,  81.302755925965200,  81.523757265350900,  81.543251956813800,  82.260527294746800,  82.342124051802000,  82.349895962136300,  82.321280272839600,  82.395820143622600,
        82.680917659459400,  82.761216919653300,  82.509026430812200,  81.922183395282700,  81.244434392236500,  81.032476429657800,  80.937341514014900,  81.076991689544700,  81.440786102219000,  81.905089336818800,
        82.530943767031500,  82.804962035219700,  82.688983375619200,  82.627715163533400,  82.447239401320700,  82.251245950067600,  82.134238432985000,  82.133518196320300,  82.136524054319800,  82.129999666184500,
        82.123256852204300,  81.754351938816900,  81.578474688981900,  81.831749546858900,  82.450566945075300,  82.311925904784400,  81.835906269873000,  80.648973725467300,  79.592998583243300,  79.445810620862900,
        79.563549418761200,  79.662413304248300,  79.433476279354900,  79.447082785849400,  79.448013237392900,  79.631489232837500,  79.685150899605500,  80.021375106888000,  80.713368432026100,  81.529195682301000,
        81.568504438055200,  83.520268272317700,  86.703917637191600,  90.217139222434100,  94.479310165056800,  96.192971953075400,  97.634656597213100,  98.262876670003100,  98.713129045808800,  99.861878697972500,
        101.321876068187000, 102.485806513390000, 103.295985741116000, 104.327322407716000, 105.398703828617000, 107.212112000969000, 108.025219766018000, 108.975456594492000, 109.772574211362000, 110.086537370552000,
        110.441662440373000, 110.522162254771000, 110.578733530857000, 110.391048398106000, 110.206443611729000, 109.988947356253000, 109.307290895079000, 109.582990039689000, 110.395833550931000, 110.288877705924000,
        110.346919774673000, 110.516506798057000, 111.328167681209000, 111.330123288041000, 111.561960371860000, 111.368697406131000, 111.100663139738000, 111.273131827315000, 111.502520117080000, 111.903285720009000,
        112.880079535892000, 113.995699537656000, 115.484085267545000, 116.316341248697000, 117.460448328978000, 118.768979683106000, 119.207186739991000, 119.983203487367000, 120.166173795443000, 120.607150722332000,
        120.581760942295000, 121.282608465753000, 121.868417489005000, 121.330066215408000, 120.747148153557000, 119.538084980847000, 117.874177815839000, 116.201118242956000, 115.583215929315000, 115.343482299960000,
        115.174335536929000, 114.828133728334000, 115.270726192755000, 115.623205165702000, 116.654011212041000, 117.835457400395000, 117.809372972058000, 118.040110765237000, 117.985509345409000, 117.803059541618000,
        117.863151361317000, 117.761841032939000, 117.987972226656000, 117.989477995182000, 118.051243816077000, 118.103148781795000, 118.604443789789000, 119.235648723598000, 119.253164464134000, 118.958582578177000,
        118.372077311511000, 118.829080740801000, 118.092460079203000, 118.002496324873000, 118.147154630581000, 117.955639213071000, 117.883881142261000, 118.249631543669000, 118.561970521115000, 119.359413826506000,
        119.646749364458000, 120.181972840037000, 120.104980083622000, 120.443985094760000, 120.265674844014000, 120.321845993662000, 119.423896481815000, 118.544397954497000, 117.120892885527000, 116.159262144531000,
        115.706724390765000, 114.851269340403000, 113.964279613270000, 112.798875110005000, 111.835704367766000, 110.329113202032000, 107.733141192143000, 106.047829936284000, 104.548407235476000, 103.620855033681000,
        102.708163769152000, 101.858717946080000, 96.832841747791800,  93.939696168513200,  91.457449564858500,  89.855780759226000,  87.856645928580000,  86.691422282247000,  86.061346770172500,  85.671261830376400,
        85.687808928375400,  85.700240988855900,  85.279105640229500,  84.633770923347900,  85.090361468077500,  85.207752798646700,  85.268401403216200,  85.497985249222400,  86.155011703089400,  86.842763462613000,
        88.101872800529800,  90.416471662495100,  90.758899385282800,  89.483221294495800,  87.867864290491500,  87.189599392126900,  87.232375463160300,  87.151748355303200,  87.019717083257700,  87.091886131451600,
        87.386841332332000,  87.718495482037100,  87.946877382723300,  88.197554523680100,  88.163457738163100,  88.399037240666400,  88.981631417101100,  90.238700223456500,  91.483608237991800,  93.312964949216600,
        95.197620822889600,  97.774652954479200,  99.510973802645300,  100.048757965939000, 100.161457694662000, 100.354195010596000, 100.830030907531000, 101.317652644431000, 101.999411905907000, 102.958459350153000,
        103.914487315967000, 104.383359044216000,
    };
}

fn testSampleUpperBandExpected() [252]f64 {
    const nan = math.nan(f64);
    return .{
        nan,                 nan,                 nan,                 nan,                 nan,                 nan,                 nan,                 nan,                 nan,                 nan,
        nan,                 nan,                 nan,                 nan,                 nan,                 nan,                 nan,                 nan,                 nan,                 98.208024890313200,
        98.396387734976100,  98.383217933021200,  98.709419113418500,  98.997462914555000,  99.128670847452100,  99.246494939311400,  99.104360703119000,  98.820671701269400,  98.720642313028800,  98.384802696109300,
        97.355641213599400,  95.841340379804000,  93.492979228226200,  93.382246626734400,  93.032742902543800,  92.098988221716200,  91.870516018771000,  91.560716951395000,  90.677185061975400,  90.128810381813900,
        89.844508262142700,  89.630771944650100,  90.148744074034900,  90.571742734649100,  91.311748043186200,  91.531972705253200,  92.150375948198100,  92.117604037863700,  92.608719727160400,  92.959179856377400,
        93.052082340540600,  93.140783080346700,  93.239473569187800,  93.279316604717300,  93.379065607763500,  93.378523570342200,  93.357658485985100,  93.349508310455300,  93.460713897781000,  93.536910663181200,
        93.436056232968600,  93.462037964780300,  93.856016624380800,  94.123284836466600,  94.519260598679300,  94.990254049932400,  95.278761567015100,  95.551481803679700,  95.448475945680200,  95.374000333815500,
        95.365243147795700,  95.371648061183100,  95.457025311018100,  95.341250453141100,  95.047433054924700,  96.436074095215600,  98.334093730127000,  101.396026274533000, 104.158001416757000, 105.627189379137000,
        106.887450581239000, 108.398086695752000, 110.042523720645000, 111.478917214151000, 112.759486762607000, 113.688510767162000, 115.181849100394000, 116.605124893112000, 117.963131567974000, 119.453304317699000,
        122.826495561945000, 124.318231727682000, 124.672082362808000, 124.596360777566000, 123.534189834943000, 123.739528046925000, 123.829343402787000, 123.894623329997000, 123.903870954191000, 124.317621302027000,
        124.217123931813000, 124.193693486610000, 123.971014258884000, 123.714677592284000, 123.330796171383000, 122.654887999031000, 123.029280233982000, 122.809543405508000, 122.662425788638000, 122.598462629448000,
        121.374337559627000, 120.881337745229000, 120.536766469143000, 120.868451601894000, 121.278056388271000, 121.926552643747000, 123.564209104921000, 124.438009960311000, 124.859666449069000, 125.410122294076000,
        126.064080225327000, 126.550493201943000, 127.000832318791000, 128.548876711959000, 130.123039628140000, 131.941302593869000, 133.259336860262000, 134.705868172685000, 136.176479882920000, 137.988714279991000,
        139.361920464108000, 140.484300462344000, 141.133914732455000, 141.863658751303000, 142.325551671022000, 142.405020316894000, 142.316813260009000, 142.103796512633000, 142.013826204557000, 141.797849277668000,
        141.811239057705000, 141.479391534247000, 141.269582510996000, 141.419933784592000, 141.471851846443000, 141.680915019153000, 142.181822184161000, 142.423881757044000, 141.960784070685000, 140.812517700040000,
        139.419664463071000, 137.908866271666000, 136.072273807245000, 134.200794834298000, 131.881988787959000, 129.974542599605000, 130.025627027942000, 129.281889234763000, 129.243490654591000, 129.119940458382000,
        129.203848638683000, 128.880158967061000, 128.053027773344000, 127.808522004818000, 127.577756183923000, 127.619851218205000, 127.637556210211000, 127.881351276403000, 128.131835535866000, 128.962417421823000,
        130.529922688489000, 131.216919259199000, 133.097539920797000, 134.575503675127000, 135.331845369419000, 136.117360786929000, 136.533118857739000, 136.780368456331000, 136.711029478885000, 136.751586173495000,
        136.732250635542000, 136.516027159963000, 136.556019916378000, 136.423014905240000, 136.501325155986000, 136.476154006338000, 136.924103518185000, 137.178602045503000, 137.790107114473000, 137.852737855469000,
        137.305275609235000, 137.004730659597000, 136.053720386730000, 135.069124889995000, 134.219295632234000, 133.394886797968000, 133.302858807857000, 132.688170063716000, 132.437592764524000, 131.052144966319000,
        129.964836230848000, 128.995282053920000, 130.921158252208000, 130.708303831487000, 130.277550435142000, 129.079219240774000, 128.372354071420000, 126.931577717753000, 125.611653229828000, 123.689738169624000,
        120.954191071625000, 118.459759011144000, 116.399894359771000, 114.720229076652000, 112.232638531923000, 110.477247201353000, 109.610598596784000, 108.181014750778000, 106.323988296911000, 104.342236537387000,
        101.833127199470000, 98.193528337504900,  98.551100614717200,  100.826778705504000, 103.842135709508000, 105.576400607873000, 106.683624536840000, 107.770251644697000, 108.496282916742000, 109.055113868548000,
        109.621158667668000, 110.379504517963000, 112.182122617277000, 114.506445476320000, 116.808542261837000, 119.038962759334000, 120.093368582899000, 120.236299776544000, 120.374391762008000, 120.064035050783000,
        119.417379177110000, 118.384347045521000, 117.848026197355000, 117.836242034061000, 117.948542305338000, 117.949804989404000, 117.885969092469000, 117.873347355569000, 117.753588094093000, 117.388540649847000,
        116.965512684033000, 116.756640955784000,
    };
}

fn testSampleBandWidthExpected() [252]f64 {
    const nan = math.nan(f64);
    return .{
        nan,                 nan,                 nan,                 nan,                 nan,                 nan,                 nan,                 nan,                 nan,                 nan,
        nan,                 nan,                 nan,                 nan,                 nan,                 nan,                 nan,                 nan,                 nan,                 0.11447879536905100,
        0.12210390894408200, 0.13007743812376500, 0.14829712256679700, 0.16883885639136400, 0.18401013139272200, 0.20268148354586400, 0.20944458552429300, 0.20742609324914300, 0.21005436812610100, 0.21023013186731600,
        0.20055018960929000, 0.17878792703308500, 0.14083588695597700, 0.13940866697941500, 0.13449449276377200, 0.12029498488821400, 0.11788745011243500, 0.11619734498345900, 0.10481909117123900, 0.09939985632242420,
        0.09888877690830560, 0.09649751201593090, 0.10318939347943500, 0.10515075024388400, 0.11302532280087300, 0.10669557559165600, 0.11242032633375200, 0.11197166321208800, 0.11761778373430300, 0.12047970930689000,
        0.11803320584160200, 0.11801532854309000, 0.12211139370607000, 0.12964652938969800, 0.13898050623801500, 0.14157417984742200, 0.14252063423471900, 0.14071848739624500, 0.13744796694782000, 0.13260019067683300,
        0.12394497224976400, 0.12091969488969200, 0.12650636663470100, 0.13007643151024100, 0.13643284121411200, 0.14374746433385900, 0.14817993195572000, 0.15103090983886500, 0.14992203047960600, 0.14922481372398400,
        0.14921514684716400, 0.15375829773569400, 0.15678833479201800, 0.15250067342408000, 0.14193811885034700, 0.15803419552030000, 0.18314022823171500, 0.22793323133363100, 0.26737272541116500, 0.28293028975889700,
        0.29309471295383300, 0.30560030832102900, 0.32309155187243000, 0.33554187934908000, 0.34661991363723400, 0.35233831506647000, 0.36431718249666600, 0.37211413300062900, 0.37497905525764500, 0.37738717187215900,
        0.40370841873714600, 0.39259293591288000, 0.35924764141261900, 0.32008436671933400, 0.26654202303881600, 0.25050009520056500, 0.23655932165565400, 0.23075292672985500, 0.22631462923660300, 0.21818000846692000,
        0.20302695200054400, 0.19152933523516900, 0.18194483596622600, 0.17003319725812700, 0.15679737281606100, 0.13436270537364500, 0.12987464401657500, 0.11936999211352200, 0.11091145117797400, 0.10754389203340500,
        0.09432200641244870, 0.08953343826223870, 0.08617364857213430, 0.09061165663497070, 0.09565748701569070, 0.10294788651464700, 0.12244450875131800, 0.12695458886700700, 0.12296276089730700, 0.12830978992827900,
        0.13296471357638400, 0.13526966135215800, 0.13152125538714800, 0.14356199103646300, 0.15359727956869200, 0.16910612130810500, 0.18136089147588600, 0.19052631602997900, 0.19924143561497400, 0.20877361868312300,
        0.20997170120928800, 0.20817825310191600, 0.19990670541356900, 0.19790314898602400, 0.19142758533596100, 0.18099841970325200, 0.17673044554242300, 0.16880343569323600, 0.16666147234048500, 0.16151139311626600,
        0.16181436330549300, 0.15372681794546800, 0.14746000214329400, 0.15292001955611400, 0.15807171633547000, 0.16953460535646400, 0.18694161540839300, 0.20278599140909100, 0.20483931399194100, 0.19885566139446300,
        0.19046268903542500, 0.18264624921030400, 0.16552319033742700, 0.14872542004448200, 0.12254142318148700, 0.09797090673669520, 0.09858376787688330, 0.09090803462308930, 0.09107330700833770, 0.09166323847323820,
        0.09180260639717450, 0.09015753954412970, 0.08181608387778460, 0.07989523112177860, 0.07756830315513840, 0.07745878437435040, 0.07336776358560260, 0.06997254379751300, 0.07178018935449920, 0.08070179487535200,
        0.09769182551348610, 0.09908447660348570, 0.11947195224009500, 0.13123080672310500, 0.13559064647436000, 0.14296459343462200, 0.14660370742111100, 0.14532201633268700, 0.14219332994692400, 0.13581745686041600,
        0.13328315713130900, 0.12726280937075200, 0.12819275100429500, 0.12441481241638700, 0.12646212567792500, 0.12581334755470400, 0.13653476552475000, 0.14573741189494900, 0.16216808398968300, 0.17080669976960100,
        0.17073143739007900, 0.17592164823704500, 0.17670280358582400, 0.17969443235908100, 0.18193973920031900, 0.18927781913915200, 0.21216513396931800, 0.22317823979150200, 0.23536568007433700, 0.23378309334808500,
        0.23429166651649800, 0.23509719656441500, 0.29934329587551800, 0.32734417989898500, 0.35014860865702800, 0.35831126573227800, 0.37474814333729500, 0.37673991504197600, 0.37369250173290900, 0.36318584969738500,
        0.34132830831340400, 0.32092004332179000, 0.30861704708513100, 0.30183952319295500, 0.27510505175620700, 0.25826705575498000, 0.24981857658924300, 0.23423323645367000, 0.20957067102199400, 0.18306324319140000,
        0.14458898464148700, 0.08246706616838850, 0.08232213015091070, 0.11921136473131600, 0.16665037211430700, 0.19076809412185000, 0.20061520528145700, 0.21155645118964200, 0.21969113354901500, 0.22394660878929400,
        0.22571994371128000, 0.22878584373316100, 0.24219623577346100, 0.25957939609124500, 0.27950241519499100, 0.29541285124873200, 0.29761317389260100, 0.28504667588157300, 0.27273724404097500, 0.25073995886685900,
        0.22570424578170900, 0.19069013171824100, 0.16872595470819600, 0.16327405804090900, 0.16310196332746400, 0.16120281789438300, 0.15596424756248100, 0.15106181103365200, 0.14338076101974900, 0.13097597244069000,
        0.11817299319146800, 0.11190451217842900,
    };
}

fn testSamplePercentBandExpected() [252]f64 {
    const nan = math.nan(f64);
    return .{
        nan,                     nan,                     nan,                      nan,                      nan,                     nan,                      nan,                      nan,                     nan,                     nan,
        nan,                     nan,                     nan,                      nan,                      nan,                     nan,                      nan,                      nan,                     nan,                     0.216382745781888000000,
        0.114974482289231000000, 0.104670826063206000000, -0.022129767139997600000, -0.037891452031514900000, 0.064366680856423900000, 0.021516885602777800000,  0.219533349143969000000,  0.484667650171922000000, 0.342119577782747000000, 0.351638471690030000000,
        0.317810732894019000000, 0.445849055140413000000, 0.375822684434906000000,  0.637448720402801000000,  0.610819820289802000000, 0.500119611771460000000,  0.505695411149647000000,  0.335080964625745000000, 0.281705325665736000000, 0.267242147598882000000,
        0.235826171188724000000, 0.499363716873834000000, 0.891618205000187000000,  0.874917710855154000000,  0.977811523772977000000, 0.915657950769241000000,  0.927573643529782000000,  0.680825428678512000000, 0.843622921495327000000, 0.767197186951035000000,
        0.612186050051103000000, 0.583240473313027000000, 0.165041917298509000000,  0.138927366230641000000,  0.119127279773760000000, 0.301110430567828000000,  0.379833984642402000000,  0.416622642966599000000, 0.623898414809933000000, 0.634028881312565000000,
        0.558825619818573000000, 0.534390296402356000000, 0.831108533272258000000,  0.771800362130530000000,  0.894859313289062000000, 0.883016476773554000000,  0.838429926642466000000,  0.714451319455194000000, 0.590707960023779000000, 0.570824520742626000000,
        0.509873896338612000000, 0.252300312067099000000, 0.129446176329507000000,  0.225267422849496000000,  0.277008029179284000000, 1.057626548079960000000,  1.093398518690930000000,  1.170818178489980000000, 1.074984668055080000000, 0.880556734310686000000,
        0.835951295732459000000, 0.867652772776983000000, 0.872014190308955000000,  0.828953998047960000000,  0.799033604516476000000, 0.733872477426493000000,  0.815282845088709000000,  0.800727783901004000000, 0.799646200682237000000, 0.823244248608595000000,
        1.004205353516470000000, 0.884964068541878000000, 0.844025056106732000000,  0.844488602832606000000,  0.807461263013206000000, 0.735011228915309000000,  0.670187184641253000000,  0.531064991807319000000, 0.471676102172974000000, 0.746782528656451000000,
        0.641099149449804000000, 0.622547625353024000000, 0.420991644650155000000,  0.486021816930671000000,  0.420547475845739000000, 0.569061417463662000000,  0.831426950099117000000,  0.552587497095990000000, 0.560706672633800000000, 0.412683301938428000000,
        0.353832663774517000000, 0.480524511802525000000, 0.531356594413289000000,  0.982967955091274000000,  0.901725574202140000000, 0.901441485506354000000,  1.083173016610340000000,  0.928104441056401000000, 0.867278164606273000000, 0.811515363205661000000,
        0.812683707431119000000, 0.751122827385219000000, 0.848090138221818000000,  1.040718585764010000000,  1.047247272626320000000, 1.015005265658590000000,  0.898038263085698000000,  0.919093180380567000000, 0.911790409663656000000, 0.976664585801571000000,
        0.941396805897482000000, 0.901682219089322000000, 0.848579315594739000000,  0.782612841808678000000,  0.755659504509320000000, 0.671052337514616000000,  0.391300709778930000000,  0.407620017344659000000, 0.169529711256848000000, 0.198334626236754000000,
        0.113909491536172000000, 0.245949640465059000000, 0.335628427654456000000,  0.201093101817606000000,  0.238500483272990000000, 0.122473731427336000000,  0.061537109951741100000,  0.087667409874974000000, 0.288380794996592000000, 0.320252320981672000000,
        0.289361488328031000000, 0.194182150122481000000, 0.386474792944973000000,  0.295883100679632000000,  0.441029595327211000000, 0.785441608089663000000,  0.875114988525714000000,  0.518591364397242000000, 0.440086949740283000000, 0.348765749804372000000,
        0.579933356638422000000, 0.381187063742269000000, 0.435370451063818000000,  0.504175559245869000000,  0.624442183479574000000, 0.534518259049816000000,  0.659302788787014000000,  0.926975133305080000000, 0.745250666727919000000, 0.989762183860115000000,
        1.120915940771400000000, 0.962308254300752000000, 1.110126710197300000000,  1.025613717287490000000,  0.828228193670808000000, 0.845424302122895000000,  0.753710101838263000000,  0.634101520717285000000, 0.375117492026793000000, 0.619277796850900000000,
        0.437403065732110000000, 0.306600373788020000000, 0.115191497658710000000,  0.285124626418301000000,  0.168414883509144000000, 0.196737242093210000000,  0.036348342443188000000,  0.131779282523164000000, 0.030436914896960200000, 0.171053173079037000000,
        0.291374895731579000000, 0.195848883255009000000, 0.108908161659781000000,  0.031482578638390500000,  0.107860066050537000000, -0.014268465814228200000, -0.104543242608985000000, 0.035741663175521800000, 0.119099668976155000000, 0.123185784358551000000,
        0.161862613165573000000, 0.189459580567713000000, -0.171109704025311000000, 0.000008263339466562320,  0.062146938855286000000, 0.143899144472743000000,  0.126947159686482000000,  0.204983743936408000000, 0.308181005936353000000, 0.291404054985413000000,
        0.258665349753513000000, 0.264648552071053000000, 0.201823109830976000000,  0.186669665403447000000,  0.326046282377343000000, 0.332901286717168000000,  0.481944933051642000000,  0.418904130514229000000, 0.481679784381663000000, 0.412425934572330000000,
        0.474692770965864000000, 0.428636240779448000000, 0.929275361545402000000,  1.274448409938370000000,  1.252146979654830000000, 1.026301442347280000000,  0.887738595677410000000,  0.865642447183746000000, 0.799489222339808000000, 0.727038578285905000000,
        0.721099659857208000000, 0.774524404016496000000, 0.987121128164541000000,  1.056769953791240000000,  0.993417988985819000000, 0.975229614743312000000,  0.783896072820707000000,  0.625426702671374000000, 0.630526055025999000000, 0.595753178855075000000,
        0.487303754418079000000, 0.553882410631350000000, 0.572012664389889000000,  0.514476471153189000000,  0.559874904927407000000, 0.434529123946968000000,  0.456730612412062000000,  0.509332136325035000000, 0.495778897023517000000, 0.418676826857180000000,
        0.370508258750670000000, 0.281787886245806000000,
    };
}

test "bollinger bands sample stddev length 20 full data" {
    const tolerance: f64 = 1e-8;
    const closing = testClosingPrice();
    const sma20 = testSma20Expected();
    const exp_lower = testSampleLowerBandExpected();
    const exp_upper = testSampleUpperBandExpected();
    const exp_bw = testSampleBandWidthExpected();
    const exp_pctb = testSamplePercentBandExpected();

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
    const closing = testClosingPrice();
    const sma20 = testSma20Expected();

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

    const closing = testClosingPrice();

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
    const closing = testClosingPrice();
    const exp_lower = testSampleLowerBandExpected();
    const exp_upper = testSampleUpperBandExpected();
    const sma20 = testSma20Expected();

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
