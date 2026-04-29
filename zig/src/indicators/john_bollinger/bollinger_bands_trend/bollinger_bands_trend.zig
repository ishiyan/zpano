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

fn testSampleExpected() [252]f64 {
    const nan = math.nan(f64);
    return .{
        nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,
        nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,
        nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,
        nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,
        nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,           -0.0376219669,
        -0.0321146284, -0.0274334573, -0.0246090294, -0.0256356253, -0.0272597903, -0.0252002454, -0.0233925242, -0.0188583730, -0.0127466031, -0.0050911412,
        0.0043008064,  0.0116686617,  0.0028033170,  0.0007458673,  -0.0025200779, -0.0088581862, -0.0091349199, -0.0101970164, -0.0105075890, -0.0098627871,
        -0.0096998956, -0.0130545096, -0.0145474851, -0.0124079193, -0.0042496669, -0.0145322942, -0.0242971129, -0.0398474173, -0.0531342459, -0.0556068980,
        -0.0554649993, -0.0541248691, -0.0570826821, -0.0561292886, -0.0552836501, -0.0533353196, -0.0528652701, -0.0505577415, -0.0442314972, -0.0362411609,
        -0.0356183167, -0.0087279783, 0.0375600611,  0.0882413048,  0.1503131540,  0.1736786887,  0.1933669951,  0.2013446822,  0.2077950061,  0.2219289127,
        0.2326338239,  0.2319318686,  0.2267447540,  0.2221590760,  0.2169226095,  0.2148703381,  0.2119802904,  0.2071126259,  0.2024686471,  0.1954711305,
        0.1798426338,  0.1672636758,  0.1567069279,  0.1477569570,  0.1402469712,  0.1343032268,  0.1311083580,  0.1294687229,  0.1280828716,  0.1204485382,
        0.1145208979,  0.1071275209,  0.1033059342,  0.1007624677,  0.0996631360,  0.0812381757,  0.0668902634,  0.0620584068,  0.0587380455,  0.0535497811,
        0.0549432044,  0.0601251184,  0.0709145991,  0.0733021534,  0.0787887195,  0.0838753557,  0.0827818272,  0.0855226332,  0.0820415190,  0.0831410505,
        0.0828492690,  0.0905083971,  0.0966492716,  0.0910220729,  0.0853101651,  0.0727897695,  0.0542953255,  0.0350910038,  0.0264346507,  0.0320718358,
        0.0390199595,  0.0292430471,  0.0147129620,  0.0002817984,  -0.0134386970, -0.0230079496, -0.0253862449, -0.0319324605, -0.0346189161, -0.0392462428,
        -0.0413377748, -0.0469636153, -0.0540771660, -0.0568401696, -0.0594962321, -0.0595027734, -0.0551002672, -0.0490795858, -0.0478072640, -0.0454128533,
        -0.0402053820, -0.0334914376, -0.0274565070, -0.0181361797, -0.0113981829, -0.0070294758, -0.0046113271, 0.0010884994,  0.0056958629,  0.0144859065,
        0.0198971055,  0.0228734643,  0.0155514161,  0.0121624943,  0.0030101506,  -0.0021251645, -0.0107975520, -0.0191896111, -0.0309791720, -0.0300678708,
        -0.0225981376, -0.0157645639, -0.0050204385, 0.0084528862,  0.0195322184,  0.0139712426,  0.0149294436,  0.0101880054,  0.0082437586,  -0.0084337782,
        -0.0216731636, -0.0334862048, -0.0140330854, -0.0196492849, -0.0275783337, -0.0436197151, -0.0530311626, -0.0739893710, -0.0932691078, -0.1222886259,
        -0.1644567814, -0.2052431189, -0.2407805532, -0.2709123842, -0.3137310459, -0.3297932902, -0.3281153947, -0.3292189654, -0.3312714634, -0.3324884274,
        -0.3320946640, -0.3327840517, -0.3086281760, -0.2834407020, -0.2565374785, -0.2328667919, -0.2083046267, -0.1866469665, -0.1713373842, -0.1533115470,
        -0.1342767806, -0.1144887884, -0.0889806075, -0.0588917831, -0.0336202018, -0.0074316181, 0.0108458687,  0.0286380805,  0.0445468191,  0.0624490925,
        0.0792172029,  0.0966381229,  0.1094907503,  0.1156334764,  0.1178400609,  0.1182254146,  0.1237330602,  0.1267953520,  0.1308796694,  0.1351922195,
        0.1391017747,  0.1409749480,
    };
}

fn testPopulationExpected() [252]f64 {
    const nan = math.nan(f64);
    return .{
        nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,
        nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,
        nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,
        nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,
        nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,           nan,           -0.0376219669,
        -0.0321146284, -0.0274334573, -0.0246090294, -0.0256356253, -0.0272597903, -0.0252002454, -0.0233925242, -0.0188583730, -0.0127466031, -0.0050911412,
        0.0043008064,  0.0116686617,  0.0047069019,  0.0027246458,  -0.0004114123, -0.0065741247, -0.0067803930, -0.0077882803, -0.0081126642, -0.0074849893,
        -0.0073238826, -0.0105754069, -0.0120071083, -0.0099545191, -0.0020395522, -0.0119730458, -0.0212563525, -0.0359663766, -0.0485174225, -0.0507276583,
        -0.0504319743, -0.0489143520, -0.0515753451, -0.0504414165, -0.0494351148, -0.0474190457, -0.0467708025, -0.0443674057, -0.0380609959, -0.0301141944,
        -0.0290956832, -0.0026453403, 0.0426682962,  0.0922421422,  0.1528725431,  0.1757582812,  0.1950358337,  0.2028446771,  0.2091623989,  0.2230300364,
        0.2326338239,  0.2319318686,  0.2267447540,  0.2221590760,  0.2169226095,  0.2148703381,  0.2119802904,  0.2071126259,  0.2024686471,  0.1954711305,
        0.1798426338,  0.1672636758,  0.1567069279,  0.1477569570,  0.1402469712,  0.1343032268,  0.1311083580,  0.1294687229,  0.1280828716,  0.1204485382,
        0.1145208979,  0.1071275209,  0.1033059342,  0.1007624677,  0.1000393074,  0.0830039514,  0.0689873739,  0.0643440365,  0.0611901269,  0.0561995642,
        0.0575972776,  0.0626997248,  0.0732544587,  0.0755874224,  0.0809199635,  0.0857962225,  0.0846485123,  0.0872407258,  0.0837618908,  0.0847717294,
        0.0844875068,  0.0919461595,  0.0979296204,  0.0924423527,  0.0868665192,  0.0746469989,  0.0566042395,  0.0378348755,  0.0292968775,  0.0347860354,
        0.0415361657,  0.0292430471,  0.0147129620,  0.0002817984,  -0.0134386970, -0.0230079496, -0.0253862449, -0.0319324605, -0.0346189161, -0.0392462428,
        -0.0413377748, -0.0469636153, -0.0540771660, -0.0568401696, -0.0594962321, -0.0595027734, -0.0551002672, -0.0490795858, -0.0478072640, -0.0454128533,
        -0.0402053820, -0.0334914376, -0.0274565070, -0.0181361797, -0.0113981829, -0.0070294758, -0.0046113271, 0.0010884994,  0.0056958629,  0.0144859065,
        0.0198971055,  0.0245868851,  0.0173526273,  0.0139400766,  0.0049109812,  -0.0001826283, -0.0086041372, -0.0167713269, -0.0281914994, -0.0300678708,
        -0.0225981376, -0.0157645639, -0.0050204385, 0.0084528862,  0.0179103214,  0.0109405410,  0.0115396210,  0.0066776659,  0.0045668568,  -0.0119188961,
        -0.0250329849, -0.0367395996, -0.0184630356, -0.0244503647, -0.0326479475, -0.0486527491, -0.0582205989, -0.0789985811, -0.0980380127, -0.1266054327,
        -0.1680160110, -0.2080807897, -0.2430731873, -0.2727986850, -0.3147787605, -0.3297932902, -0.3281153947, -0.3292189654, -0.3312714634, -0.3324884274,
        -0.3320946640, -0.3327840517, -0.3086281760, -0.2834407020, -0.2565374785, -0.2328667919, -0.2083046267, -0.1866469665, -0.1713373842, -0.1533115470,
        -0.1342767806, -0.1144887884, -0.0889806075, -0.0588917831, -0.0336202018, -0.0074316181, 0.0108458687,  0.0286380805,  0.0445468191,  0.0624490925,
        0.0792172029,  0.0966381229,  0.1094907503,  0.1156334764,  0.1190922012,  0.1194987659,  0.1237330602,  0.1267953520,  0.1308796694,  0.1351922195,
        0.1391017747,  0.1409749480,
    };
}

test "bollinger bands trend sample stddev full data" {
    const tolerance: f64 = 1e-8;
    const closing = testClosingPrice();
    const expected = testSampleExpected();

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
    const closing = testClosingPrice();
    const expected = testPopulationExpected();

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

    const closing = testClosingPrice();

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
    const closing = testClosingPrice();
    const expected = testSampleExpected();

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
