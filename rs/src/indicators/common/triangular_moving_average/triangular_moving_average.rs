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

/// Parameters to create an instance of the triangular moving average indicator.
pub struct TriangularMovingAverageParams {
    /// The length (number of time periods) of the moving window.
    /// Must be greater than 1.
    pub length: usize,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for TriangularMovingAverageParams {
    fn default() -> Self {
        Self {
            length: 20,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the triangular moving average indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum TriangularMovingAverageOutput {
    /// The scalar value of the moving average.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Computes the triangular moving average (TRIMA).
///
/// The TRIMA puts more weight on the data in the middle of the window,
/// equivalent to computing an SMA of an SMA.
///
/// Uses an optimised incremental algorithm with four adjustments per step.
pub struct TriangularMovingAverage {
    line: LineIndicator,
    factor: f64,
    numerator: f64,
    numerator_sub: f64,
    numerator_add: f64,
    window: Vec<f64>,
    window_length: usize,
    window_length_half: usize,
    window_count: usize,
    is_odd: bool,
    primed: bool,
}

impl TriangularMovingAverage {
    /// Creates a new TriangularMovingAverage from the given parameters.
    pub fn new(params: &TriangularMovingAverageParams) -> Result<Self, String> {
        if params.length < 2 {
            return Err("invalid triangular moving average parameters: length should be greater than 1".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let length = params.length;
        let length_half = length >> 1;
        let l = 1 + length_half;
        let is_odd = length % 2 == 1;

        let (factor, window_length_half) = if is_odd {
            // Odd: 1+2+...+(l)+...+2+1 = l*l where l = (length+1)/2 = length_half+1.
            (1.0 / (l * l) as f64, length_half)
        } else {
            // Even: 1+2+...+l+l+...+2+1 = length_half * l.
            (1.0 / (length_half * l) as f64, length_half - 1)
        };

        let mnemonic = format!("trima({}{})", length, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Triangular moving average {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        Ok(Self {
            line,
            factor,
            numerator: 0.0,
            numerator_sub: 0.0,
            numerator_add: 0.0,
            window: vec![0.0; length],
            window_length: length,
            window_length_half,
            window_count: 0,
            is_odd,
            primed: false,
        })
    }

    /// Core update logic. Returns the TRIMA value or NaN if not yet primed.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        let temp = sample;

        if self.primed {
            self.numerator -= self.numerator_sub;
            self.numerator_sub -= self.window[0];

            let j = self.window_length - 1;
            for i in 0..j {
                self.window[i] = self.window[i + 1];
            }

            self.window[j] = temp;
            let mid = self.window[self.window_length_half];
            self.numerator_sub += mid;

            if self.is_odd {
                self.numerator += self.numerator_add;
                self.numerator_add -= mid;
            } else {
                self.numerator_add -= mid;
                self.numerator += self.numerator_add;
            }

            self.numerator_add += sample;
            self.numerator += sample;
        } else {
            self.window[self.window_count] = temp;
            self.window_count += 1;

            if self.window_length > self.window_count {
                return f64::NAN;
            }

            // Initialise numerator_sub from the middle going left.
            let half = self.window_length_half;
            for i in (0..=half).rev() {
                self.numerator_sub += self.window[i];
                self.numerator += self.numerator_sub;
            }

            // Initialise numerator_add from the middle+1 going right.
            for i in (half + 1)..self.window_length {
                self.numerator_add += self.window[i];
                self.numerator += self.numerator_add;
            }

            self.primed = true;
        }

        self.numerator * self.factor
    }
}

impl Indicator for TriangularMovingAverage {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::TriangularMovingAverage,
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
    use crate::indicators::core::outputs::shape::Shape;
    fn create_trima(length: usize) -> TriangularMovingAverage {
        TriangularMovingAverage::new(&TriangularMovingAverageParams { length, ..Default::default() }).unwrap()
    }

    #[test]
    fn test_update_length_9() {
        let mut trima = create_trima(9);
        let input = testdata::test_input();

        for i in 0..8 {
            assert!(trima.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        let act = trima.update(input[8]);
        assert!((93.8176 - act).abs() < 1e-4, "[8] expected 93.8176, got {}", act);

        for i in 9..input.len() - 1 {
            trima.update(input[i]);
        }

        let act = trima.update(input[251]);
        assert!((109.1312 - act).abs() < 1e-4, "[251] expected 109.1312, got {}", act);

        assert!(trima.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_length_10() {
        let mut trima = create_trima(10);
        let input = testdata::test_input();

        for i in 0..9 {
            assert!(trima.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        let act = trima.update(input[9]);
        assert!((93.6043 - act).abs() < 1e-4, "[9] expected 93.6043, got {}", act);

        let act = trima.update(input[10]);
        assert!((93.4252 - act).abs() < 1e-4, "[10] expected 93.4252, got {}", act);

        for i in 11..250 {
            trima.update(input[i]);
        }

        let act = trima.update(input[250]);
        assert!((109.1850 - act).abs() < 1e-4, "[250] expected 109.1850, got {}", act);

        let act = trima.update(input[251]);
        assert!((109.1407 - act).abs() < 1e-4, "[251] expected 109.1407, got {}", act);

        assert!(trima.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_length_12() {
        let mut trima = create_trima(12);
        let input = testdata::test_input();

        for i in 0..10 {
            assert!(trima.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        // index 10 is still NaN for length 12 (need 11 NaN values: indices 0..=10)
        assert!(trima.update(input[10]).is_nan(), "[10] expected NaN");

        let act = trima.update(input[11]);
        assert!((93.5329 - act).abs() < 1e-4, "[11] expected 93.5329, got {}", act);

        for i in 12..251 {
            trima.update(input[i]);
        }

        let act = trima.update(input[251]);
        assert!((109.1157 - act).abs() < 1e-4, "[251] expected 109.1157, got {}", act);

        assert!(trima.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_xls12() {
        let mut trima = create_trima(12);
        let input = testdata::test_input();
        let expected = testdata::expected_xls12();

        for i in 0..11 {
            assert!(trima.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        for i in 11..input.len() {
            let act = trima.update(input[i]);
            assert!(
                (expected[i] - act).abs() < 1e-12,
                "[{}] expected {}, got {}", i, expected[i], act
            );
        }
    }

    #[test]
    fn test_update_entity() {
        let length = 12;
        let input = testdata::test_input();
        let inp = 97.250000; // input[11]
        let exp = 93.5329761904762;
        let time = 1617235200_i64;

        // scalar
        let mut trima = create_trima(length);
        for i in 0..11 {
            trima.update(input[i]);
        }
        let out = trima.update_scalar(&Scalar::new(time, inp));
        assert_eq!(out.len(), 1);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s.time, time);
        assert!((exp - s.value).abs() < 1e-12);

        // bar
        let mut trima = create_trima(length);
        for i in 0..11 {
            trima.update(input[i]);
        }
        let bar = Bar::new(time, 0.0, 0.0, 0.0, inp, 0.0);
        let out = trima.update_bar(&bar);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((exp - s.value).abs() < 1e-12);

        // quote
        let mut trima = create_trima(length);
        for i in 0..11 {
            trima.update(input[i]);
        }
        let quote = Quote::new(time, inp, inp, 0.0, 0.0);
        let out = trima.update_quote(&quote);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((exp - s.value).abs() < 1e-12);

        // trade
        let mut trima = create_trima(length);
        for i in 0..11 {
            trima.update(input[i]);
        }
        let trade = Trade::new(time, inp, 0.0);
        let out = trima.update_trade(&trade);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((exp - s.value).abs() < 1e-12);
    }

    #[test]
    fn test_is_primed() {
        let input = testdata::test_input();

        // length 9: primed after 9 samples (index 8)
        let mut trima = create_trima(9);
        assert!(!trima.is_primed());
        for i in 0..8 {
            trima.update(input[i]);
            assert!(!trima.is_primed(), "[{}] should not be primed", i);
        }
        for i in 8..input.len() {
            trima.update(input[i]);
            assert!(trima.is_primed(), "[{}] should be primed", i);
        }

        // length 12: primed after 12 samples (index 11)
        let mut trima = create_trima(12);
        assert!(!trima.is_primed());
        for i in 0..11 {
            trima.update(input[i]);
            assert!(!trima.is_primed(), "[{}] should not be primed", i);
        }
        for i in 11..input.len() {
            trima.update(input[i]);
            assert!(trima.is_primed(), "[{}] should be primed", i);
        }
    }

    #[test]
    fn test_metadata() {
        let trima = create_trima(5);
        let m = trima.metadata();
        assert_eq!(m.identifier, Identifier::TriangularMovingAverage);
        assert_eq!(m.outputs.len(), 1);
        assert_eq!(m.outputs[0].kind, TriangularMovingAverageOutput::Value as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
        assert_eq!(m.outputs[0].mnemonic, "trima(5)");
        assert_eq!(m.outputs[0].description, "Triangular moving average trima(5)");
    }

    #[test]
    fn test_new_invalid_length() {
        let r = TriangularMovingAverage::new(&TriangularMovingAverageParams { length: 1, ..Default::default() });
        assert!(r.is_err());
        assert_eq!(r.err().unwrap(), "invalid triangular moving average parameters: length should be greater than 1");

        let r = TriangularMovingAverage::new(&TriangularMovingAverageParams { length: 0, ..Default::default() });
        assert!(r.is_err());
    }

    #[test]
    fn test_mnemonic_components() {
        // all defaults -> no component suffix
        let trima = create_trima(5);
        assert_eq!(trima.line.mnemonic, "trima(5)");

        // bar component set
        let trima = TriangularMovingAverage::new(&TriangularMovingAverageParams {
            length: 5, bar_component: Some(BarComponent::Median), ..Default::default()
        }).unwrap();
        assert_eq!(trima.line.mnemonic, "trima(5, hl/2)");

        // bar and trade set
        let trima = TriangularMovingAverage::new(&TriangularMovingAverageParams {
            length: 5,
            bar_component: Some(BarComponent::High),
            quote_component: None,
            trade_component: Some(TradeComponent::Volume),
        }).unwrap();
        assert_eq!(trima.line.mnemonic, "trima(5, h, v)");
    }
}
