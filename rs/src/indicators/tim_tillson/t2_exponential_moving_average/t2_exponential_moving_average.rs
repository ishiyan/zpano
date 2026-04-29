use crate::entities::bar::Bar;
use crate::entities::bar_component::{component_value as bar_component_value, BarComponent, DEFAULT_BAR_COMPONENT};
use crate::entities::quote::Quote;
use crate::entities::quote_component::{component_value as quote_component_value, QuoteComponent, DEFAULT_QUOTE_COMPONENT};
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::entities::trade_component::{component_value as trade_component_value, TradeComponent, DEFAULT_TRADE_COMPONENT};
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::component_triple_mnemonic::component_triple_mnemonic;
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::line_indicator::LineIndicator;
use crate::indicators::core::metadata::Metadata;

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create a T2 EMA from length.
pub struct T2ExponentialMovingAverageLengthParams {
    pub length: i64,
    pub volume_factor: f64,
    pub first_is_average: bool,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for T2ExponentialMovingAverageLengthParams {
    fn default() -> Self {
        Self {
            length: 5,
            volume_factor: 0.7,
            first_is_average: true,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

/// Parameters to create a T2 EMA from smoothing factor.
pub struct T2ExponentialMovingAverageSmoothingFactorParams {
    pub smoothing_factor: f64,
    pub volume_factor: f64,
    pub first_is_average: bool,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for T2ExponentialMovingAverageSmoothingFactorParams {
    fn default() -> Self {
        Self {
            smoothing_factor: 0.3333,
            volume_factor: 0.7,
            first_is_average: true,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum T2ExponentialMovingAverageOutput {
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// T2 Exponential Moving Average (T2, T2EMA).
///
/// A four-pole non-linear Kalman filter developed by Tim Tillson.
///
/// T2 = c1*e4 + c2*e3 + c3*e2 where:
///   c1 = v^2, c2 = -2v(1+v), c3 = (1+v)^2
pub struct T2ExponentialMovingAverage {
    line: LineIndicator,
    smoothing_factor: f64,
    c1: f64,
    c2: f64,
    c3: f64,
    sum: f64,
    ema1: f64,
    ema2: f64,
    ema3: f64,
    ema4: f64,
    length: i64,
    length2: i64,
    length3: i64,
    length4: i64,
    count: i64,
    first_is_average: bool,
    primed: bool,
}

impl T2ExponentialMovingAverage {
    pub fn new_from_length(params: &T2ExponentialMovingAverageLengthParams) -> Result<Self, String> {
        Self::new_internal(params.length, f64::NAN, params.volume_factor, params.first_is_average,
            params.bar_component, params.quote_component, params.trade_component)
    }

    pub fn new_from_smoothing_factor(params: &T2ExponentialMovingAverageSmoothingFactorParams) -> Result<Self, String> {
        Self::new_internal(0, params.smoothing_factor, params.volume_factor, params.first_is_average,
            params.bar_component, params.quote_component, params.trade_component)
    }

    fn new_internal(
        length: i64,
        alpha: f64,
        v: f64,
        first_is_average: bool,
        bc_opt: Option<BarComponent>,
        qc_opt: Option<QuoteComponent>,
        tc_opt: Option<TradeComponent>,
    ) -> Result<Self, String> {
        const INVALID: &str = "invalid t2 exponential moving average parameters";
        const EPSILON: f64 = 0.00000001;

        if v < 0.0 || v > 1.0 {
            return Err(format!("{}: volume factor should be in range [0, 1]", INVALID));
        }

        let bc = bc_opt.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = qc_opt.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = tc_opt.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let (actual_length, actual_alpha, mnemonic);

        if alpha.is_nan() {
            if length < 2 {
                return Err(format!("{}: length should be greater than 1", INVALID));
            }
            actual_alpha = 2.0 / (1 + length) as f64;
            actual_length = length;
            mnemonic = format!("t2({}, {:.8}{})", length, v, component_triple_mnemonic(bc, qc, tc));
        } else {
            if alpha < 0.0 || alpha > 1.0 {
                return Err(format!("{}: smoothing factor should be in range [0, 1]", INVALID));
            }
            let clamped = if alpha < EPSILON { EPSILON } else { alpha };
            actual_length = (2.0_f64 / clamped).round() as i64 - 1;
            actual_alpha = clamped;
            mnemonic = format!("t2({}, {:.8}, {:.8}{})", actual_length, clamped, v, component_triple_mnemonic(bc, qc, tc));
        }

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let description = format!("T2 exponential moving average {}", mnemonic);
        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        let v1 = v + 1.0;
        let c1 = v * v;
        let c2 = -2.0 * v * v1;
        let c3 = v1 * v1;

        Ok(Self {
            line,
            smoothing_factor: actual_alpha,
            c1,
            c2,
            c3,
            sum: 0.0,
            ema1: 0.0,
            ema2: 0.0,
            ema3: 0.0,
            ema4: 0.0,
            length: actual_length,
            length2: 2 * actual_length - 1,
            length3: 3 * actual_length - 2,
            length4: 4 * actual_length - 3,
            count: 0,
            first_is_average,
            primed: false,
        })
    }

    /// Core update logic.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        let sf = self.smoothing_factor;

        if self.primed {
            let mut v1 = self.ema1;
            let mut v2 = self.ema2;
            let mut v3 = self.ema3;
            let mut v4 = self.ema4;
            v1 += (sample - v1) * sf;
            v2 += (v1 - v2) * sf;
            v3 += (v2 - v3) * sf;
            v4 += (v3 - v4) * sf;
            self.ema1 = v1;
            self.ema2 = v2;
            self.ema3 = v3;
            self.ema4 = v4;
            return self.c1 * v4 + self.c2 * v3 + self.c3 * v2;
        }

        self.count += 1;

        if self.first_is_average {
            if self.count == 1 {
                self.sum = sample;
            } else if self.length >= self.count {
                self.sum += sample;
                if self.length == self.count {
                    self.ema1 = self.sum / self.length as f64;
                    self.sum = self.ema1;
                }
            } else if self.length2 >= self.count {
                self.ema1 += (sample - self.ema1) * sf;
                self.sum += self.ema1;
                if self.length2 == self.count {
                    self.ema2 = self.sum / self.length as f64;
                    self.sum = self.ema2;
                }
            } else if self.length3 >= self.count {
                self.ema1 += (sample - self.ema1) * sf;
                self.ema2 += (self.ema1 - self.ema2) * sf;
                self.sum += self.ema2;
                if self.length3 == self.count {
                    self.ema3 = self.sum / self.length as f64;
                    self.sum = self.ema3;
                }
            } else {
                self.ema1 += (sample - self.ema1) * sf;
                self.ema2 += (self.ema1 - self.ema2) * sf;
                self.ema3 += (self.ema2 - self.ema3) * sf;
                self.sum += self.ema3;
                if self.length4 == self.count {
                    self.primed = true;
                    self.ema4 = self.sum / self.length as f64;
                    return self.c1 * self.ema4 + self.c2 * self.ema3 + self.c3 * self.ema2;
                }
            }
        } else {
            // Metastock
            if self.count == 1 {
                self.ema1 = sample;
            } else if self.length >= self.count {
                self.ema1 += (sample - self.ema1) * sf;
                if self.length == self.count {
                    self.ema2 = self.ema1;
                }
            } else if self.length2 >= self.count {
                self.ema1 += (sample - self.ema1) * sf;
                self.ema2 += (self.ema1 - self.ema2) * sf;
                if self.length2 == self.count {
                    self.ema3 = self.ema2;
                }
            } else if self.length3 >= self.count {
                self.ema1 += (sample - self.ema1) * sf;
                self.ema2 += (self.ema1 - self.ema2) * sf;
                self.ema3 += (self.ema2 - self.ema3) * sf;
                if self.length3 == self.count {
                    self.ema4 = self.ema3;
                }
            } else {
                self.ema1 += (sample - self.ema1) * sf;
                self.ema2 += (self.ema1 - self.ema2) * sf;
                self.ema3 += (self.ema2 - self.ema3) * sf;
                self.ema4 += (self.ema3 - self.ema4) * sf;
                if self.length4 == self.count {
                    self.primed = true;
                    return self.c1 * self.ema4 + self.c2 * self.ema3 + self.c3 * self.ema2;
                }
            }
        }

        f64::NAN
    }
}

impl Indicator for T2ExponentialMovingAverage {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::T2ExponentialMovingAverage,
            &self.line.mnemonic,
            &self.line.description,
            &[OutputText {
                mnemonic: self.line.mnemonic.clone(),
                description: self.line.description.clone(),
            }],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let value = self.update(sample.value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let sample_value = (self.line.bar_func)(sample);
        let value = self.update(sample_value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let sample_value = (self.line.quote_func)(sample);
        let value = self.update(sample_value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let sample_value = (self.line.trade_func)(sample);
        let value = self.update(sample_value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;
    use crate::entities::bar_component::BarComponent;
    use crate::entities::quote_component::QuoteComponent;
    use crate::entities::trade_component::TradeComponent;
    use crate::indicators::core::outputs::shape::Shape;

    #[allow(clippy::excessive_precision)]
    fn test_input() -> Vec<f64> {
        vec![
            91.500000, 94.815000, 94.375000, 95.095000, 93.780000, 94.625000, 92.530000, 92.750000, 90.315000, 92.470000,
            96.125000, 97.250000, 98.500000, 89.875000, 91.000000, 92.815000, 89.155000, 89.345000, 91.625000, 89.875000,
            88.375000, 87.625000, 84.780000, 83.000000, 83.500000, 81.375000, 84.440000, 89.250000, 86.375000, 86.250000,
            85.250000, 87.125000, 85.815000, 88.970000, 88.470000, 86.875000, 86.815000, 84.875000, 84.190000, 83.875000,
            83.375000, 85.500000, 89.190000, 89.440000, 91.095000, 90.750000, 91.440000, 89.000000, 91.000000, 90.500000,
            89.030000, 88.815000, 84.280000, 83.500000, 82.690000, 84.750000, 85.655000, 86.190000, 88.940000, 89.280000,
            88.625000, 88.500000, 91.970000, 91.500000, 93.250000, 93.500000, 93.155000, 91.720000, 90.000000, 89.690000,
            88.875000, 85.190000, 83.375000, 84.875000, 85.940000, 97.250000, 99.875000, 104.940000, 106.000000, 102.500000,
            102.405000, 104.595000, 106.125000, 106.000000, 106.065000, 104.625000, 108.625000, 109.315000, 110.500000,
            112.750000, 123.000000, 119.625000, 118.750000, 119.250000, 117.940000, 116.440000, 115.190000, 111.875000,
            110.595000, 118.125000, 116.000000, 116.000000, 112.000000, 113.750000, 112.940000, 116.000000, 120.500000,
            116.620000, 117.000000, 115.250000, 114.310000, 115.500000, 115.870000, 120.690000, 120.190000, 120.750000,
            124.750000, 123.370000, 122.940000, 122.560000, 123.120000, 122.560000, 124.620000, 129.250000, 131.000000,
            132.250000, 131.000000, 132.810000, 134.000000, 137.380000, 137.810000, 137.880000, 137.250000, 136.310000,
            136.250000, 134.630000, 128.250000, 129.000000, 123.870000, 124.810000, 123.000000, 126.250000, 128.380000,
            125.370000, 125.690000, 122.250000, 119.370000, 118.500000, 123.190000, 123.500000, 122.190000, 119.310000,
            123.310000, 121.120000, 123.370000, 127.370000, 128.500000, 123.870000, 122.940000, 121.750000, 124.440000,
            122.000000, 122.370000, 122.940000, 124.000000, 123.190000, 124.560000, 127.250000, 125.870000, 128.860000,
            132.000000, 130.750000, 134.750000, 135.000000, 132.380000, 133.310000, 131.940000, 130.000000, 125.370000,
            130.130000, 127.120000, 125.190000, 122.000000, 125.000000, 123.000000, 123.500000, 120.060000, 121.000000,
            117.750000, 119.870000, 122.000000, 119.190000, 116.370000, 113.500000, 114.250000, 110.000000, 105.060000,
            107.000000, 107.870000, 107.000000, 107.120000, 107.000000, 91.000000, 93.940000, 93.870000, 95.500000, 93.000000,
            94.940000, 98.250000, 96.750000, 94.810000, 94.370000, 91.560000, 90.250000, 93.940000, 93.620000, 97.000000,
            95.000000, 95.870000, 94.060000, 94.620000, 93.750000, 98.000000, 103.940000, 107.870000, 106.060000, 104.500000,
            105.000000, 104.190000, 103.060000, 103.420000, 105.270000, 111.870000, 116.000000, 116.620000, 118.280000,
            113.370000, 109.000000, 109.700000, 109.250000, 107.000000, 109.190000, 110.000000, 109.200000, 110.120000,
            108.000000, 108.620000, 109.750000, 109.810000, 109.000000, 108.750000, 107.870000,
        ]
    }

    /// Expected data from test_T3.xls: test_T2.xls, T2(5,0.7) column.
    #[allow(clippy::excessive_precision)]
    fn test_expected() -> Vec<f64> {
        vec![
            f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN,
            f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN,
            92.00445682439470, 90.91065523008910, 90.64635861170230, 90.30847892058210, 89.67203184711450, 88.90373302682710,
            87.58685794329080, 85.95483674752900, 84.70056892267930, 83.35230597480420, 83.01639757937170, 84.26606438522640,
            85.11761359114740, 85.64273951236740, 85.71656813204080, 86.07519575846890, 86.14040917330100, 86.84731118619480,
            87.49707083315330, 87.60487694128920, 87.48806230196600, 86.85037109765170, 86.00510559137550, 85.19270124163530,
            84.44671220986690, 84.39562430058820, 85.47846964605910, 86.77863391237980, 88.27480023486270, 89.43282744282170,
            90.38671380457610, 90.47144856898390, 90.70844123290330, 90.80974940198440, 90.47870378112050, 90.01014971788120,
            88.50241981733770, 86.77695943257320, 85.15371443592010, 84.41215046544750, 84.35591075095750, 84.68800316786790,
            85.75412809581360, 86.93319023556810, 87.74938215169110, 88.24027778265170, 89.34170209097540, 90.29228263445890,
            91.39771790246850, 92.36534535927510, 92.98185469287250, 92.98804953756390, 92.35873030252150, 91.56001444442250,
            90.66809450828110, 89.04851123746310, 87.09681376947570, 85.81489615554240, 85.29909582783180, 87.87278112697240,
            91.57930166956580, 96.09718145905370, 100.17004554905500, 102.36263587267900, 103.39235683035700, 104.28507155830600,
            105.25701353369700, 105.97310800537900, 106.42554279253100, 106.31201565835700, 106.95673692587200, 107.85537745407500,
            108.90779472362600, 110.30293857497500, 113.94246316874900, 116.72555020318500, 118.38770601306100, 119.39753212263300,
            119.61486726510600, 119.11850573798600, 118.15216910748100, 116.41234112714800, 114.45865964016000, 114.64061162124700,
            114.99246753582000, 115.33333018479700, 114.64625300711600, 114.18195589036400, 113.69963506110000, 114.05025810623800,
            115.68652710013500, 116.48211079384600, 116.91736166744200, 116.71412173592500, 116.10716021480600, 115.77116121989300,
            115.68159156874600, 116.84472084196100, 118.05979798379400, 119.16496881132300, 120.94593869721900, 122.23096010279300,
            122.95608471296700, 123.22656209567300, 123.39820574539500, 123.34529134741000, 123.70294192360900, 125.21356151217400,
            127.19355195747100, 129.19266295422300, 130.45262728011500, 131.60256664339700, 132.72482116981200, 134.37601221394000,
            135.92461267673900, 137.09424525891600, 137.70388483421700, 137.73781691122000, 137.52589963778700, 136.86260850470000,
            134.64367631336900, 132.54215730025200, 129.69729772115400, 127.44287309781500, 125.45301927369100, 124.78149677056900,
            125.23737028614000, 125.28834765521900, 125.29811473230200, 124.49118133074600, 122.91167204427900, 121.21791414741600,
            120.98626501620100, 121.39013089649900, 121.63114611836900, 121.09667781630300, 121.38604904393200, 121.37683602700200,
            121.81714963968400, 123.27680991970300, 125.04065301695500, 125.42231478992600, 125.02871976565900, 124.16184698358900,
            123.95348210659900, 123.42096848684700, 122.98054879392000, 122.80175403486300, 123.00506164290600, 123.10341273782800,
            123.47654621169500, 124.50686275946200, 125.20780013410700, 126.36056179556500, 128.16193338760800, 129.48687497042000,
            131.31008039649500, 132.94986405572100, 133.52752016483300, 133.82004244574000, 133.59810463804000, 132.76202894406400,
            130.77137039821000, 129.94273481747400, 128.95593356768700, 127.66988673048200, 125.80391233666600, 124.87150365869300,
            123.99029524198700, 123.45575776078900, 122.35139824579100, 121.53836332656800, 120.24081578533100, 119.58685232736100,
            119.81980247526100, 119.65181893967000, 118.73183613794600, 117.08643352117200, 115.74380263433200, 113.79159129198400,
            110.89385419602800, 108.78809510901800, 107.63783653194000, 106.88467024071700, 106.49446844969200, 106.31788640074800,
            102.45936340772700, 98.92286025711640, 96.26320146598560, 94.89830130673550, 93.71214241680890, 93.35493402958790,
            94.19091222655660, 94.93222890927330, 95.05352335355920, 94.85884233631840, 93.94374172767060, 92.68201046821450,
            92.46778442700140, 92.58855930337340, 93.61995692101500, 94.27141015939870, 94.87276841707030, 94.89831480519190,
            94.85776894321340, 94.59245060865460, 95.32163415260000, 97.64623243404460, 100.94595782878400, 103.42636980462600,
            104.70917474084900, 105.40458887459900, 105.52523492436600, 105.12339759255700, 104.68875644984300, 104.75482770270600,
            106.53547902116200, 109.48515731469700, 112.33128725513400, 114.90966925830500, 115.64161420440900, 114.51441441748600,
            113.13113208320500, 111.81663297883600, 110.24787839832700, 109.44340073972700, 109.26032936270900, 109.13994050792700,
            109.29451387663700, 109.00897104476500, 108.79138478114900, 108.91808951962300, 109.15266693557400, 109.18435708789600,
            109.08448284304800, 108.75304884708700,
        ]
    }

    const L: i64 = 5;
    const LPRIMED: usize = (4 * L - 4) as usize; // 16

    #[test]
    fn test_update_first_is_average_true() {
        let mut t2 = create_length(L, true, 0.7);
        let input = test_input();
        let exp = test_expected();

        for i in 0..LPRIMED {
            assert!(t2.update(input[i]).is_nan(), "[{}] should be NaN", i);
        }

        for i in LPRIMED..input.len() {
            let act = t2.update(input[i]);
            assert!(
                (exp[i] - act).abs() < 1e-8,
                "[{}] expected {}, got {}", i, exp[i], act
            );
        }

        assert!(t2.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_first_is_average_false() {
        let mut t2 = create_length(L, false, 0.7);
        let input = test_input();
        let exp = test_expected();

        for i in 0..LPRIMED {
            assert!(t2.update(input[i]).is_nan(), "[{}] should be NaN", i);
        }

        // Metastock converges after warmup
        let first_check = LPRIMED + 43;
        for i in LPRIMED..input.len() {
            let act = t2.update(input[i]);
            if i >= first_check {
                assert!(
                    (exp[i] - act).abs() < 1e-8,
                    "[{}] expected {}, got {}", i, exp[i], act
                );
            }
        }

        assert!(t2.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_entity() {
        let time = 1617235200;
        let inp = 3.0;
        let exp_false = 2.0281481481481483;
        let exp_true = 1.9555555555555555;
        let l: i64 = 2;
        let lprimed = (4 * l - 4) as usize;

        // scalar
        {
            let mut t2 = create_length(l, false, 0.7);
            for _ in 0..lprimed { t2.update(0.0); }
            let s = Scalar::new(time, inp);
            let out = t2.update_scalar(&s);
            assert_eq!(out.len(), 1);
            let sv = out[0].downcast_ref::<Scalar>().unwrap();
            assert!((sv.value - exp_false).abs() < 1e-13, "scalar: expected {}, got {}", exp_false, sv.value);
        }

        // bar
        {
            let mut t2 = create_length(l, true, 0.7);
            for _ in 0..lprimed { t2.update(0.0); }
            let b = Bar { time, open: 0.0, high: 0.0, low: 0.0, close: inp, volume: 0.0 };
            let out = t2.update_bar(&b);
            let sv = out[0].downcast_ref::<Scalar>().unwrap();
            assert!((sv.value - exp_true).abs() < 1e-13, "bar: expected {}, got {}", exp_true, sv.value);
        }

        // quote
        {
            let mut t2 = create_length(l, false, 0.7);
            for _ in 0..lprimed { t2.update(0.0); }
            let q = Quote { time, bid_price: inp, ask_price: inp, bid_size: 0.0, ask_size: 0.0 };
            let out = t2.update_quote(&q);
            let sv = out[0].downcast_ref::<Scalar>().unwrap();
            assert!((sv.value - exp_false).abs() < 1e-13, "quote: expected {}, got {}", exp_false, sv.value);
        }

        // trade
        {
            let mut t2 = create_length(l, true, 0.7);
            for _ in 0..lprimed { t2.update(0.0); }
            let r = Trade { time, price: inp, volume: 0.0 };
            let out = t2.update_trade(&r);
            let sv = out[0].downcast_ref::<Scalar>().unwrap();
            assert!((sv.value - exp_true).abs() < 1e-13, "trade: expected {}, got {}", exp_true, sv.value);
        }
    }

    #[test]
    fn test_is_primed() {
        let input = test_input();

        for &first_is_avg in &[true, false] {
            let mut t2 = create_length(L, first_is_avg, 0.7);
            assert!(!t2.is_primed());

            for i in 0..LPRIMED {
                t2.update(input[i]);
                assert!(!t2.is_primed(), "[{}] should not be primed", i);
            }

            for i in LPRIMED..input.len() {
                t2.update(input[i]);
                assert!(t2.is_primed(), "[{}] should be primed", i);
            }
        }
    }

    #[test]
    fn test_metadata_length() {
        let t2 = create_length(10, true, 0.3333);
        let m = t2.metadata();
        assert_eq!(m.identifier, Identifier::T2ExponentialMovingAverage);
        assert_eq!(m.mnemonic, "t2(10, 0.33330000)");
        assert_eq!(m.description, "T2 exponential moving average t2(10, 0.33330000)");
        assert_eq!(m.outputs.len(), 1);
        assert_eq!(m.outputs[0].kind, T2ExponentialMovingAverageOutput::Value as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
        assert_eq!(m.outputs[0].mnemonic, "t2(10, 0.33330000)");
    }

    #[test]
    fn test_metadata_alpha() {
        let alpha = 2.0 / 11.0;
        let t2 = create_alpha(alpha, false, 0.3333333);
        let m = t2.metadata();
        assert_eq!(m.identifier, Identifier::T2ExponentialMovingAverage);
        assert_eq!(m.mnemonic, "t2(10, 0.18181818, 0.33333330)");
        assert_eq!(m.description, "T2 exponential moving average t2(10, 0.18181818, 0.33333330)");
    }

    #[test]
    fn test_metadata_non_default_bar_component() {
        let params = T2ExponentialMovingAverageLengthParams {
            length: 10, volume_factor: 0.7, first_is_average: true,
            bar_component: Some(BarComponent::Median), quote_component: None, trade_component: None,
        };
        let t2 = T2ExponentialMovingAverage::new_from_length(&params).unwrap();
        let m = t2.metadata();
        assert_eq!(m.mnemonic, "t2(10, 0.70000000, hl/2)");
    }

    #[test]
    fn test_metadata_non_default_quote_component() {
        let params = T2ExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: 2.0 / 11.0, volume_factor: 0.7, first_is_average: false,
            bar_component: None, quote_component: Some(QuoteComponent::Bid), trade_component: None,
        };
        let t2 = T2ExponentialMovingAverage::new_from_smoothing_factor(&params).unwrap();
        let m = t2.metadata();
        assert_eq!(m.mnemonic, "t2(10, 0.18181818, 0.70000000, b)");
    }

    #[test]
    fn test_new_length_errors() {
        // length < 2
        let p = T2ExponentialMovingAverageLengthParams { length: 1, volume_factor: 0.7, ..Default::default() };
        assert!(T2ExponentialMovingAverage::new_from_length(&p).is_err());

        let p = T2ExponentialMovingAverageLengthParams { length: 0, volume_factor: 0.7, ..Default::default() };
        assert!(T2ExponentialMovingAverage::new_from_length(&p).is_err());

        let p = T2ExponentialMovingAverageLengthParams { length: -1, volume_factor: 0.7, ..Default::default() };
        assert!(T2ExponentialMovingAverage::new_from_length(&p).is_err());
    }

    #[test]
    fn test_new_alpha_errors() {
        let p = T2ExponentialMovingAverageSmoothingFactorParams { smoothing_factor: -1.0, volume_factor: 0.7, ..Default::default() };
        assert!(T2ExponentialMovingAverage::new_from_smoothing_factor(&p).is_err());

        let p = T2ExponentialMovingAverageSmoothingFactorParams { smoothing_factor: 2.0, volume_factor: 0.7, ..Default::default() };
        assert!(T2ExponentialMovingAverage::new_from_smoothing_factor(&p).is_err());
    }

    #[test]
    fn test_new_volume_factor_errors() {
        let p = T2ExponentialMovingAverageLengthParams { length: 3, volume_factor: -0.7, ..Default::default() };
        assert!(T2ExponentialMovingAverage::new_from_length(&p).is_err());

        let p = T2ExponentialMovingAverageLengthParams { length: 3, volume_factor: 1.7, ..Default::default() };
        assert!(T2ExponentialMovingAverage::new_from_length(&p).is_err());
    }

    #[test]
    fn test_new_alpha_clamped_to_epsilon() {
        // alpha = 0 gets clamped to epsilon
        let p = T2ExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: 0.0, volume_factor: 0.7, ..Default::default()
        };
        let t2 = T2ExponentialMovingAverage::new_from_smoothing_factor(&p).unwrap();
        assert_eq!(t2.smoothing_factor, 0.00000001);
        assert_eq!(t2.length, 199999999);
    }

    #[test]
    fn test_new_alpha_one() {
        let p = T2ExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: 1.0, volume_factor: 0.7, ..Default::default()
        };
        let t2 = T2ExponentialMovingAverage::new_from_smoothing_factor(&p).unwrap();
        assert_eq!(t2.smoothing_factor, 1.0);
        assert_eq!(t2.length, 1);
    }

    fn create_length(length: i64, first_is_average: bool, volume: f64) -> T2ExponentialMovingAverage {
        let params = T2ExponentialMovingAverageLengthParams {
            length, volume_factor: volume, first_is_average,
            bar_component: None, quote_component: None, trade_component: None,
        };
        T2ExponentialMovingAverage::new_from_length(&params).unwrap()
    }

    fn create_alpha(alpha: f64, first_is_average: bool, volume: f64) -> T2ExponentialMovingAverage {
        let params = T2ExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: alpha, volume_factor: volume, first_is_average,
            bar_component: None, quote_component: None, trade_component: None,
        };
        T2ExponentialMovingAverage::new_from_smoothing_factor(&params).unwrap()
    }
}
