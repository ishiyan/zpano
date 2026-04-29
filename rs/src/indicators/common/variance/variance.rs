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

/// Parameters to create an instance of the variance indicator.
pub struct VarianceParams {
    /// The length (number of time periods) of the moving window. Must be > 1.
    pub length: usize,
    /// Whether to use unbiased sample variance (true) or population variance (false).
    pub is_unbiased: bool,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for VarianceParams {
    fn default() -> Self {
        Self { length: 20, is_unbiased: true, bar_component: None, quote_component: None, trade_component: None }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum VarianceOutput {
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Computes the variance of samples within a moving window.
pub struct Variance {
    line: LineIndicator,
    window: Vec<f64>,
    window_sum: f64,
    window_squared_sum: f64,
    window_length: usize,
    window_count: usize,
    last_index: usize,
    primed: bool,
    unbiased: bool,
}

impl Variance {
    pub fn new(params: &VarianceParams) -> Result<Self, String> {
        if params.length < 2 {
            return Err("invalid variance parameters: length should be greater than 1".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let c = if params.is_unbiased { 's' } else { 'p' };
        let mnemonic = format!("var.{}({}{})", c, params.length, component_triple_mnemonic(bc, qc, tc));
        let description = if params.is_unbiased {
            format!("Unbiased estimation of the sample variance {}", mnemonic)
        } else {
            format!("Estimation of the population variance {}", mnemonic)
        };

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        Ok(Self {
            line,
            window: vec![0.0; params.length],
            window_sum: 0.0,
            window_squared_sum: 0.0,
            window_length: params.length,
            window_count: 0,
            last_index: params.length - 1,
            primed: false,
            unbiased: params.is_unbiased,
        })
    }

    /// Core update logic.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        let mut temp = sample;
        let wlen = self.window_length as f64;

        if self.primed {
            self.window_sum += temp;
            temp *= temp;
            self.window_squared_sum += temp;
            temp = self.window[0];
            self.window_sum -= temp;
            temp *= temp;
            self.window_squared_sum -= temp;

            let value = if self.unbiased {
                temp = self.window_sum;
                temp *= temp;
                temp /= wlen;
                (self.window_squared_sum - temp) / self.last_index as f64
            } else {
                temp = self.window_sum / wlen;
                temp *= temp;
                self.window_squared_sum / wlen - temp
            };

            for i in 0..self.last_index {
                self.window[i] = self.window[i + 1];
            }

            self.window[self.last_index] = sample;

            value
        } else {
            self.window_sum += temp;
            self.window[self.window_count] = temp;
            temp *= temp;
            self.window_squared_sum += temp;

            self.window_count += 1;
            if self.window_length == self.window_count {
                self.primed = true;
                if self.unbiased {
                    temp = self.window_sum;
                    temp *= temp;
                    temp /= wlen;
                    (self.window_squared_sum - temp) / self.last_index as f64
                } else {
                    temp = self.window_sum / wlen;
                    temp *= temp;
                    self.window_squared_sum / wlen - temp
                }
            } else {
                f64::NAN
            }
        }
    }
}

impl Indicator for Variance {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::Variance,
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
    use crate::indicators::core::outputs::shape::Shape;

    fn test_input() -> Vec<f64> {
        vec![1.0, 2.0, 8.0, 4.0, 9.0, 6.0, 7.0, 13.0, 9.0, 10.0, 3.0, 12.0]
    }

    fn expected_len3_population() -> Vec<f64> {
        vec![
            f64::NAN, f64::NAN,
            9.55555555555556000, 6.22222222222222000, 4.66666666666667000, 4.22222222222222000, 1.55555555555556000,
            9.55555555555556000, 6.22222222222222000, 2.88888888888889000, 9.55555555555556000, 14.88888888888890000,
        ]
    }

    fn expected_len5_population() -> Vec<f64> {
        vec![
            f64::NAN, f64::NAN, f64::NAN, f64::NAN,
            10.16000, 6.56000, 2.96000, 9.36000, 5.76000, 6.00000, 11.04000, 12.24000,
        ]
    }

    fn expected_len3_sample() -> Vec<f64> {
        vec![
            f64::NAN, f64::NAN,
            14.3333333333333000, 9.3333333333333400, 7.0000000000000000, 6.3333333333333400, 2.3333333333333300,
            14.3333333333333000, 9.3333333333333400, 4.3333333333333400, 14.3333333333333000, 22.3333333333333000,
        ]
    }

    fn create_variance(length: usize, unbiased: bool) -> Variance {
        Variance::new(&VarianceParams { length, is_unbiased: unbiased, ..Default::default() }).unwrap()
    }

    #[test]
    fn test_population_variance_length_3() {
        let mut v = create_variance(3, false);
        let input = test_input();
        let expected = expected_len3_population();

        for i in 0..2 {
            assert!(v.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        for i in 2..input.len() {
            let act = v.update(input[i]);
            assert!((expected[i] - act).abs() < 1e-13, "[{}] expected {}, got {}", i, expected[i], act);
        }

        assert!(v.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_population_variance_length_5() {
        let mut v = create_variance(5, false);
        let input = test_input();
        let expected = expected_len5_population();

        for i in 0..4 {
            assert!(v.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        for i in 4..input.len() {
            let act = v.update(input[i]);
            assert!((expected[i] - act).abs() < 1e-13, "[{}] expected {}, got {}", i, expected[i], act);
        }
    }

    #[test]
    fn test_sample_variance_length_3() {
        let mut v = create_variance(3, true);
        let input = test_input();
        let expected = expected_len3_sample();

        for i in 0..2 {
            assert!(v.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        for i in 2..input.len() {
            let act = v.update(input[i]);
            assert!((expected[i] - act).abs() < 1e-13, "[{}] expected {}, got {}", i, expected[i], act);
        }

        assert!(v.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_is_primed() {
        let input = test_input();
        let mut v = create_variance(3, false);

        assert!(!v.is_primed());
        for i in 0..2 {
            v.update(input[i]);
            assert!(!v.is_primed());
        }
        for i in 2..input.len() {
            v.update(input[i]);
            assert!(v.is_primed());
        }
    }

    #[test]
    fn test_metadata_population() {
        let v = create_variance(7, false);
        let m = v.metadata();
        assert_eq!(m.identifier, Identifier::Variance);
        assert_eq!(m.mnemonic, "var.p(7)");
        assert_eq!(m.description, "Estimation of the population variance var.p(7)");
        assert_eq!(m.outputs.len(), 1);
        assert_eq!(m.outputs[0].kind, VarianceOutput::Value as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
    }

    #[test]
    fn test_metadata_sample() {
        let v = create_variance(7, true);
        let m = v.metadata();
        assert_eq!(m.mnemonic, "var.s(7)");
        assert_eq!(m.description, "Unbiased estimation of the sample variance var.s(7)");
    }

    #[test]
    fn test_new_invalid_length() {
        assert!(Variance::new(&VarianceParams { length: 1, ..Default::default() }).is_err());
        assert!(Variance::new(&VarianceParams { length: 0, ..Default::default() }).is_err());
    }
}
