const std = @import("std");
const math = std.math;

const Bar = @import("bar").Bar;
const Quote = @import("quote").Quote;
const Trade = @import("trade").Trade;
const Scalar = @import("scalar").Scalar;
const quote_component = @import("quote_component");
const trade_component = @import("trade_component");

const indicator_mod = @import("../../core/indicator.zig");
const line_indicator_mod = @import("../../core/line_indicator.zig");
const build_metadata_mod = @import("../../core/build_metadata.zig");
const identifier_mod = @import("../../core/identifier.zig");
const metadata_mod = @import("../../core/metadata.zig");

const OutputArray = indicator_mod.OutputArray;
const LineIndicator = line_indicator_mod.LineIndicator;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the Advance-Decline indicator.
pub const AdvanceDeclineOutput = enum(u8) {
    /// The scalar value of the A/D line.
    value = 1,
};

/// Marc Chaikin's Advance-Decline (A/D) Line.
///
/// The Accumulation/Distribution Line is a cumulative indicator that uses volume
/// and price to assess whether a stock is being accumulated or distributed.
///
/// CLV = ((Close - Low) - (High - Close)) / (High - Low)
/// AD  = AD_previous + CLV × Volume
///
/// When High equals Low, the A/D value is unchanged (no division by zero).
pub const AdvanceDecline = struct {
    line: LineIndicator,
    ad: f64,
    value: f64,
    primed: bool,

    const mnemonic_str = "ad";
    const description_str = "Advance-Decline";

    pub fn init() AdvanceDecline {
        return .{
            .line = LineIndicator.new(
                mnemonic_str,
                description_str,
                null,
                null,
                null,
            ),
            .ad = 0,
            .value = math.nan(f64),
            .primed = false,
        };
    }

    pub fn deinit(_: *AdvanceDecline) void {}

    pub fn fixSlices(_: *AdvanceDecline) void {}

    /// Core update logic. For scalar/quote/trade, H=L=C so AD is unchanged.
    pub fn update(self: *AdvanceDecline, sample: f64) f64 {
        if (math.isNan(sample)) {
            return math.nan(f64);
        }
        return self.updateHlcv(sample, sample, sample, 1);
    }

    /// Updates the indicator with the given high, low, close, and volume values.
    pub fn updateHlcv(self: *AdvanceDecline, high: f64, low: f64, close: f64, volume: f64) f64 {
        if (math.isNan(high) or math.isNan(low) or math.isNan(close) or math.isNan(volume)) {
            return math.nan(f64);
        }

        const temp = high - low;
        if (temp > 0) {
            self.ad += ((close - low) - (high - close)) / temp * volume;
        }

        self.value = self.ad;
        self.primed = true;

        return self.value;
    }

    pub fn isPrimed(self: *const AdvanceDecline) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const AdvanceDecline, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .advance_decline,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *AdvanceDecline, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    /// Shadows LineIndicator.updateBar to extract HLCV directly from the bar.
    pub fn updateBar(self: *AdvanceDecline, sample: *const Bar) OutputArray {
        const value = self.updateHlcv(sample.high, sample.low, sample.close, sample.volume);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *AdvanceDecline, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *AdvanceDecline, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *AdvanceDecline) indicator_mod.Indicator {
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
        const self: *AdvanceDecline = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const AdvanceDecline = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *AdvanceDecline = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *AdvanceDecline = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *AdvanceDecline = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *AdvanceDecline = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;

fn roundTo(v: f64, comptime digits: comptime_int) f64 {
    const p = comptime blk: {
        var result: f64 = 1.0;
        for (0..digits) |_| result *= 10.0;
        break :blk result;
    };
    return @round(v * p) / p;
}

// High test data, 252 entries.
fn testHighs() [252]f64 {
    return .{
        93.25,  94.94,  96.375,  96.19,   96,      94.72,  95,     93.72,   92.47,   92.75,
        96.25,  99.625, 99.125,  92.75,   91.315,  93.25,  93.405, 90.655,  91.97,   92.25,
        90.345, 88.5,   88.25,   85.5,    84.44,   84.75,  84.44,  89.405,  88.125,  89.125,
        87.155, 87.25,  87.375,  88.97,   90,      89.845, 86.97,  85.94,   84.75,   85.47,
        84.47,  88.5,   89.47,   90,      92.44,   91.44,  92.97,  91.72,   91.155,  91.75,
        90,     88.875, 89,      85.25,   83.815,  85.25,  86.625, 87.94,   89.375,  90.625,
        90.75,  88.845, 91.97,   93.375,  93.815,  94.03,  94.03,  91.815,  92,      91.94,
        89.75,  88.75,  86.155,  84.875,  85.94,   99.375, 103.28, 105.375, 107.625, 105.25,
        104.5,  105.5,  106.125, 107.94,  106.25,  107,    108.75, 110.94,  110.94,  114.22,
        123,    121.75, 119.815, 120.315, 119.375, 118.19, 116.69, 115.345, 113,     118.315,
        116.87, 116.75, 113.87,  114.62,  115.31,  116,    121.69, 119.87,  120.87,  116.75,
        116.5,  116,    118.31,  121.5,   122,     121.44, 125.75, 127.75,  124.19,  124.44,
        125.75, 124.69, 125.31,  132,     131.31,  132.25, 133.88, 133.5,   135.5,   137.44,
        138.69, 139.19, 138.5,   138.13,  137.5,   138.88, 132.13, 129.75,  128.5,   125.44,
        125.12, 126.5,  128.69,  126.62,  126.69,  126,    123.12, 121.87,  124,     127,
        124.44, 122.5,  123.75,  123.81,  124.5,   127.87, 128.56, 129.63,  124.87,  124.37,
        124.87, 123.62, 124.06,  125.87,  125.19,  125.62, 126,    128.5,   126.75,  129.75,
        132.69, 133.94, 136.5,   137.69,  135.56,  133.56, 135,    132.38,  131.44,  130.88,
        129.63, 127.25, 127.81,  125,     126.81,  124.75, 122.81, 122.25,  121.06,  120,
        123.25, 122.75, 119.19,  115.06,  116.69,  114.87, 110.87, 107.25,  108.87,  109,
        108.5,  113.06, 93,      94.62,   95.12,   96,     95.56,  95.31,   99,      98.81,
        96.81,  95.94,  94.44,   92.94,   93.94,   95.5,   97.06,  97.5,    96.25,   96.37,
        95,     94.87,  98.25,   105.12,  108.44,  109.87, 105,    106,     104.94,  104.5,
        104.44, 106.31, 112.87,  116.5,   119.19,  121,    122.12, 111.94,  112.75,  110.19,
        107.94, 109.69, 111.06,  110.44,  110.12,  110.31, 110.44, 110,     110.75,  110.5,
        110.5,  109.5,
    };
}

// Low test data, 252 entries.
fn testLows() [252]f64 {
    return .{
        90.75,  91.405, 94.25,   93.5,   92.815,  93.5,   92,      89.75,   89.44,  90.625,
        92.75,  96.315, 96.03,   88.815, 86.75,   90.94,  88.905,  88.78,   89.25,  89.75,
        87.5,   86.53,  84.625,  82.28,  81.565,  80.875, 81.25,   84.065,  85.595, 85.97,
        84.405, 85.095, 85.5,    85.53,  87.875,  86.565, 84.655,  83.25,   82.565, 83.44,
        82.53,  85.065, 86.875,  88.53,  89.28,   90.125, 90.75,   89,      88.565, 90.095,
        89,     86.47,  84,      83.315, 82,      83.25,  84.75,   85.28,   87.19,  88.44,
        88.25,  87.345, 89.28,   91.095, 89.53,   91.155, 92,      90.53,   89.97,  88.815,
        86.75,  85.065, 82.03,   81.5,   82.565,  96.345, 96.47,   101.155, 104.25, 101.75,
        101.72, 101.72, 103.155, 105.69, 103.655, 104,    105.53,  108.53,  108.75, 107.75,
        117,    118,    116,     118.5,  116.53,  116.25, 114.595, 110.875, 110.5,  110.72,
        112.62, 114.19, 111.19,  109.44, 111.56,  112.44, 117.5,   116.06,  116.56, 113.31,
        112.56, 114,    114.75,  118.87, 119,     119.75, 122.62,  123,     121.75, 121.56,
        123.12, 122.19, 122.75,  124.37, 128,     129.5,  130.81,  130.63,  132.13, 133.88,
        135.38, 135.75, 136.19,  134.5,  135.38,  133.69, 126.06,  126.87,  123.5,  122.62,
        122.75, 123.56, 125.81,  124.62, 124.37,  121.81, 118.19,  118.06,  117.56, 121,
        121.12, 118.94, 119.81,  121,    122,     124.5,  126.56,  123.5,   121.25, 121.06,
        122.31, 121,    120.87,  122.06, 122.75,  122.69, 122.87,  125.5,   124.25, 128,
        128.38, 130.69, 131.63,  134.38, 132,     131.94, 131.94,  129.56,  123.75, 126,
        126.25, 124.37, 121.44,  120.44, 121.37,  121.69, 120,     119.62,  115.5,  116.75,
        119.06, 119.06, 115.06,  111.06, 113.12,  110,    105,     104.69,  103.87, 104.69,
        105.44, 107,    89,      92.5,   92.12,   94.62,  92.81,   94.25,   96.25,  96.37,
        93.69,  93.5,   90,      90.19,  90.5,    92.12,  94.12,   94.87,   93,     93.87,
        93,     92.62,  93.56,   98.37,  104.44,  106,    101.81,  104.12,  103.37, 102.12,
        102.25, 103.37, 107.94,  112.5,  115.44,  115.5,  112.25,  107.56,  106.56, 106.87,
        104.5,  105.75, 108.62,  107.75, 108.06,  108,    108.19,  108.12,  109.06, 108.75,
        108.56, 106.62,
    };
}

// Close test data, 252 entries.
fn testCloses() [252]f64 {
    return .{
        91.5,    94.815,  94.375,  95.095, 93.78,   94.625,  92.53,   92.75,   90.315,  92.47,
        96.125,  97.25,   98.5,    89.875, 91,      92.815,  89.155,  89.345,  91.625,  89.875,
        88.375,  87.625,  84.78,   83,     83.5,    81.375,  84.44,   89.25,   86.375,  86.25,
        85.25,   87.125,  85.815,  88.97,  88.47,   86.875,  86.815,  84.875,  84.19,   83.875,
        83.375,  85.5,    89.19,   89.44,  91.095,  90.75,   91.44,   89,      91,      90.5,
        89.03,   88.815,  84.28,   83.5,   82.69,   84.75,   85.655,  86.19,   88.94,   89.28,
        88.625,  88.5,    91.97,   91.5,   93.25,   93.5,    93.155,  91.72,   90,      89.69,
        88.875,  85.19,   83.375,  84.875, 85.94,   97.25,   99.875,  104.94,  106,     102.5,
        102.405, 104.595, 106.125, 106,    106.065, 104.625, 108.625, 109.315, 110.5,   112.75,
        123,     119.625, 118.75,  119.25, 117.94,  116.44,  115.19,  111.875, 110.595, 118.125,
        116,     116,     112,     113.75, 112.94,  116,     120.5,   116.62,  117,     115.25,
        114.31,  115.5,   115.87,  120.69, 120.19,  120.75,  124.75,  123.37,  122.94,  122.56,
        123.12,  122.56,  124.62,  129.25, 131,     132.25,  131,     132.81,  134,     137.38,
        137.81,  137.88,  137.25,  136.31, 136.25,  134.63,  128.25,  129,     123.87,  124.81,
        123,     126.25,  128.38,  125.37, 125.69,  122.25,  119.37,  118.5,   123.19,  123.5,
        122.19,  119.31,  123.31,  121.12, 123.37,  127.37,  128.5,   123.87,  122.94,  121.75,
        124.44,  122,     122.37,  122.94, 124,     123.19,  124.56,  127.25,  125.87,  128.86,
        132,     130.75,  134.75,  135,    132.38,  133.31,  131.94,  130,     125.37,  130.13,
        127.12,  125.19,  122,     125,    123,     123.5,   120.06,  121,     117.75,  119.87,
        122,     119.19,  116.37,  113.5,  114.25,  110,     105.06,  107,     107.87,  107,
        107.12,  107,     91,      93.94,  93.87,   95.5,    93,      94.94,   98.25,   96.75,
        94.81,   94.37,   91.56,   90.25,  93.94,   93.62,   97,      95,      95.87,   94.06,
        94.62,   93.75,   98,      103.94, 107.87,  106.06,  104.5,   105,     104.19,  103.06,
        103.42,  105.27,  111.87,  116,    116.62,  118.28,  113.37,  109,     109.7,   109.25,
        107,     109.19,  110,     109.2,  110.12,  108,     108.62,  109.75,  109.81,  109,
        108.75,  107.87,
    };
}

// Volume test data, 252 entries.
fn testVolumes() [252]f64 {
    return .{
        4077500,  4955900,  4775300,  4155300,  4593100,  3631300,  3382800,  4954200,  4500000,  3397500,
        4204500,  6321400,  10203600, 19043900, 11692000, 9553300,  8920300,  5970900,  5062300,  3705600,
        5865600,  5603000,  5811900,  8483800,  5995200,  5408800,  5430500,  6283800,  5834800,  4515500,
        4493300,  4346100,  3700300,  4600200,  4557200,  4323600,  5237500,  7404100,  4798400,  4372800,
        3872300,  10750800, 5804800,  3785500,  5014800,  3507700,  4298800,  4842500,  3952200,  3304700,
        3462000,  7253900,  9753100,  5953000,  5011700,  5910800,  4916900,  4135000,  4054200,  3735300,
        2921900,  2658400,  4624400,  4372200,  5831600,  4268600,  3059200,  4495500,  3425000,  3630800,
        4168100,  5966900,  7692800,  7362500,  6581300,  19587700, 10378600, 9334700,  10467200, 5671400,
        5645000,  4518600,  4519500,  5569700,  4239700,  4175300,  4995300,  4776600,  4190000,  6035300,
        12168900, 9040800,  5780300,  4320800,  3899100,  3221400,  3455500,  4304200,  4703900,  8316300,
        10553900, 6384800,  7163300,  7007800,  5114100,  5263800,  6666100,  7398400,  5575000,  4852300,
        4298100,  4900500,  4887700,  6964800,  4679200,  9165000,  6469800,  6792000,  4423800,  5231900,
        4565600,  6235200,  5225900,  8261400,  5912500,  3545600,  5714500,  6653900,  6094500,  4799200,
        5050800,  5648900,  4726300,  5585600,  5124800,  7630200,  14311600, 8793600,  8874200,  6966600,
        5525500,  6515500,  5291900,  5711700,  4327700,  4568000,  6859200,  5757500,  7367000,  6144100,
        4052700,  5849700,  5544700,  5032200,  4400600,  4894100,  5140000,  6610900,  7585200,  5963100,
        6045500,  8443300,  6464700,  6248300,  4357200,  4774700,  6216900,  6266900,  5584800,  5284500,
        7554500,  7209500,  8424800,  5094500,  4443600,  4591100,  5658400,  6094100,  14862200, 7544700,
        6985600,  8093000,  7590000,  7451300,  7078000,  7105300,  8778800,  6643900,  10563900, 7043100,
        6438900,  8057700,  14240000, 17872300, 7831100,  8277700,  15017800, 14183300, 13921100, 9683000,
        9187300,  11380500, 69447300, 26673600, 13768400, 11371600, 9872200,  9450500,  11083300, 9552800,
        11108400, 10374200, 16701900, 13741900, 8523600,  9551900,  8680500,  7151700,  9673100,  6264700,
        8541600,  8358000,  18720800, 19683100, 13682500, 10668100, 9710600,  3113100,  5682000,  5763600,
        5340000,  6220800,  14680500, 9933000,  11329500, 8145300,  16644700, 12593800, 7138100,  7442300,
        9442300,  7123600,  7680600,  4839800,  4775500,  4008800,  4533600,  3741100,  4084800,  2685200,
        3438000,  2870500,
    };
}

// Expected AD output, 252 entries.
fn testExpectedAD() [252]f64 {
    return .{
        -1631000.0000000000,  2974412.0226308300,   -1239087.9773691700,  -466727.3825736260,   -2276567.4139707900,
        789202.2581603610,    -1398341.7418396400,  1134914.1775558300,   -766075.9214540700,   1736082.9020753400,
        5640261.4735039100,   2890166.0052259700,   8972764.3897170800,   188856.2575697040,    10267283.4207680000,
        16222587.3168718000,  8293431.7613162900,   5920994.1613162700,   9699107.7642574600,   6364067.7642574600,
        4106481.8240114100,   4732197.5600520200,   -582684.7847755650,   -5272487.2692476200,  -3197626.7475084800,
        -7210607.3926697700,  -1780107.3926697700,  4138902.7196897800,   1901845.0121799000,   -1812171.6280736600,
        -3544134.5371645800,  297777.5278006210,    -2159221.6721993900,  2440978.3278006100,   435810.3278006080,
        -3070523.8185408500,  1465626.2894505000,   3007000.2671456800,   5345808.5051319400,   2847065.6479890900,
        2348057.9160303300,   -5679832.0403015400,  -1127705.6433843900,  -226396.1195748840,   519476.0323238490,
        346091.6216774670,    -1280481.3512955100,  -6122981.3512955100,  -2643824.5945387600,  -4331118.5522426800,
        -7585398.5522426800,  -693439.5085836460,   -9354192.3085836400,  -14168892.5669816000, -15370043.8066510000,
        -12414643.8066510000, -12585096.3399843000, -13890885.8136685000, -11450943.9372383000, -12314251.2598928000,
        -14359581.2598928000, -12924045.2598928000, -8299645.2598927900,  -11118563.6809454000, -6824818.5234191600,
        -4130032.7842887200,  -3708074.1635990700,  122721.5562452970,    -3201046.9166611000,  -4798598.9166611000,
        -3061890.5833277700,  -8623980.5426222100,  -11300142.4820162000, -3937642.4820161500,  2643657.5179838500,
        -5243139.1816861100,  -5243139.1816861100,  2167108.4486456300,   2554782.5227197100,   -686017.4772802920,
        -3549128.9880716500,  -1194197.2420399100,  3325302.7579600900,   -709635.4642621180,   2925560.1041386400,
        489968.4374719780,    5097434.5865403000,   3432561.5574946500,   5938908.5894581200,   9231738.4194426700,
        21400638.4194427000,  20195198.4194427000,  22748227.3840560000,  21998336.4749651000,  21964073.7333131000,
        19373669.6096018000,  17880959.5857354000,  15502576.1405453000,  11156172.5405452000,  19056383.7979514000,
        25289392.9744220000,  27933099.2244220000,  25099853.7020340000,  29753682.2734625000,  28403559.8734625000,
        33667359.8734625000,  36546987.7970902000,  31323445.5398723000,  26886728.6025174000,  27507371.6257732000,
        27027380.7628290000,  29477630.7628290000,  27665337.5044020000,  30340032.5614362000,  29372997.8947696000,
        31054151.7409234000,  33389894.2329362000,  27656016.3381994000,  27547234.3709863000,  25948598.2598752000,
        21382998.2598752000,  16993417.4598752000,  19402230.7411252000,  21708493.1264463000,  26513515.7850566000,
        30059115.7850566000,  25051948.0326135000,  28506411.7956798000,  29175540.8757985000,  33812970.0892816000,
        36178148.3370158000,  37524688.4532948000,  37135945.1632515000,  37120557.8354279000,  36201961.6090128000,
        31335687.6205735000,  27351074.1115125000,  31564674.1115125000,  24003855.7115125000,  27857719.5412998000,
        23497936.8408778000,  28905358.6095853000,  33058030.1373631000,  31630105.1373631000,  32227029.2752941000,
        28618418.2967738000,  25042729.8586399000,  20615046.1316058000,  26128856.6906120000,  25104840.0239454000,
        23664422.5540658000,  19030671.4304703000,  23336961.2781861000,  18734557.7194672000,  19157015.3194672000,
        22598860.1265889000,  27430460.1265889000,  21617613.3076656000,  21114727.1198203000,  17637753.4038082000,
        21652343.2475581000,  19654310.4231306000,  19269265.5955444000,  15907319.4013187000,  16014463.6636138000,
        12869354.1072998000,  13365911.6152870000,  14410394.9486203000,  16063495.7486204000,  15972904.3200490000,
        21108562.0926708000,  14165259.0157478000,  16535274.6214973000,  13349288.2166634000,  9854321.9245286000,
        13028415.7516891000,  7370015.7516890900,   3177620.7162280900,   -5422742.0926145400,  -197109.7155653600,
        -3586572.4374588300,  -7071058.5485699700,  -13326553.0540645000, -5875253.0540644700,  -8711657.4658291900,
        -7411341.1259599000,  -15815245.7522944000, -15486840.0488724000, -17500820.9841242000, -11021168.9841242000,
        -8424094.7597804800,  -15914043.5402683000, -21120435.7920843000, -17188529.7920843000, -20062126.7108518000,
        -28339826.7108518000, -43050618.8743952000, -31637494.6556452000, -23284834.6556452000, -22588377.5790791000,
        -21687661.8928045000, -33068161.8928045000, -33068161.8928045000, -23505927.9305405000, -21211194.5972072000,
        -18079884.4522797000, -26587925.9068252000, -23734944.7747498000, -18697081.1383861000, -25274418.8433042000,
        -28407557.3048426000, -31383762.2228754000, -36349191.9526051000, -49491445.4071505000, -40967845.4071505000,
        -42041727.6556712000, -33715533.7781202000, -40160221.6108198000, -32749138.8415890000, -38061604.4415890000,
        -32765812.4415890000, -32728665.7749224000, -16003686.6704448000, -3202381.6334077700,  6580605.8665922800,
        -3756700.3349581000,  2909824.4299321800,   2711115.9192938800,   2964453.4989116900,   1753613.1627772300,
        2119366.5874347800,   3939056.3833531200,   12663978.2900468000,  20113728.2900468000,  15914260.2900469000,
        16003118.1082287000,  3135938.6756046000,   -1177006.5298748600,  -1073221.5541074900,  2154763.9880611900,
        6436737.2438751500,   11752316.9393066000,  12759608.7425853000,  13137436.9953734000,  17912936.9953734000,
        13904136.9953734000,  11103379.6620401000,  13849506.2577848000,  13390268.3879623000,  11472268.3879623000,
        8707691.0683746700,   8328944.5405968900,
    };
}

test "AdvanceDecline with volume" {
    const highs = testHighs();
    const lows = testLows();
    const closes = testCloses();
    const volumes = testVolumes();
    const expected = testExpectedAD();

    var ad = AdvanceDecline.init();

    for (0..252) |i| {
        const v = ad.updateHlcv(highs[i], lows[i], closes[i], volumes[i]);
        try testing.expect(!math.isNan(v));
        try testing.expect(ad.isPrimed());

        const got = roundTo(v, 2);
        const exp = roundTo(expected[i], 2);
        try testing.expectEqual(exp, got);
    }
}

test "AdvanceDecline TA-Lib spot checks" {
    const highs = testHighs();
    const lows = testLows();
    const closes = testCloses();
    const volumes = testVolumes();

    var ad = AdvanceDecline.init();
    var values: [252]f64 = undefined;

    for (0..252) |i| {
        values[i] = ad.updateHlcv(highs[i], lows[i], closes[i], volumes[i]);
    }

    try testing.expectEqual(roundTo(-1631000.00, 2), roundTo(values[0], 2));
    try testing.expectEqual(roundTo(2974412.02, 2), roundTo(values[1], 2));
    try testing.expectEqual(roundTo(8707691.07, 2), roundTo(values[250], 2));
    try testing.expectEqual(roundTo(8328944.54, 2), roundTo(values[251], 2));
}

test "AdvanceDecline updateBar" {
    var ad = AdvanceDecline.init();

    const highs = testHighs();
    const lows = testLows();
    const closes = testCloses();
    const volumes = testVolumes();
    const expected = testExpectedAD();

    for (0..10) |i| {
        const bar = Bar{
            .time = 0,
            .open = highs[i],
            .high = highs[i],
            .low = lows[i],
            .close = closes[i],
            .volume = volumes[i],
        };

        const output = ad.updateBar(&bar);
        const scalar_val = output.slice()[0].scalar.value;

        const got = roundTo(scalar_val, 2);
        const exp = roundTo(expected[i], 2);
        try testing.expectEqual(exp, got);
    }
}

test "AdvanceDecline scalar update" {
    var ad = AdvanceDecline.init();

    // Scalar update: H=L=C, so range=0, AD unchanged (remains 0 after primed).
    const v = ad.update(100.0);
    try testing.expectEqual(@as(f64, 0), v);
    try testing.expect(ad.isPrimed());
}

test "AdvanceDecline NaN" {
    var ad = AdvanceDecline.init();

    try testing.expect(math.isNan(ad.update(math.nan(f64))));
    try testing.expect(math.isNan(ad.updateHlcv(math.nan(f64), 1, 2, 3)));
    try testing.expect(math.isNan(ad.updateHlcv(1, math.nan(f64), 2, 3)));
    try testing.expect(math.isNan(ad.updateHlcv(1, 2, math.nan(f64), 3)));
    try testing.expect(math.isNan(ad.updateHlcv(1, 2, 3, math.nan(f64))));
}

test "AdvanceDecline not primed initially" {
    var ad = AdvanceDecline.init();

    try testing.expect(!ad.isPrimed());
    try testing.expect(math.isNan(ad.update(math.nan(f64))));
    try testing.expect(!ad.isPrimed());
}

test "AdvanceDecline metadata" {
    var ad = AdvanceDecline.init();
    var meta: Metadata = undefined;
    ad.getMetadata(&meta);

    try testing.expectEqual(Identifier.advance_decline, meta.identifier);
    try testing.expectEqualStrings("ad", meta.mnemonic);
    try testing.expectEqualStrings("Advance-Decline", meta.description);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
}
