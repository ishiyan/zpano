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
const metadata_mod = @import("../../core/metadata.zig");

const OutputArray = indicator_mod.OutputArray;
const LineIndicator = line_indicator_mod.LineIndicator;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the Kaufman Adaptive Moving Average indicator.
pub const KaufmanAdaptiveMovingAverageOutput = enum(u8) {
    /// The scalar value of the moving average.
    value = 1,
};

/// Parameters to create an instance of the indicator based on lengths.
pub const KaufmanAdaptiveMovingAverageLengthParams = struct {
    /// Efficiency ratio length. Must be >= 2.
    efficiency_ratio_length: u32 = 10,
    /// Fastest boundary length. Must be >= 2.
    fastest_length: u32 = 2,
    /// Slowest boundary length. Must be >= 2.
    slowest_length: u32 = 30,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Parameters to create an instance of the indicator based on smoothing factors.
pub const KaufmanAdaptiveMovingAverageSmoothingFactorParams = struct {
    /// Efficiency ratio length. Must be >= 2.
    efficiency_ratio_length: u32 = 10,
    /// Fastest smoothing factor in [0, 1].
    fastest_smoothing_factor: f64 = 2.0 / 3.0,
    /// Slowest smoothing factor in [0, 1].
    slowest_smoothing_factor: f64 = 2.0 / 31.0,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Kaufman's Adaptive Moving Average (KAMA).
///
/// An EMA with the smoothing factor adapted by the efficiency ratio.
pub const KaufmanAdaptiveMovingAverage = struct {
    line: LineIndicator,
    efficiency_ratio_length: u32,
    window: []f64,
    absolute_delta: []f64,
    absolute_delta_sum: f64,
    alpha_fastest: f64,
    alpha_slowest: f64,
    alpha_diff: f64,
    value: f64,
    efficiency_ratio: f64,
    window_count: u32,
    primed: bool,
    allocator: std.mem.Allocator,
    mnemonic_buf: [128]u8,
    mnemonic_len: usize,

    pub fn initLength(allocator: std.mem.Allocator, params: KaufmanAdaptiveMovingAverageLengthParams) !KaufmanAdaptiveMovingAverage {
        if (params.efficiency_ratio_length < 2) return error.InvalidEfficiencyRatioLength;
        if (params.fastest_length < 2) return error.InvalidFastestLength;
        if (params.slowest_length < 2) return error.InvalidSlowestLength;

        const fastest_alpha = 2.0 / @as(f64, @floatFromInt(1 + params.fastest_length));
        const slowest_alpha = 2.0 / @as(f64, @floatFromInt(1 + params.slowest_length));

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [128]u8 = undefined;
        const mnemonic = std.fmt.bufPrint(&mnemonic_buf, "kama({d}, {d}, {d}{s})", .{
            params.efficiency_ratio_length, params.fastest_length, params.slowest_length, triple,
        }) catch unreachable;
        const mnemonic_len = mnemonic.len;

        return initInternal(allocator, params.efficiency_ratio_length, fastest_alpha, slowest_alpha, params.bar_component, params.quote_component, params.trade_component, mnemonic_buf, mnemonic_len);
    }

    pub fn initSmoothingFactor(allocator: std.mem.Allocator, params: KaufmanAdaptiveMovingAverageSmoothingFactorParams) !KaufmanAdaptiveMovingAverage {
        if (params.efficiency_ratio_length < 2) return error.InvalidEfficiencyRatioLength;
        if (params.fastest_smoothing_factor < 0.0 or params.fastest_smoothing_factor > 1.0) return error.InvalidFastestSmoothingFactor;
        if (params.slowest_smoothing_factor < 0.0 or params.slowest_smoothing_factor > 1.0) return error.InvalidSlowestSmoothingFactor;

        const epsilon = 0.00000001;
        var fastest_alpha = params.fastest_smoothing_factor;
        var slowest_alpha = params.slowest_smoothing_factor;
        if (fastest_alpha < epsilon) fastest_alpha = epsilon;
        if (slowest_alpha < epsilon) slowest_alpha = epsilon;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [128]u8 = undefined;
        const mnemonic = std.fmt.bufPrint(&mnemonic_buf, "kama({d}, {d:.4}, {d:.4}{s})", .{
            params.efficiency_ratio_length, fastest_alpha, slowest_alpha, triple,
        }) catch unreachable;
        const mnemonic_len = mnemonic.len;

        return initInternal(allocator, params.efficiency_ratio_length, fastest_alpha, slowest_alpha, params.bar_component, params.quote_component, params.trade_component, mnemonic_buf, mnemonic_len);
    }

    fn initInternal(
        allocator: std.mem.Allocator,
        efficiency_ratio_length: u32,
        fastest_alpha: f64,
        slowest_alpha: f64,
        bc_opt: ?bar_component.BarComponent,
        qc_opt: ?quote_component.QuoteComponent,
        tc_opt: ?trade_component.TradeComponent,
        mnemonic_buf: [128]u8,
        mnemonic_len: usize,
    ) !KaufmanAdaptiveMovingAverage {
        const buf_len = efficiency_ratio_length + 1;
        const window = try allocator.alloc(f64, buf_len);
        @memset(window, 0);
        const absolute_delta = try allocator.alloc(f64, buf_len);
        @memset(absolute_delta, 0);

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                "Kaufman adaptive moving average ",
                bc_opt,
                qc_opt,
                tc_opt,
            ),
            .efficiency_ratio_length = efficiency_ratio_length,
            .window = window,
            .absolute_delta = absolute_delta,
            .absolute_delta_sum = 0,
            .alpha_fastest = fastest_alpha,
            .alpha_slowest = slowest_alpha,
            .alpha_diff = fastest_alpha - slowest_alpha,
            .value = math.nan(f64),
            .efficiency_ratio = math.nan(f64),
            .window_count = 0,
            .primed = false,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
        };
    }

    pub fn deinit(self: *KaufmanAdaptiveMovingAverage) void {
        self.allocator.free(self.window);
        self.allocator.free(self.absolute_delta);
    }

    pub fn fixSlices(self: *KaufmanAdaptiveMovingAverage) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
    }

    pub fn update(self: *KaufmanAdaptiveMovingAverage, sample: f64) f64 {
        if (math.isNan(sample)) return sample;

        const epsilon = 0.00000001;
        const er_len = self.efficiency_ratio_length;

        if (self.primed) {
            var temp = @abs(sample - self.window[er_len]);
            self.absolute_delta_sum += temp - self.absolute_delta[1];

            // Shift window and absolute_delta left by 1.
            for (0..er_len) |i| {
                const j = i + 1;
                self.window[i] = self.window[j];
                self.absolute_delta[i] = self.absolute_delta[j];
            }

            self.window[er_len] = sample;
            self.absolute_delta[er_len] = temp;
            const delta = @abs(sample - self.window[0]);

            if (self.absolute_delta_sum <= delta or self.absolute_delta_sum < epsilon) {
                temp = 1.0;
            } else {
                temp = delta / self.absolute_delta_sum;
            }

            self.efficiency_ratio = temp;
            temp = self.alpha_slowest + temp * self.alpha_diff;
            self.value += (sample - self.value) * temp * temp;

            return self.value;
        } else {
            self.window[self.window_count] = sample;
            if (self.window_count > 0) {
                const temp = @abs(sample - self.window[self.window_count - 1]);
                self.absolute_delta[self.window_count] = temp;
                self.absolute_delta_sum += temp;
            }

            if (er_len == self.window_count) {
                self.primed = true;
                const delta = @abs(sample - self.window[0]);

                var temp: f64 = undefined;
                if (self.absolute_delta_sum <= delta or self.absolute_delta_sum < epsilon) {
                    temp = 1.0;
                } else {
                    temp = delta / self.absolute_delta_sum;
                }

                self.efficiency_ratio = temp;
                temp = self.alpha_slowest + temp * self.alpha_diff;
                self.value = self.window[er_len - 1];
                self.value += (sample - self.value) * temp * temp;

                return self.value;
            } else {
                self.window_count += 1;
            }
        }

        return math.nan(f64);
    }

    pub fn isPrimed(self: *const KaufmanAdaptiveMovingAverage) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const KaufmanAdaptiveMovingAverage, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .kaufman_adaptive_moving_average,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *KaufmanAdaptiveMovingAverage, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *KaufmanAdaptiveMovingAverage, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *KaufmanAdaptiveMovingAverage, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *KaufmanAdaptiveMovingAverage, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *KaufmanAdaptiveMovingAverage) indicator_mod.Indicator {
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
        const self: *KaufmanAdaptiveMovingAverage = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const KaufmanAdaptiveMovingAverage = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *KaufmanAdaptiveMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *KaufmanAdaptiveMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *KaufmanAdaptiveMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *KaufmanAdaptiveMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const Error = error{
        InvalidEfficiencyRatioLength,
        InvalidFastestLength,
        InvalidSlowestLength,
        InvalidFastestSmoothingFactor,
        InvalidSlowestSmoothingFactor,
    } || std.mem.Allocator.Error;
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;

fn testInput() [252]f64 {
    return .{
        91.500000,  94.815000,  94.375000,  95.095000,  93.780000,  94.625000,  92.530000,  92.750000,  90.315000,  92.470000,
        96.125000,  97.250000,  98.500000,  89.875000,  91.000000,  92.815000,  89.155000,  89.345000,  91.625000,  89.875000,
        88.375000,  87.625000,  84.780000,  83.000000,  83.500000,  81.375000,  84.440000,  89.250000,  86.375000,  86.250000,
        85.250000,  87.125000,  85.815000,  88.970000,  88.470000,  86.875000,  86.815000,  84.875000,  84.190000,  83.875000,
        83.375000,  85.500000,  89.190000,  89.440000,  91.095000,  90.750000,  91.440000,  89.000000,  91.000000,  90.500000,
        89.030000,  88.815000,  84.280000,  83.500000,  82.690000,  84.750000,  85.655000,  86.190000,  88.940000,  89.280000,
        88.625000,  88.500000,  91.970000,  91.500000,  93.250000,  93.500000,  93.155000,  91.720000,  90.000000,  89.690000,
        88.875000,  85.190000,  83.375000,  84.875000,  85.940000,  97.250000,  99.875000,  104.940000, 106.000000, 102.500000,
        102.405000, 104.595000, 106.125000, 106.000000, 106.065000, 104.625000, 108.625000, 109.315000, 110.500000, 112.750000,
        123.000000, 119.625000, 118.750000, 119.250000, 117.940000, 116.440000, 115.190000, 111.875000, 110.595000, 118.125000,
        116.000000, 116.000000, 112.000000, 113.750000, 112.940000, 116.000000, 120.500000, 116.620000, 117.000000, 115.250000,
        114.310000, 115.500000, 115.870000, 120.690000, 120.190000, 120.750000, 124.750000, 123.370000, 122.940000, 122.560000,
        123.120000, 122.560000, 124.620000, 129.250000, 131.000000, 132.250000, 131.000000, 132.810000, 134.000000, 137.380000,
        137.810000, 137.880000, 137.250000, 136.310000, 136.250000, 134.630000, 128.250000, 129.000000, 123.870000, 124.810000,
        123.000000, 126.250000, 128.380000, 125.370000, 125.690000, 122.250000, 119.370000, 118.500000, 123.190000, 123.500000,
        122.190000, 119.310000, 123.310000, 121.120000, 123.370000, 127.370000, 128.500000, 123.870000, 122.940000, 121.750000,
        124.440000, 122.000000, 122.370000, 122.940000, 124.000000, 123.190000, 124.560000, 127.250000, 125.870000, 128.860000,
        132.000000, 130.750000, 134.750000, 135.000000, 132.380000, 133.310000, 131.940000, 130.000000, 125.370000, 130.130000,
        127.120000, 125.190000, 122.000000, 125.000000, 123.000000, 123.500000, 120.060000, 121.000000, 117.750000, 119.870000,
        122.000000, 119.190000, 116.370000, 113.500000, 114.250000, 110.000000, 105.060000, 107.000000, 107.870000, 107.000000,
        107.120000, 107.000000, 91.000000,  93.940000,  93.870000,  95.500000,  93.000000,  94.940000,  98.250000,  96.750000,
        94.810000,  94.370000,  91.560000,  90.250000,  93.940000,  93.620000,  97.000000,  95.000000,  95.870000,  94.060000,
        94.620000,  93.750000,  98.000000,  103.940000, 107.870000, 106.060000, 104.500000, 105.000000, 104.190000, 103.060000,
        103.420000, 105.270000, 111.870000, 116.000000, 116.620000, 118.280000, 113.370000, 109.000000, 109.700000, 109.250000,
        107.000000, 109.190000, 110.000000, 109.200000, 110.120000, 108.000000, 108.620000, 109.750000, 109.810000, 109.000000,
        108.750000, 107.870000,
    };
}

fn testExpected() [252]f64 {
    return .{
        math.nan(f64),     math.nan(f64),     math.nan(f64),     math.nan(f64),     math.nan(f64),
        math.nan(f64),     math.nan(f64),     math.nan(f64),     math.nan(f64),     math.nan(f64),
        92.6574744421924,  92.7783471257434,  93.0592520064115,  92.9356368995325,  92.9000149644911,
        92.8990048732289,  92.8229942018608,  92.7516051928620,  92.7414384525517,  92.6960363223993,
        92.3934372123882,  91.9139380062599,  90.7658162726830,  90.0740111936089,  89.3620815288014,
        87.6656280861040,  87.4895131032692,  87.4974604839614,  87.4487997113532,  87.4134797590652,
        87.3586513546248,  87.3571985565411,  87.3428271277309,  87.4342339727455,  87.4790967331831,
        87.4478089486627,  87.4341052772180,  87.2779545841798,  87.1866387951289,  87.0799098978843,
        86.9861110535034,  86.9549433796085,  87.0479997922396,  87.0668566957271,  87.2090146571776,
        87.4600776240503,  87.8014795040326,  87.8826076877600,  88.2803844203263,  88.5454141018648,
        88.5859031486005,  88.5965040436874,  88.2719621445720,  87.8163354339468,  86.8611444903465,
        86.6741610056912,  86.5906930013157,  86.5766752991618,  86.6296450514704,  86.6650208354184,
        86.6783504731998,  86.6895963952268,  87.6981988794437,  88.5095835057360,  89.9508715587081,
        90.9585930437125,  91.4794679492180,  91.5092409530174,  91.4856744284233,  91.4717808315536,
        91.4557387469302,  91.1940009725015,  89.4266294004067,  88.8455374050859,  88.3697094609281,
        88.5930899916723,  89.1316678888979,  90.8601116442358,  93.2091460910382,  94.0581656977510,
        94.9201636069605,  96.8889752566530,  99.4062425239817,  101.1201449462390, 102.3769237660390,
        102.6006738368170, 103.3003850710980, 103.6578508957870, 104.0764855627630, 106.4159093020280,
        112.1346727325330, 113.5057358502340, 114.2548283428500, 115.0085673230990, 115.3491682211620,
        115.4744042357010, 115.4586954188130, 115.4033778968360, 115.3819703222920, 115.4596680866820,
        115.4927139908920, 115.5083211482970, 115.3016588863670, 115.2382416224770, 115.1532481002890,
        115.1580191296150, 115.3257950434630, 115.3602912952500, 115.4272550190370, 115.4236654978450,
        115.4094918992810, 115.4100431369950, 115.4265778341240, 115.7744740794160, 116.0930627623780,
        116.3101967717570, 116.6603109196670, 117.3487018143020, 117.8153888221880, 118.4531290804430,
        119.3499419409230, 119.8086689971510, 120.6175024210070, 122.0458817467430, 123.9704416533650,
        125.8138480326600, 126.3738969105690, 127.6872486354350, 129.2393432164220, 131.6880947713340,
        133.5239638088170, 135.0004207395880, 135.6288233403940, 135.7374059656390, 135.8007904215550,
        135.7583248045180, 135.5543718432480, 135.2569852680960, 133.6204824276490, 131.3192797761920,
        128.7932379609940, 128.4062405870340, 128.4039316032540, 128.0791656483760, 127.8414201748350,
        127.1988985844810, 126.5381546649790, 125.6607070438540, 125.6440698902700, 125.6229493897650,
        125.5972771029140, 125.1856884028260, 125.1156207098550, 124.9914050152240, 124.9677440635400,
        125.0508437113440, 125.3554407671800, 125.3059272985400, 125.2940386783170, 125.2530757692210,
        125.2419747210570, 125.1887237516160, 125.1656598262800, 125.1342643444030, 125.1261708430550,
        125.0293527295390, 125.0082100078360, 125.1058124672220, 125.1321388339230, 125.5284397017590,
        126.2554117345480, 126.9803557764160, 128.5646940398630, 129.8559054638140, 130.0995104273400,
        130.5156892070650, 130.6273781337970, 130.6136632314180, 130.5821372483140, 130.5780360175850,
        130.4619826221790, 130.2592097652620, 129.0901503140520, 128.7592330158310, 128.3218396854650,
        127.9194919253990, 127.1326782278630, 126.7107330400510, 126.1909025410680, 125.5077119513560,
        125.3652360592940, 125.0689417277010, 124.6785367307510, 123.1715118076970, 122.3246069304410,
        120.4996045001390, 118.0226226271800, 116.5389084881180, 115.7700047414230, 114.4762055991300,
        112.8691910705370, 111.7330463494810, 105.8813879559000, 103.7386265802100, 101.7705073498860,
        100.9556429673090, 100.0740835866110, 99.5051792798608,  99.4197548401710,  99.2260466472373,
        98.8377738185378,  98.4351675572326,  98.3887252314702,  98.0891751313173,  98.0708172638065,
        98.0047820815841,  97.9717872707032,  97.9587393847739,  97.9160266616328,  97.8272391679346,
        97.8109932013579,  97.7811643727499,  97.7968786191168,  98.8421055702164,  100.3972096134300,
        101.1278312905150, 101.3486183367770, 101.7632588756100, 101.9699249107700, 102.0803180404650,
        102.2131955779830, 102.6495717799380, 104.1660350536590, 105.9174582846280, 107.1295132390960,
        109.3610815395210, 109.7246822740860, 109.7071337912410, 109.7068748325140, 109.6867591775540,
        109.6319778699710, 109.6221417907160, 109.6271816752350, 109.5930223785590, 109.6314010730650,
        109.3937985883840, 109.3445353771140, 109.3487688924230, 109.3510517081720, 109.3489501843720,
        109.3310159853090, 109.2940150671190,
    };
}

fn testExpectedEr() [252]f64 {
    return .{
        math.nan(f64),       math.nan(f64),       math.nan(f64),       math.nan(f64),       math.nan(f64),
        math.nan(f64),       math.nan(f64),       math.nan(f64),       math.nan(f64),       math.nan(f64),
        0.26897353881942400, 0.16227924025324900, 0.26082832753714800, 0.22006745362563200, 0.11814704632384200,
        0.07387755102040830, 0.12948398235181300, 0.13078548108315700, 0.05061823802163840, 0.10186457311089300,
        0.33233276157804500, 0.41948136848986700, 0.55908720456397700, 0.38852783272110800, 0.43936731107205600,
        0.65822784810126600, 0.28090557044980700, 0.00443821537024054, 0.23863636363636400, 0.17791411042944800,
        0.15723270440251600, 0.02380952380952380, 0.05317236064731550, 0.28646833013435700, 0.23848368522072900,
        0.27080256031511600, 0.13724357122219000, 0.30308278489781800, 0.17844017966517000, 0.19099316445516700,
        0.15710096355257600, 0.13336068937217900, 0.23171987641606600, 0.04030874785591760, 0.20483808037456100,
        0.33506268914829200, 0.37925379253792600, 0.32493107522646700, 0.48608137044967900, 0.46671363156040900,
        0.37289812067260200, 0.25009430403621300, 0.34822695035461000, 0.40601503759398500, 0.60972071091766400,
        0.38709677419354800, 0.36811963092586700, 0.20347574221578600, 0.14148351648351700, 0.08472222222222210,
        0.02981229297018780, 0.02334197851055930, 0.61866452131938800, 0.66006600660066000, 0.80857580398162300,
        0.77777777777777800, 0.70159027128157200, 0.47713546160483200, 0.10037878787878800, 0.03893637226970530,
        0.02338634237605240, 0.23228070175438600, 0.68241365621278300, 0.48623853211009200, 0.56491499227202500,
        0.15625000000000000, 0.25570776255707800, 0.44199264460046800, 0.54700854700854700, 0.39488286066584500,
        0.42654476670870100, 0.64201819685690700, 0.75985303941215800, 0.73954139681428300, 0.73009250861599900,
        0.41678440237355200, 0.45883586785527000, 0.29772031303164400, 0.30364372469635600, 0.75534266764922700,
        0.86807165437302400, 0.60337213970293100, 0.52051123479694900, 0.53796183516037400, 0.45893719806763300,
        0.45556198187777100, 0.28315721371576400, 0.09918636187524220, 0.00366724570546222, 0.17235850569183900,
        0.30355594102341700, 0.18415036830073700, 0.29592284085927200, 0.22859517871986700, 0.21222410865874400,
        0.01751592356687890, 0.18716954529432500, 0.16398824952479700, 0.22846441947565500, 0.12918445293192500,
        0.08020882771713330, 0.02246181491464510, 0.20772946859903400, 0.31981566820276500, 0.33894343151005200,
        0.25145579671784000, 0.23110386079391000, 0.42479546884833300, 0.37264742785445400, 0.50171585449553900,
        0.62085976039464400, 0.52064896755162200, 0.57377049180327900, 0.56839309428950900, 0.66278356836296800,
        0.67647058823529400, 0.43859649122807000, 0.64305177111716600, 0.71632124352331600, 0.80368763557483700,
        0.80229382850901100, 0.85970819304152600, 0.77059182428310000, 0.55590551181102400, 0.47683923705722100,
        0.20913884007029800, 0.16656571774682000, 0.24660194174757300, 0.52243424445590500, 0.74159292035398200,
        0.80796508456083000, 0.54067875406787500, 0.38548457192525000, 0.43620414673046300, 0.41673243883188700,
        0.45581737849779100, 0.37531699070160600, 0.44154751892346600, 0.02913453299057440, 0.05768383971818600,
        0.03647005853219280, 0.31776556776556800, 0.21383382539013100, 0.18567059851463500, 0.09347300564061220,
        0.20173364854215900, 0.38637325433770600, 0.19605695509310000, 0.01057977147693610, 0.07139942880456960,
        0.08690614136732330, 0.10569744597249500, 0.04307974335472030, 0.09009900990098980, 0.03314045239347690,
        0.26422250316055700, 0.24533001245330000, 0.23937677053824300, 0.20109814687714500, 0.43433109346365400,
        0.44946492271105900, 0.55982085732565600, 0.64278296988577400, 0.63674762407603000, 0.40878048780487800,
        0.49078564500485000, 0.35790494665373400, 0.13839959738298900, 0.02162629757785470, 0.05102450783447100,
        0.19709208400646200, 0.21855345911949700, 0.51766138855054800, 0.36523009495982500, 0.35052316890881900,
        0.37257880744398100, 0.41830985915493000, 0.32846715328467200, 0.29285165257494300, 0.43883661248930700,
        0.22755555555555600, 0.25662959794696300, 0.24467622772707500, 0.50262237762237800, 0.40453074433657000,
        0.53191489361702100, 0.55803571428571400, 0.50215208034433300, 0.38745098039215700, 0.53072164948453600,
        0.66906474820143800, 0.62352941176470500, 0.77512985029025300, 0.59634146341463400, 0.63449564134495600,
        0.49152542372881400, 0.44567627494456800, 0.44567627494456800, 0.32610169491525400, 0.34019249917026200,
        0.38528951486698000, 0.39138518748063200, 0.02935010482180310, 0.21146131805157600, 0.00332225913621230,
        0.09514170040485810, 0.19379844961240300, 0.00289855072463779, 0.13033953997809400, 0.14485729671513200,
        0.01105293775450830, 0.03518728717366650, 0.33788037775445900, 0.57788096243140600, 0.58211450062682800,
        0.48937844217151800, 0.31779661016949100, 0.45248868778280500, 0.37749546279491800, 0.42134831460674100,
        0.41587901701323200, 0.52032520325203200, 0.56635361371988600, 0.53174603174603200, 0.45172947857511600,
        0.63579604578564000, 0.39299955693398300, 0.15128593040847200, 0.20926699582225600, 0.24132553606237800,
        0.12999273783587500, 0.14060258249641300, 0.08465368945224100, 0.36247334754797400, 0.34102833158447000,
        0.52663934426229500, 0.31188443860801000, 0.06255212677231020, 0.00969162995594708, 0.02134927412467970,
        0.18022657054582900, 0.15714285714285600,
    };
}

fn createKamaLength(allocator: std.mem.Allocator, er_len: u32, fastest: u32, slowest: u32) !KaufmanAdaptiveMovingAverage {
    var kama = try KaufmanAdaptiveMovingAverage.initLength(allocator, .{
        .efficiency_ratio_length = er_len,
        .fastest_length = fastest,
        .slowest_length = slowest,
    });
    kama.fixSlices();
    return kama;
}

fn createKamaAlpha(allocator: std.mem.Allocator, er_len: u32, fastest_alpha: f64, slowest_alpha: f64) !KaufmanAdaptiveMovingAverage {
    var kama = try KaufmanAdaptiveMovingAverage.initSmoothingFactor(allocator, .{
        .efficiency_ratio_length = er_len,
        .fastest_smoothing_factor = fastest_alpha,
        .slowest_smoothing_factor = slowest_alpha,
    });
    kama.fixSlices();
    return kama;
}

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) <= epsilon;
}

test "kaufman adaptive moving average value" {
    const input = testInput();
    const expected = testExpected();

    var kama = try createKamaLength(testing.allocator, 10, 2, 30);
    defer kama.deinit();

    for (0..10) |i| {
        const v = kama.update(input[i]);
        try testing.expect(math.isNan(v));
    }

    for (10..252) |i| {
        const v = kama.update(input[i]);
        try testing.expect(!math.isNan(v));
        try testing.expect(almostEqual(v, expected[i], 1e-8));
    }

    // NaN passthrough
    try testing.expect(math.isNan(kama.update(math.nan(f64))));
}

test "kaufman adaptive moving average efficiency ratio" {
    const input = testInput();
    const expected_er = testExpectedEr();

    var kama = try createKamaLength(testing.allocator, 10, 2, 30);
    defer kama.deinit();

    for (0..10) |_| {
        _ = kama.update(input[0]);
    }

    // Re-create to get clean state
    kama.deinit();
    kama = try createKamaLength(testing.allocator, 10, 2, 30);

    for (0..10) |i| {
        _ = kama.update(input[i]);
    }

    for (10..252) |i| {
        _ = kama.update(input[i]);
        try testing.expect(almostEqual(kama.efficiency_ratio, expected_er[i], 1e-8));
    }
}

test "kaufman adaptive moving average is primed" {
    const input = testInput();

    var kama = try createKamaLength(testing.allocator, 10, 2, 30);
    defer kama.deinit();

    try testing.expect(!kama.isPrimed());

    for (0..10) |_| {
        _ = kama.update(input[0]);
        try testing.expect(!kama.isPrimed());
    }

    // Re-create for clean state
    kama.deinit();
    kama = try createKamaLength(testing.allocator, 10, 2, 30);

    for (0..10) |i| {
        _ = kama.update(input[i]);
        try testing.expect(!kama.isPrimed());
    }

    _ = kama.update(input[10]);
    try testing.expect(kama.isPrimed());
}

test "kaufman adaptive moving average metadata length" {
    var kama = try createKamaLength(testing.allocator, 10, 2, 30);
    defer kama.deinit();

    var m: Metadata = undefined;
    kama.getMetadata(&m);

    try testing.expectEqual(@import("../../core/identifier.zig").Identifier.kaufman_adaptive_moving_average, m.identifier);
    try testing.expectEqual(@as(usize, 1), m.outputs_len);
    try testing.expectEqual(@as(i32, 1), m.outputs_buf[0].kind);
    try testing.expectEqualStrings("kama(10, 2, 30)", m.mnemonic);
}

test "kaufman adaptive moving average metadata alpha" {
    var kama = try createKamaAlpha(testing.allocator, 10, 0.666666666, 0.064516129);
    defer kama.deinit();

    var m: Metadata = undefined;
    kama.getMetadata(&m);

    try testing.expectEqualStrings("kama(10, 0.6667, 0.0645)", m.mnemonic);
}

test "kaufman adaptive moving average update scalar" {
    var kama = try createKamaLength(testing.allocator, 10, 2, 30);
    defer kama.deinit();

    for (0..10) |_| {
        _ = kama.update(0);
    }

    const s = Scalar{ .time = 1617235200, .value = 3.0 };
    const out = kama.updateScalar(&s);
    const slice = out.slice();
    try testing.expectEqual(@as(usize, 1), slice.len);
    try testing.expectEqual(@as(i64, 1617235200), slice[0].scalar.time);
    try testing.expectEqual(@as(f64, 1.3333333333333328), slice[0].scalar.value);
}

test "kaufman adaptive moving average update bar" {
    var kama = try createKamaLength(testing.allocator, 10, 2, 30);
    defer kama.deinit();

    for (0..10) |_| {
        _ = kama.update(0);
    }

    const bar = Bar{ .time = 1617235200, .open = 0, .high = 0, .low = 0, .close = 3.0, .volume = 0 };
    const out = kama.updateBar(&bar);
    const slice = out.slice();
    try testing.expectEqual(@as(usize, 1), slice.len);
    try testing.expectEqual(@as(f64, 1.3333333333333328), slice[0].scalar.value);
}

test "kaufman adaptive moving average invalid params" {
    // ER length < 2
    try testing.expectError(error.InvalidEfficiencyRatioLength, KaufmanAdaptiveMovingAverage.initLength(testing.allocator, .{
        .efficiency_ratio_length = 1,
    }));

    // Fastest length < 2
    try testing.expectError(error.InvalidFastestLength, KaufmanAdaptiveMovingAverage.initLength(testing.allocator, .{
        .fastest_length = 1,
    }));

    // Slowest length < 2
    try testing.expectError(error.InvalidSlowestLength, KaufmanAdaptiveMovingAverage.initLength(testing.allocator, .{
        .slowest_length = 1,
    }));

    // Fastest alpha out of range
    try testing.expectError(error.InvalidFastestSmoothingFactor, KaufmanAdaptiveMovingAverage.initSmoothingFactor(testing.allocator, .{
        .fastest_smoothing_factor = -0.00000001,
    }));
    try testing.expectError(error.InvalidFastestSmoothingFactor, KaufmanAdaptiveMovingAverage.initSmoothingFactor(testing.allocator, .{
        .fastest_smoothing_factor = 1.00000001,
    }));

    // Slowest alpha out of range
    try testing.expectError(error.InvalidSlowestSmoothingFactor, KaufmanAdaptiveMovingAverage.initSmoothingFactor(testing.allocator, .{
        .slowest_smoothing_factor = -0.00000001,
    }));
    try testing.expectError(error.InvalidSlowestSmoothingFactor, KaufmanAdaptiveMovingAverage.initSmoothingFactor(testing.allocator, .{
        .slowest_smoothing_factor = 1.00000001,
    }));
}
