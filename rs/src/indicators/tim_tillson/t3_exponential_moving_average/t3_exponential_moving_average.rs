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
    use super::super::testdata::testdata;
    use crate::entities::bar_component::BarComponent;
    use crate::entities::quote_component::QuoteComponent;
    use crate::entities::trade_component::TradeComponent;
    use crate::indicators::core::outputs::shape::Shape;
    const L: i64 = 5;
    const LPRIMED: usize = (6 * L - 6) as usize; // 24

    #[test]
    fn test_update_first_is_average_true_xls() {
        let mut t3 = create_length(L, true, 0.7);
        let input = testdata::test_input();
        let exp = testdata::test_expected();

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
        let input = testdata::test_input();

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
        let input = testdata::test_input();

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
        let input = testdata::test_input();

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
