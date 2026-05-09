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

/// Parameters to create a TEMA from length.
pub struct TripleExponentialMovingAverageLengthParams {
    pub length: i64,
    pub first_is_average: bool,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for TripleExponentialMovingAverageLengthParams {
    fn default() -> Self {
        Self {
            length: 20,
            first_is_average: true,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

/// Parameters to create a TEMA from smoothing factor.
pub struct TripleExponentialMovingAverageSmoothingFactorParams {
    pub smoothing_factor: f64,
    pub first_is_average: bool,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for TripleExponentialMovingAverageSmoothingFactorParams {
    fn default() -> Self {
        Self {
            smoothing_factor: 0.0952,
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
pub enum TripleExponentialMovingAverageOutput {
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Computes the Triple Exponential Moving Average (TEMA).
///
/// TEMA = 3*(EMA1 - EMA2) + EMA3, where EMA2 = EMA(EMA1), EMA3 = EMA(EMA2).
pub struct TripleExponentialMovingAverage {
    line: LineIndicator,
    smoothing_factor: f64,
    sum: f64,
    ema1: f64,
    ema2: f64,
    ema3: f64,
    length: i64,
    length2: i64,
    length3: i64,
    count: i64,
    first_is_average: bool,
    primed: bool,
}

impl TripleExponentialMovingAverage {
    pub fn new_from_length(params: &TripleExponentialMovingAverageLengthParams) -> Result<Self, String> {
        Self::new_internal(params.length, f64::NAN, params.first_is_average,
            params.bar_component, params.quote_component, params.trade_component)
    }

    pub fn new_from_smoothing_factor(params: &TripleExponentialMovingAverageSmoothingFactorParams) -> Result<Self, String> {
        Self::new_internal(0, params.smoothing_factor, params.first_is_average,
            params.bar_component, params.quote_component, params.trade_component)
    }

    fn new_internal(
        length: i64,
        alpha: f64,
        first_is_average: bool,
        bc_opt: Option<BarComponent>,
        qc_opt: Option<QuoteComponent>,
        tc_opt: Option<TradeComponent>,
    ) -> Result<Self, String> {
        const INVALID: &str = "invalid triple exponential moving average parameters";
        const EPSILON: f64 = 0.00000001;

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
            mnemonic = format!("tema({}{})", length, component_triple_mnemonic(bc, qc, tc));
        } else {
            if alpha < 0.0 || alpha > 1.0 {
                return Err(format!("{}: smoothing factor should be in range [0, 1]", INVALID));
            }
            let clamped = if alpha < EPSILON { EPSILON } else { alpha };
            actual_length = (2.0_f64 / clamped).round() as i64 - 1;
            actual_alpha = clamped;
            mnemonic = format!("tema({}, {:.8}{})", actual_length, clamped, component_triple_mnemonic(bc, qc, tc));
        }

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let description = format!("Triple exponential moving average {}", mnemonic);
        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        Ok(Self {
            line,
            smoothing_factor: actual_alpha,
            sum: 0.0,
            ema1: 0.0,
            ema2: 0.0,
            ema3: 0.0,
            length: actual_length,
            length2: 2 * actual_length - 1,
            length3: 3 * actual_length - 2,
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

        if self.primed {
            let sf = self.smoothing_factor;
            let mut v1 = self.ema1;
            let mut v2 = self.ema2;
            let mut v3 = self.ema3;
            v1 += (sample - v1) * sf;
            v2 += (v1 - v2) * sf;
            v3 += (v2 - v3) * sf;
            self.ema1 = v1;
            self.ema2 = v2;
            self.ema3 = v3;
            return 3.0 * (v1 - v2) + v3;
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
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.sum += self.ema1;

                if self.length2 == self.count {
                    self.ema2 = self.sum / self.length as f64;
                    self.sum = self.ema2;
                }
            } else {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;
                self.sum += self.ema2;

                if self.length3 == self.count {
                    self.primed = true;
                    self.ema3 = self.sum / self.length as f64;
                    return 3.0 * (self.ema1 - self.ema2) + self.ema3;
                }
            }
        } else {
            // Metastock
            if self.count == 1 {
                self.ema1 = sample;
            } else if self.length >= self.count {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                if self.length == self.count {
                    self.ema2 = self.ema1;
                }
            } else if self.length2 >= self.count {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;

                if self.length2 == self.count {
                    self.ema3 = self.ema2;
                }
            } else {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;
                self.ema3 += (self.ema2 - self.ema3) * self.smoothing_factor;

                if self.length3 == self.count {
                    self.primed = true;
                    return 3.0 * (self.ema1 - self.ema2) + self.ema3;
                }
            }
        }

        f64::NAN
    }
}

impl Indicator for TripleExponentialMovingAverage {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::TripleExponentialMovingAverage,
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
    fn create_tema_length(length: i64, first_is_average: bool) -> TripleExponentialMovingAverage {
        TripleExponentialMovingAverage::new_from_length(&TripleExponentialMovingAverageLengthParams {
            length,
            first_is_average,
            ..Default::default()
        }).unwrap()
    }

    fn create_tema_alpha(alpha: f64, first_is_average: bool) -> TripleExponentialMovingAverage {
        TripleExponentialMovingAverage::new_from_smoothing_factor(&TripleExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: alpha,
            first_is_average,
            ..Default::default()
        }).unwrap()
    }

    #[test]
    fn test_update_length_14_first_is_average_true() {
        let mut tema = create_tema_length(14, true);
        let input = testdata::test_input();
        let lprimed: usize = 3 * 14 - 3; // 39

        for i in 0..lprimed {
            assert!(tema.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        for i in lprimed..input.len() {
            let act = tema.update(input[i]);
            match i {
                39 => assert!((84.8629 - act).abs() < 1e-3, "[39] got {}", act),
                40 => assert!((84.2246 - act).abs() < 1e-3, "[40] got {}", act),
                251 => assert!((108.418 - act).abs() < 1e-3, "[251] got {}", act),
                _ => {}
            }
        }

        assert!(tema.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_length_14_first_is_average_false() {
        let mut tema = create_tema_length(14, false);
        let input = testdata::test_input();
        let lprimed: usize = 3 * 14 - 3;

        for i in 0..lprimed {
            assert!(tema.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        for i in lprimed..input.len() {
            let act = tema.update(input[i]);
            match i {
                39 => assert!((84.721 - act).abs() < 1e-3, "[39] got {}", act),
                40 => assert!((84.089 - act).abs() < 1e-3, "[40] got {}", act),
                251 => assert!((108.418 - act).abs() < 1e-3, "[251] got {}", act),
                _ => {}
            }
        }

        assert!(tema.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_length_26_first_is_average_false_tasc() {
        let mut tema = create_tema_length(26, false);
        let input = testdata::test_tasc_input();
        let expected = testdata::test_tasc_expected();
        let lprimed: usize = 3 * 26 - 3; // 75

        for i in 0..lprimed {
            assert!(tema.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        let first_check = 216;
        for i in lprimed..input.len() {
            let act = tema.update(input[i]);
            if i >= first_check {
                assert!((expected[i] - act).abs() < 1e-3, "[{}] expected {}, got {}", i, expected[i], act);
            }
        }

        assert!(tema.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_entity() {
        let l: i64 = 2;
        let lprimed = 3 * l - 3; // 3
        let inp = 3.0_f64;
        let exp_false = 2.888888888888889;
        let exp_true = 2.6666666666666665;
        let time = 1617235200_i64;

        // scalar (firstIsAverage = false)
        let mut tema = create_tema_length(l, false);
        for _ in 0..lprimed { tema.update(0.0); }
        let out = tema.update_scalar(&Scalar::new(time, inp));
        assert_eq!(out.len(), 1);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s.time, time);
        assert!((s.value - exp_false).abs() < 1e-13, "scalar value {} != {}", s.value, exp_false);

        // bar (firstIsAverage = true)
        let mut tema = create_tema_length(l, true);
        for _ in 0..lprimed { tema.update(0.0); }
        let bar = Bar::new(time, 0.0, 0.0, 0.0, inp, 0.0);
        let out = tema.update_bar(&bar);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((s.value - exp_true).abs() < 1e-13, "bar value {} != {}", s.value, exp_true);

        // quote (firstIsAverage = false)
        let mut tema = create_tema_length(l, false);
        for _ in 0..lprimed { tema.update(0.0); }
        let quote = Quote::new(time, inp, inp, 0.0, 0.0);
        let out = tema.update_quote(&quote);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((s.value - exp_false).abs() < 1e-13);

        // trade (firstIsAverage = true)
        let mut tema = create_tema_length(l, true);
        for _ in 0..lprimed { tema.update(0.0); }
        let trade = Trade::new(time, inp, 0.0);
        let out = tema.update_trade(&trade);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((s.value - exp_true).abs() < 1e-13);
    }

    #[test]
    fn test_is_primed() {
        let input = testdata::test_input();
        let l: usize = 14;
        let lprimed = 3 * l - 3;

        // firstIsAverage = true
        let mut tema = create_tema_length(l as i64, true);
        assert!(!tema.is_primed());
        for i in 0..lprimed {
            tema.update(input[i]);
            assert!(!tema.is_primed(), "[{}] should not be primed", i);
        }
        for i in lprimed..input.len() {
            tema.update(input[i]);
            assert!(tema.is_primed(), "[{}] should be primed", i);
        }

        // firstIsAverage = false
        let mut tema = create_tema_length(l as i64, false);
        assert!(!tema.is_primed());
        for i in 0..lprimed {
            tema.update(input[i]);
            assert!(!tema.is_primed(), "[{}] should not be primed", i);
        }
        for i in lprimed..input.len() {
            tema.update(input[i]);
            assert!(tema.is_primed(), "[{}] should be primed", i);
        }
    }

    #[test]
    fn test_metadata_length() {
        let tema = create_tema_length(10, true);
        let m = tema.metadata();
        assert_eq!(m.identifier, Identifier::TripleExponentialMovingAverage);
        assert_eq!(m.mnemonic, "tema(10)");
        assert_eq!(m.description, "Triple exponential moving average tema(10)");
        assert_eq!(m.outputs.len(), 1);
        assert_eq!(m.outputs[0].kind, TripleExponentialMovingAverageOutput::Value as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
        assert_eq!(m.outputs[0].mnemonic, "tema(10)");
        assert_eq!(m.outputs[0].description, "Triple exponential moving average tema(10)");
    }

    #[test]
    fn test_metadata_alpha() {
        let alpha = 2.0 / 11.0;
        let tema = create_tema_alpha(alpha, false);
        let m = tema.metadata();
        assert_eq!(m.identifier, Identifier::TripleExponentialMovingAverage);
        assert_eq!(m.mnemonic, "tema(10, 0.18181818)");
        assert_eq!(m.description, "Triple exponential moving average tema(10, 0.18181818)");
    }

    #[test]
    fn test_metadata_length_with_bar_component() {
        let tema = TripleExponentialMovingAverage::new_from_length(&TripleExponentialMovingAverageLengthParams {
            length: 10,
            first_is_average: true,
            bar_component: Some(BarComponent::Median),
            ..Default::default()
        }).unwrap();
        let m = tema.metadata();
        assert_eq!(m.mnemonic, "tema(10, hl/2)");
        assert_eq!(m.description, "Triple exponential moving average tema(10, hl/2)");
    }

    #[test]
    fn test_metadata_alpha_with_quote_component() {
        let tema = TripleExponentialMovingAverage::new_from_smoothing_factor(&TripleExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: 2.0 / 11.0,
            first_is_average: false,
            quote_component: Some(QuoteComponent::Bid),
            ..Default::default()
        }).unwrap();
        let m = tema.metadata();
        assert_eq!(m.mnemonic, "tema(10, 0.18181818, b)");
        assert_eq!(m.description, "Triple exponential moving average tema(10, 0.18181818, b)");
    }

    #[test]
    fn test_new_length_zero() {
        let r = TripleExponentialMovingAverage::new_from_length(&TripleExponentialMovingAverageLengthParams {
            length: 0, ..Default::default()
        });
        assert!(r.is_err());
        assert_eq!(r.err().unwrap(), "invalid triple exponential moving average parameters: length should be greater than 1");
    }

    #[test]
    fn test_new_length_one() {
        let r = TripleExponentialMovingAverage::new_from_length(&TripleExponentialMovingAverageLengthParams {
            length: 1, ..Default::default()
        });
        assert!(r.is_err());
        assert_eq!(r.err().unwrap(), "invalid triple exponential moving average parameters: length should be greater than 1");
    }

    #[test]
    fn test_new_length_negative() {
        let r = TripleExponentialMovingAverage::new_from_length(&TripleExponentialMovingAverageLengthParams {
            length: -1, ..Default::default()
        });
        assert!(r.is_err());
    }

    #[test]
    fn test_new_alpha_negative() {
        let r = TripleExponentialMovingAverage::new_from_smoothing_factor(&TripleExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: -1.0, ..Default::default()
        });
        assert!(r.is_err());
        assert_eq!(r.err().unwrap(), "invalid triple exponential moving average parameters: smoothing factor should be in range [0, 1]");
    }

    #[test]
    fn test_new_alpha_greater_than_1() {
        let r = TripleExponentialMovingAverage::new_from_smoothing_factor(&TripleExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: 2.0, ..Default::default()
        });
        assert!(r.is_err());
    }

    #[test]
    fn test_new_alpha_zero_clamped() {
        let tema = TripleExponentialMovingAverage::new_from_smoothing_factor(&TripleExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: 0.0, ..Default::default()
        }).unwrap();
        assert_eq!(tema.smoothing_factor, 0.00000001);
        assert_eq!(tema.length, 199999999);
    }

    #[test]
    fn test_mnemonic_components() {
        let tema = create_tema_length(10, true);
        assert_eq!(tema.line.mnemonic, "tema(10)");

        let tema = TripleExponentialMovingAverage::new_from_length(&TripleExponentialMovingAverageLengthParams {
            length: 10, bar_component: Some(BarComponent::Median), ..Default::default()
        }).unwrap();
        assert_eq!(tema.line.mnemonic, "tema(10, hl/2)");

        let tema = TripleExponentialMovingAverage::new_from_length(&TripleExponentialMovingAverageLengthParams {
            length: 10, quote_component: Some(QuoteComponent::Bid), ..Default::default()
        }).unwrap();
        assert_eq!(tema.line.mnemonic, "tema(10, b)");

        let tema = TripleExponentialMovingAverage::new_from_length(&TripleExponentialMovingAverageLengthParams {
            length: 10, trade_component: Some(TradeComponent::Volume), ..Default::default()
        }).unwrap();
        assert_eq!(tema.line.mnemonic, "tema(10, v)");
    }
}
