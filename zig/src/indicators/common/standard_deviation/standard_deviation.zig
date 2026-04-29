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
const variance_mod = @import("../variance/variance.zig");

const OutputArray = indicator_mod.OutputArray;
const LineIndicator = line_indicator_mod.LineIndicator;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the standard deviation indicator.
pub const StandardDeviationOutput = enum(u8) {
    /// The scalar value of the standard deviation.
    value = 1,
};

/// Parameters to create an instance of the standard deviation indicator.
pub const StandardDeviationParams = struct {
    /// The length (number of time periods). Must be >= 2.
    length: usize,
    /// Whether to compute based on unbiased sample variance (true) or population variance (false).
    is_unbiased: bool = false,
    /// Bar component to extract. `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract. `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract. `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the standard deviation as the square root of variance.
///
/// The indicator is not primed during the first l-1 updates.
pub const StandardDeviation = struct {
    line: LineIndicator,
    variance: variance_mod.Variance,
    mnemonic_buf: [96]u8,
    mnemonic_len: usize,
    description_buf: [160]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: StandardDeviationParams) !StandardDeviation {
        // Create the underlying variance indicator (validates length >= 2).
        var v = try variance_mod.Variance.init(allocator, .{
            .length = params.length,
            .is_unbiased = params.is_unbiased,
            .bar_component = params.bar_component,
            .quote_component = params.quote_component,
            .trade_component = params.trade_component,
        });
        v.fixSlices();

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        const c: u8 = if (params.is_unbiased) 's' else 'p';

        var mnemonic_buf: [96]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "stdev.{c}({d}{s})", .{ c, params.length, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [160]u8 = undefined;
        const desc_prefix: []const u8 = if (params.is_unbiased)
            "Standard deviation based on unbiased estimation of the sample variance "
        else
            "Standard deviation based on estimation of the population variance ";
        const desc_slice = std.fmt.bufPrint(&description_buf, "{s}{s}", .{ desc_prefix, mnemonic_slice }) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .variance = v,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *StandardDeviation) void {
        self.variance.deinit();
    }

    pub fn fixSlices(self: *StandardDeviation) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    /// Core update logic. Returns sqrt(variance) or NaN if not yet primed.
    pub fn update(self: *StandardDeviation, sample: f64) f64 {
        const v = self.variance.update(sample);
        if (math.isNan(v)) {
            return v;
        }
        return @sqrt(v);
    }

    pub fn isPrimed(self: *const StandardDeviation) bool {
        return self.variance.isPrimed();
    }

    pub fn getMetadata(self: *const StandardDeviation, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .standard_deviation,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *StandardDeviation, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *StandardDeviation, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *StandardDeviation, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *StandardDeviation, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *StandardDeviation) indicator_mod.Indicator {
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
        const self: *StandardDeviation = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const StandardDeviation = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *StandardDeviation = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *StandardDeviation = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *StandardDeviation = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *StandardDeviation = @ptrCast(@alignCast(ptr));
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

fn testInput() [252]f64 {
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

fn expectedPopulation() [252]f64 {
    return .{
        0,                 0,                 0,                 0, // NaN placeholders
        1.285646141051260, 0.446246568614257, 0.883942305809604, 1.006356795575010,
        1.452008953140440, 1.367254914052240, 1.865900854815180, 2.542562880245050,
        3.062846062080170, 3.197591906419580, 3.459226792218170, 3.410854145225210,
        3.351479374843290, 1.350827894292980, 1.385870123785050, 1.423985954986920,
        1.087363784572580, 1.369358974118910, 2.291995636994100, 2.494532821992930,
        2.169791234197430, 2.091600822336800, 1.204559670585060, 2.661085492801760,
        2.670454268471940, 2.589309946684640, 1.629225582907410, 1.340242515368020,
        0.621712152044658, 1.297700273560890, 1.445321417540060, 1.136685532590260,
        1.160971145205600, 1.442585872660620, 1.534428232274160, 1.281859586694270,
        1.197732858362000, 0.748215209682348, 2.103001188777600, 2.580118989504170,
        2.839624975238810, 1.987012833375770, 0.902372428656816, 0.954316509340585,
        0.857540669589495, 0.829322615150461, 1.007603096462090, 0.899702172943914,
        2.374253566913190, 2.795009838980890, 2.709309137031060, 2.122968676170230,
        1.019107452627050, 1.304797302265760, 2.034069811977950, 1.815642035204080,
        1.506288816927220, 1.092472425281300, 1.282316653561050, 1.468795424829480,
        1.890884449140140, 1.788055927536940, 0.789848086659707, 0.839261580200119,
        1.318847982141990, 1.565227140066260, 1.543206402267690, 2.161531864211120,
        2.655600497062760, 2.444997750510210, 1.815079612579020, 5.031641282921510,
        6.875640042934190, 7.890907045454280, 7.189955771769390, 3.222045313151260,
        2.146139324461480, 1.413713549485890, 1.620836203939190, 1.620836203939190,
        1.434620507311950, 0.713145146516472, 1.295096907571010, 1.759961363212270,
        2.161171904314880, 2.666838202816210, 5.270695020583150, 5.348179690324550,
        4.609501057598320, 3.318132004607410, 1.737663949099480, 1.131266546840310,
        1.501820228922220, 2.533567445322900, 2.762992218592010, 2.810866414470810,
        2.753329983855910, 2.820583982085980, 2.797497810544270, 2.105943968865270,
        1.618077872044480, 1.618077872044480, 3.034761275619550, 2.647764339966830,
        2.413515278592620, 1.812706264125550, 2.113978240190750, 0.970002061853479,
        0.875045141692701, 2.243065759178720, 2.610305729220240, 2.390882682190820,
        2.817147493476330, 1.787825494840030, 1.694792022638770, 1.294582558201680,
        0.749063415205949, 0.316733326317267, 0.761787371908987, 2.528955515623000,
        3.392179240547290, 3.731555171774900, 2.678153094951820, 1.229461670813690,
        1.139638539186880, 2.172200727373050, 2.628938949462310, 2.139538267944740,
        1.452234140901530, 0.562515777556504, 0.704499822569175, 1.100538050228160,
        3.254943317478820, 3.540488101943010, 4.501972900851360, 3.799728411347310,
        2.399438267595150, 2.103830791675030, 1.895082056270910, 1.765258054789720,
        1.724916229850020, 1.971480661837690, 3.104579842748450, 2.963913628970990,
        2.610042145253600, 2.042561137395890, 2.037478834245890, 2.049247666828000,
        1.562331590924280, 1.545814995398870, 1.519315635409570, 2.699122820473350,
        2.759728972199990, 2.710591079451120, 2.272610833380850, 2.603264104926740,
        2.288956093943260, 1.038903267874350, 0.957768239189418, 0.957768239189418,
        0.933980727852560, 0.690883492348745, 0.777364779238164, 1.542924495884360,
        1.434274729610750, 1.985654552030640, 2.578615132197900, 2.232797348618990,
        2.982848303216240, 2.350092764126550, 1.643564419181680, 1.569858592357920,
        1.227495010173160, 1.641993909854720, 2.829169489443850, 2.685553946581600,
        2.349173471670410, 2.152444192075600, 2.659213417535340, 2.679726851751870,
        1.793503833282770, 1.210196678230440, 1.641881847149790, 1.772167035016730,
        2.082156574323840, 1.864538548810400, 1.413543066199260, 1.464805789174800,
        1.911330426692360, 2.954553096493610, 3.161843765906220, 3.055299657971370,
        3.960770632086640, 3.569562438170820, 3.137786480944810, 1.598469267768390,
        0.933209515596579, 0.339199056602463, 6.507092745612280, 7.198239784836290,
        6.999970285651220, 5.562144910014480, 1.471276996353850, 0.876538647179917,
        1.790546285355390, 1.761356295585880, 1.797342482667120, 1.460597138159590,
        2.270774317275940, 2.338806533255800, 1.770407862612460, 1.578827412987250,
        2.303420065902010, 2.200149085857590, 1.244822878967120, 1.223470473693580,
        1.029990291216380, 0.744231146889192, 1.549799987095110, 3.845899634675870,
        5.455362866024590, 5.267593758064490, 3.330066665999350, 1.386414079559210,
        1.333695617448000, 0.983268020429832, 0.707152034572480, 0.859334626324342,
        3.243660894729910, 5.133827422109160, 5.421588697051820, 4.672564178264440,
        2.305501247017660, 3.238552763195310, 3.666974774933690, 3.553471542027600,
        2.070986238486390, 0.942409677369668, 1.056851929079940, 1.011205221505510,
        1.120578422066030, 0.758218965734834, 0.807722724701986, 0.762375235694340,
        0.809864186144813, 0.686515841040831, 0.500503746239725, 0.714355653718789,
    };
}

fn expectedSample() [252]f64 {
    return .{
        0,                  0,                  0,                  0,
        1.4373960832004500, 0.4989188310737520, 0.9882775419890900, 1.1251411022622900,
        1.6233953615801700, 1.5286374651957200, 2.0861405753208500, 2.8426717186477900,
        3.4243659997144000, 3.5750164335286600, 3.8675331284941800, 3.8134508650302500,
        3.7470678536690500, 1.5102714987710000, 1.5494499023847100, 1.5920646971778500,
        1.2157096692878600, 1.5309898758646300, 2.5625290242258700, 2.7889724810402800,
        2.4259003483243100, 2.3384808102697800, 1.3467386531914800, 2.9751840279216300,
        2.9856586375538600, 2.8949365278016000, 1.8215295770313500, 1.4984366853491000,
        0.6950953172047700, 1.4508730130511100, 1.6159184694779600, 1.2708530599561900,
        1.2980052002977500, 1.6128600373250000, 1.7155429169799300, 1.4331625867290800,
        1.3391060450912800, 0.8365300353244950, 2.3512268074347900, 2.8846607252846900,
        3.1747972376200700, 2.2215478837963400, 1.0088830457491100, 1.0669582934679300,
        0.9587596153363990, 0.9272108713771640, 1.1265345090142600, 1.0058976091034300,
        2.6544961857196200, 3.1249159988710100, 3.0290997012313700, 2.3735511370096900,
        1.1393967702253700, 1.4588077323622900, 2.2741591852814500, 2.0299495067611900,
        1.6840820941985000, 1.2214213032365200, 1.4336736030212700, 1.6421632074796900,
        2.1140730829372900, 1.9991073007720200, 0.8830770068346250, 0.9383229721156790,
        1.4745168700289600, 1.7499771427078700, 1.7253572093917300, 2.4166660919539500,
        2.9690516162572900, 2.7335905874874500, 2.0293206991503300, 5.6255459735033700,
        7.6871992624102100, 8.8223022788839000, 8.0386149304466600, 3.6023561733954100,
        2.3994567093406800, 1.5805797986814800, 1.8121499662003700, 1.8121499662003700,
        1.6039544881323800, 0.7973205127174390, 1.4479623613892700, 1.9676966229579200,
        2.4162636445553700, 2.9816157532452100, 5.8928161773467900, 5.9794466717247300,
        5.1535788535735000, 3.7097843603099100, 1.9427673561185900, 1.2647944497031900,
        1.6790860609271900, 2.8326145166612400, 3.0891192110373500, 3.1426441892139200,
        3.0783165041951100, 3.1535087600956500, 3.1276976356419100, 2.3545169355942200,
        1.8090660573898300, 1.8090660573898300, 3.3929662538846500, 2.9602905262828500,
        2.6983921138337200, 2.0266672149122100, 2.3634995240109500, 1.0844952743096700,
        0.9783302101029080, 2.5078237577628900, 2.9184105262968000, 2.6730881018028500,
        3.1496666490281100, 1.9988496691847600, 1.8948350851723200, 1.4473873013122600,
        0.8374783579293250, 0.3541186241925160, 0.8517041739947050, 2.8274582225030300,
        3.7925716868636800, 4.1720055129397900, 2.9942661872318500, 1.3745799358349500,
        1.2741546217002100, 2.4285942435903100, 2.9392430998473000, 2.3920765037933000,
        1.6236471291509100, 0.6289117585162470, 0.7876547467006080, 1.2304389460676200,
        3.6391372603956600, 3.9583860347368800, 5.0333587195827800, 4.2482255119049400,
        2.6826535370785400, 2.3521543316712900, 2.1187661503809200, 1.9736185041694300,
        1.9285149727186400, 2.2041823880976800, 3.4710257849805700, 3.3137561769086100,
        2.9181158304632100, 2.2836527757082500, 2.2779705880454200, 2.2911285428801200,
        1.7467398203510400, 1.7282737051751900, 1.6986465200270500, 3.0177110530996800,
        3.0854707906574000, 3.0305329564286200, 2.5408561549210100, 2.9105377510006600,
        2.5591307117847700, 1.1615291645068600, 1.0708174447589100, 1.0708174447589100,
        1.0442221985765300, 0.7724312267121250, 0.8691202448453260, 1.7250420284735100,
        1.6035678969098900, 2.2200292790862100, 2.8829793617020600, 2.4963433257466800,
        3.3349257862807100, 2.6274835870086800, 1.8375608833450900, 1.7551552637872200,
        1.3723811423944900, 1.8358050005379100, 3.1631076491324100, 3.0025405908996500,
        2.6264557867971000, 2.4065057656278300, 2.9730909841442800, 2.9960257008243400,
        2.0051982445633700, 1.3530410193338600, 1.8356797106249200, 1.9813429788908300,
        2.3279218199931000, 2.0846174709044300, 1.5803891925725100, 1.6377026592150400,
        2.1369323807739000, 3.3032907834461100, 3.5350487974001100, 3.4159288634279300,
        4.4282761883152700, 3.9908921308399200, 3.5081519351362200, 1.7871429713372100,
        1.0433599570618000, 0.3792360742334530, 7.2751508575424100, 8.0478767386187000,
        7.8262046995973700, 6.2186670597484200, 1.6449376887894600, 0.9799999999999990,
        2.0018916054572000, 1.9692562047636200, 2.0094899850459600, 1.6329972443332500,
        2.5388028674948300, 2.6148651972902900, 1.9793761643507800, 1.7651827100898100,
        2.5753019240469600, 2.4598414583058000, 1.3917542886587400, 1.3678815738213600,
        1.1515641536623100, 0.8320757177084320, 1.7327290613364800, 4.2998465088884300,
        6.0992811051795300, 5.8893488604428900, 3.7231277173903100, 1.5500580634286000,
        1.4911170309536400, 1.0993270668913800, 0.7906200098656740, 0.9607653199403050,
        3.6265231282869300, 5.7397935502943000, 6.0615204363261900, 5.2240855659148600,
        2.5776287552710100, 3.6208120636122500, 4.0998024342643600, 3.9729019620423600,
        2.3154330048610800, 1.0536460506261100, 1.1815963777872700, 1.1305618072445200,
        1.2528447629295500, 0.8477145746063360, 0.9030614597024950, 0.8523614256874840,
        0.9054556863811730, 0.7675480440988690, 0.5595801997926650, 0.7986739009132570,
    };
}

fn createStdDev(allocator: std.mem.Allocator, length: usize, unbiased: bool) !StandardDeviation {
    var sd = try StandardDeviation.init(allocator, .{ .length = length, .is_unbiased = unbiased });
    sd.fixSlices();
    return sd;
}

test "stdev population" {
    const input = testInput();
    var sd = try createStdDev(testing.allocator, 5, false);
    defer sd.deinit();
    const expected = expectedPopulation();

    for (0..4) |i| {
        try testing.expect(math.isNan(sd.update(input[i])));
    }
    for (4..252) |i| {
        const act = sd.update(input[i]);
        try testing.expect(almostEqual(act, expected[i], 1e-10));
    }
    try testing.expect(math.isNan(sd.update(math.nan(f64))));
}

test "stdev sample" {
    const input = testInput();
    var sd = try createStdDev(testing.allocator, 5, true);
    defer sd.deinit();
    const expected = expectedSample();

    for (0..4) |i| {
        try testing.expect(math.isNan(sd.update(input[i])));
    }
    for (4..252) |i| {
        const act = sd.update(input[i]);
        try testing.expect(almostEqual(act, expected[i], 1e-10));
    }
    try testing.expect(math.isNan(sd.update(math.nan(f64))));
}

test "stdev is primed" {
    const input = testInput();
    var sd = try createStdDev(testing.allocator, 5, true);
    defer sd.deinit();

    try testing.expect(!sd.isPrimed());
    for (0..4) |i| {
        _ = sd.update(input[i]);
        try testing.expect(!sd.isPrimed());
    }
    _ = sd.update(input[4]);
    try testing.expect(sd.isPrimed());
}

test "stdev metadata population" {
    var sd = try createStdDev(testing.allocator, 7, false);
    defer sd.deinit();
    var m: Metadata = undefined;
    sd.getMetadata(&m);

    try testing.expectEqual(Identifier.standard_deviation, m.identifier);
    try testing.expectEqualStrings("stdev.p(7)", m.outputs_buf[0].mnemonic);
    try testing.expectEqualStrings("Standard deviation based on estimation of the population variance stdev.p(7)", m.outputs_buf[0].description);
}

test "stdev metadata sample" {
    var sd = try createStdDev(testing.allocator, 7, true);
    defer sd.deinit();
    var m: Metadata = undefined;
    sd.getMetadata(&m);

    try testing.expectEqual(Identifier.standard_deviation, m.identifier);
    try testing.expectEqualStrings("stdev.s(7)", m.outputs_buf[0].mnemonic);
    try testing.expectEqualStrings("Standard deviation based on unbiased estimation of the sample variance stdev.s(7)", m.outputs_buf[0].description);
}

test "stdev update entity" {
    const length: usize = 3;
    const inp: f64 = 3.0;
    const exp: f64 = @sqrt(inp * inp / @as(f64, @floatFromInt(length)));
    const time: i64 = 1617235200;

    var sd = try createStdDev(testing.allocator, length, true);
    defer sd.deinit();
    _ = sd.update(0.0);
    _ = sd.update(0.0);
    const out = sd.updateScalar(&.{ .time = time, .value = inp });
    try testing.expectEqual(@as(usize, 1), out.len);
    const s = out.slice()[0].scalar;
    try testing.expectEqual(time, s.time);
    try testing.expect(almostEqual(s.value, exp, 1e-13));
}

test "stdev init invalid length" {
    const r1 = StandardDeviation.init(testing.allocator, .{ .length = 1 });
    try testing.expectError(error.InvalidLength, r1);
    const r0 = StandardDeviation.init(testing.allocator, .{ .length = 0 });
    try testing.expectError(error.InvalidLength, r0);
}

test "stdev mnemonic components" {
    {
        var sd = try createStdDev(testing.allocator, 5, true);
        defer sd.deinit();
        try testing.expectEqualStrings("stdev.s(5)", sd.line.mnemonic);
    }
    {
        var sd = try StandardDeviation.init(testing.allocator, .{
            .length = 5,
            .is_unbiased = true,
            .bar_component = .median,
        });
        defer sd.deinit();
        sd.fixSlices();
        try testing.expectEqualStrings("stdev.s(5, hl/2)", sd.line.mnemonic);
    }
}
