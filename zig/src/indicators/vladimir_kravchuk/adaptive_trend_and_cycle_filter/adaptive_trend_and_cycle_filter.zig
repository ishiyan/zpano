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
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;
const LineIndicator = line_indicator_mod.LineIndicator;

// =============================================================================
// FIR Coefficients
// =============================================================================

/// 39 taps of the Fast Adaptive Trend Line filter.
const fatl_coefficients = [39]f64{
    0.0040364019004036386962421862,  0.0130129076013012957968308448,  0.000786016000078601746116832,   0.0005541115000554108210219855,  -0.0047717710004771784587179668,
    -0.0072003400007200276742901798, -0.0067093714006709378328730376, -0.002382462300238249230464677,  0.0040444064004044386936567327,  0.009571141900957106908521166,
    0.0110573605011056964284725581,  0.0069480557006948077557780087,  -0.0016060704001606094812392607, -0.0108597376010859964923047548, -0.0160483392016047948163864379,
    -0.0136744850013673955831413446, -0.0036771622003677188122766093, 0.0100299086010029967603395219,  0.0208778257020877932564622982,  0.0226522218022651926833323579,
    0.0128149838012814958607602322,  -0.0055774838005577481984727324, -0.0244141482024413921142301306, -0.0338917071033891890529786056, -0.027243253702724291200429054,
    -0.0047706151004770584590913225, 0.0249252327024924919491498371,  0.0477818607047781845664589924,  0.0502044896050203837839498576,  0.0259609206025960916146226454,
    -0.0190795053019079938373197875, -0.0670110374067010783554349176, -0.0933058722093305698622032764, -0.0760367731076036754401222862, -0.0054034585005403482546829043,
    0.1104506886110449643244275786,  0.2460452079246049205273978404,  0.3658689069365868818243430595,  0.4360409450436038591587747509,
};

/// 65 taps of the Slow Adaptive Trend Line filter.
const satl_coefficients = [65]f64{
    0.016138097598386190240161381,   0.0049516077995048392200495161,  0.0056078228994392177100560782,  0.0062325476993767452300623255,  0.0068163568993183643100681636,
    0.0073260525992673947400732605,  0.0077543819992245618000775438,  0.0080741358991925864100807414,  0.008290102199170989780082901,   0.0083694797991630520200836948,
    0.0083037665991696233400830377,  0.0080376627991962337200803766,  0.0076266452992373354700762665,  0.0070340084992965991500703401,  0.0062194590993780540900621946,
    0.0052380200994761979900523802,  0.0040471368995952863100404714,  0.0026845692997315430700268457,  0.0011421468998857853100114215,  -0.0005535179999446482000055352,
    -0.0023956943997604305600239569, -0.0043466730995653326900434667, -0.0063841849993615815000638418, -0.0084736769991526323000847368, -0.0105938330989406166901059383,
    -0.0126796775987320322401267968, -0.0147139427985286057201471394, -0.0166377698983362230101663777, -0.018412699198158730080184127,  -0.0199924533980007546601999245,
    -0.0213300462978669953702133005, -0.0223796899977620310002237969, -0.0231017776976898222302310178, -0.0234566314976543368502345663, -0.0234080862976591913702340809,
    -0.0229204860977079513902292049, -0.0219739145978026085402197391, -0.0205446726979455327302054467, -0.0186164871981383512801861649, -0.0161875264983812473501618753,
    -0.0132507214986749278501325072, -0.0098190255990180974400981903, -0.0059060081994093991800590601, -0.0015350358998464964100153504, 0.00326399789967360021003264,
    0.0084512447991548755200845124,  0.0139807862986019213701398079,  0.0198005182980199481701980052,  0.0258537720974146227902585377,  0.0320735367967926463203207354,
    0.0383959949961604005003839599,  0.0447468228955253177104474682,  0.0510534241948946575805105342,  0.0572428924942757107505724289,  0.0632381577936761842206323816,
    0.0689666681931033331806896667,  0.0743569345925643065407435693,  0.0793406349920659365007934063,  0.0838544302916145569708385443,  0.087839100591216089940878391,
    0.0912437089908756291009124371,  0.0940230543905976945609402305,  0.0961401077903859892209614011,  0.0975682268902431773109756823,  0.0982862173901713782609828622,
};

/// 44 taps of the Reference Fast Trend Line filter.
const rftl_coefficients = [44]f64{
    0.0018747783,  0.0060440751,  0.0003650790,  0.0002573669,  -0.0022163335,
    -0.0033443253, -0.0031162862, -0.0011065767, 0.0018784961,  0.0044454862,
    0.0051357867,  0.0032271474,  -0.0007459678, -0.0050439973, -0.0074539350,
    -0.0063513565, -0.0017079230, 0.0046585685,  0.0096970755,  0.0105212252,
    0.0059521459,  -0.0025905610, -0.0113395830, -0.0157416029, -0.0126536111,
    -0.0022157966, 0.0115769653,  0.0221931304,  0.0233183633,  0.0120580088,
    -0.0088618137, -0.0311244617, -0.0433375629, -0.0353166244, -0.0025097319,
    0.0513007762,  0.1142800493,  0.1699342860,  0.2025269304,  0.2025269304,
    0.1699342860,  0.1142800493,  0.0513007762,  -0.0025097319,
};

/// 91 taps of the Reference Slow Trend Line filter.
const rstl_coefficients = [91]f64{
    0.0073925494970429788,  0.0022682354990927055,   0.0025688348989724658,  0.002855009198857996,    0.0031224408987510226,
    0.00335592259865763,    0.0035521319985791465,   0.0036986050985205569,  0.0037975349984809849,   0.0038338963984664407,
    0.0038037943984784812,  0.0036818973985272402,   0.003493618298602552,   0.0032221428987111419,   0.0028490135988603941,
    0.0023994353990402255,  0.0018539148992584337,   0.0012297490995081001,  0.00052319529979072182,  -0.00025355589989857757,
    -0.0010974210995610314, -0.001991126699203549,   -0.0029244712988302111, -0.0038816270984473483,  -0.0048528294980588671,
    -0.005808314397676673,  -0.0067401717973039291,  -0.0076214396969514226, -0.0084345003966261982,  -0.0091581550963367366,
    -0.0097708804960916461, -0.010251701895899317,   -0.010582476295767008,  -0.010745027995701987,   -0.010722790395710882,
    -0.010499430195800226,  -0.010065824095973669,   -0.0094111160962355514, -0.0085278516965888573,  -0.0074151918970339218,
    -0.0060698984975720389, -0.0044979051982008368,  -0.0027054277989178284, -0.00070317019971873182, 0.00149517409940193,
    0.0038713512984514587,  0.0064043270974382671,   0.0090702333963719045,  0.011843111595262752,    0.01469226519412309,
    0.017588460592964612,   0.020497651691800935,    0.023386583490645364,   0.026221858789511249,    0.028968173588412725,
    0.031592293087363076,   0.034061469586375404,    0.03634440608546223,    0.038412088184635158,    0.040237388383905039,
    0.0417969734832812,     0.043070137682771938,    0.044039918782384023,   0.044694112382122349,    0.04502300998199079,
    0.04502300998199079,    0.044694112382122349,    0.044039918782384023,   0.043070137682771938,    0.0417969734832812,
    0.040237388383905039,   0.038412088184635158,    0.03634440608546223,    0.034061469586375404,    0.031592293087363076,
    0.028968173588412725,   0.026221858789511249,    0.023386583490645364,   0.020497651691800935,    0.017588460592964612,
    0.01469226519412309,    0.011843111595262752,    0.0090702333963719045,  0.0064043270974382671,   0.0038713512984514587,
    0.00149517409940193,    -0.00070317019971873182, -0.0027054277989178284, -0.0044979051982008368,  -0.0060698984975720389,
    -0.0074151918970339218,
};

/// 56 taps of the Range Bound Channel Index filter.
const rbci_coefficients = [56]f64{
    1.6156174062090192153914095277,  1.3775160858518416893554976293,  1.5136918536280435656798483244,  1.2766707742770234133790334563,  0.6386689877404132301203554117,
    -0.3089253210608743300469836813, -1.3536792507159717290810388558, -2.2289941407052666020200196315, -2.6973742493750332214376893622, -2.6270409969741336827525619917,
    -2.0577410867291241943560079078, -1.1887841547760696822235971887, -0.3278853541689465187629951569, 0.2245901590801639067569342685,  0.2797065817943275162276668425,
    -0.1561848847902538433044469068, -0.8771442472997222096084165948, -1.5412722887852520460759366626, -1.7969987452428928478844892329, -1.4202166850952351050428400987,
    -0.4132650218556106245769805601, 0.9760510632634910606018990454,  2.332625807295967101587012479,   3.2216514733634133981714563696,  3.3589597011460702965326006902,
    2.7322958715740864679722928674,  1.627491649276702400877203685,   0.5359717984550392511937237318,  -0.026072229548611708427086738,  0.2740437898620496022136827326,
    1.4310126661567721970936015234,  3.0671459994827321970515735232,  4.5422535558908452685778180309,  5.18085572453087762982600249,    4.5358834718545357895708540006,
    2.5919387157740506799120888755,  -0.1815496242348328581385472914, -2.9604409038745131520847249669, -4.8510863196511920220117945255, -5.2342243578350788396599493861,
    -4.0433304530469835823678064195, -1.8617342916118854621877471345, 0.2191111443489335227889210799,  0.9559212015487508488278798383,  -0.581752756415990711571147056,
    -4.5964240181996169037378163513, -10.352401329008687575349519179, -16.270239152740363170620070073, -20.326611695861686666411613999, -20.656621157742740599133621415,
    -16.17628165220480541756739088,  -7.0231637350320332896825897512, 5.3418475974485313054566284411,  18.427745065038146870717437163,  29.333989817203741958061329161,
    35.524182142487838212180677809,
};

// =============================================================================
// FIR Filter
// =============================================================================

/// Internal FIR engine shared by all five ATCF lines.
const FirFilter = struct {
    window: []f64,
    coeffs: []const f64,
    count: usize,
    primed: bool,
    value: f64,

    fn init(allocator: std.mem.Allocator, coeffs: []const f64) !FirFilter {
        const window = try allocator.alloc(f64, coeffs.len);
        return FirFilter{
            .window = window,
            .coeffs = coeffs,
            .count = 0,
            .primed = false,
            .value = math.nan(f64),
        };
    }

    fn deinit(self: *FirFilter, allocator: std.mem.Allocator) void {
        allocator.free(self.window);
    }

    fn isPrimed(self: *const FirFilter) bool {
        return self.primed;
    }

    fn update(self: *FirFilter, sample: f64) f64 {
        if (self.primed) {
            // Shift window left by 1.
            const len = self.window.len;
            for (0..len - 1) |i| {
                self.window[i] = self.window[i + 1];
            }
            self.window[len - 1] = sample;

            var sum: f64 = 0.0;
            for (self.window, self.coeffs) |w, c| {
                sum += w * c;
            }
            self.value = sum;
            return self.value;
        }

        self.window[self.count] = sample;
        self.count += 1;

        if (self.count == self.window.len) {
            self.primed = true;

            var sum: f64 = 0.0;
            for (self.window, self.coeffs) |w, c| {
                sum += w * c;
            }
            self.value = sum;
        }

        return self.value;
    }
};

// =============================================================================
// Adaptive Trend and Cycle Filter
// =============================================================================

/// Vladimir Kravchuk's Adaptive Trend and Cycle Filter (ATCF) suite.
///
/// Exposes eight scalar outputs: five FIR filters (FATL, SATL, RFTL, RSTL,
/// RBCI) plus three composites (FTLM = FATL - RFTL, STLM = SATL - RSTL,
/// PCCI = sample - FATL).
pub const AdaptiveTrendAndCycleFilter = struct {
    fatl: FirFilter,
    satl: FirFilter,
    rftl: FirFilter,
    rstl: FirFilter,
    rbci: FirFilter,

    ftlm_value: f64,
    stlm_value: f64,
    pcci_value: f64,

    line: LineIndicator,
    allocator: std.mem.Allocator,

    // Mnemonic/description buffers.
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    // Sub-mnemonics and sub-descriptions (8 outputs).
    sub_mnemonic_bufs: [8][64]u8,
    sub_mnemonic_lens: [8]usize,
    sub_description_bufs: [8][128]u8,
    sub_description_lens: [8]usize,

    pub const Error = error{
        OutOfMemory,
        MnemonicTooLong,
    };

    pub fn init(allocator: std.mem.Allocator) Error!AdaptiveTrendAndCycleFilter {
        return initWithComponents(allocator, null, null, null);
    }

    pub fn initWithComponents(
        allocator: std.mem.Allocator,
        bc: ?bar_component.BarComponent,
        qc: ?quote_component.QuoteComponent,
        tc: ?trade_component.TradeComponent,
    ) Error!AdaptiveTrendAndCycleFilter {
        const actual_bc = bc orelse bar_component.default_bar_component;
        const actual_qc = qc orelse quote_component.default_quote_component;
        const actual_tc = tc orelse trade_component.default_trade_component;

        // Build component mnemonic.
        var ctm_buf: [64]u8 = undefined;
        const ctm = component_triple_mnemonic_mod.componentTripleMnemonic(&ctm_buf, actual_bc, actual_qc, actual_tc);

        // Top-level mnemonic: "atcf()" or "atcf(hl/2)".
        var top_arg_buf: [64]u8 = undefined;
        var top_arg_len: usize = 0;
        var sub_arg_buf: [64]u8 = undefined;
        var sub_arg_len: usize = 0;

        if (ctm.len > 0) {
            // ctm starts with ", " — skip it.
            const stripped = ctm[2..];
            @memcpy(top_arg_buf[0..stripped.len], stripped);
            top_arg_len = stripped.len;
            @memcpy(sub_arg_buf[0..stripped.len], stripped);
            sub_arg_len = stripped.len;
        }

        const top_arg = top_arg_buf[0..top_arg_len];
        const sub_arg = sub_arg_buf[0..sub_arg_len];

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "atcf({s})", .{top_arg}) catch return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Adaptive trend and cycle filter atcf({s})", .{top_arg}) catch return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        // Build sub-mnemonics and sub-descriptions for all 8 outputs.
        const sub_names = [8][]const u8{ "fatl", "satl", "rftl", "rstl", "rbci", "ftlm", "stlm", "pcci" };
        const sub_fulls = [8][]const u8{
            "Fast Adaptive Trend Line",
            "Slow Adaptive Trend Line",
            "Reference Fast Trend Line",
            "Reference Slow Trend Line",
            "Range Bound Channel Index",
            "Fast Trend Line Momentum",
            "Slow Trend Line Momentum",
            "Perfect Commodity Channel Index",
        };

        var sub_mnemonic_bufs: [8][64]u8 = undefined;
        var sub_mnemonic_lens: [8]usize = undefined;
        var sub_description_bufs: [8][128]u8 = undefined;
        var sub_description_lens: [8]usize = undefined;

        for (0..8) |i| {
            const sm = std.fmt.bufPrint(&sub_mnemonic_bufs[i], "{s}({s})", .{ sub_names[i], sub_arg }) catch return error.MnemonicTooLong;
            sub_mnemonic_lens[i] = sm.len;

            const sd = std.fmt.bufPrint(&sub_description_bufs[i], "{s} {s}({s})", .{ sub_fulls[i], sub_names[i], sub_arg }) catch return error.MnemonicTooLong;
            sub_description_lens[i] = sd.len;
        }

        // Create FIR filters.
        var fatl_filter = FirFilter.init(allocator, &fatl_coefficients) catch return error.OutOfMemory;
        errdefer fatl_filter.deinit(allocator);
        var satl_filter = FirFilter.init(allocator, &satl_coefficients) catch return error.OutOfMemory;
        errdefer satl_filter.deinit(allocator);
        var rftl_filter = FirFilter.init(allocator, &rftl_coefficients) catch return error.OutOfMemory;
        errdefer rftl_filter.deinit(allocator);
        var rstl_filter = FirFilter.init(allocator, &rstl_coefficients) catch return error.OutOfMemory;
        errdefer rstl_filter.deinit(allocator);
        var rbci_filter = FirFilter.init(allocator, &rbci_coefficients) catch return error.OutOfMemory;
        errdefer rbci_filter.deinit(allocator);

        const line = LineIndicator.new(
            mnemonic_buf[0..mnemonic_len],
            description_buf[0..description_len],
            bc,
            qc,
            tc,
        );

        return AdaptiveTrendAndCycleFilter{
            .fatl = fatl_filter,
            .satl = satl_filter,
            .rftl = rftl_filter,
            .rstl = rstl_filter,
            .rbci = rbci_filter,
            .ftlm_value = math.nan(f64),
            .stlm_value = math.nan(f64),
            .pcci_value = math.nan(f64),
            .line = line,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
            .sub_mnemonic_bufs = sub_mnemonic_bufs,
            .sub_mnemonic_lens = sub_mnemonic_lens,
            .sub_description_bufs = sub_description_bufs,
            .sub_description_lens = sub_description_lens,
        };
    }

    pub fn deinit(self: *AdaptiveTrendAndCycleFilter) void {
        self.fatl.deinit(self.allocator);
        self.satl.deinit(self.allocator);
        self.rftl.deinit(self.allocator);
        self.rstl.deinit(self.allocator);
        self.rbci.deinit(self.allocator);
    }

    pub fn fixSlices(self: *AdaptiveTrendAndCycleFilter) void {
        // Fix line indicator mnemonic/description to point at our embedded buffers.
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
        // Fix FIR filter coefficient pointers.
        self.fatl.coeffs = &fatl_coefficients;
        self.satl.coeffs = &satl_coefficients;
        self.rftl.coeffs = &rftl_coefficients;
        self.rstl.coeffs = &rstl_coefficients;
        self.rbci.coeffs = &rbci_coefficients;
    }

    /// Core update: feed the next sample to all five FIR filters and recompute composites.
    /// Returns all 8 outputs as a struct.
    pub fn update(self: *AdaptiveTrendAndCycleFilter, sample: f64) struct {
        fatl: f64,
        satl: f64,
        rftl: f64,
        rstl: f64,
        rbci: f64,
        ftlm: f64,
        stlm: f64,
        pcci: f64,
    } {
        if (math.isNan(sample)) {
            const nan = math.nan(f64);
            return .{ .fatl = nan, .satl = nan, .rftl = nan, .rstl = nan, .rbci = nan, .ftlm = nan, .stlm = nan, .pcci = nan };
        }

        const fatl = self.fatl.update(sample);
        const satl = self.satl.update(sample);
        const rftl = self.rftl.update(sample);
        const rstl = self.rstl.update(sample);
        const rbci = self.rbci.update(sample);

        if (self.fatl.isPrimed() and self.rftl.isPrimed()) {
            self.ftlm_value = fatl - rftl;
        }

        if (self.satl.isPrimed() and self.rstl.isPrimed()) {
            self.stlm_value = satl - rstl;
        }

        if (self.fatl.isPrimed()) {
            self.pcci_value = sample - fatl;
        }

        return .{
            .fatl = fatl,
            .satl = satl,
            .rftl = rftl,
            .rstl = rstl,
            .rbci = rbci,
            .ftlm = self.ftlm_value,
            .stlm = self.stlm_value,
            .pcci = self.pcci_value,
        };
    }

    pub fn isPrimed(self: *const AdaptiveTrendAndCycleFilter) bool {
        return self.rstl.isPrimed();
    }

    pub fn getMetadata(self: *const AdaptiveTrendAndCycleFilter, out: *Metadata) void {
        const mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        const description = self.description_buf[0..self.description_len];

        build_metadata_mod.buildMetadata(out, Identifier.adaptive_trend_and_cycle_filter, mnemonic, description, &.{
            .{ .mnemonic = self.sub_mnemonic_bufs[0][0..self.sub_mnemonic_lens[0]], .description = self.sub_description_bufs[0][0..self.sub_description_lens[0]] },
            .{ .mnemonic = self.sub_mnemonic_bufs[1][0..self.sub_mnemonic_lens[1]], .description = self.sub_description_bufs[1][0..self.sub_description_lens[1]] },
            .{ .mnemonic = self.sub_mnemonic_bufs[2][0..self.sub_mnemonic_lens[2]], .description = self.sub_description_bufs[2][0..self.sub_description_lens[2]] },
            .{ .mnemonic = self.sub_mnemonic_bufs[3][0..self.sub_mnemonic_lens[3]], .description = self.sub_description_bufs[3][0..self.sub_description_lens[3]] },
            .{ .mnemonic = self.sub_mnemonic_bufs[4][0..self.sub_mnemonic_lens[4]], .description = self.sub_description_bufs[4][0..self.sub_description_lens[4]] },
            .{ .mnemonic = self.sub_mnemonic_bufs[5][0..self.sub_mnemonic_lens[5]], .description = self.sub_description_bufs[5][0..self.sub_description_lens[5]] },
            .{ .mnemonic = self.sub_mnemonic_bufs[6][0..self.sub_mnemonic_lens[6]], .description = self.sub_description_bufs[6][0..self.sub_description_lens[6]] },
            .{ .mnemonic = self.sub_mnemonic_bufs[7][0..self.sub_mnemonic_lens[7]], .description = self.sub_description_bufs[7][0..self.sub_description_lens[7]] },
        });
    }

    fn makeOutput(_: *const AdaptiveTrendAndCycleFilter, time: i64, r: anytype) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = Scalar{ .time = time, .value = r.fatl } });
        out.append(.{ .scalar = Scalar{ .time = time, .value = r.satl } });
        out.append(.{ .scalar = Scalar{ .time = time, .value = r.rftl } });
        out.append(.{ .scalar = Scalar{ .time = time, .value = r.rstl } });
        out.append(.{ .scalar = Scalar{ .time = time, .value = r.rbci } });
        out.append(.{ .scalar = Scalar{ .time = time, .value = r.ftlm } });
        out.append(.{ .scalar = Scalar{ .time = time, .value = r.stlm } });
        out.append(.{ .scalar = Scalar{ .time = time, .value = r.pcci } });
        return out;
    }

    pub fn updateScalar(self: *AdaptiveTrendAndCycleFilter, sample: *const Scalar) OutputArray {
        const r = self.update(sample.value);
        return self.makeOutput(sample.time, r);
    }

    pub fn updateBar(self: *AdaptiveTrendAndCycleFilter, sample: *const Bar) OutputArray {
        const v = self.line.extractBar(sample);
        const r = self.update(v);
        return self.makeOutput(sample.time, r);
    }

    pub fn updateQuote(self: *AdaptiveTrendAndCycleFilter, sample: *const Quote) OutputArray {
        const v = self.line.extractQuote(sample);
        const r = self.update(v);
        return self.makeOutput(sample.time, r);
    }

    pub fn updateTrade(self: *AdaptiveTrendAndCycleFilter, sample: *const Trade) OutputArray {
        const v = self.line.extractTrade(sample);
        const r = self.update(v);
        return self.makeOutput(sample.time, r);
    }

    // --- Indicator interface ---

    pub fn indicator(self: *AdaptiveTrendAndCycleFilter) indicator_mod.Indicator {
        return indicator_mod.Indicator{
            .ptr = @ptrCast(self),
            .vtable = &.{
                .isPrimed = vtableIsPrimed,
                .metadata = vtableMetadata,
                .updateScalar = vtableUpdateScalar,
                .updateBar = vtableUpdateBar,
                .updateQuote = vtableUpdateQuote,
                .updateTrade = vtableUpdateTrade,
            },
        };
    }

    fn vtableIsPrimed(ptr: *anyopaque) bool {
        const self: *const AdaptiveTrendAndCycleFilter = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const AdaptiveTrendAndCycleFilter = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *AdaptiveTrendAndCycleFilter = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *AdaptiveTrendAndCycleFilter = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *AdaptiveTrendAndCycleFilter = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *AdaptiveTrendAndCycleFilter = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }
};

// =============================================================================
// Tests
// =============================================================================

fn closeEnough(exp: f64, got: f64) bool {
    if (math.isNan(exp)) return math.isNan(got);
    return @abs(exp - got) <= testdata.test_tolerance;
}

const testdata = @import("testdata.zig");

test "atcf update" {
    const allocator = std.testing.allocator;
    var ind = try AdaptiveTrendAndCycleFilter.init(allocator);
    defer ind.deinit();

    const input = testdata.testATCFInput();
    const snaps = testdata.testATCFSnapshots();

    var si: usize = 0;

    for (0..252) |i| {
        const r = ind.update(input[i]);

        if (si < snaps.len and snaps[si].i == i) {
            const s = snaps[si];
            if (!closeEnough(s.fatl, r.fatl)) {
                std.debug.print("[{d}] fatl: expected {d}, got {d}\n", .{ i, s.fatl, r.fatl });
                return error.TestUnexpectedResult;
            }
            if (!closeEnough(s.satl, r.satl)) {
                std.debug.print("[{d}] satl: expected {d}, got {d}\n", .{ i, s.satl, r.satl });
                return error.TestUnexpectedResult;
            }
            if (!closeEnough(s.rftl, r.rftl)) {
                std.debug.print("[{d}] rftl: expected {d}, got {d}\n", .{ i, s.rftl, r.rftl });
                return error.TestUnexpectedResult;
            }
            if (!closeEnough(s.rstl, r.rstl)) {
                std.debug.print("[{d}] rstl: expected {d}, got {d}\n", .{ i, s.rstl, r.rstl });
                return error.TestUnexpectedResult;
            }
            if (!closeEnough(s.rbci, r.rbci)) {
                std.debug.print("[{d}] rbci: expected {d}, got {d}\n", .{ i, s.rbci, r.rbci });
                return error.TestUnexpectedResult;
            }
            if (!closeEnough(s.ftlm, r.ftlm)) {
                std.debug.print("[{d}] ftlm: expected {d}, got {d}\n", .{ i, s.ftlm, r.ftlm });
                return error.TestUnexpectedResult;
            }
            if (!closeEnough(s.stlm, r.stlm)) {
                std.debug.print("[{d}] stlm: expected {d}, got {d}\n", .{ i, s.stlm, r.stlm });
                return error.TestUnexpectedResult;
            }
            if (!closeEnough(s.pcci, r.pcci)) {
                std.debug.print("[{d}] pcci: expected {d}, got {d}\n", .{ i, s.pcci, r.pcci });
                return error.TestUnexpectedResult;
            }
            si += 1;
        }
    }

    try std.testing.expectEqual(snaps.len, si);
}

test "atcf primes at bar 90" {
    const allocator = std.testing.allocator;
    var ind = try AdaptiveTrendAndCycleFilter.init(allocator);
    defer ind.deinit();

    try std.testing.expect(!ind.isPrimed());

    const input = testdata.testATCFInput();
    var primed_at: ?usize = null;

    for (0..252) |i| {
        _ = ind.update(input[i]);
        if (ind.isPrimed() and primed_at == null) {
            primed_at = i;
        }
    }

    try std.testing.expectEqual(@as(?usize, 90), primed_at);
}

test "atcf nan input" {
    const allocator = std.testing.allocator;
    var ind = try AdaptiveTrendAndCycleFilter.init(allocator);
    defer ind.deinit();

    const r = ind.update(math.nan(f64));
    try std.testing.expect(math.isNan(r.fatl));
    try std.testing.expect(math.isNan(r.satl));
    try std.testing.expect(math.isNan(r.rftl));
    try std.testing.expect(math.isNan(r.rstl));
    try std.testing.expect(math.isNan(r.rbci));
    try std.testing.expect(math.isNan(r.ftlm));
    try std.testing.expect(math.isNan(r.stlm));
    try std.testing.expect(math.isNan(r.pcci));
    try std.testing.expect(!ind.isPrimed());
}

test "atcf metadata" {
    const allocator = std.testing.allocator;
    var ind = try AdaptiveTrendAndCycleFilter.init(allocator);
    defer ind.deinit();

    var m: Metadata = undefined;
    ind.getMetadata(&m);

    try std.testing.expectEqual(Identifier.adaptive_trend_and_cycle_filter, m.identifier);
    try std.testing.expectEqualStrings("atcf()", m.mnemonic);
    try std.testing.expectEqual(@as(usize, 8), m.outputs_len);

    const outs = m.outputs_buf[0..m.outputs_len];
    try std.testing.expectEqual(@as(i32, 1), outs[0].kind);
    try std.testing.expectEqual(@as(i32, 2), outs[1].kind);
    try std.testing.expectEqual(@as(i32, 3), outs[2].kind);
    try std.testing.expectEqual(@as(i32, 4), outs[3].kind);
    try std.testing.expectEqual(@as(i32, 5), outs[4].kind);
    try std.testing.expectEqual(@as(i32, 6), outs[5].kind);
    try std.testing.expectEqual(@as(i32, 7), outs[6].kind);
    try std.testing.expectEqual(@as(i32, 8), outs[7].kind);

    try std.testing.expectEqualStrings("fatl()", outs[0].mnemonic);
    try std.testing.expectEqualStrings("satl()", outs[1].mnemonic);
    try std.testing.expectEqualStrings("rftl()", outs[2].mnemonic);
    try std.testing.expectEqualStrings("rstl()", outs[3].mnemonic);
    try std.testing.expectEqualStrings("rbci()", outs[4].mnemonic);
    try std.testing.expectEqualStrings("ftlm()", outs[5].mnemonic);
    try std.testing.expectEqualStrings("stlm()", outs[6].mnemonic);
    try std.testing.expectEqualStrings("pcci()", outs[7].mnemonic);
}

test "atcf update bar" {
    const allocator = std.testing.allocator;
    var ind = try AdaptiveTrendAndCycleFilter.init(allocator);
    defer ind.deinit();

    const input = testdata.testATCFInput();

    // Prime with 100 samples.
    for (0..100) |i| {
        _ = ind.update(input[i]);
    }

    const bar = Bar{
        .time = 12345,
        .open = 100.0,
        .high = 100.0,
        .low = 100.0,
        .close = 100.0,
        .volume = 0,
    };
    const out = ind.updateBar(&bar);
    const items = out.slice();
    try std.testing.expectEqual(@as(usize, 8), items.len);
    try std.testing.expectEqual(@as(i64, 12345), items[0].scalar.time);
}

test "atcf update scalar" {
    const allocator = std.testing.allocator;
    var ind = try AdaptiveTrendAndCycleFilter.init(allocator);
    defer ind.deinit();

    const input = testdata.testATCFInput();

    for (0..100) |i| {
        _ = ind.update(input[i]);
    }

    const s = Scalar{ .time = 99999, .value = 100.0 };
    const out = ind.updateScalar(&s);
    const items = out.slice();
    try std.testing.expectEqual(@as(usize, 8), items.len);
    try std.testing.expectEqual(@as(i64, 99999), items[0].scalar.time);
}
