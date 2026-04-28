const std = @import("std");
const math = std.math;

const Bar = @import("bar").Bar;
const Quote = @import("quote").Quote;
const Trade = @import("trade").Trade;
const Scalar = @import("scalar").Scalar;
const bar_component = @import("bar_component");
const quote_component = @import("quote_component");
const trade_component = @import("trade_component");

const indicator_mod = @import("../../core/indicator.zig");
const line_indicator_mod = @import("../../core/line_indicator.zig");
const build_metadata_mod = @import("../../core/build_metadata.zig");
const component_triple_mnemonic_mod = @import("../../core/component_triple_mnemonic.zig");
const identifier_mod = @import("../../core/identifier.zig");
const metadata_mod = @import("../../core/metadata.zig");
const ema_mod = @import("../../common/exponential_moving_average/exponential_moving_average.zig");

const OutputArray = indicator_mod.OutputArray;
const LineIndicator = line_indicator_mod.LineIndicator;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;
const ExponentialMovingAverage = ema_mod.ExponentialMovingAverage;

/// Enumerates the outputs of the TRIX indicator.
pub const TripleExponentialMovingAverageOscillatorOutput = enum(u8) {
    /// The scalar value of the TRIX oscillator.
    value = 1,
};

/// Parameters to create an instance of the TRIX indicator.
pub const TripleExponentialMovingAverageOscillatorParams = struct {
    /// The length (number of time periods). Must be >= 1.
    length: usize,
    /// Bar component to extract. `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract. `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract. `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

/// Jack Hutson's Triple Exponential Moving Average Oscillator (TRIX).
///
/// TRIX is a 1-day rate-of-change of a triple-smoothed exponential moving average:
///
///   TRIX = ((EMA3[i] - EMA3[i-1]) / EMA3[i-1]) * 100
///
/// The indicator oscillates around zero. Positive values indicate upward momentum,
/// negative values indicate downward momentum.
pub const TripleExponentialMovingAverageOscillator = struct {
    line: LineIndicator,
    ema1: ExponentialMovingAverage,
    ema2: ExponentialMovingAverage,
    ema3: ExponentialMovingAverage,
    previous_ema3: f64,
    has_previous_ema: bool,
    primed: bool,
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub fn init(params: TripleExponentialMovingAverageOscillatorParams) !TripleExponentialMovingAverageOscillator {
        if (params.length < 1) {
            return error.InvalidLength;
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "trix({d}{s})", .{ params.length, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Triple exponential moving average oscillator {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const ema_params = ema_mod.ExponentialMovingAverageLengthParams{
            .length = params.length,
            .first_is_average = true,
        };

        var ema1 = try ExponentialMovingAverage.initLength(ema_params);
        ema1.fixSlices();
        var ema2 = try ExponentialMovingAverage.initLength(ema_params);
        ema2.fixSlices();
        var ema3 = try ExponentialMovingAverage.initLength(ema_params);
        ema3.fixSlices();

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .ema1 = ema1,
            .ema2 = ema2,
            .ema3 = ema3,
            .previous_ema3 = math.nan(f64),
            .has_previous_ema = false,
            .primed = false,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn fixSlices(self: *TripleExponentialMovingAverageOscillator) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
        // Internal EMAs don't need fixSlices since we don't use their mnemonics.
    }

    /// Core update logic.
    pub fn update(self: *TripleExponentialMovingAverageOscillator, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        const v1 = self.ema1.update(sample);
        if (math.isNan(v1)) return math.nan(f64);

        const v2 = self.ema2.update(v1);
        if (math.isNan(v2)) return math.nan(f64);

        const v3 = self.ema3.update(v2);
        if (math.isNan(v3)) return math.nan(f64);

        if (!self.has_previous_ema) {
            self.previous_ema3 = v3;
            self.has_previous_ema = true;
            return math.nan(f64);
        }

        const result = ((v3 - self.previous_ema3) / self.previous_ema3) * 100.0;
        self.previous_ema3 = v3;

        if (!self.primed) {
            self.primed = true;
        }

        return result;
    }

    pub fn isPrimed(self: *const TripleExponentialMovingAverageOscillator) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const TripleExponentialMovingAverageOscillator, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .triple_exponential_moving_average_oscillator,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *TripleExponentialMovingAverageOscillator, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *TripleExponentialMovingAverageOscillator, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *TripleExponentialMovingAverageOscillator, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *TripleExponentialMovingAverageOscillator, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *TripleExponentialMovingAverageOscillator) indicator_mod.Indicator {
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
        const self: *TripleExponentialMovingAverageOscillator = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const TripleExponentialMovingAverageOscillator = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *TripleExponentialMovingAverageOscillator = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *TripleExponentialMovingAverageOscillator = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *TripleExponentialMovingAverageOscillator = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *TripleExponentialMovingAverageOscillator = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;

fn createTrix(length: usize) !TripleExponentialMovingAverageOscillator {
    var trix = try TripleExponentialMovingAverageOscillator.init(.{ .length = length });
    trix.fixSlices();
    return trix;
}

fn almostEqual(a: f64, b: f64, tolerance: f64) bool {
    return @abs(a - b) <= tolerance;
}

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

fn testExpected() [252]f64 {
    return .{
        math.nan(f64),            math.nan(f64),            math.nan(f64),            math.nan(f64),            math.nan(f64),
        math.nan(f64),            math.nan(f64),            math.nan(f64),            math.nan(f64),            math.nan(f64),
        math.nan(f64),            math.nan(f64),            math.nan(f64),            2.58962720696066184e-01,  1.04946633050522244e-02,
        -1.07946280630203409e-01, -2.98063213610859579e-01, -4.42746334131551789e-01, -4.30599171391310154e-01, -4.28687591865216655e-01,
        -4.75069823235181321e-01, -5.37433840276123731e-01, -6.84986079749266730e-01, -8.69836447385321843e-01, -9.69285259665234378e-01,
        -1.07266083502368459e+00, -9.85962305677537620e-01, -6.23333572107235390e-01, -3.67763648819158684e-01, -2.02182398005543107e-01,
        -1.41898857279466523e-01, -4.27772161723139685e-02, -1.25621377616182018e-02, 1.25618688579349935e-01,  2.33491688393708319e-01,
        2.26884410158885358e-01,  1.77283197977560342e-01,  3.85167494517237516e-02,  -1.21186481780683114e-01, -2.55053582263658374e-01,
        -3.59199918906476401e-01, -3.23581501921914039e-01, -8.40813672802582790e-02, 1.68450568659050198e-01,  4.23606129456622071e-01,
        5.80760614825099775e-01,  6.74423245908670088e-01,  5.95054583997587705e-01,  5.48385036268770665e-01,  4.81655497491729589e-01,
        3.47132806917158143e-01,  2.06273946077274561e-01,  -9.55176536428308937e-02, -3.95956186981914759e-01, -6.39556567499502515e-01,
        -6.96069099622400933e-01, -6.18780210801841646e-01, -4.75349956278855568e-01, -2.13580887134896868e-01, 3.89103698839637843e-02,
        1.93263990458743595e-01,  2.65410590578911643e-01,  4.32220997131950946e-01,  5.46114679436054296e-01,  6.67104647605417655e-01,
        7.42428207780259752e-01,  7.41181968824710657e-01,  6.30286542745132516e-01,  4.22657469374050521e-01,  2.13201875998381390e-01,
        1.66920006824442073e-02,  -2.76064370450727792e-01, -5.86976444719135571e-01, -7.42923087506186630e-01, -7.42816682545351159e-01,
        -1.96753404734580251e-01, 4.91439970841801577e-01,  1.22936584008139516e+00,  1.77090447366538251e+00,  1.88835285764635019e+00,
        1.77284269884264445e+00,  1.63836435177078799e+00,  1.52826351203718169e+00,  1.38905067176893882e+00,  1.22692744884666860e+00,
        1.00114584212491131e+00,  9.25117393795746978e-01,  9.00970052761961182e-01,  9.05368830978650063e-01,  9.61116008872658578e-01,
        1.33855797173158098e+00,  1.53214679247618757e+00,  1.52774085173358332e+00,  1.42421451188451065e+00,  1.22106584858479339e+00,
        9.48153586076694976e-01,  6.51803048386640826e-01,  2.95918548613889665e-01,  -3.62457733471194254e-02, -3.26907614280814232e-02,
        2.12792456265050323e-03,  3.71100994960601102e-02,  -6.68646520502907404e-02, -1.26410997023991278e-01, -1.78882630512154583e-01,
        -1.10531624325180419e-01, 1.25585605695406416e-01,  2.19994223991791360e-01,  2.51706655321016659e-01,  1.91021168666089175e-01,
        8.18252982065182655e-02,  2.17972395096944424e-02,  2.87113685950853665e-03,  1.55155351229340627e-01,  2.96478405291518310e-01,
        4.03948668739852690e-01,  5.84597360648575481e-01,  6.72279360176057939e-01,  6.70070616386483575e-01,  6.04851783955095423e-01,
        5.32976523442800443e-01,  4.40690147966624979e-01,  4.13281343400153101e-01,  5.37588419613239998e-01,  7.05476862178885722e-01,
        8.51232865063755173e-01,  8.80574727800292556e-01,  8.87837630859612315e-01,  8.87104829025468278e-01,  9.48035765431419875e-01,
        9.85294600874028759e-01,  9.68971422876985455e-01,  8.86009653158735144e-01,  7.45587475671371513e-01,  5.95929041204476206e-01,
        4.15842517769955400e-01,  8.30170062395004099e-02,  -1.90514398908667060e-01, -5.11164852599737207e-01, -7.20970937521774435e-01,
        -8.70165969625510405e-01, -8.42005874287407363e-01, -6.76385710570910459e-01, -5.72208143287373794e-01, -4.83359126356914304e-01,
        -5.03671854850489376e-01, -6.17408228626891153e-01, -7.34295316436145984e-01, -6.58979307861029762e-01, -5.11333633989799696e-01,
        -3.99694860917008066e-01, -3.99042175478594219e-01, -2.97483676300644151e-01, -2.46646581161520329e-01, -1.47133627857009563e-01,
        6.76367355418673044e-02,  2.92337175775879643e-01,  3.10132235643382881e-01,  2.22879302225391512e-01,  8.37024450626865646e-02,
        4.28654387525587155e-02,  -3.29810235965555046e-02, -8.72780154622277288e-02, -1.00891413629615179e-01, -6.33978732831631253e-02,
        -4.23680684137147032e-02, 1.10261940800207067e-02,  1.40513702872840152e-01,  2.12042477635689453e-01,  3.29198387507458612e-01,
        5.10015858722597470e-01,  6.04673430899255737e-01,  7.43672061141589236e-01,  8.36554796407949941e-01,  7.82403707382248736e-01,
        6.97102273214048274e-01,  5.59641397872549407e-01,  3.67134639147453312e-01,  6.42807006875088793e-02,  -5.99320597209111253e-02,
        -1.82560443235562558e-01, -3.21398101636197864e-01, -5.08938437226262974e-01, -5.59733371959351578e-01, -5.94015579733394916e-01,
        -5.78985203119367964e-01, -6.32279260492530248e-01, -6.42853120947694867e-01, -7.10793043685214010e-01, -6.91019287495097445e-01,
        -5.61193612676043285e-01, -4.95066007153342480e-01, -5.32478884507267658e-01, -6.59940959362007806e-01, -7.36372123625530062e-01,
        -8.83221212890419216e-01, -1.13798656358759320e+00, -1.26364612600900417e+00, -1.24797331187935945e+00, -1.17736713273478877e+00,
        -1.06237491602584222e+00, -9.28915212854969918e-01, -1.32329028490183109e+00, -1.64132964315893726e+00, -1.80819327477252934e+00,
        -1.77148532949395676e+00, -1.70885356502759422e+00, -1.52562851931094889e+00, -1.17489645643349250e+00, -8.72620010128195123e-01,
        -7.02895489251532668e-01, -6.08201266369661098e-01, -6.45386057647670208e-01, -7.40494086178827615e-01, -6.61030511591078063e-01,
        -5.37163777839567036e-01, -2.79902094107505939e-01, -1.16326826827063620e-01, 1.51949717990906459e-02,  3.14846011562785882e-02,
        3.01394979933595451e-02,  -1.12564815893076171e-02, 1.13045980249598277e-01,  4.82076435266708736e-01,  9.59350350043020383e-01,
        1.23005820083969297e+00,  1.25831508252485191e+00,  1.17882101225620572e+00,  1.01387822086194146e+00,  7.88219340251123057e-01,
        5.88525242439529306e-01,  4.94021254182158809e-01,  6.74547470578328867e-01,  1.00852388193493825e+00,  1.27771792173385368e+00,
        1.46100645443146226e+00,  1.34019672118515776e+00,  9.64864568849165827e-01,  6.05482522430805470e-01,  3.07805301570223622e-01,
        1.89910160436106944e-02,  -1.20048717679914985e-01, -1.47573924217660940e-01, -1.56298732605740176e-01, -1.20517838432609992e-01,
        -1.47611315182772818e-01, -1.60147519194369287e-01, -1.21101023753211398e-01, -7.02606434189095713e-02, -5.38583141622932041e-02,
        -5.84078945373598668e-02, -9.56468061647420681e-02,
    };
}

test "trix values" {
    const tolerance = 1e-10;
    const closes = testCloses();
    const expected = testExpected();

    var trix = try createTrix(5);

    for (closes, expected, 0..) |c, exp, i| {
        const result = trix.update(c);
        _ = i;

        if (math.isNan(exp)) {
            try testing.expect(math.isNan(result));
        } else {
            try testing.expect(!math.isNan(result));
            try testing.expect(almostEqual(result, exp, tolerance));
        }
    }
}

test "trix is primed" {
    const closes = testCloses();

    var trix = try createTrix(5);

    // Lookback = 3*(5-1) + 1 = 13. First primed at index 13.
    for (0..13) |i| {
        _ = trix.update(closes[i]);
        try testing.expect(!trix.isPrimed());
    }

    _ = trix.update(closes[13]);
    try testing.expect(trix.isPrimed());
}

test "trix metadata" {
    var trix = try TripleExponentialMovingAverageOscillator.init(.{ .length = 30 });
    trix.fixSlices();

    var m: Metadata = undefined;
    trix.getMetadata(&m);

    try testing.expectEqual(Identifier.triple_exponential_moving_average_oscillator, m.identifier);
    try testing.expectEqualStrings("trix(30)", m.mnemonic);
    try testing.expectEqualStrings("Triple exponential moving average oscillator trix(30)", m.description);
    try testing.expectEqual(@as(usize, 1), m.outputs_len);
    try testing.expectEqual(@as(i32, 1), m.outputs_buf[0].kind);
}

test "trix invalid params" {
    const result = TripleExponentialMovingAverageOscillator.init(.{ .length = 0 });
    try testing.expect(result == error.InvalidLength);
}

test "trix NaN" {
    var trix = try createTrix(5);
    const result = trix.update(math.nan(f64));
    try testing.expect(math.isNan(result));
}

test "trix update entity" {
    var trix = try createTrix(5);
    const time: i64 = 1617235200;

    const bar = Bar{ .time = time, .open = 0, .high = 0, .low = 0, .close = 100.0, .volume = 0 };
    const out = trix.updateBar(&bar);
    try testing.expectEqual(@as(usize, 1), out.len);
}
