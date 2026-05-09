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
use crate::indicators::john_ehlers::hilbert_transformer::{
    new_cycle_estimator, estimator_moniker, CycleEstimator, CycleEstimatorParams,
    CycleEstimatorType,
};

/// Output describes the outputs of the indicator.
#[repr(u8)]
#[derive(Debug, Clone, Copy, PartialEq)]
pub enum HilbertTransformerInstantaneousTrendLineOutput {
    /// Value is the instantaneous trend line value.
    Value = 1,
    /// DominantCyclePeriod is the smoothed dominant cycle period.
    DominantCyclePeriod = 2,
}

/// Params describes parameters to create an instance of the indicator.
pub struct HilbertTransformerInstantaneousTrendLineParams {
    pub estimator_type: CycleEstimatorType,
    pub estimator_params: CycleEstimatorParams,
    pub alpha_ema_period_additional: f64,
    pub trend_line_smoothing_length: usize,
    pub cycle_part_multiplier: f64,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for HilbertTransformerInstantaneousTrendLineParams {
    fn default() -> Self {
        Self {
            estimator_type: CycleEstimatorType::HomodyneDiscriminator,
            estimator_params: CycleEstimatorParams {
                smoothing_length: 4,
                alpha_ema_quadrature_in_phase: 0.2,
                alpha_ema_period: 0.2,
                warm_up_period: 100,
            },
            alpha_ema_period_additional: 0.33,
            trend_line_smoothing_length: 4,
            cycle_part_multiplier: 1.0,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

/// HilbertTransformerInstantaneousTrendLine is Ehlers' Instantaneous Trend Line indicator
/// built on top of a Hilbert transformer cycle estimator.
pub struct HilbertTransformerInstantaneousTrendLine {
    mnemonic: String,
    description: String,
    mnemonic_dcp: String,
    description_dcp: String,
    htce: Box<dyn CycleEstimator>,
    alpha_ema_period_additional: f64,
    one_min_alpha_ema_period_additional: f64,
    cycle_part_multiplier: f64,
    coeff0: f64,
    coeff1: f64,
    coeff2: f64,
    coeff3: f64,
    smoothed_period: f64,
    value: f64,
    average1: f64,
    average2: f64,
    average3: f64,
    input: Vec<f64>,
    input_length: usize,
    input_length_min1: usize,
    primed: bool,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl HilbertTransformerInstantaneousTrendLine {
    /// Creates an instance with default parameters.
    pub fn new_default() -> Result<Self, String> {
        Self::new(&HilbertTransformerInstantaneousTrendLineParams::default())
    }

    /// Creates an instance with the given parameters.
    pub fn new(p: &HilbertTransformerInstantaneousTrendLineParams) -> Result<Self, String> {
        let invalid = "invalid hilbert transformer instantaneous trend line parameters";

        if p.alpha_ema_period_additional <= 0.0 || p.alpha_ema_period_additional > 1.0 {
            return Err(format!(
                "{}: \u{03B1} for additional smoothing should be in range (0, 1]",
                invalid
            ));
        }

        if p.trend_line_smoothing_length < 2 || p.trend_line_smoothing_length > 4 {
            return Err(format!(
                "{}: trend line smoothing length should be 2, 3, or 4",
                invalid
            ));
        }

        if p.cycle_part_multiplier <= 0.0 || p.cycle_part_multiplier > 10.0 {
            return Err(format!(
                "{}: cycle part multiplier should be in range (0, 10]",
                invalid
            ));
        }

        // Resolve defaults. Default bar component is Median (hl/2).
        let bc = p.bar_component.unwrap_or(BarComponent::Median);
        let qc = p.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = p.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let estimator = new_cycle_estimator(p.estimator_type, &p.estimator_params)?;

        // Build estimator moniker (only if non-default).
        let estimator_moniker_str = {
            let is_default = p.estimator_type == CycleEstimatorType::HomodyneDiscriminator
                && p.estimator_params.smoothing_length == 4
                && p.estimator_params.alpha_ema_quadrature_in_phase == 0.2
                && p.estimator_params.alpha_ema_period == 0.2;
            if is_default {
                String::new()
            } else {
                let m = estimator_moniker(p.estimator_type, estimator.as_ref());
                if m.is_empty() {
                    String::new()
                } else {
                    format!(", {}", m)
                }
            }
        };

        let component_mnemonic = component_triple_mnemonic(bc, qc, tc);

        let mnemonic = format!(
            "htitl({:.3}, {}, {:.3}{}{})",
            p.alpha_ema_period_additional,
            p.trend_line_smoothing_length,
            p.cycle_part_multiplier,
            estimator_moniker_str,
            component_mnemonic
        );
        let mnemonic_dcp = format!(
            "dcp({:.3}{}{})",
            p.alpha_ema_period_additional,
            estimator_moniker_str,
            component_mnemonic
        );

        let description = format!("Hilbert transformer instantaneous trend line {}", mnemonic);
        let description_dcp = format!("Dominant cycle period {}", mnemonic_dcp);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let (c0, c1, c2, c3) = match p.trend_line_smoothing_length {
            2 => (2.0 / 3.0, 1.0 / 3.0, 0.0, 0.0),
            3 => (3.0 / 6.0, 2.0 / 6.0, 1.0 / 6.0, 0.0),
            _ => (4.0 / 10.0, 3.0 / 10.0, 2.0 / 10.0, 1.0 / 10.0),
        };

        let max_period = estimator.max_period();

        Ok(Self {
            mnemonic,
            description,
            mnemonic_dcp,
            description_dcp,
            htce: estimator,
            alpha_ema_period_additional: p.alpha_ema_period_additional,
            one_min_alpha_ema_period_additional: 1.0 - p.alpha_ema_period_additional,
            cycle_part_multiplier: p.cycle_part_multiplier,
            coeff0: c0,
            coeff1: c1,
            coeff2: c2,
            coeff3: c3,
            smoothed_period: 0.0,
            value: 0.0,
            average1: 0.0,
            average2: 0.0,
            average3: 0.0,
            input: vec![0.0; max_period],
            input_length: max_period,
            input_length_min1: max_period - 1,
            primed: false,
            bar_func,
            quote_func,
            trade_func,
        })
    }

    /// Updates the indicator given the next sample, returning (value, period).
    /// Returns NaN values if not yet primed.
    pub fn update(&mut self, sample: f64) -> (f64, f64) {
        if sample.is_nan() {
            return (f64::NAN, f64::NAN);
        }

        self.htce.update(sample);
        self.push_input(sample);

        if self.primed {
            self.smoothed_period = self.alpha_ema_period_additional * self.htce.period()
                + self.one_min_alpha_ema_period_additional * self.smoothed_period;
            let average = self.calculate_average();
            self.value = self.coeff0 * average
                + self.coeff1 * self.average1
                + self.coeff2 * self.average2
                + self.coeff3 * self.average3;
            self.average3 = self.average2;
            self.average2 = self.average1;
            self.average1 = average;

            return (self.value, self.smoothed_period);
        }

        if self.htce.primed() {
            self.primed = true;
            self.smoothed_period = self.htce.period();
            let average = self.calculate_average();
            self.value = average;
            self.average1 = average;
            self.average2 = average;
            self.average3 = average;

            return (self.value, self.smoothed_period);
        }

        (f64::NAN, f64::NAN)
    }

    fn push_input(&mut self, value: f64) {
        for i in (1..self.input_length).rev() {
            self.input[i] = self.input[i - 1];
        }
        self.input[0] = value;
    }

    fn calculate_average(&self) -> f64 {
        let length = ((self.smoothed_period * self.cycle_part_multiplier + 0.5).floor() as usize)
            .clamp(1, self.input_length);

        let sum: f64 = self.input[..length].iter().sum();
        sum / length as f64
    }

    fn update_entity(&mut self, time: i64, sample: f64) -> Output {
        let (value, period) = self.update(sample);
        vec![
            Box::new(Scalar::new(time, value)) as Box<dyn Any>,
            Box::new(Scalar::new(time, period)) as Box<dyn Any>,
        ]
    }
}

impl Indicator for HilbertTransformerInstantaneousTrendLine {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::HilbertTransformerInstantaneousTrendLine,
            &self.mnemonic,
            &self.description,
            &[
                OutputText {
                    mnemonic: self.mnemonic.clone(),
                    description: self.description.clone(),
                },
                OutputText {
                    mnemonic: self.mnemonic_dcp.clone(),
                    description: self.description_dcp.clone(),
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
    const TOLERANCE: f64 = 1e-4;
    const SKIP: usize = 9;
    const SETTLE_SKIP: usize = 177;

    fn create_default() -> HilbertTransformerInstantaneousTrendLine {
        HilbertTransformerInstantaneousTrendLine::new_default().unwrap()
    }

    #[test]
    fn test_reference_value() {
        let mut x = create_default();
        let input = testdata::test_input();
        let exp_value = testdata::test_expected_value();

        for i in SKIP..input.len() {
            let (value, _) = x.update(input[i]);
            if value.is_nan() || i < SETTLE_SKIP {
                continue;
            }
            if exp_value[i].is_nan() {
                continue;
            }
            assert!(
                (exp_value[i] - value).abs() <= TOLERANCE,
                "[{}] value: expected {}, actual {}", i, exp_value[i], value
            );
        }
    }

    #[test]
    fn test_reference_period() {
        let mut x = create_default();
        let input = testdata::test_input();
        let exp_period = testdata::test_expected_period();

        for i in SKIP..input.len() {
            let (_, period) = x.update(input[i]);
            if period.is_nan() || i < SETTLE_SKIP {
                continue;
            }
            assert!(
                (exp_period[i] - period).abs() <= TOLERANCE,
                "[{}] period: expected {}, actual {}", i, exp_period[i], period
            );
        }
    }

    #[test]
    fn test_nan_input() {
        let mut x = create_default();
        let (value, period) = x.update(f64::NAN);
        assert!(value.is_nan());
        assert!(period.is_nan());
    }

    #[test]
    fn test_is_primed() {
        let mut x = create_default();
        let input = testdata::test_input();

        assert!(!x.is_primed());

        let mut primed_at: Option<usize> = None;

        for i in 0..input.len() {
            x.update(input[i]);
            if x.is_primed() && primed_at.is_none() {
                primed_at = Some(i);
            }
        }

        assert!(primed_at.is_some(), "expected indicator to become primed");
        assert!(x.is_primed());
    }

    #[test]
    fn test_metadata() {
        let x = create_default();
        let m = x.metadata();

        let mnemonic = "htitl(0.330, 4, 1.000, hl/2)";
        let mnemonic_dcp = "dcp(0.330, hl/2)";

        assert_eq!(m.identifier, Identifier::HilbertTransformerInstantaneousTrendLine);
        assert_eq!(m.mnemonic, mnemonic);
        assert_eq!(
            m.description,
            format!("Hilbert transformer instantaneous trend line {}", mnemonic)
        );
        assert_eq!(m.outputs.len(), 2);
        assert_eq!(m.outputs[0].mnemonic, mnemonic);
        assert_eq!(m.outputs[1].mnemonic, mnemonic_dcp);
    }

    #[test]
    fn test_update_entity_scalar() {
        let mut x = create_default();
        let input = testdata::test_input();

        for i in 0..200 {
            x.update(input[i % input.len()]);
        }

        let s = Scalar::new(1000, 100.0);
        let out = x.update_scalar(&s);
        assert_eq!(out.len(), 2);

        let s0 = out[0].downcast_ref::<Scalar>().unwrap();
        let s1 = out[1].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s0.time, 1000);
        assert_eq!(s1.time, 1000);
    }

    #[test]
    fn test_update_entity_bar() {
        let mut x = create_default();
        let input = testdata::test_input();

        for i in 0..200 {
            x.update(input[i % input.len()]);
        }

        let bar = Bar::new(1000, 0.0, 100.0, 100.0, 0.0, 0.0);
        let out = x.update_bar(&bar);
        assert_eq!(out.len(), 2);

        let s0 = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s0.time, 1000);
    }

    #[test]
    fn test_update_entity_quote() {
        let mut x = create_default();
        let input = testdata::test_input();

        for i in 0..200 {
            x.update(input[i % input.len()]);
        }

        let q = Quote::new(1000, 100.0, 100.0, 0.0, 0.0);
        let out = x.update_quote(&q);
        assert_eq!(out.len(), 2);

        let s0 = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s0.time, 1000);
    }

    #[test]
    fn test_update_entity_trade() {
        let mut x = create_default();
        let input = testdata::test_input();

        for i in 0..200 {
            x.update(input[i % input.len()]);
        }

        let t = Trade::new(1000, 100.0, 0.0);
        let out = x.update_trade(&t);
        assert_eq!(out.len(), 2);

        let s0 = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s0.time, 1000);
    }

    #[test]
    fn test_new_validation_alpha() {
        let mut p = HilbertTransformerInstantaneousTrendLineParams::default();
        p.alpha_ema_period_additional = 0.0;
        assert!(HilbertTransformerInstantaneousTrendLine::new(&p).is_err());

        p.alpha_ema_period_additional = 1.00000001;
        assert!(HilbertTransformerInstantaneousTrendLine::new(&p).is_err());

        p.alpha_ema_period_additional = 1.0;
        assert!(HilbertTransformerInstantaneousTrendLine::new(&p).is_ok());
    }

    #[test]
    fn test_new_validation_tlsl() {
        let mut p = HilbertTransformerInstantaneousTrendLineParams::default();
        p.trend_line_smoothing_length = 1;
        assert!(HilbertTransformerInstantaneousTrendLine::new(&p).is_err());

        p.trend_line_smoothing_length = 5;
        assert!(HilbertTransformerInstantaneousTrendLine::new(&p).is_err());

        p.trend_line_smoothing_length = 2;
        assert!(HilbertTransformerInstantaneousTrendLine::new(&p).is_ok());
    }

    #[test]
    fn test_new_validation_cpm() {
        let mut p = HilbertTransformerInstantaneousTrendLineParams::default();
        p.cycle_part_multiplier = 0.0;
        assert!(HilbertTransformerInstantaneousTrendLine::new(&p).is_err());

        p.cycle_part_multiplier = 10.00001;
        assert!(HilbertTransformerInstantaneousTrendLine::new(&p).is_err());

        p.cycle_part_multiplier = 10.0;
        assert!(HilbertTransformerInstantaneousTrendLine::new(&p).is_ok());
    }

    #[test]
    fn test_tlsl_2_coefficients() {
        let mut p = HilbertTransformerInstantaneousTrendLineParams::default();
        p.trend_line_smoothing_length = 2;
        let x = HilbertTransformerInstantaneousTrendLine::new(&p).unwrap();
        assert_eq!(x.coeff0, 2.0 / 3.0);
        assert_eq!(x.coeff1, 1.0 / 3.0);
    }

    #[test]
    fn test_tlsl_3_coefficients() {
        let mut p = HilbertTransformerInstantaneousTrendLineParams::default();
        p.trend_line_smoothing_length = 3;
        let x = HilbertTransformerInstantaneousTrendLine::new(&p).unwrap();
        assert_eq!(x.coeff0, 3.0 / 6.0);
        assert_eq!(x.coeff1, 2.0 / 6.0);
        assert_eq!(x.coeff2, 1.0 / 6.0);
    }
}
