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

/// Parameters to create a T3 EMA from length.
pub struct T3ExponentialMovingAverageLengthParams {
    pub length: i64,
    pub volume_factor: f64,
    pub first_is_average: bool,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for T3ExponentialMovingAverageLengthParams {
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

/// Parameters to create a T3 EMA from smoothing factor.
pub struct T3ExponentialMovingAverageSmoothingFactorParams {
    pub smoothing_factor: f64,
    pub volume_factor: f64,
    pub first_is_average: bool,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for T3ExponentialMovingAverageSmoothingFactorParams {
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
pub enum T3ExponentialMovingAverageOutput {
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// T3 Exponential Moving Average (T3, T3EMA).
///
/// A six-pole non-linear Kalman filter developed by Tim Tillson.
///
/// T3 = c1*e6 + c2*e5 + c3*e4 + c4*e3 where:
///   c1 = -v^3, c2 = 3v^2+3v^3, c3 = -6v^2-3v-3v^3, c4 = 1+3v+v^3+3v^2
pub struct T3ExponentialMovingAverage {
    line: LineIndicator,
    smoothing_factor: f64,
    c1: f64,
    c2: f64,
    c3: f64,
    c4: f64,
    sum: f64,
    ema1: f64,
    ema2: f64,
    ema3: f64,
    ema4: f64,
    ema5: f64,
    ema6: f64,
    length: i64,
    length2: i64,
    length3: i64,
    length4: i64,
    length5: i64,
    length6: i64,
    count: i64,
    first_is_average: bool,
    primed: bool,
}

impl T3ExponentialMovingAverage {
    pub fn new_from_length(params: &T3ExponentialMovingAverageLengthParams) -> Result<Self, String> {
        Self::new_internal(params.length, f64::NAN, params.volume_factor, params.first_is_average,
            params.bar_component, params.quote_component, params.trade_component)
    }

    pub fn new_from_smoothing_factor(params: &T3ExponentialMovingAverageSmoothingFactorParams) -> Result<Self, String> {
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
        const INVALID: &str = "invalid t3 exponential moving average parameters";
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
            mnemonic = format!("t3({}, {:.8}{})", length, v, component_triple_mnemonic(bc, qc, tc));
        } else {
            if alpha < 0.0 || alpha > 1.0 {
                return Err(format!("{}: smoothing factor should be in range [0, 1]", INVALID));
            }
            let clamped = if alpha < EPSILON { EPSILON } else { alpha };
            actual_length = (2.0_f64 / clamped).round() as i64 - 1;
            actual_alpha = clamped;
            mnemonic = format!("t3({}, {:.8}, {:.8}{})", actual_length, clamped, v, component_triple_mnemonic(bc, qc, tc));
        }

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let description = format!("T3 exponential moving average {}", mnemonic);
        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        let vv = v * v;
        let c1 = -vv * v;
        let c2 = 3.0 * (vv - c1);
        let c3 = -6.0 * vv - 3.0 * (v - c1);
        let c4 = 1.0 + 3.0 * v - c1 + 3.0 * vv;

        Ok(Self {
            line,
            smoothing_factor: actual_alpha,
            c1,
            c2,
            c3,
            c4,
            sum: 0.0,
            ema1: 0.0,
            ema2: 0.0,
            ema3: 0.0,
            ema4: 0.0,
            ema5: 0.0,
            ema6: 0.0,
            length: actual_length,
            length2: 2 * actual_length - 1,
            length3: 3 * actual_length - 2,
            length4: 4 * actual_length - 3,
            length5: 5 * actual_length - 4,
            length6: 6 * actual_length - 5,
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
            let mut v5 = self.ema5;
            let mut v6 = self.ema6;
            v1 += (sample - v1) * sf;
            v2 += (v1 - v2) * sf;
            v3 += (v2 - v3) * sf;
            v4 += (v3 - v4) * sf;
            v5 += (v4 - v5) * sf;
            v6 += (v5 - v6) * sf;
            self.ema1 = v1;
            self.ema2 = v2;
            self.ema3 = v3;
            self.ema4 = v4;
            self.ema5 = v5;
            self.ema6 = v6;
            return self.c1 * v6 + self.c2 * v5 + self.c3 * v4 + self.c4 * v3;
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
            } else if self.length4 >= self.count {
                self.ema1 += (sample - self.ema1) * sf;
                self.ema2 += (self.ema1 - self.ema2) * sf;
                self.ema3 += (self.ema2 - self.ema3) * sf;
                self.sum += self.ema3;
                if self.length4 == self.count {
                    self.ema4 = self.sum / self.length as f64;
                    self.sum = self.ema4;
                }
            } else if self.length5 >= self.count {
                self.ema1 += (sample - self.ema1) * sf;
                self.ema2 += (self.ema1 - self.ema2) * sf;
                self.ema3 += (self.ema2 - self.ema3) * sf;
                self.ema4 += (self.ema3 - self.ema4) * sf;
                self.sum += self.ema4;
                if self.length5 == self.count {
                    self.ema5 = self.sum / self.length as f64;
                    self.sum = self.ema5;
                }
            } else {
                self.ema1 += (sample - self.ema1) * sf;
                self.ema2 += (self.ema1 - self.ema2) * sf;
                self.ema3 += (self.ema2 - self.ema3) * sf;
                self.ema4 += (self.ema3 - self.ema4) * sf;
                self.ema5 += (self.ema4 - self.ema5) * sf;
                self.sum += self.ema5;
                if self.length6 == self.count {
                    self.primed = true;
                    self.ema6 = self.sum / self.length as f64;
                    return self.c1 * self.ema6 + self.c2 * self.ema5 + self.c3 * self.ema4 + self.c4 * self.ema3;
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
            } else if self.length4 >= self.count {
                self.ema1 += (sample - self.ema1) * sf;
                self.ema2 += (self.ema1 - self.ema2) * sf;
                self.ema3 += (self.ema2 - self.ema3) * sf;
                self.ema4 += (self.ema3 - self.ema4) * sf;
                if self.length4 == self.count {
                    self.ema5 = self.ema4;
                }
            } else if self.length5 >= self.count {
                self.ema1 += (sample - self.ema1) * sf;
                self.ema2 += (self.ema1 - self.ema2) * sf;
                self.ema3 += (self.ema2 - self.ema3) * sf;
                self.ema4 += (self.ema3 - self.ema4) * sf;
                self.ema5 += (self.ema4 - self.ema5) * sf;
                if self.length5 == self.count {
                    self.ema6 = self.ema5;
                }
            } else {
                self.ema1 += (sample - self.ema1) * sf;
                self.ema2 += (self.ema1 - self.ema2) * sf;
                self.ema3 += (self.ema2 - self.ema3) * sf;
                self.ema4 += (self.ema3 - self.ema4) * sf;
                self.ema5 += (self.ema4 - self.ema5) * sf;
                self.ema6 += (self.ema5 - self.ema6) * sf;
                if self.length6 == self.count {
                    self.primed = true;
                    return self.c1 * self.ema6 + self.c2 * self.ema5 + self.c3 * self.ema4 + self.c4 * self.ema3;
                }
            }
        }

        f64::NAN
    }
}

impl Indicator for T3ExponentialMovingAverage {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::T3ExponentialMovingAverage,
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

    /// Expected data from test_T3.xls, T3(5,0.7) column.
    #[allow(clippy::excessive_precision)]
    fn test_expected() -> Vec<f64> {
        vec![
            f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN,
            f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN,
            f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN,
            f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN,
            85.72987567371790, 84.36536238957790, 83.48212833499610, 83.64878022075270, 84.20774780003610, 84.81400226243770,
            85.21162459867090, 85.61977953584720, 85.88366564657920, 86.37507196062170, 86.96360606377290, 87.33715584797100,
            87.47581538320310, 87.22704246772630, 86.66138343888210, 85.93988420535540, 85.17149180117680, 84.72447668908230,
            85.01865520943790, 85.83612757497730, 87.03563064369770, 88.27257526381620, 89.41758978061620, 90.07559145334440,
            90.52788384532180, 90.79849305637170, 90.75816999615070, 90.48012595902110, 89.56721885323770, 88.19961294746030,
            86.64012904487460, 85.41932552383290, 84.73734286979860, 84.55421441417930, 85.00890603797630, 85.87022403441470,
            86.77016654750900, 87.51893615115180, 88.47034581811800, 89.45474018834570, 90.52574408904200, 91.57169411417090,
            92.42614882677630, 92.87068016379410, 92.77362350686670, 92.29390466156780, 91.56310318186090, 90.34706085030460,
            88.70815109494770, 87.17399400070270, 86.08896648247290, 86.77721477010180, 89.01690920285690, 92.50578045524320,
            96.44661245628320, 99.68374118082500, 101.92353915902500, 103.51864520203800, 104.78791133759700, 105.75424271690200,
            106.42923203569800, 106.67331521810400, 107.05684930109700, 107.65552066563500, 108.45831824703700, 109.54712933016300,
            111.89434730682000, 114.52574405349600, 116.75228613125600, 118.41726647291900, 119.37242887772400, 119.58121746519900,
            119.15016873725100, 117.99712010891400, 116.35343432693000, 115.50133109382700, 115.19900149595300, 115.20832852971600,
            114.89045855286400, 114.49505935654600, 114.04873581207800, 113.98299653094000, 114.76120631137200, 115.59884277390800,
            116.28205585323000, 116.55693390132200, 116.39983049399400, 116.13322166220300, 115.93071030281200, 116.38166637287000,
            117.22757950324600, 118.23631535248800, 119.65852250290400, 121.06189350073800, 122.17001559831100, 122.88208870818400,
            123.32075513993400, 123.49845436008200, 123.73526507154800, 124.57544794795000, 125.98775939563200, 127.72677347766200,
            129.27852779306300, 130.66084445385100, 131.92850098151000, 133.38763285852100, 134.90304940294100, 136.26189527105800,
            137.25635611457500, 137.76465314041100, 137.88397712608300, 137.57294795595100, 136.27097486999100, 134.48409991908300,
            132.08334163347200, 129.64905881153300, 127.34920616402000, 125.78806210911800, 125.20525121026800, 124.97471519964800,
            124.91187781656600, 124.53236452951700, 123.58691150647000, 122.26362413973900, 121.44605573567300, 121.22904968429300,
            121.26833012683900, 121.06480537490900, 121.11936168993500, 121.16873672766900, 121.42906776612200, 122.30150235050600,
            123.65994185435500, 124.60369529647400, 124.92167475815800, 124.65108838183800, 124.37460034158500, 123.94458586126200,
            123.47997231186300, 123.13057512793100, 123.03814664421900, 123.04147396438600, 123.23225566873300, 123.84790124323500,
            124.53741993908800, 125.48193174896000, 126.88001089241300, 128.29156445167800, 129.94307874395900, 131.63064617976500,
            132.80966556857300, 133.55644670051600, 133.80913491375100, 133.48924486083300, 132.29305889081000, 131.18723345437400,
            130.07504168974000, 128.82765488259600, 127.23236569221000, 125.90878699219100, 124.77446308473800, 123.92185993031100,
            122.95122696884300, 122.05570316734200, 120.96585380865100, 120.08120106718900, 119.74698117219900, 119.52440496847700,
            118.98986994944600, 117.91920934928700, 116.69443222363700, 115.09004689050000, 112.81604242902500, 110.56231508883200,
            108.80117430219600, 107.52097068017500, 106.69052592372500, 106.21332619981500, 104.12646621492700, 101.28478214721100,
            98.45909805366300, 96.29196542448850, 94.59404389669660, 93.56678784171560, 93.49539073445030, 93.90336034189440,
            94.25829840322930, 94.40922959431940, 94.07888457524260, 93.30020230129970, 92.78070057794470, 92.57408795701960,
            92.98810688816150, 93.56080911108470, 94.18829647483600, 94.55572153094430, 94.73726505383050, 94.70162919926040,
            95.03014981975750, 96.34501640344490, 98.68718019257840, 101.19176302440600, 103.18034967184000, 104.57401565486800,
            105.34469118495800, 105.51082025160400, 105.32911720489200, 105.20951060163700, 105.97637243377600, 107.82138781238700,
            110.21769922957300, 112.77811195271300, 114.51075261504700, 114.85539096796300, 114.29440592244600, 113.27318258016400,
            111.89052001075400, 110.70018738987500, 109.94200614625900, 109.47043806636400, 109.29969950462100, 109.08364080071000,
            108.87294256387700, 108.83218146601300, 108.93829305405100, 109.02476604337200, 109.03210341275800, 108.87915000449300,
        ]
    }

    const L: i64 = 5;
    const LPRIMED: usize = (6 * L - 6) as usize; // 24

    #[test]
    fn test_update_first_is_average_true_xls() {
        let mut t3 = create_length(L, true, 0.7);
        let input = test_input();
        let exp = test_expected();

        for i in 0..LPRIMED {
            assert!(t3.update(input[i]).is_nan(), "[{}] should be NaN", i);
        }

        for i in LPRIMED..input.len() {
            let act = t3.update(input[i]);
            assert!(
                (exp[i] - act).abs() < 1e-3,
                "[{}] expected {}, got {}", i, exp[i], act
            );
        }

        assert!(t3.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_first_is_average_true_spot_check() {
        let mut t3 = create_length(L, true, 0.7);
        let input = test_input();

        for i in 0..LPRIMED {
            assert!(t3.update(input[i]).is_nan());
        }

        let mut values = Vec::new();
        for i in LPRIMED..input.len() {
            values.push(t3.update(input[i]));
        }

        // Index 250 and 251 spot checks from TA-Lib
        let i250 = 250 - LPRIMED;
        let i251 = 251 - LPRIMED;
        assert!((values[i250] - 109.032).abs() < 1e-3, "i250: {}", values[i250]);
        assert!((values[i251] - 108.88).abs() < 1e-3, "i251: {}", values[i251]);
    }

    #[test]
    fn test_update_first_is_average_false() {
        let mut t3 = create_length(L, false, 0.7);
        let input = test_input();

        for i in 0..LPRIMED {
            assert!(t3.update(input[i]).is_nan(), "[{}] should be NaN", i);
        }

        let mut values = Vec::new();
        for i in LPRIMED..input.len() {
            values.push(t3.update(input[i]));
        }

        // TA-Lib Metastock spot checks
        let i24 = 24 - LPRIMED;
        let i25 = 25 - LPRIMED;
        let i250 = 250 - LPRIMED;
        let i251 = 251 - LPRIMED;
        assert!((values[i24] - 85.749).abs() < 1e-3, "i24: {}", values[i24]);
        assert!((values[i25] - 84.380).abs() < 1e-3, "i25: {}", values[i25]);
        assert!((values[i250] - 109.032).abs() < 1e-3, "i250: {}", values[i250]);
        assert!((values[i251] - 108.88).abs() < 1e-3, "i251: {}", values[i251]);

        assert!(t3.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_entity() {
        let time = 1617235200;
        let inp = 3.0;
        let exp_false = 1.6675884773662544;
        let exp_true = 1.6901728395061721;
        let l: i64 = 2;
        let lprimed = (6 * l - 6) as usize;

        // scalar
        {
            let mut t3 = create_length(l, false, 0.7);
            for _ in 0..lprimed { t3.update(0.0); }
            let s = Scalar::new(time, inp);
            let out = t3.update_scalar(&s);
            assert_eq!(out.len(), 1);
            let sv = out[0].downcast_ref::<Scalar>().unwrap();
            assert_eq!(sv.value, exp_false);
        }

        // bar
        {
            let mut t3 = create_length(l, true, 0.7);
            for _ in 0..lprimed { t3.update(0.0); }
            let b = Bar { time, open: 0.0, high: 0.0, low: 0.0, close: inp, volume: 0.0 };
            let out = t3.update_bar(&b);
            let sv = out[0].downcast_ref::<Scalar>().unwrap();
            assert_eq!(sv.value, exp_true);
        }

        // quote
        {
            let mut t3 = create_length(l, false, 0.7);
            for _ in 0..lprimed { t3.update(0.0); }
            let q = Quote { time, bid_price: inp, ask_price: inp, bid_size: 0.0, ask_size: 0.0 };
            let out = t3.update_quote(&q);
            let sv = out[0].downcast_ref::<Scalar>().unwrap();
            assert_eq!(sv.value, exp_false);
        }

        // trade
        {
            let mut t3 = create_length(l, true, 0.7);
            for _ in 0..lprimed { t3.update(0.0); }
            let r = Trade { time, price: inp, volume: 0.0 };
            let out = t3.update_trade(&r);
            let sv = out[0].downcast_ref::<Scalar>().unwrap();
            assert_eq!(sv.value, exp_true);
        }
    }

    #[test]
    fn test_is_primed() {
        let input = test_input();

        for &first_is_avg in &[true, false] {
            let mut t3 = create_length(L, first_is_avg, 0.7);
            assert!(!t3.is_primed());

            for i in 0..LPRIMED {
                t3.update(input[i]);
                assert!(!t3.is_primed(), "[{}] should not be primed", i);
            }

            for i in LPRIMED..input.len() {
                t3.update(input[i]);
                assert!(t3.is_primed(), "[{}] should be primed", i);
            }
        }
    }

    #[test]
    fn test_metadata_length() {
        let t3 = create_length(10, true, 0.3333);
        let m = t3.metadata();
        assert_eq!(m.identifier, Identifier::T3ExponentialMovingAverage);
        assert_eq!(m.mnemonic, "t3(10, 0.33330000)");
        assert_eq!(m.description, "T3 exponential moving average t3(10, 0.33330000)");
        assert_eq!(m.outputs.len(), 1);
        assert_eq!(m.outputs[0].kind, T3ExponentialMovingAverageOutput::Value as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
    }

    #[test]
    fn test_metadata_alpha() {
        let alpha = 2.0 / 11.0;
        let t3 = create_alpha(alpha, false, 0.3333333);
        let m = t3.metadata();
        assert_eq!(m.identifier, Identifier::T3ExponentialMovingAverage);
        assert_eq!(m.mnemonic, "t3(10, 0.18181818, 0.33333330)");
    }

    #[test]
    fn test_metadata_non_default_bar_component() {
        let params = T3ExponentialMovingAverageLengthParams {
            length: 10, volume_factor: 0.7, first_is_average: true,
            bar_component: Some(BarComponent::Median), quote_component: None, trade_component: None,
        };
        let t3 = T3ExponentialMovingAverage::new_from_length(&params).unwrap();
        let m = t3.metadata();
        assert_eq!(m.mnemonic, "t3(10, 0.70000000, hl/2)");
    }

    #[test]
    fn test_metadata_non_default_quote_component() {
        let params = T3ExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: 2.0 / 11.0, volume_factor: 0.7, first_is_average: false,
            bar_component: None, quote_component: Some(QuoteComponent::Bid), trade_component: None,
        };
        let t3 = T3ExponentialMovingAverage::new_from_smoothing_factor(&params).unwrap();
        let m = t3.metadata();
        assert_eq!(m.mnemonic, "t3(10, 0.18181818, 0.70000000, b)");
    }

    #[test]
    fn test_new_length_errors() {
        let p = T3ExponentialMovingAverageLengthParams { length: 1, volume_factor: 0.7, ..Default::default() };
        assert!(T3ExponentialMovingAverage::new_from_length(&p).is_err());

        let p = T3ExponentialMovingAverageLengthParams { length: 0, volume_factor: 0.7, ..Default::default() };
        assert!(T3ExponentialMovingAverage::new_from_length(&p).is_err());

        let p = T3ExponentialMovingAverageLengthParams { length: -1, volume_factor: 0.7, ..Default::default() };
        assert!(T3ExponentialMovingAverage::new_from_length(&p).is_err());
    }

    #[test]
    fn test_new_alpha_errors() {
        let p = T3ExponentialMovingAverageSmoothingFactorParams { smoothing_factor: -1.0, volume_factor: 0.7, ..Default::default() };
        assert!(T3ExponentialMovingAverage::new_from_smoothing_factor(&p).is_err());

        let p = T3ExponentialMovingAverageSmoothingFactorParams { smoothing_factor: 2.0, volume_factor: 0.7, ..Default::default() };
        assert!(T3ExponentialMovingAverage::new_from_smoothing_factor(&p).is_err());
    }

    #[test]
    fn test_new_volume_factor_errors() {
        let p = T3ExponentialMovingAverageLengthParams { length: 3, volume_factor: -0.7, ..Default::default() };
        assert!(T3ExponentialMovingAverage::new_from_length(&p).is_err());

        let p = T3ExponentialMovingAverageLengthParams { length: 3, volume_factor: 1.7, ..Default::default() };
        assert!(T3ExponentialMovingAverage::new_from_length(&p).is_err());
    }

    #[test]
    fn test_new_alpha_clamped_to_epsilon() {
        let p = T3ExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: 0.0, volume_factor: 0.7, ..Default::default()
        };
        let t3 = T3ExponentialMovingAverage::new_from_smoothing_factor(&p).unwrap();
        assert_eq!(t3.smoothing_factor, 0.00000001);
        assert_eq!(t3.length, 199999999);
    }

    #[test]
    fn test_new_alpha_one() {
        let p = T3ExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: 1.0, volume_factor: 0.7, ..Default::default()
        };
        let t3 = T3ExponentialMovingAverage::new_from_smoothing_factor(&p).unwrap();
        assert_eq!(t3.smoothing_factor, 1.0);
        assert_eq!(t3.length, 1);
    }

    fn create_length(length: i64, first_is_average: bool, volume: f64) -> T3ExponentialMovingAverage {
        let params = T3ExponentialMovingAverageLengthParams {
            length, volume_factor: volume, first_is_average,
            bar_component: None, quote_component: None, trade_component: None,
        };
        T3ExponentialMovingAverage::new_from_length(&params).unwrap()
    }

    fn create_alpha(alpha: f64, first_is_average: bool, volume: f64) -> T3ExponentialMovingAverage {
        let params = T3ExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: alpha, volume_factor: volume, first_is_average,
            bar_component: None, quote_component: None, trade_component: None,
        };
        T3ExponentialMovingAverage::new_from_smoothing_factor(&params).unwrap()
    }
}
