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

/// Parameters for the Relative Strength Index indicator.
pub struct RelativeStrengthIndexParams {
    /// Number of periods. Must be >= 2. Default is 14.
    pub length: usize,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for RelativeStrengthIndexParams {
    fn default() -> Self {
        Self {
            length: 14,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the RSI indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum RelativeStrengthIndexOutput {
    /// The scalar value of the RSI.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

const EPSILON: f64 = 1e-8;

/// Welles Wilder's Relative Strength Index (RSI).
///
/// RSI measures the magnitude of recent price changes to evaluate overbought
/// or oversold conditions. It oscillates between 0 and 100.
pub struct RelativeStrengthIndex {
    line: LineIndicator,
    length: usize,
    count: i32,
    previous_sample: f64,
    previous_gain: f64,
    previous_loss: f64,
    value: f64,
    primed: bool,
}

impl RelativeStrengthIndex {
    /// Creates a new RelativeStrengthIndex from the given parameters.
    pub fn new(params: &RelativeStrengthIndexParams) -> Result<Self, String> {
        if params.length < 2 {
            return Err("invalid relative strength index parameters: length should be greater than 1".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!("rsi({}{})", params.length, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Relative Strength Index {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        Ok(Self {
            line,
            length: params.length,
            count: -1,
            previous_sample: 0.0,
            previous_gain: 0.0,
            previous_loss: 0.0,
            value: f64::NAN,
            primed: false,
        })
    }

    /// Core update with a single sample value.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        self.count += 1;

        if self.count == 0 {
            self.previous_sample = sample;
            return self.value;
        }

        let temp = sample - self.previous_sample;
        self.previous_sample = sample;

        if !self.primed {
            // Accumulation phase: count 1..length-1.
            if temp < 0.0 {
                self.previous_loss -= temp;
            } else {
                self.previous_gain += temp;
            }

            if (self.count as usize) < self.length {
                return self.value;
            }

            // Priming: count == length.
            self.previous_gain /= self.length as f64;
            self.previous_loss /= self.length as f64;
            self.primed = true;
        } else {
            // Wilder's smoothing.
            self.previous_gain *= (self.length - 1) as f64;
            self.previous_loss *= (self.length - 1) as f64;

            if temp < 0.0 {
                self.previous_loss -= temp;
            } else {
                self.previous_gain += temp;
            }

            self.previous_gain /= self.length as f64;
            self.previous_loss /= self.length as f64;
        }

        let sum = self.previous_gain + self.previous_loss;
        if sum > EPSILON {
            self.value = 100.0 * self.previous_gain / sum;
        } else {
            self.value = 0.0;
        }

        self.value
    }
}

impl Indicator for RelativeStrengthIndex {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::RelativeStrengthIndex,
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

#[cfg(test)]
mod tests {
    use super::*;

    fn test_input_1() -> Vec<f64> {
        vec![
            91.15, 90.50, 92.55, 94.70, 95.55, 94.00, 91.30, 91.95, 92.45, 93.80,
            92.50, 94.55, 96.75, 97.80, 98.40, 98.15, 96.70, 98.85, 98.90, 100.50,
            102.60, 104.80, 103.80, 103.10, 102.00,
        ]
    }

    fn test_expected_1() -> Vec<f64> {
        vec![
            f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN,
            f64::NAN, f64::NAN, f64::NAN, f64::NAN,
            60.6425702811244, 54.2677448337826, 61.4558190165176, 67.6034767388667,
            70.1590191481383, 71.5992400904851, 70.0152589447766, 61.1833361324987,
            67.9312249318593, 68.076417836971, 72.5504646296262, 77.2568847385616,
            81.0801123570899, 74.6619680507228, 70.2808713845906, 63.6754215506388,
        ]
    }

    fn test_input_2() -> Vec<f64> {
        vec![
            44.34, 44.09, 43.61, 44.33, 44.83, 45.10, 45.42, 45.84, 46.08, 45.89,
            46.03, 45.61, 46.28, 46.28, 46.00, 46.03, 46.41, 46.22, 45.64, 46.21,
            46.25, 45.71, 46.45, 45.78, 45.35, 44.03, 44.18, 44.22, 44.57, 43.42,
            42.66, 43.13, 44.94, 43.61, 44.33, 44.83, 45.10, 45.42, 45.84, 46.08,
            45.89, 46.03, 45.61, 46.28, 46.28, 46.00, 46.03, 46.41, 46.22, 45.64,
            46.21, 46.25, 45.71, 46.45, 45.78, 45.35, 44.03, 44.18, 44.22, 44.57,
            43.42, 42.66, 43.13, 44.94, 43.61, 44.33, 44.83, 45.10, 45.42, 45.84,
            46.08, 45.89, 46.03, 45.61, 46.28, 46.28, 46.00, 46.03, 46.41, 46.22,
            45.64, 46.21, 46.25, 45.71, 46.45, 45.78, 45.35, 44.03, 44.18, 44.22,
            44.57, 43.42, 42.66, 43.13, 44.94, 43.61, 44.33, 44.83, 45.10, 45.42,
            45.84, 46.08, 45.89, 46.03, 45.61, 46.28, 46.28, 46.00, 46.03, 46.41,
            46.22, 45.64, 46.21, 46.25, 45.71, 46.45, 45.78, 45.35, 44.03, 44.18,
            44.22, 44.57, 43.42, 42.66, 43.13, 44.94, 43.61, 44.33, 44.83, 45.10,
            45.42, 45.84, 46.08, 45.89, 46.03, 45.61, 46.28, 46.28, 46.00, 46.03,
            46.41, 46.22, 45.64, 46.21, 46.25, 45.71, 46.45, 45.78, 45.35, 44.03,
            44.18, 44.22, 44.57, 43.42, 42.66, 43.13, 44.94, 43.61, 44.33, 44.83,
            45.10, 45.42, 45.84, 46.08, 45.89, 46.03, 45.61, 46.28, 46.28, 46.00,
            46.03, 46.41, 46.22, 45.64, 46.21, 46.25, 45.71, 46.45, 45.78, 45.35,
            44.03, 44.18, 44.22, 44.57, 43.42, 42.66, 43.13, 44.94, 43.61, 44.33,
            44.83, 45.10, 45.42, 45.84, 46.08, 45.89, 46.03, 45.61, 46.28, 46.28,
            46.00, 46.03, 46.41, 46.22, 45.64, 46.21, 46.25, 45.71, 46.45, 45.78,
            45.35, 44.03, 44.18, 44.22, 44.57, 43.42, 42.66, 43.13, 44.94, 43.61,
            44.33, 44.83, 45.10, 45.42, 45.84, 46.08, 45.89, 46.03, 45.61, 46.28,
            46.28, 46.00, 46.03, 46.41, 46.22, 45.64, 46.21, 46.25, 45.71, 46.45,
            45.78, 45.35, 44.03, 44.18, 44.22, 44.57, 43.42, 42.66, 43.13, 44.94,
            43.61, 44.33,
        ]
    }

    #[test]
    fn test_rsi_update() {
        let tolerance = 1e-9;
        let input = test_input_1();
        let expected = test_expected_1();
        let params = RelativeStrengthIndexParams { length: 9, ..Default::default() };
        let mut rsi = RelativeStrengthIndex::new(&params).unwrap();

        for i in 0..input.len() {
            let act = rsi.update(input[i]);

            if i < 9 {
                assert!(act.is_nan(), "[{}] expected NaN, got {}", i, act);
                continue;
            }

            assert!(
                (act - expected[i]).abs() < tolerance,
                "[{}] expected {}, got {}",
                i,
                expected[i],
                act
            );
        }

        assert!(rsi.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_rsi_update_2() {
        let tolerance = 0.5;
        let input = test_input_2();
        let params = RelativeStrengthIndexParams { length: 14, ..Default::default() };
        let mut rsi = RelativeStrengthIndex::new(&params).unwrap();

        let mut act = f64::NAN;
        for i in 0..input.len() {
            act = rsi.update(input[i]);
            if i < 14 {
                assert!(act.is_nan(), "[{}] expected NaN, got {}", i, act);
            }
        }

        assert!(act >= 0.0 && act <= 100.0, "final RSI should be in [0,100], got {}", act);
    }

    #[test]
    fn test_rsi_is_primed() {
        let params = RelativeStrengthIndexParams { length: 5, ..Default::default() };
        let mut rsi = RelativeStrengthIndex::new(&params).unwrap();

        assert!(!rsi.is_primed());

        for i in 1..=5 {
            rsi.update(i as f64);
            assert!(!rsi.is_primed(), "[{}] should not be primed", i);
        }

        rsi.update(6.0);
        assert!(rsi.is_primed());

        for i in 7..=11 {
            rsi.update(i as f64);
            assert!(rsi.is_primed(), "[{}] should be primed", i);
        }
    }

    #[test]
    fn test_rsi_metadata() {
        let params = RelativeStrengthIndexParams { length: 9, ..Default::default() };
        let rsi = RelativeStrengthIndex::new(&params).unwrap();
        let meta = rsi.metadata();

        assert_eq!(meta.identifier, Identifier::RelativeStrengthIndex);
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].mnemonic, "rsi(9)");
        assert_eq!(meta.outputs[0].description, "Relative Strength Index rsi(9)");
    }

    #[test]
    fn test_rsi_constructor_validation() {
        let params = RelativeStrengthIndexParams { length: 1, ..Default::default() };
        assert!(RelativeStrengthIndex::new(&params).is_err());

        let params = RelativeStrengthIndexParams { length: 14, ..Default::default() };
        let rsi = RelativeStrengthIndex::new(&params).unwrap();
        assert_eq!(rsi.line.mnemonic, "rsi(14)");
    }

    #[test]
    fn test_rsi_bar_component_mnemonic() {
        let params = RelativeStrengthIndexParams {
            length: 14,
            bar_component: Some(BarComponent::Open),
            ..Default::default()
        };
        let rsi = RelativeStrengthIndex::new(&params).unwrap();
        assert_eq!(rsi.line.mnemonic, "rsi(14, o)");
    }
}
