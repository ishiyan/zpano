use std::any::Any;

use crate::entities::bar::Bar;
use crate::entities::bar_component::{component_value as bar_component_value, BarComponent};
use crate::entities::quote::Quote;
use crate::entities::quote_component::{
    component_value as quote_component_value, QuoteComponent, DEFAULT_QUOTE_COMPONENT,
};
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::entities::trade_component::{
    component_value as trade_component_value, TradeComponent, DEFAULT_TRADE_COMPONENT,
};
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::component_triple_mnemonic::component_triple_mnemonic;
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::metadata::Metadata;

/// Output describes the outputs of the indicator.
#[repr(u8)]
#[derive(Debug, Clone, Copy, PartialEq)]
pub enum InstantaneousTrendLineOutput {
    /// Value is the scalar value of the instantaneous trend line.
    Value = 1,
    /// Trigger is the scalar value of the trigger line.
    Trigger = 2,
}

/// LengthParams describes parameters based on length.
pub struct LengthParams {
    pub length: i32,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for LengthParams {
    fn default() -> Self {
        Self {
            length: 28,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

/// SmoothingFactorParams describes parameters based on smoothing factor.
pub struct SmoothingFactorParams {
    pub smoothing_factor: f64,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for SmoothingFactorParams {
    fn default() -> Self {
        Self {
            smoothing_factor: 0.07,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

/// InstantaneousTrendLine (Ehler's Instantaneous Trend Line, iTrend).
pub struct InstantaneousTrendLine {
    length: i64,
    smoothing_factor: f64,
    mnemonic: String,
    description: String,
    mnemonic_trig: String,
    description_trig: String,
    coeff1: f64,
    coeff2: f64,
    coeff3: f64,
    coeff4: f64,
    coeff5: f64,
    count: i32,
    previous_sample1: f64,
    previous_sample2: f64,
    previous_trend_line1: f64,
    previous_trend_line2: f64,
    trend_line: f64,
    trigger_line: f64,
    primed: bool,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl InstantaneousTrendLine {
    /// Creates an instance based on length parameters.
    pub fn new_length(p: &LengthParams) -> Result<Self, String> {
        Self::new_internal(p.length, f64::NAN, p.bar_component, p.quote_component, p.trade_component)
    }

    /// Creates an instance based on smoothing factor parameters.
    pub fn new_smoothing_factor(p: &SmoothingFactorParams) -> Result<Self, String> {
        Self::new_internal(0, p.smoothing_factor, p.bar_component, p.quote_component, p.trade_component)
    }

    fn new_internal(
        length: i32,
        alpha: f64,
        bc: Option<BarComponent>,
        qc: Option<QuoteComponent>,
        tc: Option<TradeComponent>,
    ) -> Result<Self, String> {
        let invalid = "invalid instantaneous trend line parameters";
        let epsilon = 0.00000001;

        let (length, alpha) = if alpha.is_nan() {
            // Length-based construction.
            if length < 1 {
                return Err(format!("{}: length should be a positive integer", invalid));
            }
            let a = 2.0 / (1 + length) as f64;
            (length as i64, a)
        } else {
            // Smoothing-factor-based construction.
            if alpha < 0.0 || alpha > 1.0 {
                return Err(format!("{}: smoothing factor should be in range [0, 1]", invalid));
            }
            let l = if alpha < epsilon {
                i64::MAX
            } else {
                (2.0_f64 / alpha).round() as i64 - 1
            };
            (l, alpha)
        };

        // Resolve defaults. Default bar component is Median (hl/2), not Close.
        let bc = bc.unwrap_or(BarComponent::Median);
        let qc = qc.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = tc.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let component_mnemonic = component_triple_mnemonic(bc, qc, tc);
        let mnemonic = format!("iTrend({}{component_mnemonic})", length);
        let mnemonic_trig = format!("iTrendTrigger({}{component_mnemonic})", length);
        let description = format!("Instantaneous Trend Line {}", mnemonic);
        let description_trig = format!("Instantaneous Trend Line trigger {}", mnemonic_trig);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        // Calculate coefficients.
        let a2 = alpha * alpha;
        let c1 = alpha - a2 / 4.0;
        let c2 = a2 / 2.0;
        let c3 = -(alpha - 3.0 * a2 / 4.0);
        let x = 1.0 - alpha;
        let c4 = 2.0 * x;
        let c5 = -(x * x);

        Ok(Self {
            length,
            smoothing_factor: alpha,
            mnemonic,
            description,
            mnemonic_trig,
            description_trig,
            coeff1: c1,
            coeff2: c2,
            coeff3: c3,
            coeff4: c4,
            coeff5: c5,
            count: 0,
            previous_sample1: 0.0,
            previous_sample2: 0.0,
            previous_trend_line1: 0.0,
            previous_trend_line2: 0.0,
            trend_line: f64::NAN,
            trigger_line: f64::NAN,
            primed: false,
            bar_func,
            quote_func,
            trade_func,
        })
    }

    /// Updates the indicator with a new sample value.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return f64::NAN;
        }

        if self.primed {
            self.trend_line = self.coeff1 * sample
                + self.coeff2 * self.previous_sample1
                + self.coeff3 * self.previous_sample2
                + self.coeff4 * self.previous_trend_line1
                + self.coeff5 * self.previous_trend_line2;
            self.trigger_line = 2.0 * self.trend_line - self.previous_trend_line2;

            self.previous_sample2 = self.previous_sample1;
            self.previous_sample1 = sample;
            self.previous_trend_line2 = self.previous_trend_line1;
            self.previous_trend_line1 = self.trend_line;

            return self.trend_line;
        }

        self.count += 1;

        match self.count {
            1 => {
                self.previous_sample2 = sample;
                f64::NAN
            }
            2 => {
                self.previous_sample1 = sample;
                f64::NAN
            }
            3 => {
                self.previous_trend_line2 =
                    (sample + 2.0 * self.previous_sample1 + self.previous_sample2) / 4.0;
                self.previous_sample2 = self.previous_sample1;
                self.previous_sample1 = sample;
                f64::NAN
            }
            4 => {
                self.previous_trend_line1 =
                    (sample + 2.0 * self.previous_sample1 + self.previous_sample2) / 4.0;
                self.previous_sample2 = self.previous_sample1;
                self.previous_sample1 = sample;
                f64::NAN
            }
            5 => {
                self.trend_line = self.coeff1 * sample
                    + self.coeff2 * self.previous_sample1
                    + self.coeff3 * self.previous_sample2
                    + self.coeff4 * self.previous_trend_line1
                    + self.coeff5 * self.previous_trend_line2;
                self.trigger_line = 2.0 * self.trend_line - self.previous_trend_line2;

                self.previous_sample2 = self.previous_sample1;
                self.previous_sample1 = sample;
                self.previous_trend_line2 = self.previous_trend_line1;
                self.previous_trend_line1 = self.trend_line;
                self.primed = true;

                self.trend_line
            }
            _ => f64::NAN,
        }
    }

    fn update_entity(&mut self, time: i64, sample: f64) -> Output {
        let v = self.update(sample);
        let trig = if v.is_nan() { f64::NAN } else { self.trigger_line };
        vec![
            Box::new(Scalar::new(time, v)) as Box<dyn Any>,
            Box::new(Scalar::new(time, trig)) as Box<dyn Any>,
        ]
    }
}

impl Indicator for InstantaneousTrendLine {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::InstantaneousTrendLine,
            &self.mnemonic,
            &self.description,
            &[
                OutputText {
                    mnemonic: self.mnemonic.clone(),
                    description: self.description.clone(),
                },
                OutputText {
                    mnemonic: self.mnemonic_trig.clone(),
                    description: self.description_trig.clone(),
                },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        self.update_entity(sample.time, sample.value)
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let v = (self.bar_func)(sample);
        self.update_entity(sample.time, v)
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (self.quote_func)(sample);
        self.update_entity(sample.time, v)
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let v = (self.trade_func)(sample);
        self.update_entity(sample.time, v)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::testdata;
    fn create_default() -> InstantaneousTrendLine {
        let params = SmoothingFactorParams {
            smoothing_factor: 0.07,
            ..Default::default()
        };
        InstantaneousTrendLine::new_smoothing_factor(&params).unwrap()
    }

    const LPRIMED: usize = 4;
    const TOLERANCE: f64 = 1e-8;

    #[test]
    fn test_update_trend_line() {
        let mut itl = create_default();
        let input = testdata::test_input();
        let exp = testdata::test_expected_trend_line();

        for i in 0..LPRIMED {
            assert!(itl.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        for i in LPRIMED..input.len() {
            let act = itl.update(input[i]);
            assert!(
                (exp[i] - act).abs() < TOLERANCE,
                "[{}] expected {}, actual {}", i, exp[i], act
            );
        }

        assert!(itl.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_trigger_line() {
        let mut itl = create_default();
        let input = testdata::test_input();
        let exp_trig = testdata::test_expected_trigger();

        for i in 0..LPRIMED {
            itl.update(input[i]);
        }

        for i in LPRIMED..input.len() {
            itl.update(input[i]);
            let act = itl.trigger_line;
            assert!(
                (exp_trig[i] - act).abs() < TOLERANCE,
                "[{}] trigger expected {}, actual {}", i, exp_trig[i], act
            );
        }
    }

    #[test]
    fn test_update_entity_scalar() {
        let mut itl = create_default();
        let input = testdata::test_input();
        let exp_trend = testdata::test_expected_trend_line();
        let exp_trigger = testdata::test_expected_trigger();

        for i in 0..input.len() {
            let s = Scalar::new(1000, input[i]);
            let out = itl.update_scalar(&s);
            assert_eq!(out.len(), 2);

            let s0 = out[0].downcast_ref::<Scalar>().unwrap();
            let s1 = out[1].downcast_ref::<Scalar>().unwrap();
            assert_eq!(s0.time, 1000);
            assert_eq!(s1.time, 1000);

            if exp_trend[i].is_nan() {
                assert!(s0.value.is_nan(), "[{}] expected NaN for trend", i);
                assert!(s1.value.is_nan(), "[{}] expected NaN for trigger", i);
            } else {
                assert!(
                    (exp_trend[i] - s0.value).abs() < TOLERANCE,
                    "[{}] trend expected {}, actual {}", i, exp_trend[i], s0.value
                );
                assert!(
                    (exp_trigger[i] - s1.value).abs() < TOLERANCE,
                    "[{}] trigger expected {}, actual {}", i, exp_trigger[i], s1.value
                );
            }
        }
    }

    #[test]
    fn test_update_entity_bar() {
        let mut itl = create_default();
        let input_high = testdata::test_input_high();
        let input_low = testdata::test_input_low();
        let exp_trend = testdata::test_expected_trend_line();
        let exp_trigger = testdata::test_expected_trigger();

        for i in 0..input_high.len() {
            let bar = Bar::new(1000, 0.0, input_high[i], input_low[i], 0.0, 0.0);
            let out = itl.update_bar(&bar);
            assert_eq!(out.len(), 2);

            let s0 = out[0].downcast_ref::<Scalar>().unwrap();
            let s1 = out[1].downcast_ref::<Scalar>().unwrap();

            if exp_trend[i].is_nan() {
                assert!(s0.value.is_nan());
                assert!(s1.value.is_nan());
            } else {
                assert!(
                    (exp_trend[i] - s0.value).abs() < TOLERANCE,
                    "[{}] bar trend expected {}, actual {}", i, exp_trend[i], s0.value
                );
                assert!(
                    (exp_trigger[i] - s1.value).abs() < TOLERANCE,
                    "[{}] bar trigger expected {}, actual {}", i, exp_trigger[i], s1.value
                );
            }
        }
    }

    #[test]
    fn test_is_primed() {
        let mut itl = create_default();
        let input = testdata::test_input();

        assert!(!itl.is_primed());

        for i in 0..LPRIMED {
            itl.update(input[i]);
            assert!(!itl.is_primed(), "[{}] should not be primed yet", i);
        }

        for i in LPRIMED..input.len() {
            itl.update(input[i]);
            assert!(itl.is_primed(), "[{}] should be primed", i);
        }
    }

    #[test]
    fn test_metadata() {
        let itl = create_default();
        let m = itl.metadata();

        assert_eq!(m.identifier, Identifier::InstantaneousTrendLine);
        assert_eq!(m.mnemonic, "iTrend(28, hl/2)");
        assert_eq!(m.description, "Instantaneous Trend Line iTrend(28, hl/2)");
        assert_eq!(m.outputs.len(), 2);
        assert_eq!(m.outputs[0].mnemonic, "iTrend(28, hl/2)");
        assert_eq!(m.outputs[1].mnemonic, "iTrendTrigger(28, hl/2)");
    }

    #[test]
    fn test_new_length_valid() {
        let params = LengthParams { length: 28, ..Default::default() };
        let itl = InstantaneousTrendLine::new_length(&params).unwrap();
        assert_eq!(itl.length, 28);
        assert!(!itl.primed);
        assert!(itl.trend_line.is_nan());
        assert!(itl.trigger_line.is_nan());
    }

    #[test]
    fn test_new_length_errors() {
        let params = LengthParams { length: 0, ..Default::default() };
        assert!(InstantaneousTrendLine::new_length(&params).is_err());

        let params = LengthParams { length: -8, ..Default::default() };
        assert!(InstantaneousTrendLine::new_length(&params).is_err());
    }

    #[test]
    fn test_new_smoothing_factor_valid() {
        let params = SmoothingFactorParams { smoothing_factor: 0.07, ..Default::default() };
        let itl = InstantaneousTrendLine::new_smoothing_factor(&params).unwrap();
        assert_eq!(itl.smoothing_factor, 0.07);
        assert_eq!(itl.length, 28);

        let params = SmoothingFactorParams { smoothing_factor: 0.06, ..Default::default() };
        let itl = InstantaneousTrendLine::new_smoothing_factor(&params).unwrap();
        assert_eq!(itl.length, 32);

        // Near-zero alpha
        let params = SmoothingFactorParams { smoothing_factor: 0.000000001, ..Default::default() };
        let itl = InstantaneousTrendLine::new_smoothing_factor(&params).unwrap();
        assert_eq!(itl.length, i64::MAX);

        // Alpha=0
        let params = SmoothingFactorParams { smoothing_factor: 0.0, ..Default::default() };
        let itl = InstantaneousTrendLine::new_smoothing_factor(&params).unwrap();
        assert_eq!(itl.length, i64::MAX);

        // Alpha=1
        let params = SmoothingFactorParams { smoothing_factor: 1.0, ..Default::default() };
        let itl = InstantaneousTrendLine::new_smoothing_factor(&params).unwrap();
        assert_eq!(itl.length, 1);
    }

    #[test]
    fn test_new_smoothing_factor_errors() {
        let params = SmoothingFactorParams { smoothing_factor: -0.0001, ..Default::default() };
        assert!(InstantaneousTrendLine::new_smoothing_factor(&params).is_err());

        let params = SmoothingFactorParams { smoothing_factor: 1.0001, ..Default::default() };
        assert!(InstantaneousTrendLine::new_smoothing_factor(&params).is_err());
    }

    #[test]
    fn test_metadata_non_default_component() {
        let params = LengthParams {
            length: 3,
            trade_component: Some(TradeComponent::Volume),
            ..Default::default()
        };
        let itl = InstantaneousTrendLine::new_length(&params).unwrap();
        let m = itl.metadata();

        assert_eq!(m.mnemonic, "iTrend(3, hl/2, v)");
        assert_eq!(m.description, "Instantaneous Trend Line iTrend(3, hl/2, v)");
    }
}
