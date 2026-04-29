use crate::entities::bar::Bar;
use crate::entities::quote::Quote;
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::metadata::Metadata;

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum AdaptiveTrendAndCycleFilterOutput {
    Fatl = 1,
    Satl = 2,
    Rftl = 3,
    Rstl = 4,
    Rbci = 5,
    Ftlm = 6,
    Stlm = 7,
    Pcci = 8,
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

pub struct AdaptiveTrendAndCycleFilterParams;

impl Default for AdaptiveTrendAndCycleFilterParams {
    fn default() -> Self {
        Self
    }
}

// ---------------------------------------------------------------------------
// FIR filter
// ---------------------------------------------------------------------------

struct FirFilter {
    window: Vec<f64>,
    coeffs: &'static [f64],
    count: usize,
    primed: bool,
    value: f64,
}

impl FirFilter {
    fn new(coeffs: &'static [f64]) -> Self {
        Self {
            window: vec![0.0; coeffs.len()],
            coeffs,
            count: 0,
            primed: false,
            value: f64::NAN,
        }
    }

    fn is_primed(&self) -> bool {
        self.primed
    }

    fn update(&mut self, sample: f64) -> f64 {
        if self.primed {
            self.window.copy_within(1.., 0);
            let last = self.window.len() - 1;
            self.window[last] = sample;

            let mut sum = 0.0;
            for i in 0..self.window.len() {
                sum += self.window[i] * self.coeffs[i];
            }
            self.value = sum;
            self.value
        } else {
            self.window[self.count] = sample;
            self.count += 1;

            if self.count == self.window.len() {
                self.primed = true;
                let mut sum = 0.0;
                for i in 0..self.window.len() {
                    sum += self.window[i] * self.coeffs[i];
                }
                self.value = sum;
            }

            self.value
        }
    }
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Vladimir Kravchuk's Adaptive Trend and Cycle Filter (ATCF) suite.
pub struct AdaptiveTrendAndCycleFilter {
    fatl: FirFilter,
    satl: FirFilter,
    rftl: FirFilter,
    rstl: FirFilter,
    rbci: FirFilter,

    ftlm_value: f64,
    stlm_value: f64,
    pcci_value: f64,

    mnemonic: String,
    description: String,

    mnemonic_fatl: String,
    description_fatl: String,
    mnemonic_satl: String,
    description_satl: String,
    mnemonic_rftl: String,
    description_rftl: String,
    mnemonic_rstl: String,
    description_rstl: String,
    mnemonic_rbci: String,
    description_rbci: String,
    mnemonic_ftlm: String,
    description_ftlm: String,
    mnemonic_stlm: String,
    description_stlm: String,
    mnemonic_pcci: String,
    description_pcci: String,
}

impl AdaptiveTrendAndCycleFilter {
    /// Creates a new ATCF instance with default parameters.
    pub fn new() -> Result<Self, String> {
        let mnemonic = "atcf()".to_string();
        let description = format!("Adaptive trend and cycle filter {}", &mnemonic);

        let mk_sub = |name: &str, full: &str| -> (String, String) {
            let m = format!("{}()", name);
            let d = format!("{} {}", full, m);
            (m, d)
        };

        let (m_fatl, d_fatl) = mk_sub("fatl", "Fast Adaptive Trend Line");
        let (m_satl, d_satl) = mk_sub("satl", "Slow Adaptive Trend Line");
        let (m_rftl, d_rftl) = mk_sub("rftl", "Reference Fast Trend Line");
        let (m_rstl, d_rstl) = mk_sub("rstl", "Reference Slow Trend Line");
        let (m_rbci, d_rbci) = mk_sub("rbci", "Range Bound Channel Index");
        let (m_ftlm, d_ftlm) = mk_sub("ftlm", "Fast Trend Line Momentum");
        let (m_stlm, d_stlm) = mk_sub("stlm", "Slow Trend Line Momentum");
        let (m_pcci, d_pcci) = mk_sub("pcci", "Perfect Commodity Channel Index");

        Ok(Self {
            fatl: FirFilter::new(&FATL_COEFFICIENTS),
            satl: FirFilter::new(&SATL_COEFFICIENTS),
            rftl: FirFilter::new(&RFTL_COEFFICIENTS),
            rstl: FirFilter::new(&RSTL_COEFFICIENTS),
            rbci: FirFilter::new(&RBCI_COEFFICIENTS),
            ftlm_value: f64::NAN,
            stlm_value: f64::NAN,
            pcci_value: f64::NAN,
            mnemonic,
            description,
            mnemonic_fatl: m_fatl,
            description_fatl: d_fatl,
            mnemonic_satl: m_satl,
            description_satl: d_satl,
            mnemonic_rftl: m_rftl,
            description_rftl: d_rftl,
            mnemonic_rstl: m_rstl,
            description_rstl: d_rstl,
            mnemonic_rbci: m_rbci,
            description_rbci: d_rbci,
            mnemonic_ftlm: m_ftlm,
            description_ftlm: d_ftlm,
            mnemonic_stlm: m_stlm,
            description_stlm: d_stlm,
            mnemonic_pcci: m_pcci,
            description_pcci: d_pcci,
        })
    }

    /// Updates all filters with the next sample value.
    /// Returns (fatl, satl, rftl, rstl, rbci, ftlm, stlm, pcci).
    pub fn update(&mut self, sample: f64) -> (f64, f64, f64, f64, f64, f64, f64, f64) {
        if sample.is_nan() {
            let nan = f64::NAN;
            return (nan, nan, nan, nan, nan, nan, nan, nan);
        }

        let fatl = self.fatl.update(sample);
        let satl = self.satl.update(sample);
        let rftl = self.rftl.update(sample);
        let rstl = self.rstl.update(sample);
        let rbci = self.rbci.update(sample);

        if self.fatl.is_primed() && self.rftl.is_primed() {
            self.ftlm_value = fatl - rftl;
        }

        if self.satl.is_primed() && self.rstl.is_primed() {
            self.stlm_value = satl - rstl;
        }

        if self.fatl.is_primed() {
            self.pcci_value = sample - fatl;
        }

        (fatl, satl, rftl, rstl, rbci, self.ftlm_value, self.stlm_value, self.pcci_value)
    }
}

impl Indicator for AdaptiveTrendAndCycleFilter {
    fn is_primed(&self) -> bool {
        self.rstl.is_primed()
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::AdaptiveTrendAndCycleFilter,
            &self.mnemonic,
            &self.description,
            &[
                OutputText { mnemonic: self.mnemonic_fatl.clone(), description: self.description_fatl.clone() },
                OutputText { mnemonic: self.mnemonic_satl.clone(), description: self.description_satl.clone() },
                OutputText { mnemonic: self.mnemonic_rftl.clone(), description: self.description_rftl.clone() },
                OutputText { mnemonic: self.mnemonic_rstl.clone(), description: self.description_rstl.clone() },
                OutputText { mnemonic: self.mnemonic_rbci.clone(), description: self.description_rbci.clone() },
                OutputText { mnemonic: self.mnemonic_ftlm.clone(), description: self.description_ftlm.clone() },
                OutputText { mnemonic: self.mnemonic_stlm.clone(), description: self.description_stlm.clone() },
                OutputText { mnemonic: self.mnemonic_pcci.clone(), description: self.description_pcci.clone() },
            ],
        )
    }

    fn update_scalar(&mut self, scalar: &Scalar) -> Output {
        let (fatl, satl, rftl, rstl, rbci, ftlm, stlm, pcci) = self.update(scalar.value);
        let t = scalar.time;
        vec![
            Box::new(Scalar::new(t, fatl)),
            Box::new(Scalar::new(t, satl)),
            Box::new(Scalar::new(t, rftl)),
            Box::new(Scalar::new(t, rstl)),
            Box::new(Scalar::new(t, rbci)),
            Box::new(Scalar::new(t, ftlm)),
            Box::new(Scalar::new(t, stlm)),
            Box::new(Scalar::new(t, pcci)),
        ]
    }

    fn update_bar(&mut self, bar: &Bar) -> Output {
        let sample = bar.close;
        let (fatl, satl, rftl, rstl, rbci, ftlm, stlm, pcci) = self.update(sample);
        let t = bar.time;
        vec![
            Box::new(Scalar::new(t, fatl)),
            Box::new(Scalar::new(t, satl)),
            Box::new(Scalar::new(t, rftl)),
            Box::new(Scalar::new(t, rstl)),
            Box::new(Scalar::new(t, rbci)),
            Box::new(Scalar::new(t, ftlm)),
            Box::new(Scalar::new(t, stlm)),
            Box::new(Scalar::new(t, pcci)),
        ]
    }

    fn update_quote(&mut self, quote: &Quote) -> Output {
        let sample = (quote.bid_price + quote.ask_price) / 2.0;
        let (fatl, satl, rftl, rstl, rbci, ftlm, stlm, pcci) = self.update(sample);
        let t = quote.time;
        vec![
            Box::new(Scalar::new(t, fatl)),
            Box::new(Scalar::new(t, satl)),
            Box::new(Scalar::new(t, rftl)),
            Box::new(Scalar::new(t, rstl)),
            Box::new(Scalar::new(t, rbci)),
            Box::new(Scalar::new(t, ftlm)),
            Box::new(Scalar::new(t, stlm)),
            Box::new(Scalar::new(t, pcci)),
        ]
    }

    fn update_trade(&mut self, trade: &Trade) -> Output {
        let sample = trade.price;
        let (fatl, satl, rftl, rstl, rbci, ftlm, stlm, pcci) = self.update(sample);
        let t = trade.time;
        vec![
            Box::new(Scalar::new(t, fatl)),
            Box::new(Scalar::new(t, satl)),
            Box::new(Scalar::new(t, rftl)),
            Box::new(Scalar::new(t, rstl)),
            Box::new(Scalar::new(t, rbci)),
            Box::new(Scalar::new(t, ftlm)),
            Box::new(Scalar::new(t, stlm)),
            Box::new(Scalar::new(t, pcci)),
        ]
    }
}

// ---------------------------------------------------------------------------
// Coefficients
// ---------------------------------------------------------------------------

static FATL_COEFFICIENTS: [f64; 39] = [
    0.0040364019004036386962421862, 0.0130129076013012957968308448, 0.000786016000078601746116832, 0.0005541115000554108210219855, -0.0047717710004771784587179668,
    -0.0072003400007200276742901798, -0.0067093714006709378328730376, -0.002382462300238249230464677, 0.0040444064004044386936567327, 0.009571141900957106908521166,
    0.0110573605011056964284725581, 0.0069480557006948077557780087, -0.0016060704001606094812392607, -0.0108597376010859964923047548, -0.0160483392016047948163864379,
    -0.0136744850013673955831413446, -0.0036771622003677188122766093, 0.0100299086010029967603395219, 0.0208778257020877932564622982, 0.0226522218022651926833323579,
    0.0128149838012814958607602322, -0.0055774838005577481984727324, -0.0244141482024413921142301306, -0.0338917071033891890529786056, -0.027243253702724291200429054,
    -0.0047706151004770584590913225, 0.0249252327024924919491498371, 0.0477818607047781845664589924, 0.0502044896050203837839498576, 0.0259609206025960916146226454,
    -0.0190795053019079938373197875, -0.0670110374067010783554349176, -0.0933058722093305698622032764, -0.0760367731076036754401222862, -0.0054034585005403482546829043,
    0.1104506886110449643244275786, 0.2460452079246049205273978404, 0.3658689069365868818243430595, 0.4360409450436038591587747509,
];

static SATL_COEFFICIENTS: [f64; 65] = [
    0.016138097598386190240161381, 0.0049516077995048392200495161, 0.0056078228994392177100560782, 0.0062325476993767452300623255, 0.0068163568993183643100681636,
    0.0073260525992673947400732605, 0.0077543819992245618000775438, 0.0080741358991925864100807414, 0.008290102199170989780082901, 0.0083694797991630520200836948,
    0.0083037665991696233400830377, 0.0080376627991962337200803766, 0.0076266452992373354700762665, 0.0070340084992965991500703401, 0.0062194590993780540900621946,
    0.0052380200994761979900523802, 0.0040471368995952863100404714, 0.0026845692997315430700268457, 0.0011421468998857853100114215, -0.0005535179999446482000055352,
    -0.0023956943997604305600239569, -0.0043466730995653326900434667, -0.0063841849993615815000638418, -0.0084736769991526323000847368, -0.0105938330989406166901059383,
    -0.0126796775987320322401267968, -0.0147139427985286057201471394, -0.0166377698983362230101663777, -0.018412699198158730080184127, -0.0199924533980007546601999245,
    -0.0213300462978669953702133005, -0.0223796899977620310002237969, -0.0231017776976898222302310178, -0.0234566314976543368502345663, -0.0234080862976591913702340809,
    -0.0229204860977079513902292049, -0.0219739145978026085402197391, -0.0205446726979455327302054467, -0.0186164871981383512801861649, -0.0161875264983812473501618753,
    -0.0132507214986749278501325072, -0.0098190255990180974400981903, -0.0059060081994093991800590601, -0.0015350358998464964100153504, 0.00326399789967360021003264,
    0.0084512447991548755200845124, 0.0139807862986019213701398079, 0.0198005182980199481701980052, 0.0258537720974146227902585377, 0.0320735367967926463203207354,
    0.0383959949961604005003839599, 0.0447468228955253177104474682, 0.0510534241948946575805105342, 0.0572428924942757107505724289, 0.0632381577936761842206323816,
    0.0689666681931033331806896667, 0.0743569345925643065407435693, 0.0793406349920659365007934063, 0.0838544302916145569708385443, 0.087839100591216089940878391,
    0.0912437089908756291009124371, 0.0940230543905976945609402305, 0.0961401077903859892209614011, 0.0975682268902431773109756823, 0.0982862173901713782609828622,
];

static RFTL_COEFFICIENTS: [f64; 44] = [
    0.0018747783, 0.0060440751, 0.0003650790, 0.0002573669, -0.0022163335,
    -0.0033443253, -0.0031162862, -0.0011065767, 0.0018784961, 0.0044454862,
    0.0051357867, 0.0032271474, -0.0007459678, -0.0050439973, -0.0074539350,
    -0.0063513565, -0.0017079230, 0.0046585685, 0.0096970755, 0.0105212252,
    0.0059521459, -0.0025905610, -0.0113395830, -0.0157416029, -0.0126536111,
    -0.0022157966, 0.0115769653, 0.0221931304, 0.0233183633, 0.0120580088,
    -0.0088618137, -0.0311244617, -0.0433375629, -0.0353166244, -0.0025097319,
    0.0513007762, 0.1142800493, 0.1699342860, 0.2025269304, 0.2025269304,
    0.1699342860, 0.1142800493, 0.0513007762, -0.0025097319,
];

static RSTL_COEFFICIENTS: [f64; 91] = [
    0.0073925494970429788, 0.0022682354990927055, 0.0025688348989724658, 0.002855009198857996, 0.0031224408987510226,
    0.00335592259865763, 0.0035521319985791465, 0.0036986050985205569, 0.0037975349984809849, 0.0038338963984664407,
    0.0038037943984784812, 0.0036818973985272402, 0.003493618298602552, 0.0032221428987111419, 0.0028490135988603941,
    0.0023994353990402255, 0.0018539148992584337, 0.0012297490995081001, 0.00052319529979072182, -0.00025355589989857757,
    -0.0010974210995610314, -0.001991126699203549, -0.0029244712988302111, -0.0038816270984473483, -0.0048528294980588671,
    -0.005808314397676673, -0.0067401717973039291, -0.0076214396969514226, -0.0084345003966261982, -0.0091581550963367366,
    -0.0097708804960916461, -0.010251701895899317, -0.010582476295767008, -0.010745027995701987, -0.010722790395710882,
    -0.010499430195800226, -0.010065824095973669, -0.0094111160962355514, -0.0085278516965888573, -0.0074151918970339218,
    -0.0060698984975720389, -0.0044979051982008368, -0.0027054277989178284, -0.00070317019971873182, 0.00149517409940193,
    0.0038713512984514587, 0.0064043270974382671, 0.0090702333963719045, 0.011843111595262752, 0.01469226519412309,
    0.017588460592964612, 0.020497651691800935, 0.023386583490645364, 0.026221858789511249, 0.028968173588412725,
    0.031592293087363076, 0.034061469586375404, 0.03634440608546223, 0.038412088184635158, 0.040237388383905039,
    0.0417969734832812, 0.043070137682771938, 0.044039918782384023, 0.044694112382122349, 0.04502300998199079,
    0.04502300998199079, 0.044694112382122349, 0.044039918782384023, 0.043070137682771938, 0.0417969734832812,
    0.040237388383905039, 0.038412088184635158, 0.03634440608546223, 0.034061469586375404, 0.031592293087363076,
    0.028968173588412725, 0.026221858789511249, 0.023386583490645364, 0.020497651691800935, 0.017588460592964612,
    0.01469226519412309, 0.011843111595262752, 0.0090702333963719045, 0.0064043270974382671, 0.0038713512984514587,
    0.00149517409940193, -0.00070317019971873182, -0.0027054277989178284, -0.0044979051982008368, -0.0060698984975720389,
    -0.0074151918970339218,
];

static RBCI_COEFFICIENTS: [f64; 56] = [
    1.6156174062090192153914095277, 1.3775160858518416893554976293, 1.5136918536280435656798483244, 1.2766707742770234133790334563, 0.6386689877404132301203554117,
    -0.3089253210608743300469836813, -1.3536792507159717290810388558, -2.2289941407052666020200196315, -2.6973742493750332214376893622, -2.6270409969741336827525619917,
    -2.0577410867291241943560079078, -1.1887841547760696822235971887, -0.3278853541689465187629951569, 0.2245901590801639067569342685, 0.2797065817943275162276668425,
    -0.1561848847902538433044469068, -0.8771442472997222096084165948, -1.5412722887852520460759366626, -1.7969987452428928478844892329, -1.4202166850952351050428400987,
    -0.4132650218556106245769805601, 0.9760510632634910606018990454, 2.332625807295967101587012479, 3.2216514733634133981714563696, 3.3589597011460702965326006902,
    2.7322958715740864679722928674, 1.627491649276702400877203685, 0.5359717984550392511937237318, -0.026072229548611708427086738, 0.2740437898620496022136827326,
    1.4310126661567721970936015234, 3.0671459994827321970515735232, 4.5422535558908452685778180309, 5.18085572453087762982600249, 4.5358834718545357895708540006,
    2.5919387157740506799120888755, -0.1815496242348328581385472914, -2.9604409038745131520847249669, -4.8510863196511920220117945255, -5.2342243578350788396599493861,
    -4.0433304530469835823678064195, -1.8617342916118854621877471345, 0.2191111443489335227889210799, 0.9559212015487508488278798383, -0.581752756415990711571147056,
    -4.5964240181996169037378163513, -10.352401329008687575349519179, -16.270239152740363170620070073, -20.326611695861686666411613999, -20.656621157742740599133621415,
    -16.17628165220480541756739088, -7.0231637350320332896825897512, 5.3418475974485313054566284411, 18.427745065038146870717437163, 29.333989817203741958061329161,
    35.524182142487838212180677809,
];

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;

    fn test_input() -> Vec<f64> {
        vec![
            92.0000, 93.1725, 95.3125, 94.8450, 94.4075, 94.1100, 93.5000, 91.7350, 90.9550, 91.6875,
            94.5000, 97.9700, 97.5775, 90.7825, 89.0325, 92.0950, 91.1550, 89.7175, 90.6100, 91.0000,
            88.9225, 87.5150, 86.4375, 83.8900, 83.0025, 82.8125, 82.8450, 86.7350, 86.8600, 87.5475,
            85.7800, 86.1725, 86.4375, 87.2500, 88.9375, 88.2050, 85.8125, 84.5950, 83.6575, 84.4550,
            83.5000, 86.7825, 88.1725, 89.2650, 90.8600, 90.7825, 91.8600, 90.3600, 89.8600, 90.9225,
            89.5000, 87.6725, 86.5000, 84.2825, 82.9075, 84.2500, 85.6875, 86.6100, 88.2825, 89.5325,
            89.5000, 88.0950, 90.6250, 92.2350, 91.6725, 92.5925, 93.0150, 91.1725, 90.9850, 90.3775,
            88.2500, 86.9075, 84.0925, 83.1875, 84.2525, 97.8600, 99.8750, 103.2650, 105.9375, 103.5000,
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
            106.9700, 110.0300, 91.0000, 93.5600, 93.6200, 95.3100, 94.1850, 94.7800, 97.6250, 97.5900,
            95.2500, 94.7200, 92.2200, 91.5650, 92.2200, 93.8100, 95.5900, 96.1850, 94.6250, 95.1200,
            94.0000, 93.7450, 95.9050, 101.7450, 106.4400, 107.9350, 103.4050, 105.0600, 104.1550, 103.3100,
            103.3450, 104.8400, 110.4050, 114.5000, 117.3150, 118.2500, 117.1850, 109.7500, 109.6550, 108.5300,
            106.2200, 107.7200, 109.8400, 109.0950, 109.0900, 109.1550, 109.3150, 109.0600, 109.9050, 109.6250,
            109.5300, 108.0600,
        ]
    }

    struct AtcfSnap {
        i: usize,
        fatl: f64,
        satl: f64,
        rftl: f64,
        rstl: f64,
        rbci: f64,
        ftlm: f64,
        stlm: f64,
        pcci: f64,
    }

    fn test_snapshots() -> Vec<AtcfSnap> {
        let n = f64::NAN;
        vec![
            AtcfSnap { i: 0, fatl: n, satl: n, rftl: n, rstl: n, rbci: n, ftlm: n, stlm: n, pcci: n },
            AtcfSnap { i: 38, fatl: 84.9735715498821, satl: n, rftl: n, rstl: n, rbci: n, ftlm: n, stlm: n, pcci: -1.3160715498821 },
            AtcfSnap { i: 39, fatl: 84.4518660416872, satl: n, rftl: n, rstl: n, rbci: n, ftlm: n, stlm: n, pcci: 0.0031339583128 },
            AtcfSnap { i: 43, fatl: 88.2793028340854, satl: n, rftl: 84.9781981272507, rstl: n, rbci: n, ftlm: 3.3011047068347, stlm: n, pcci: 0.9856971659146 },
            AtcfSnap { i: 44, fatl: 90.3071933727095, satl: n, rftl: 85.3111711946473, rstl: n, rbci: n, ftlm: 4.9960221780622, stlm: n, pcci: 0.5528066272905 },
            AtcfSnap { i: 55, fatl: 83.5737547263234, satl: n, rftl: 87.6545375029340, rstl: n, rbci: -701.3930208567576, ftlm: -4.0807827766106, stlm: n, pcci: 0.6762452736766 },
            AtcfSnap { i: 56, fatl: 84.2004074439195, satl: n, rftl: 86.4101353078987, rstl: n, rbci: -596.7632782263086, ftlm: -2.2097278639792, stlm: n, pcci: 1.4870925560805 },
            AtcfSnap { i: 64, fatl: 91.3026041176860, satl: 89.8909098632724, rftl: 89.2605446508615, rstl: n, rbci: 260.0958399205915, ftlm: 2.0420594668245, stlm: n, pcci: 0.3698958823140 },
            AtcfSnap { i: 65, fatl: 91.9122247829182, satl: 90.3013166280409, rftl: 90.0608560382592, rstl: n, rbci: 271.4055284612814, ftlm: 1.8513687446590, stlm: n, pcci: 0.6802752170818 },
            AtcfSnap { i: 90, fatl: 115.0676036598003, satl: 109.5130909788342, rftl: 106.9904903948140, rstl: 91.0255929287335, rbci: 648.4101282691054, ftlm: 8.0771132649863, stlm: 18.4874980501007, pcci: 4.9323963401997 },
            AtcfSnap { i: 91, fatl: 117.8447026727287, satl: 111.5377810965825, rftl: 108.9908122410267, rstl: 91.4218609612485, rbci: 750.5214819459538, ftlm: 8.8538904317020, stlm: 20.1159201353340, pcci: 2.0302973272713 },
            AtcfSnap { i: 100, fatl: 112.8634350429428, satl: 119.4023289602100, rftl: 115.8265249211198, rstl: 97.7871686087879, rbci: -617.3149799371608, ftlm: -2.9630898781769, stlm: 21.6151603514221, pcci: 1.8815649570572 },
            AtcfSnap { i: 150, fatl: 121.5097808704445, satl: 124.0945687443045, rftl: 123.2003217712845, rstl: 127.9357790331669, rbci: -268.9358266646477, ftlm: -1.6905409008400, stlm: -3.8412102888624, pcci: 1.2702191295555 },
            AtcfSnap { i: 200, fatl: 106.1833142820738, satl: 109.8912725552509, rftl: 109.8071754394800, rstl: 127.4173713354640, rbci: -592.7380669351005, ftlm: -3.6238611574062, stlm: -17.5260987802131, pcci: 0.7866857179262 },
            AtcfSnap { i: 251, fatl: 108.1030068950443, satl: 114.1981767327412, rftl: 110.1319723971535, rstl: 102.4461386298790, rbci: -312.3373212974634, ftlm: -2.0289655021092, stlm: 11.7520381028621, pcci: -0.0430068950443 },
        ]
    }

    const TOLERANCE: f64 = 1e-10;

    fn close_enough(exp: f64, got: f64) -> bool {
        if exp.is_nan() {
            return got.is_nan();
        }
        (exp - got).abs() <= TOLERANCE
    }

    #[test]
    fn test_atcf_update() {
        let input = test_input();
        let snaps = test_snapshots();

        let mut x = AdaptiveTrendAndCycleFilter::new().unwrap();
        let mut si = 0;

        for i in 0..input.len() {
            let (fatl, satl, rftl, rstl, rbci, ftlm, stlm, pcci) = x.update(input[i]);

            if si < snaps.len() && snaps[si].i == i {
                let s = &snaps[si];
                assert!(close_enough(s.fatl, fatl), "[{}] fatl: expected {}, got {}", i, s.fatl, fatl);
                assert!(close_enough(s.satl, satl), "[{}] satl: expected {}, got {}", i, s.satl, satl);
                assert!(close_enough(s.rftl, rftl), "[{}] rftl: expected {}, got {}", i, s.rftl, rftl);
                assert!(close_enough(s.rstl, rstl), "[{}] rstl: expected {}, got {}", i, s.rstl, rstl);
                assert!(close_enough(s.rbci, rbci), "[{}] rbci: expected {}, got {}", i, s.rbci, rbci);
                assert!(close_enough(s.ftlm, ftlm), "[{}] ftlm: expected {}, got {}", i, s.ftlm, ftlm);
                assert!(close_enough(s.stlm, stlm), "[{}] stlm: expected {}, got {}", i, s.stlm, stlm);
                assert!(close_enough(s.pcci, pcci), "[{}] pcci: expected {}, got {}", i, s.pcci, pcci);
                si += 1;
            }
        }

        assert_eq!(si, snaps.len(), "did not hit all snapshots");
    }

    #[test]
    fn test_atcf_primes_at_bar_90() {
        let input = test_input();
        let mut x = AdaptiveTrendAndCycleFilter::new().unwrap();

        assert!(!x.is_primed());

        let mut primed_at: Option<usize> = None;
        for i in 0..input.len() {
            x.update(input[i]);
            if x.is_primed() && primed_at.is_none() {
                primed_at = Some(i);
            }
        }

        assert_eq!(primed_at, Some(90));
    }

    #[test]
    fn test_atcf_nan_input() {
        let mut x = AdaptiveTrendAndCycleFilter::new().unwrap();
        let (fatl, satl, rftl, rstl, rbci, ftlm, stlm, pcci) = x.update(f64::NAN);

        for v in [fatl, satl, rftl, rstl, rbci, ftlm, stlm, pcci] {
            assert!(v.is_nan(), "expected NaN output for NaN input, got {}", v);
        }

        assert!(!x.is_primed());
    }

    #[test]
    fn test_atcf_metadata() {
        let x = AdaptiveTrendAndCycleFilter::new().unwrap();
        let md = x.metadata();

        assert_eq!(md.identifier, Identifier::AdaptiveTrendAndCycleFilter);
        assert_eq!(md.mnemonic, "atcf()");
        assert_eq!(md.description, "Adaptive trend and cycle filter atcf()");
        assert_eq!(md.outputs.len(), 8);

        let expected_mnemonics = ["fatl()", "satl()", "rftl()", "rstl()", "rbci()", "ftlm()", "stlm()", "pcci()"];
        let expected_kinds = [1, 2, 3, 4, 5, 6, 7, 8];

        for i in 0..8 {
            assert_eq!(md.outputs[i].kind, expected_kinds[i]);
            assert_eq!(md.outputs[i].mnemonic, expected_mnemonics[i]);
        }
    }
}
