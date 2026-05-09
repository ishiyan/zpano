use crate::entities::bar::Bar;
use crate::entities::bar_component::{component_value as bar_component_value, BarComponent, DEFAULT_BAR_COMPONENT};
use crate::entities::quote::Quote;
use crate::entities::quote_component::{component_value as quote_component_value, QuoteComponent, DEFAULT_QUOTE_COMPONENT};
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::entities::trade_component::{component_value as trade_component_value, TradeComponent, DEFAULT_TRADE_COMPONENT};
use crate::indicators::common::exponential_moving_average::{ExponentialMovingAverage, ExponentialMovingAverageLengthParams};
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::component_triple_mnemonic::component_triple_mnemonic;
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::line_indicator::LineIndicator;
use crate::indicators::core::metadata::Metadata;

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the TRIX indicator.
pub struct TripleExponentialMovingAverageOscillatorParams {
    /// The number of time periods for the three chained EMA calculations.
    /// Must be >= 1. Default is 30.
    pub length: usize,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for TripleExponentialMovingAverageOscillatorParams {
    fn default() -> Self {
        Self {
            length: 30,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the TRIX indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum TripleExponentialMovingAverageOscillatorOutput {
    /// The scalar value of the oscillator.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Jack Hutson's Triple Exponential Moving Average Oscillator (TRIX).
///
/// TRIX is a 1-day rate-of-change of a triple-smoothed EMA:
///
///   TRIX = ((EMA3[i] - EMA3[i-1]) / EMA3[i-1]) * 100
///
/// The indicator oscillates around zero.
pub struct TripleExponentialMovingAverageOscillator {
    line: LineIndicator,
    ema1: ExponentialMovingAverage,
    ema2: ExponentialMovingAverage,
    ema3: ExponentialMovingAverage,
    previous_ema3: f64,
    has_previous_ema: bool,
    primed: bool,
}

impl TripleExponentialMovingAverageOscillator {
    /// Creates a new TRIX from the given parameters.
    pub fn new(params: &TripleExponentialMovingAverageOscillatorParams) -> Result<Self, String> {
        if params.length < 1 {
            return Err("invalid triple exponential moving average oscillator parameters: length should be positive".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let ema_params = ExponentialMovingAverageLengthParams {
            length: params.length as i64,
            first_is_average: true,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        };

        let ema1 = ExponentialMovingAverage::new_from_length(&ema_params)?;
        let ema2 = ExponentialMovingAverage::new_from_length(&ema_params)?;
        let ema3 = ExponentialMovingAverage::new_from_length(&ema_params)?;

        let mnemonic = format!("trix({}{})", params.length, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Triple exponential moving average oscillator {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        Ok(Self {
            line,
            ema1,
            ema2,
            ema3,
            previous_ema3: f64::NAN,
            has_previous_ema: false,
            primed: false,
        })
    }

    /// Core update logic.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        let v1 = self.ema1.update(sample);
        if v1.is_nan() {
            return f64::NAN;
        }

        let v2 = self.ema2.update(v1);
        if v2.is_nan() {
            return f64::NAN;
        }

        let v3 = self.ema3.update(v2);
        if v3.is_nan() {
            return f64::NAN;
        }

        if !self.has_previous_ema {
            self.previous_ema3 = v3;
            self.has_previous_ema = true;
            return f64::NAN;
        }

        let result = ((v3 - self.previous_ema3) / self.previous_ema3) * 100.0;
        self.previous_ema3 = v3;

        if !self.primed {
            self.primed = true;
        }

        result
    }
}

impl Indicator for TripleExponentialMovingAverageOscillator {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::TripleExponentialMovingAverageOscillator,
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
    use crate::indicators::core::outputs::shape::Shape;
    const TOLERANCE: f64 = 1e-10;

    #[test]
    fn test_values() {
        let closes = testdata::test_closes();
        let expected = testdata::test_expected();

        let mut ind = TripleExponentialMovingAverageOscillator::new(
            &TripleExponentialMovingAverageOscillatorParams { length: 5, ..Default::default() },
        ).unwrap();

        for (i, &c) in closes.iter().enumerate() {
            let result = ind.update(c);

            if expected[i].is_nan() {
                assert!(result.is_nan(), "[{}] expected NaN, got {}", i, result);
            } else {
                assert!(!result.is_nan(), "[{}] expected {}, got NaN", i, expected[i]);
                assert!(
                    (expected[i] - result).abs() <= TOLERANCE,
                    "[{}] expected {}, got {}", i, expected[i], result,
                );
            }
        }
    }

    #[test]
    fn test_is_primed() {
        let closes = testdata::test_closes();

        let mut ind = TripleExponentialMovingAverageOscillator::new(
            &TripleExponentialMovingAverageOscillatorParams { length: 5, ..Default::default() },
        ).unwrap();

        // Lookback = 3*(5-1) + 1 = 13. First primed at index 13.
        for i in 0..13 {
            ind.update(closes[i]);
            assert!(!ind.is_primed(), "should not be primed at index {}", i);
        }

        ind.update(closes[13]);
        assert!(ind.is_primed(), "should be primed at index 13");
    }

    #[test]
    fn test_metadata() {
        let ind = TripleExponentialMovingAverageOscillator::new(
            &TripleExponentialMovingAverageOscillatorParams { length: 30, ..Default::default() },
        ).unwrap();

        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::TripleExponentialMovingAverageOscillator);
        assert_eq!(meta.mnemonic, "trix(30)");
        assert_eq!(meta.description, "Triple exponential moving average oscillator trix(30)");
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].kind, TripleExponentialMovingAverageOscillatorOutput::Value as i32);
        assert_eq!(meta.outputs[0].shape, Shape::Scalar);
    }

    #[test]
    fn test_invalid_params() {
        assert!(TripleExponentialMovingAverageOscillator::new(
            &TripleExponentialMovingAverageOscillatorParams { length: 0, ..Default::default() },
        ).is_err());
    }

    #[test]
    fn test_nan() {
        let mut ind = TripleExponentialMovingAverageOscillator::new(
            &TripleExponentialMovingAverageOscillatorParams { length: 5, ..Default::default() },
        ).unwrap();

        assert!(ind.update(f64::NAN).is_nan());
    }
}
