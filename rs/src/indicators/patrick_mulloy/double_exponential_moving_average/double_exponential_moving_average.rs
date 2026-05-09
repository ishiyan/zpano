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

/// Parameters to create a DEMA from length.
pub struct DoubleExponentialMovingAverageLengthParams {
    pub length: i64,
    pub first_is_average: bool,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for DoubleExponentialMovingAverageLengthParams {
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

/// Parameters to create a DEMA from smoothing factor.
pub struct DoubleExponentialMovingAverageSmoothingFactorParams {
    pub smoothing_factor: f64,
    pub first_is_average: bool,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for DoubleExponentialMovingAverageSmoothingFactorParams {
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
pub enum DoubleExponentialMovingAverageOutput {
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Computes the Double Exponential Moving Average (DEMA).
///
/// DEMA = 2*EMA1 - EMA2, where EMA2 = EMA(EMA1).
pub struct DoubleExponentialMovingAverage {
    line: LineIndicator,
    smoothing_factor: f64,
    sum: f64,
    ema1: f64,
    ema2: f64,
    length: i64,
    length2: i64,
    count: i64,
    first_is_average: bool,
    primed: bool,
}

impl DoubleExponentialMovingAverage {
    pub fn new_from_length(params: &DoubleExponentialMovingAverageLengthParams) -> Result<Self, String> {
        Self::new_internal(params.length, f64::NAN, params.first_is_average,
            params.bar_component, params.quote_component, params.trade_component)
    }

    pub fn new_from_smoothing_factor(params: &DoubleExponentialMovingAverageSmoothingFactorParams) -> Result<Self, String> {
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
        const INVALID: &str = "invalid double exponential moving average parameters";
        const EPSILON: f64 = 0.00000001;

        let bc = bc_opt.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = qc_opt.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = tc_opt.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let (actual_length, actual_alpha, mnemonic);

        if alpha.is_nan() {
            if length < 1 {
                return Err(format!("{}: length should be positive", INVALID));
            }
            actual_alpha = 2.0 / (1 + length) as f64;
            actual_length = length;
            mnemonic = format!("dema({}{})", length, component_triple_mnemonic(bc, qc, tc));
        } else {
            if alpha < 0.0 || alpha > 1.0 {
                return Err(format!("{}: smoothing factor should be in range [0, 1]", INVALID));
            }
            let clamped = if alpha < EPSILON { EPSILON } else { alpha };
            actual_length = (2.0_f64 / clamped).round() as i64 - 1;
            actual_alpha = clamped;
            mnemonic = format!("dema({}, {:.8}{})", actual_length, clamped, component_triple_mnemonic(bc, qc, tc));
        }

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let description = format!("Double exponential moving average {}", mnemonic);
        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        Ok(Self {
            line,
            smoothing_factor: actual_alpha,
            sum: 0.0,
            ema1: 0.0,
            ema2: 0.0,
            length: actual_length,
            length2: 2 * actual_length - 1,
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
            v1 += (sample - v1) * sf;
            v2 += (v1 - v2) * sf;
            self.ema1 = v1;
            self.ema2 = v2;
            return 2.0 * v1 - v2;
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
            } else {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.sum += self.ema1;

                if self.length2 == self.count {
                    self.primed = true;
                    self.ema2 = self.sum / self.length as f64;
                    return 2.0 * self.ema1 - self.ema2;
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
            } else {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;

                if self.length2 == self.count {
                    self.primed = true;
                    return 2.0 * self.ema1 - self.ema2;
                }
            }
        }

        f64::NAN
    }
}

impl Indicator for DoubleExponentialMovingAverage {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::DoubleExponentialMovingAverage,
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
    fn create_dema_length(length: i64, first_is_average: bool) -> DoubleExponentialMovingAverage {
        DoubleExponentialMovingAverage::new_from_length(&DoubleExponentialMovingAverageLengthParams {
            length,
            first_is_average,
            ..Default::default()
        }).unwrap()
    }

    fn create_dema_alpha(alpha: f64, first_is_average: bool) -> DoubleExponentialMovingAverage {
        DoubleExponentialMovingAverage::new_from_smoothing_factor(&DoubleExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: alpha,
            first_is_average,
            ..Default::default()
        }).unwrap()
    }

    #[test]
    fn test_update_length_2_first_is_average_true() {
        let mut dema = create_dema_length(2, true);
        let input = testdata::test_input();
        let l: usize = 2;
        let lprimed = 2 * l - 2; // 2

        for i in 0..lprimed {
            assert!(dema.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        let act = dema.update(input[lprimed]); // index 2
        // Check a few values with tolerance 1e-2
        assert!((94.013 - dema.update(input[3])).abs() < 2.0); // just feed through

        // Feed all remaining
        let mut dema2 = create_dema_length(2, true);
        for i in 0..input.len() {
            let act = dema2.update(input[i]);
            if i == 4 {
                assert!((94.013 - act).abs() < 1e-2, "[4] expected ~94.013, got {}", act);
            }
            if i == 5 {
                assert!((94.539 - act).abs() < 1e-2, "[5] expected ~94.539, got {}", act);
            }
            if i == 251 {
                assert!((107.94 - act).abs() < 1e-2, "[251] expected ~107.94, got {}", act);
            }
        }

        assert!(dema2.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_length_14_first_is_average_true() {
        let mut dema = create_dema_length(14, true);
        let input = testdata::test_input();
        let lprimed: usize = 2 * 14 - 2; // 26

        for i in 0..lprimed {
            assert!(dema.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        for i in lprimed..input.len() {
            let act = dema.update(input[i]);
            match i {
                28 => assert!((84.347 - act).abs() < 1e-2, "[28] got {}", act),
                29 => assert!((84.487 - act).abs() < 1e-2, "[29] got {}", act),
                30 => assert!((84.374 - act).abs() < 1e-2, "[30] got {}", act),
                31 => assert!((84.772 - act).abs() < 1e-2, "[31] got {}", act),
                48 => assert!((89.803 - act).abs() < 1e-2, "[48] got {}", act),
                251 => assert!((109.4676 - act).abs() < 1e-2, "[251] got {}", act),
                _ => {}
            }
        }

        assert!(dema.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_length_2_first_is_average_false() {
        let mut dema = create_dema_length(2, false);
        let input = testdata::test_input();
        let lprimed: usize = 2 * 2 - 2; // 2

        for i in 0..lprimed {
            assert!(dema.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        let mut dema2 = create_dema_length(2, false);
        for i in 0..input.len() {
            let act = dema2.update(input[i]);
            match i {
                4 => assert!((93.977 - act).abs() < 1e-2, "[4] got {}", act),
                5 => assert!((94.522 - act).abs() < 1e-2, "[5] got {}", act),
                251 => assert!((107.94 - act).abs() < 1e-2, "[251] got {}", act),
                _ => {}
            }
        }

        assert!(dema2.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_length_14_first_is_average_false() {
        let mut dema = create_dema_length(14, false);
        let input = testdata::test_input();
        let lprimed: usize = 2 * 14 - 2;

        for i in 0..lprimed {
            assert!(dema.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        for i in lprimed..input.len() {
            let act = dema.update(input[i]);
            match i {
                28 => assert!((84.87 - act).abs() < 1e-2, "[28] got {}", act),
                29 => assert!((84.94 - act).abs() < 1e-2, "[29] got {}", act),
                30 => assert!((84.77 - act).abs() < 1e-2, "[30] got {}", act),
                31 => assert!((85.12 - act).abs() < 1e-2, "[31] got {}", act),
                48 => assert!((89.83 - act).abs() < 1e-2, "[48] got {}", act),
                251 => assert!((109.4676 - act).abs() < 1e-2, "[251] got {}", act),
                _ => {}
            }
        }

        assert!(dema.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_length_26_first_is_average_false_tasc() {
        let mut dema = create_dema_length(26, false);
        let input = testdata::test_tasc_input();
        let expected = testdata::test_tasc_expected();
        let lprimed: usize = 2 * 26 - 2; // 50

        for i in 0..lprimed {
            assert!(dema.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        let first_check = 216;
        for i in lprimed..input.len() {
            let act = dema.update(input[i]);
            if i >= first_check {
                assert!((expected[i] - act).abs() < 1e-2, "[{}] expected {}, got {}", i, expected[i], act);
            }
        }

        assert!(dema.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_entity() {
        let l: i64 = 2;
        let lprimed = 2 * l - 2;
        let inp = 3.0_f64;
        let exp_false = 2.666666666666667;
        let time = 1617235200_i64;

        // scalar
        let mut dema = create_dema_length(l, false);
        for _ in 0..lprimed { dema.update(0.0); }
        let out = dema.update_scalar(&Scalar::new(time, inp));
        assert_eq!(out.len(), 1);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s.time, time);
        assert!((s.value - exp_false).abs() < 1e-13, "scalar value {} != {}", s.value, exp_false);

        // bar
        let mut dema = create_dema_length(l, false);
        for _ in 0..lprimed { dema.update(0.0); }
        let bar = Bar::new(time, 0.0, 0.0, 0.0, inp, 0.0);
        let out = dema.update_bar(&bar);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((s.value - exp_false).abs() < 1e-13);

        // quote
        let mut dema = create_dema_length(l, false);
        for _ in 0..lprimed { dema.update(0.0); }
        let quote = Quote::new(time, inp, inp, 0.0, 0.0);
        let out = dema.update_quote(&quote);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((s.value - exp_false).abs() < 1e-13);

        // trade
        let mut dema = create_dema_length(l, false);
        for _ in 0..lprimed { dema.update(0.0); }
        let trade = Trade::new(time, inp, 0.0);
        let out = dema.update_trade(&trade);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((s.value - exp_false).abs() < 1e-13);
    }

    #[test]
    fn test_is_primed() {
        let input = testdata::test_input();
        let l: usize = 14;
        let lprimed = 2 * l - 2;

        // firstIsAverage = true
        let mut dema = create_dema_length(l as i64, true);
        assert!(!dema.is_primed());
        for i in 0..lprimed {
            dema.update(input[i]);
            assert!(!dema.is_primed(), "[{}] should not be primed", i);
        }
        for i in lprimed..input.len() {
            dema.update(input[i]);
            assert!(dema.is_primed(), "[{}] should be primed", i);
        }

        // firstIsAverage = false
        let mut dema = create_dema_length(l as i64, false);
        assert!(!dema.is_primed());
        for i in 0..lprimed {
            dema.update(input[i]);
            assert!(!dema.is_primed(), "[{}] should not be primed", i);
        }
        for i in lprimed..input.len() {
            dema.update(input[i]);
            assert!(dema.is_primed(), "[{}] should be primed", i);
        }
    }

    #[test]
    fn test_metadata_length() {
        let dema = create_dema_length(10, true);
        let m = dema.metadata();
        assert_eq!(m.identifier, Identifier::DoubleExponentialMovingAverage);
        assert_eq!(m.mnemonic, "dema(10)");
        assert_eq!(m.description, "Double exponential moving average dema(10)");
        assert_eq!(m.outputs.len(), 1);
        assert_eq!(m.outputs[0].kind, DoubleExponentialMovingAverageOutput::Value as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
        assert_eq!(m.outputs[0].mnemonic, "dema(10)");
        assert_eq!(m.outputs[0].description, "Double exponential moving average dema(10)");
    }

    #[test]
    fn test_metadata_alpha() {
        let alpha = 2.0 / 11.0;
        let dema = create_dema_alpha(alpha, false);
        let m = dema.metadata();
        assert_eq!(m.identifier, Identifier::DoubleExponentialMovingAverage);
        assert_eq!(m.mnemonic, "dema(10, 0.18181818)");
        assert_eq!(m.description, "Double exponential moving average dema(10, 0.18181818)");
    }

    #[test]
    fn test_metadata_length_with_bar_component() {
        let dema = DoubleExponentialMovingAverage::new_from_length(&DoubleExponentialMovingAverageLengthParams {
            length: 10,
            first_is_average: true,
            bar_component: Some(BarComponent::Median),
            ..Default::default()
        }).unwrap();
        let m = dema.metadata();
        assert_eq!(m.mnemonic, "dema(10, hl/2)");
        assert_eq!(m.description, "Double exponential moving average dema(10, hl/2)");
    }

    #[test]
    fn test_metadata_alpha_with_quote_component() {
        let dema = DoubleExponentialMovingAverage::new_from_smoothing_factor(&DoubleExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: 2.0 / 11.0,
            first_is_average: false,
            quote_component: Some(QuoteComponent::Bid),
            ..Default::default()
        }).unwrap();
        let m = dema.metadata();
        assert_eq!(m.mnemonic, "dema(10, 0.18181818, b)");
        assert_eq!(m.description, "Double exponential moving average dema(10, 0.18181818, b)");
    }

    #[test]
    fn test_new_length_zero() {
        let r = DoubleExponentialMovingAverage::new_from_length(&DoubleExponentialMovingAverageLengthParams {
            length: 0, ..Default::default()
        });
        assert!(r.is_err());
        assert_eq!(r.err().unwrap(), "invalid double exponential moving average parameters: length should be positive");
    }

    #[test]
    fn test_new_length_negative() {
        let r = DoubleExponentialMovingAverage::new_from_length(&DoubleExponentialMovingAverageLengthParams {
            length: -1, ..Default::default()
        });
        assert!(r.is_err());
    }

    #[test]
    fn test_new_alpha_negative() {
        let r = DoubleExponentialMovingAverage::new_from_smoothing_factor(&DoubleExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: -1.0, ..Default::default()
        });
        assert!(r.is_err());
        assert_eq!(r.err().unwrap(), "invalid double exponential moving average parameters: smoothing factor should be in range [0, 1]");
    }

    #[test]
    fn test_new_alpha_greater_than_1() {
        let r = DoubleExponentialMovingAverage::new_from_smoothing_factor(&DoubleExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: 2.0, ..Default::default()
        });
        assert!(r.is_err());
    }

    #[test]
    fn test_new_alpha_zero_clamped() {
        let dema = DoubleExponentialMovingAverage::new_from_smoothing_factor(&DoubleExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: 0.0, ..Default::default()
        }).unwrap();
        assert_eq!(dema.smoothing_factor, 0.00000001);
        assert_eq!(dema.length, 199999999);
    }

    #[test]
    fn test_mnemonic_components() {
        let dema = create_dema_length(10, true);
        assert_eq!(dema.line.mnemonic, "dema(10)");

        let dema = DoubleExponentialMovingAverage::new_from_length(&DoubleExponentialMovingAverageLengthParams {
            length: 10, bar_component: Some(BarComponent::Median), ..Default::default()
        }).unwrap();
        assert_eq!(dema.line.mnemonic, "dema(10, hl/2)");

        let dema = DoubleExponentialMovingAverage::new_from_length(&DoubleExponentialMovingAverageLengthParams {
            length: 10, quote_component: Some(QuoteComponent::Bid), ..Default::default()
        }).unwrap();
        assert_eq!(dema.line.mnemonic, "dema(10, b)");

        let dema = DoubleExponentialMovingAverage::new_from_length(&DoubleExponentialMovingAverageLengthParams {
            length: 10, trade_component: Some(TradeComponent::Volume), ..Default::default()
        }).unwrap();
        assert_eq!(dema.line.mnemonic, "dema(10, v)");
    }
}
