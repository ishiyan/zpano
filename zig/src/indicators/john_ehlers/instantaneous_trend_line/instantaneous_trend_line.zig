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
const line_indicator_mod = @import("../../core/line_indicator.zig");
const build_metadata_mod = @import("../../core/build_metadata.zig");
const component_triple_mnemonic_mod = @import("../../core/component_triple_mnemonic.zig");
const identifier_mod = @import("../../core/identifier.zig");
const metadata_mod = @import("../../core/metadata.zig");

const OutputArray = indicator_mod.OutputArray;
const LineIndicator = line_indicator_mod.LineIndicator;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the instantaneous trend line indicator.
pub const InstantaneousTrendLineOutput = enum(u8) {
    value = 1,
    trigger = 2,
};

/// Parameters to create an ITL based on length.
pub const LengthParams = struct {
    /// Length ℓ of the instantaneous trend line (α = 2/(ℓ+1)). Must be >= 1. Default is 28.
    length: i32 = 28,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Parameters to create an ITL based on smoothing factor.
pub const SmoothingFactorParams = struct {
    /// Smoothing factor α in [0, 1]. Default is 0.07.
    smoothing_factor: f64 = 0.07,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Ehler's Instantaneous Trend Line (iTrend).
///
///   H(z) = ((α-α²/4) + α²z⁻¹/2 - (α-3α²/4)z⁻²) / (1 - 2(1-α)z⁻¹ + (1-α)²z⁻²)
///
/// Two outputs: trend line value and trigger line.
/// Primed after 5 samples.
pub const InstantaneousTrendLine = struct {
    coeff1: f64,
    coeff2: f64,
    coeff3: f64,
    coeff4: f64,
    coeff5: f64,
    count: usize,
    previous_sample1: f64,
    previous_sample2: f64,
    previous_trend_line1: f64,
    previous_trend_line2: f64,
    trend_line: f64,
    trigger_line: f64,
    primed: bool,
    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,
    mnemonic_buf: [128]u8,
    mnemonic_len: usize,
    description_buf: [192]u8,
    description_len: usize,
    mnemonic_trig_buf: [128]u8,
    mnemonic_trig_len: usize,
    description_trig_buf: [192]u8,
    description_trig_len: usize,

    pub fn initLength(params: LengthParams) !InstantaneousTrendLine {
        if (params.length < 1) {
            return error.InvalidLength;
        }

        const alpha: f64 = 2.0 / @as(f64, @floatFromInt(1 + params.length));
        return initCommon(params.length, alpha, params.bar_component, params.quote_component, params.trade_component);
    }

    pub fn initSmoothingFactor(params: SmoothingFactorParams) !InstantaneousTrendLine {
        const alpha = params.smoothing_factor;
        if (alpha < 0.0 or alpha > 1.0) {
            return error.InvalidSmoothingFactor;
        }

        const epsilon: f64 = 0.00000001;
        const length: i32 = if (alpha < epsilon)
            std.math.maxInt(i32)
        else
            @as(i32, @intFromFloat(@round(2.0 / alpha))) - 1;

        return initCommon(length, alpha, params.bar_component, params.quote_component, params.trade_component);
    }

    fn initCommon(
        length: i32,
        alpha: f64,
        bc_opt: ?bar_component.BarComponent,
        qc_opt: ?quote_component.QuoteComponent,
        tc_opt: ?trade_component.TradeComponent,
    ) !InstantaneousTrendLine {
        const bc = bc_opt orelse bar_component.BarComponent.median;
        const qc = qc_opt orelse quote_component.default_quote_component;
        const tc = tc_opt orelse trade_component.default_trade_component;

        // Calculate coefficients.
        const a2 = alpha * alpha;
        const c1 = alpha - a2 / 4.0;
        const c2 = a2 / 2.0;
        const c3 = -(alpha - 3.0 * a2 / 4.0);
        const x = 1.0 - alpha;
        const c4 = 2.0 * x;
        const c5 = -(x * x);

        // Build mnemonics.
        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(
            &triple_buf,
            bc_opt orelse bar_component.BarComponent.median,
            qc_opt orelse quote_component.default_quote_component,
            tc_opt orelse trade_component.default_trade_component,
        );

        var mnemonic_buf: [128]u8 = undefined;
        const mn = std.fmt.bufPrint(&mnemonic_buf, "iTrend({d}{s})", .{ length, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mn.len;

        var description_buf: [192]u8 = undefined;
        const desc = std.fmt.bufPrint(&description_buf, "Instantaneous Trend Line {s}", .{mn}) catch
            return error.MnemonicTooLong;
        const description_len = desc.len;

        var mnemonic_trig_buf: [128]u8 = undefined;
        const mn_trig = std.fmt.bufPrint(&mnemonic_trig_buf, "iTrendTrigger({d}{s})", .{ length, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_trig_len = mn_trig.len;

        var description_trig_buf: [192]u8 = undefined;
        const desc_trig = std.fmt.bufPrint(&description_trig_buf, "Instantaneous Trend Line trigger {s}", .{mn_trig}) catch
            return error.MnemonicTooLong;
        const description_trig_len = desc_trig.len;

        return .{
            .coeff1 = c1,
            .coeff2 = c2,
            .coeff3 = c3,
            .coeff4 = c4,
            .coeff5 = c5,
            .count = 0,
            .previous_sample1 = 0.0,
            .previous_sample2 = 0.0,
            .previous_trend_line1 = 0.0,
            .previous_trend_line2 = 0.0,
            .trend_line = math.nan(f64),
            .trigger_line = math.nan(f64),
            .primed = false,
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
            .mnemonic_trig_buf = mnemonic_trig_buf,
            .mnemonic_trig_len = mnemonic_trig_len,
            .description_trig_buf = description_trig_buf,
            .description_trig_len = description_trig_len,
        };
    }

    pub fn fixSlices(self: *InstantaneousTrendLine) void {
        // No slice fields to fix (mnemonics stored as owned buffers + lengths).
        _ = self;
    }

    /// Update the iTrend given the next sample. Returns the trend line value.
    pub fn update(self: *InstantaneousTrendLine, sample: f64) f64 {
        if (math.isNan(sample)) {
            return math.nan(f64);
        }

        if (self.primed) {
            self.trend_line = self.coeff1 * sample + self.coeff2 * self.previous_sample1 +
                self.coeff3 * self.previous_sample2 +
                self.coeff4 * self.previous_trend_line1 + self.coeff5 * self.previous_trend_line2;
            self.trigger_line = 2.0 * self.trend_line - self.previous_trend_line2;

            self.previous_sample2 = self.previous_sample1;
            self.previous_sample1 = sample;
            self.previous_trend_line2 = self.previous_trend_line1;
            self.previous_trend_line1 = self.trend_line;

            return self.trend_line;
        }

        self.count += 1;

        switch (self.count) {
            1 => {
                self.previous_sample2 = sample;
                return math.nan(f64);
            },
            2 => {
                self.previous_sample1 = sample;
                return math.nan(f64);
            },
            3 => {
                self.previous_trend_line2 = (sample + 2.0 * self.previous_sample1 + self.previous_sample2) / 4.0;
                self.previous_sample2 = self.previous_sample1;
                self.previous_sample1 = sample;
                return math.nan(f64);
            },
            4 => {
                self.previous_trend_line1 = (sample + 2.0 * self.previous_sample1 + self.previous_sample2) / 4.0;
                self.previous_sample2 = self.previous_sample1;
                self.previous_sample1 = sample;
                return math.nan(f64);
            },
            5 => {
                self.trend_line = self.coeff1 * sample + self.coeff2 * self.previous_sample1 +
                    self.coeff3 * self.previous_sample2 +
                    self.coeff4 * self.previous_trend_line1 + self.coeff5 * self.previous_trend_line2;
                self.trigger_line = 2.0 * self.trend_line - self.previous_trend_line2;

                self.previous_sample2 = self.previous_sample1;
                self.previous_sample1 = sample;
                self.previous_trend_line2 = self.previous_trend_line1;
                self.previous_trend_line1 = self.trend_line;
                self.primed = true;

                return self.trend_line;
            },
            else => return math.nan(f64),
        }
    }

    pub fn isPrimed(self: *const InstantaneousTrendLine) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const InstantaneousTrendLine, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .instantaneous_trend_line,
            self.mnemonic_buf[0..self.mnemonic_len],
            self.description_buf[0..self.description_len],
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.mnemonic_buf[0..self.mnemonic_len], .description = self.description_buf[0..self.description_len] },
                .{ .mnemonic = self.mnemonic_trig_buf[0..self.mnemonic_trig_len], .description = self.description_trig_buf[0..self.description_trig_len] },
            },
        );
    }

    pub fn updateScalar(self: *InstantaneousTrendLine, sample: *const Scalar) OutputArray {
        return self.updateEntity(sample.time, sample.value);
    }

    pub fn updateBar(self: *InstantaneousTrendLine, sample: *const Bar) OutputArray {
        return self.updateEntity(sample.time, self.bar_func(sample.*));
    }

    pub fn updateQuote(self: *InstantaneousTrendLine, sample: *const Quote) OutputArray {
        return self.updateEntity(sample.time, self.quote_func(sample.*));
    }

    pub fn updateTrade(self: *InstantaneousTrendLine, sample: *const Trade) OutputArray {
        return self.updateEntity(sample.time, self.trade_func(sample.*));
    }

    fn updateEntity(self: *InstantaneousTrendLine, time: i64, sample: f64) OutputArray {
        const v = self.update(sample);
        var trig = self.trigger_line;
        if (math.isNan(v)) {
            trig = math.nan(f64);
        }

        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = v } });
        out.append(.{ .scalar = .{ .time = time, .value = trig } });
        return out;
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *InstantaneousTrendLine) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(InstantaneousTrendLine);
};

// --- Tests ---
const testing = std.testing;

fn almostEqual(a: f64, b: f64, tolerance: f64) bool {
    return @abs(a - b) <= tolerance;
}

// 252-entry input data from test_iTrend.xls.
const test_input = [_]f64{
    92.0000,  93.1725,  95.3125,  94.8450,  94.4075,  94.1100,  93.5000,  91.7350,  90.9550,  91.6875,
    94.5000,  97.9700,  97.5775,  90.7825,  89.0325,  92.0950,  91.1550,  89.7175,  90.6100,  91.0000,
    88.9225,  87.5150,  86.4375,  83.8900,  83.0025,  82.8125,  82.8450,  86.7350,  86.8600,  87.5475,
    85.7800,  86.1725,  86.4375,  87.2500,  88.9375,  88.2050,  85.8125,  84.5950,  83.6575,  84.4550,
    83.5000,  86.7825,  88.1725,  89.2650,  90.8600,  90.7825,  91.8600,  90.3600,  89.8600,  90.9225,
    89.5000,  87.6725,  86.5000,  84.2825,  82.9075,  84.2500,  85.6875,  86.6100,  88.2825,  89.5325,
    89.5000,  88.0950,  90.6250,  92.2350,  91.6725,  92.5925,  93.0150,  91.1725,  90.9850,  90.3775,
    88.2500,  86.9075,  84.0925,  83.1875,  84.2525,  97.8600,  99.8750,  103.2650, 105.9375, 103.5000,
    103.1100, 103.6100, 104.6400, 106.8150, 104.9525, 105.5000, 107.1400, 109.7350, 109.8450, 110.9850,
    120.0000, 119.8750, 117.9075, 119.4075, 117.9525, 117.2200, 115.6425, 113.1100, 111.7500, 114.5175,
    114.7450, 115.4700, 112.5300, 112.0300, 113.4350, 114.2200, 119.5950, 117.9650, 118.7150, 115.0300,
    114.5300, 115.0000, 116.5300, 120.1850, 120.5000, 120.5950, 124.1850, 125.3750, 122.9700, 123.0000,
    124.4350, 123.4400, 124.0300, 128.1850, 129.6550, 130.8750, 132.3450, 132.0650, 133.8150, 135.6600,
    137.0350, 137.4700, 137.3450, 136.3150, 136.4400, 136.2850, 129.0950, 128.3100, 126.0000, 124.0300,
    123.9350, 125.0300, 127.2500, 125.6200, 125.5300, 123.9050, 120.6550, 119.9650, 120.7800, 124.0000,
    122.7800, 120.7200, 121.7800, 122.4050, 123.2500, 126.1850, 127.5600, 126.5650, 123.0600, 122.7150,
    123.5900, 122.3100, 122.4650, 123.9650, 123.9700, 124.1550, 124.4350, 127.0000, 125.5000, 128.8750,
    130.5350, 132.3150, 134.0650, 136.0350, 133.7800, 132.7500, 133.4700, 130.9700, 127.5950, 128.4400,
    127.9400, 125.8100, 124.6250, 122.7200, 124.0900, 123.2200, 121.4050, 120.9350, 118.2800, 118.3750,
    121.1550, 120.9050, 117.1250, 113.0600, 114.9050, 112.4350, 107.9350, 105.9700, 106.3700, 106.8450,
    106.9700, 110.0300, 91.0000,  93.5600,  93.6200,  95.3100,  94.1850,  94.7800,  97.6250,  97.5900,
    95.2500,  94.7200,  92.2200,  91.5650,  92.2200,  93.8100,  95.5900,  96.1850,  94.6250,  95.1200,
    94.0000,  93.7450,  95.9050,  101.7450, 106.4400, 107.9350, 103.4050, 105.0600, 104.1550, 103.3100,
    103.3450, 104.8400, 110.4050, 114.5000, 117.3150, 118.2500, 117.1850, 109.7500, 109.6550, 108.5300,
    106.2200, 107.7200, 109.8400, 109.0950, 109.0900, 109.1550, 109.3150, 109.0600, 109.9050, 109.6250,
    109.5300, 108.0600,
};

// Expected trend line values from test_iTrend.xls, 252 entries.
const test_expected_trend = [_]f64{
    math.nan(f64),     math.nan(f64),     math.nan(f64),     math.nan(f64),
    95.6783140625000,  96.5028085937501,  97.1425047142188,  97.5160834907126,
    97.6401563403976,  97.7133538695222,  97.9891484785014,  98.6357747207628,
    99.3949330240631,  99.5492649312695,  99.0687548246491,  98.6985386722923,
    98.4844531951247,  98.1021667827664,  97.6964818974821,  97.3968985914021,
    96.9894308743756,  96.3628930846349,  95.6101883741735,  94.6675583970620,
    93.5694669437127,  92.4960188201866,  91.5097777334300,  90.8840107191004,
    90.5890558008832,  90.3712503561927,  90.0930726753346,  89.7411074930513,
    89.4634645676785,  89.2819597876420,  89.2849665754289,  89.3474103099662,
    89.1826928354487,  88.7812986218447,  88.2683926532516,  87.7948552820145,
    87.3561415812497,  87.1202165702100,  87.2278192794679,  87.4928422357355,
    87.9129003511563,  88.3911077159631,  88.8853919629762,  89.2957766750993,
    89.5220581069066,  89.7593357700528,  89.9428946006348,  89.8794508246620,
    89.6119166187823,  89.1350032676849,  88.4571023318091,  87.8447263234442,
    87.4853735298246,  87.3245468433268,  87.3581575626426,  87.5886540642218,
    87.8782063960230,  88.0378025589573,  88.2569316727403,  88.7365899780548,
    89.2366859679288,  89.7071196658280,  90.2182125722784,  90.5752654604632,
    90.7519078276980,  90.8516112127636,  90.7489102755643,  90.4144026996304,
    89.8252603364769,  89.0386743309368,  88.3429065905235,  88.7275860170466,
    90.1461431440628,  91.7875167268132,  93.6720233440727,  95.3716484529544,
    96.6929936947067,  97.8757892751943,  99.0308875553095,  100.2753170837600,
    101.3998320667070, 102.3039571483300, 103.2506994789000, 104.3758386181630,
    105.5549099754820, 106.6814094835480, 108.3724236266050, 110.4812798831640,
    112.2076534505350, 113.7013912595460, 115.0207048983880, 116.0292644981210,
    116.7503266748880, 117.0897446258680, 117.1023044005030, 117.1870738705230,
    117.4465108231770, 117.7226112529950, 117.7962598446040, 117.6047708382480,
    117.4753019945440, 117.4907019118500, 117.9097256109610, 118.5247906778280,
    118.9974889798400, 119.2012494202490, 119.0782475779990, 118.9487386215050,
    118.9537760057880, 119.2991062120260, 119.8662051869630, 120.3872206849690,
    121.0920682328380, 122.0367952426500, 122.7837697117460, 123.2732920834810,
    123.7960227265840, 124.2780086984450, 124.6653559978840, 125.3221513077800,
    126.2809211549000, 127.3080974320160, 128.3959957666760, 129.4345099070680,
    130.4473954385470, 131.5829000220750, 132.8019109762600, 133.9980896867510,
    135.0699315139900, 135.9289864709500, 136.6141566945170, 137.2035375530770,
    137.2040195986350, 136.6338371988050, 135.8889642639190, 134.9075702376420,
    133.8682685751500, 132.9897113012430, 132.4161509046650, 131.9299026282320,
    131.3641284710660, 130.7284745480260, 129.8139655697020, 128.7151168230590,
    127.7321343446540, 127.1233977157930, 126.7112577566840, 126.1129417430420,
    125.5053048083030, 125.0747945048860, 124.7889321503870, 124.7920482824440,
    125.0907983884760, 125.3840521930790, 125.3364243029350, 125.0278799616640,
    124.7869828490870, 124.5427680954590, 124.2463703163780, 124.0947300377010,
    124.0627720584880, 124.0477320191810, 124.0667766272890, 124.2800314783690,
    124.5447645698240, 124.9110409992310, 125.5852702571280, 126.4251838180240,
    127.4188999061350, 128.5624863412020, 129.5637621908190, 130.2300241134180,
    130.7998251821180, 131.1815464830450, 131.1127325334490, 130.8702464590300,
    130.6702691706150, 130.3042772449290, 129.7429382449040, 129.0227467463810,
    128.3368422852520, 127.7538249896240, 127.0448282131860, 126.2512776180000,
    125.3251155728950, 124.3199977037770, 123.6190361450280, 123.1678904657560,
    122.4900556544710, 121.3472286784840, 120.1754689564280, 119.0886840499350,
    117.6428104824650, 115.9111007875960, 114.2618150536440, 112.8558760535870,
    111.6505273697750, 110.8038288340330, 108.9618937591840, 106.1946900835270,
    103.9133441430420, 102.0099799028150, 100.3656476449200, 98.8794850016057,
    97.8091437048954,  97.0687907132167,  96.2677584612191,  95.3783074000063,
    94.3987057209034,  93.3299124456148,  92.3984698208342,  91.7447857175393,
    91.4174101365837,  91.3136883119459,  91.1800231080881,  91.0118730850419,
    90.8423728269925,  90.6194228019534,  90.5735497785674,  91.1049063067259,
    92.3268496520272,  93.8670773880834,  95.0557583027968,  95.9381300852486,
    96.7935171024736,  97.4512735998693,  97.9952374538275,  98.5970656275922,
    99.6300659435061,  101.2270373436170, 103.1384979245890, 105.1121945412410,
    106.8723884917320, 107.8633337859010, 108.2299870352780, 108.4704152441910,
    108.4451676923830, 108.3623795131320, 108.5329188572830, 108.7761946586380,
    108.9384100454040, 109.0835916741940, 109.2248242907320, 109.3404926167510,
    109.4803617381020, 109.6402028936430, 109.7532703898900, 109.7425683174850,
};

// Expected trigger line values from test_iTrend.xls, 252 entries.
const test_expected_trigger = [_]f64{
    math.nan(f64),     math.nan(f64),     math.nan(f64),     math.nan(f64),
    97.9422531250001,  98.3449921875001,  98.6066953659376,  98.5293583876752,
    98.1378079665764,  97.9106242483318,  98.3381406166052,  99.5581955720035,
    100.8007175696250, 100.4627551417760, 98.7425766252352,  97.8478124133152,
    97.9001515656004,  97.5057948932404,  96.9085105998395,  96.6916304000378,
    96.2823798512691,  95.3288875778678,  94.2309458739715,  92.9722237094891,
    91.5287455132518,  90.3244792433112,  89.4500885231473,  89.2720026180142,
    89.6683338683363,  89.8584899932850,  89.5970895497861,  89.1109646299099,
    88.8338564600224,  88.8228120822326,  89.1064685831792,  89.4128608322904,
    89.0804190954684,  88.2151869337233,  87.3540924710546,  86.8084119421843,
    86.4438905092477,  86.4455778584056,  87.0994969776860,  87.8654679012610,
    88.5979814228448,  89.2893731961907,  89.8578835747961,  90.2004456342355,
    90.1587242508369,  90.2228948650064,  90.3637310943630,  89.9995658792711,
    89.2809386369298,  88.3905557107078,  87.3022880448359,  86.5544493792035,
    86.5136447278401,  86.8043673632094,  87.2309415954606,  87.8527612851168,
    88.3982552294034,  88.4869510536929,  88.6356569494577,  89.4353773971523,
    90.2164402631173,  90.6776493536011,  91.1997391766280,  91.4434112550984,
    91.2856030831175,  91.1279569650640,  90.7459127234306,  89.9771941864972,
    88.9016103973896,  87.6629459622432,  86.8605528445702,  88.4164977031563,
    91.9493796976020,  94.8474474365799,  97.1979035440825,  98.9557801790956,
    99.7139640453408,  100.3799300974340, 101.3687814159120, 102.6748448923260,
    103.7687765781040, 104.3325972129000, 105.1015668910930, 106.4477200879950,
    107.8591204720650, 108.9869803489330, 111.1899372777270, 114.2811502827800,
    116.0428832744650, 116.9215026359280, 117.8337563462420, 118.3571377366950,
    118.4799484513890, 118.1502247536150, 117.4542821261180, 117.2844031151780,
    117.7907172458520, 118.2581486354660, 118.1460088660310, 117.4869304235020,
    117.1543441444840, 117.3766329854530, 118.3441492273780, 119.5588794438050,
    120.0852523487190, 119.8777081626690, 119.1590061761580, 118.6962278227620,
    118.8293044335770, 119.6494738025470, 120.7786343681370, 121.4753351579120,
    122.3179312787140, 123.6863698003300, 124.4754711906540, 124.5097889243120,
    124.8082757414230, 125.2827253134090, 125.5346892691840, 126.3662939171150,
    127.8964863119170, 129.2940435562520, 130.5110703784520, 131.5609223821190,
    132.4987951104180, 133.7312901370820, 135.1564265139730, 136.4132793514270,
    137.3379520517190, 137.8598832551480, 138.1583818750440, 138.4780886352040,
    137.7938825027540, 136.0641368445340, 134.5739089292020, 133.1813032764780,
    131.8475728863820, 131.0718523648450, 130.9640332341800, 130.8700939552200,
    130.3121060374670, 129.5270464678190, 128.2638026683390, 126.7017590980930,
    125.6503031196060, 125.5316786085270, 125.6903811687130, 125.1024857702920,
    124.2993518599220, 124.0366472667300, 124.0725594924710, 124.5093020600020,
    125.3926646265650, 125.9760561037150, 125.5820502173940, 124.6717077302490,
    124.2375413952400, 124.0576562292540, 123.7057577836690, 123.6466919799430,
    123.8791738005980, 124.0007340006600, 124.0707811960900, 124.5123309375570,
    125.0227525123580, 125.5420505200920, 126.6257759444330, 127.9393266368180,
    129.2525295551420, 130.6997888643790, 131.7086244755030, 131.8975618856340,
    132.0358881734170, 132.1330688526710, 131.4256398847790, 130.5589464350150,
    130.2278058077810, 129.7383080308290, 128.8156073191920, 127.7412162478330,
    126.9307463256010, 126.4849032328660, 125.7528141411190, 124.7487302463760,
    123.6054029326050, 122.3887177895550, 121.9129567171620, 122.0157832277350,
    121.3610751639140, 119.5265668912120, 117.8608822583850, 116.8301394213860,
    115.1101520085020, 112.7335175252570, 110.8808196248240, 109.8006513195780,
    109.0392396859050, 108.7517816144800, 106.2732601485940, 101.5855513330210,
    98.8647945268997,  97.8252697221038,  96.8179511467976,  95.7489901003959,
    95.2526397648711,  95.2580964248278,  94.7263732175427,  93.6878240867959,
    92.5296529805876,  91.2815174912233,  90.3982339207650,  90.1596589894639,
    90.4363504523332,  90.8825909063525,  90.9426360795926,  90.7100578581379,
    90.5047225458969,  90.2269725188648,  90.3047267301423,  91.5903898114985,
    94.0801495254871,  96.6292484694409,  97.7846669535663,  98.0091827824139,
    98.5312759021504,  98.9644171144899,  99.1969578051814,  99.7428576553150,
    101.2648944331850, 103.8570090596410, 106.6469299056720, 108.9973517388660,
    110.6062790588740, 110.6144730305620, 109.5875855788240, 109.0774967024800,
    108.6603483494880, 108.2543437820720, 108.6206700221830, 109.1900098041450,
    109.3439012335250, 109.3909886897500, 109.5112385360600, 109.5973935593070,
    109.7358991854730, 109.9399131705340, 110.0261790416790, 109.8449337413270,
};

test "ITL update trend line" {
    const tolerance = 1e-8;
    const l_primed = 4;

    var itl = try InstantaneousTrendLine.initSmoothingFactor(.{ .smoothing_factor = 0.07 });
    itl.fixSlices();

    for (0..l_primed) |i| {
        try testing.expect(math.isNan(itl.update(test_input[i])));
    }

    for (l_primed..test_input.len) |i| {
        const act = itl.update(test_input[i]);
        try testing.expect(almostEqual(act, test_expected_trend[i], tolerance));
    }

    // NaN passthrough.
    try testing.expect(math.isNan(itl.update(math.nan(f64))));
}

test "ITL update trigger line" {
    const tolerance = 1e-8;
    const l_primed = 4;

    var itl = try InstantaneousTrendLine.initSmoothingFactor(.{ .smoothing_factor = 0.07 });
    itl.fixSlices();

    for (0..l_primed) |_i| {
        _ = itl.update(test_input[_i]);
    }

    for (l_primed..test_input.len) |i| {
        _ = itl.update(test_input[i]);
        try testing.expect(almostEqual(itl.trigger_line, test_expected_trigger[i], tolerance));
    }
}

test "ITL isPrimed" {
    var itl = try InstantaneousTrendLine.initSmoothingFactor(.{ .smoothing_factor = 0.07 });
    itl.fixSlices();

    try testing.expect(!itl.isPrimed());

    // First 4 updates: not primed.
    for (0..4) |i| {
        _ = itl.update(test_input[i]);
        try testing.expect(!itl.isPrimed());
    }

    // 5th update: primed.
    _ = itl.update(test_input[4]);
    try testing.expect(itl.isPrimed());
}

test "ITL metadata" {
    var itl = try InstantaneousTrendLine.initSmoothingFactor(.{ .smoothing_factor = 0.07 });
    itl.fixSlices();
    var meta: Metadata = undefined;
    itl.getMetadata(&meta);

    try testing.expectEqual(Identifier.instantaneous_trend_line, meta.identifier);
    try testing.expectEqualStrings("iTrend(28, hl/2)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 2), meta.outputs_len);
}

test "ITL constructor length" {
    // Valid length.
    _ = try InstantaneousTrendLine.initLength(.{ .length = 28 });
    _ = try InstantaneousTrendLine.initLength(.{ .length = 1 });

    // Invalid length.
    try testing.expectError(error.InvalidLength, InstantaneousTrendLine.initLength(.{ .length = 0 }));
    try testing.expectError(error.InvalidLength, InstantaneousTrendLine.initLength(.{ .length = -8 }));
}

test "ITL constructor smoothing factor" {
    // Valid.
    _ = try InstantaneousTrendLine.initSmoothingFactor(.{ .smoothing_factor = 0.07 });
    _ = try InstantaneousTrendLine.initSmoothingFactor(.{ .smoothing_factor = 0.0 });
    _ = try InstantaneousTrendLine.initSmoothingFactor(.{ .smoothing_factor = 1.0 });

    // Invalid.
    try testing.expectError(error.InvalidSmoothingFactor, InstantaneousTrendLine.initSmoothingFactor(.{ .smoothing_factor = -0.0001 }));
    try testing.expectError(error.InvalidSmoothingFactor, InstantaneousTrendLine.initSmoothingFactor(.{ .smoothing_factor = 1.0001 }));
}

test "ITL updateScalar" {
    const tolerance = 1e-8;
    var itl = try InstantaneousTrendLine.initSmoothingFactor(.{ .smoothing_factor = 0.07 });
    itl.fixSlices();

    for (0..test_input.len) |i| {
        const s = Scalar{ .time = @intCast(i), .value = test_input[i] };
        const out = itl.updateScalar(&s);
        const outputs = out.slice();
        try testing.expectEqual(@as(usize, 2), outputs.len);

        if (i < 4) {
            try testing.expect(math.isNan(outputs[0].scalar.value));
            try testing.expect(math.isNan(outputs[1].scalar.value));
        } else {
            try testing.expect(almostEqual(outputs[0].scalar.value, test_expected_trend[i], tolerance));
            try testing.expect(almostEqual(outputs[1].scalar.value, test_expected_trigger[i], tolerance));
        }
    }
}
