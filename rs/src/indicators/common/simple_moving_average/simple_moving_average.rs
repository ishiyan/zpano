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

/// Parameters to create an instance of the simple moving average indicator.
pub struct SimpleMovingAverageParams {
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

impl Default for SimpleMovingAverageParams {
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

/// Enumerates the outputs of the simple moving average indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum SimpleMovingAverageOutput {
    /// The scalar value of the moving average.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Computes the simple (arithmetic) moving average (SMA).
///
/// SMAᵢ = SMAᵢ₋₁ + (Pᵢ − Pᵢ₋ℓ) / ℓ
///
/// The indicator is not primed during the first ℓ−1 updates.
pub struct SimpleMovingAverage {
    line: LineIndicator,
    window: Vec<f64>,
    window_sum: f64,
    window_length: usize,
    window_count: usize,
    last_index: usize,
    primed: bool,
}

impl SimpleMovingAverage {
    /// Creates a new SimpleMovingAverage from the given parameters.
    pub fn new(params: &SimpleMovingAverageParams) -> Result<Self, String> {
        if params.length < 2 {
            return Err("invalid simple moving average parameters: length should be greater than 1".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!("sma({}{})", params.length, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Simple moving average {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        Ok(Self {
            line,
            window: vec![0.0; params.length],
            window_sum: 0.0,
            window_length: params.length,
            window_count: 0,
            last_index: params.length - 1,
            primed: false,
        })
    }

    /// Core update logic. Returns the SMA value or NaN if not yet primed.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        if self.primed {
            self.window_sum += sample - self.window[0];

            for i in 0..self.last_index {
                self.window[i] = self.window[i + 1];
            }

            self.window[self.last_index] = sample;
        } else {
            self.window_sum += sample;
            self.window[self.window_count] = sample;
            self.window_count += 1;

            if self.window_length > self.window_count {
                return f64::NAN;
            }

            self.primed = true;
        }

        self.window_sum / self.window_length as f64
    }
}

impl Indicator for SimpleMovingAverage {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::SimpleMovingAverage,
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
    fn create_sma(length: usize) -> SimpleMovingAverage {
        SimpleMovingAverage::new(&SimpleMovingAverageParams { length, ..Default::default() }).unwrap()
    }

    #[test]
    fn test_update_length_3() {
        let mut sma = create_sma(3);
        let input = testdata::test_input();
        let expected = testdata::expected_3();

        for i in 0..2 {
            assert!(sma.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        for i in 2..input.len() {
            let act = sma.update(input[i]);
            assert!((expected[i] - act).abs() < 1e-2, "[{}] expected {}, got {}", i, expected[i], act);
        }

        assert!(sma.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_length_5() {
        let mut sma = create_sma(5);
        let input = testdata::test_input();
        let expected = testdata::expected_5();

        for i in 0..4 {
            assert!(sma.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        for i in 4..input.len() {
            let act = sma.update(input[i]);
            assert!((expected[i] - act).abs() < 1e-2, "[{}] expected {}, got {}", i, expected[i], act);
        }

        assert!(sma.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_length_10() {
        let mut sma = create_sma(10);
        let input = testdata::test_input();
        let expected = testdata::expected_10();

        for i in 0..9 {
            assert!(sma.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        for i in 9..input.len() {
            let act = sma.update(input[i]);
            assert!((expected[i] - act).abs() < 1e-2, "[{}] expected {}, got {}", i, expected[i], act);
        }

        assert!(sma.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_entity() {
        let length = 2;
        let inp = 3.0_f64;
        let exp = inp / length as f64;
        let time = 1617235200;

        // scalar
        let mut sma = create_sma(length);
        sma.update(0.0);
        let out = sma.update_scalar(&Scalar::new(time, inp));
        assert_eq!(out.len(), 1);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s.time, time);
        assert_eq!(s.value, exp);

        // bar
        let mut sma = create_sma(length);
        sma.update(0.0);
        let bar = Bar::new(time, 0.0, 0.0, 0.0, inp, 0.0);
        let out = sma.update_bar(&bar);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s.value, exp);

        // quote
        let mut sma = create_sma(length);
        sma.update(0.0);
        let quote = Quote::new(time, inp, inp, 0.0, 0.0);
        let out = sma.update_quote(&quote);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s.value, exp);

        // trade
        let mut sma = create_sma(length);
        sma.update(0.0);
        let trade = Trade::new(time, inp, 0.0);
        let out = sma.update_trade(&trade);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s.value, exp);
    }

    #[test]
    fn test_is_primed() {
        let input = testdata::test_input();

        let mut sma = create_sma(3);
        assert!(!sma.is_primed());
        for i in 0..2 {
            sma.update(input[i]);
            assert!(!sma.is_primed(), "[{}] should not be primed", i);
        }
        for i in 2..input.len() {
            sma.update(input[i]);
            assert!(sma.is_primed(), "[{}] should be primed", i);
        }
    }

    #[test]
    fn test_metadata() {
        let sma = create_sma(5);
        let m = sma.metadata();
        assert_eq!(m.identifier, Identifier::SimpleMovingAverage);
        assert_eq!(m.outputs.len(), 1);
        assert_eq!(m.outputs[0].kind, SimpleMovingAverageOutput::Value as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
        assert_eq!(m.outputs[0].mnemonic, "sma(5)");
        assert_eq!(m.outputs[0].description, "Simple moving average sma(5)");
    }

    #[test]
    fn test_new_invalid_length() {
        let r = SimpleMovingAverage::new(&SimpleMovingAverageParams { length: 1, ..Default::default() });
        assert!(r.is_err());
        assert_eq!(r.err().unwrap(), "invalid simple moving average parameters: length should be greater than 1");

        let r = SimpleMovingAverage::new(&SimpleMovingAverageParams { length: 0, ..Default::default() });
        assert!(r.is_err());
    }

    #[test]
    fn test_mnemonic_components() {
        // all defaults -> no component suffix
        let sma = create_sma(5);
        assert_eq!(sma.line.mnemonic, "sma(5)");

        // bar component set
        let sma = SimpleMovingAverage::new(&SimpleMovingAverageParams {
            length: 5, bar_component: Some(BarComponent::Median), ..Default::default()
        }).unwrap();
        assert_eq!(sma.line.mnemonic, "sma(5, hl/2)");

        // all three set (bar non-default, quote default, trade non-default)
        let sma = SimpleMovingAverage::new(&SimpleMovingAverageParams {
            length: 5,
            bar_component: Some(BarComponent::High),
            quote_component: None,
            trade_component: Some(TradeComponent::Volume),
        }).unwrap();
        assert_eq!(sma.line.mnemonic, "sma(5, h, v)");
    }
}
