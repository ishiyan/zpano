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
    use super::super::testdata::testdata;
    const TOLERANCE: f64 = 1e-10;

    fn close_enough(exp: f64, got: f64) -> bool {
        if exp.is_nan() {
            return got.is_nan();
        }
        (exp - got).abs() <= TOLERANCE
    }

    #[test]
    fn test_atcf_update() {
        let input = testdata::test_input();
        let snaps = testdata::test_snapshots();

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
        let input = testdata::test_input();
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
