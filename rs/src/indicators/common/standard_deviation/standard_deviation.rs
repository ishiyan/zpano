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
use crate::indicators::common::variance::{Variance, VarianceParams};

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the standard deviation indicator.
pub struct StandardDeviationParams {
    pub length: usize,
    pub is_unbiased: bool,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for StandardDeviationParams {
    fn default() -> Self {
        Self { length: 20, is_unbiased: true, bar_component: None, quote_component: None, trade_component: None }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum StandardDeviationOutput {
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Computes the standard deviation as the square root of variance.
pub struct StandardDeviation {
    line: LineIndicator,
    variance: Variance,
}

impl StandardDeviation {
    pub fn new(params: &StandardDeviationParams) -> Result<Self, String> {
        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let vp = VarianceParams {
            length: params.length,
            is_unbiased: params.is_unbiased,
            bar_component: Some(bc),
            quote_component: Some(qc),
            trade_component: Some(tc),
        };

        let variance = Variance::new(&vp)?;

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let c = if params.is_unbiased { 's' } else { 'p' };
        let mnemonic = format!("stdev.{}({}{})", c, params.length, component_triple_mnemonic(bc, qc, tc));
        let description = if params.is_unbiased {
            format!("Standard deviation based on unbiased estimation of the sample variance {}", mnemonic)
        } else {
            format!("Standard deviation based on estimation of the population variance {}", mnemonic)
        };

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        Ok(Self { line, variance })
    }

    /// Core update logic.
    pub fn update(&mut self, sample: f64) -> f64 {
        let v = self.variance.update(sample);
        if v.is_nan() {
            return v;
        }
        v.sqrt()
    }
}

impl Indicator for StandardDeviation {
    fn is_primed(&self) -> bool {
        self.variance.is_primed()
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::StandardDeviation,
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
        let v = (self.line.bar_func)(sample);
        let value = self.update(v);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (self.line.quote_func)(sample);
        let value = self.update(v);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let v = (self.line.trade_func)(sample);
        let value = self.update(v);
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
    use crate::indicators::core::outputs::shape::Shape;
    fn create_stdev(length: usize, unbiased: bool) -> StandardDeviation {
        StandardDeviation::new(&StandardDeviationParams { length, is_unbiased: unbiased, ..Default::default() }).unwrap()
    }

    #[test]
    fn test_population_stdev_length_3() {
        let mut sd = create_stdev(3, false);
        let input = testdata::test_input();
        let expected = testdata::expected_len3_population();

        for i in 0..2 {
            assert!(sd.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        for i in 2..input.len() {
            let act = sd.update(input[i]);
            assert!((expected[i] - act).abs() < 1e-13, "[{}] expected {}, got {}", i, expected[i], act);
        }

        assert!(sd.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_sample_stdev_length_3() {
        let mut sd = create_stdev(3, true);
        let input = testdata::test_input();
        let expected = testdata::expected_len3_sample();

        for i in 0..2 {
            assert!(sd.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        for i in 2..input.len() {
            let act = sd.update(input[i]);
            assert!((expected[i] - act).abs() < 1e-13, "[{}] expected {}, got {}", i, expected[i], act);
        }
    }

    #[test]
    fn test_is_primed() {
        let input = testdata::test_input();
        let mut sd = create_stdev(3, false);
        assert!(!sd.is_primed());
        for i in 0..2 {
            sd.update(input[i]);
            assert!(!sd.is_primed());
        }
        for i in 2..input.len() {
            sd.update(input[i]);
            assert!(sd.is_primed());
        }
    }

    #[test]
    fn test_metadata_population() {
        let sd = create_stdev(7, false);
        let m = sd.metadata();
        assert_eq!(m.identifier, Identifier::StandardDeviation);
        assert_eq!(m.mnemonic, "stdev.p(7)");
        assert_eq!(m.description, "Standard deviation based on estimation of the population variance stdev.p(7)");
        assert_eq!(m.outputs.len(), 1);
        assert_eq!(m.outputs[0].kind, StandardDeviationOutput::Value as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
    }

    #[test]
    fn test_metadata_sample() {
        let sd = create_stdev(7, true);
        let m = sd.metadata();
        assert_eq!(m.mnemonic, "stdev.s(7)");
        assert_eq!(m.description, "Standard deviation based on unbiased estimation of the sample variance stdev.s(7)");
    }

    #[test]
    fn test_new_invalid_length() {
        assert!(StandardDeviation::new(&StandardDeviationParams { length: 1, ..Default::default() }).is_err());
    }
}
