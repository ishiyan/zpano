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

/// Parameters for the Chande Momentum Oscillator indicator.
pub struct ChandeMomentumOscillatorParams {
    /// Number of periods. Must be >= 1. Default is 14.
    pub length: usize,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for ChandeMomentumOscillatorParams {
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

/// Enumerates the outputs of the CMO indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum ChandeMomentumOscillatorOutput {
    /// The scalar value of the CMO.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

const EPSILON: f64 = 1e-12;

/// Tushar Chande's Momentum Oscillator (CMO).
///
/// CMO = 100 * (SU - SD) / (SU + SD), where SU is sum of gains and SD is
/// sum of losses over the lookback period.
pub struct ChandeMomentumOscillator {
    line: LineIndicator,
    length: usize,
    count: usize,
    ring_buffer: Vec<f64>,
    ring_head: usize,
    previous_sample: f64,
    gain_sum: f64,
    loss_sum: f64,
    primed: bool,
}

impl ChandeMomentumOscillator {
    /// Creates a new ChandeMomentumOscillator from the given parameters.
    pub fn new(params: &ChandeMomentumOscillatorParams) -> Result<Self, String> {
        if params.length < 1 {
            return Err("invalid Chande momentum oscillator parameters: length should be positive".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!("cmo({}{})", params.length, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Chande Momentum Oscillator {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        Ok(Self {
            line,
            length: params.length,
            count: 0,
            ring_buffer: vec![0.0; params.length],
            ring_head: 0,
            previous_sample: 0.0,
            gain_sum: 0.0,
            loss_sum: 0.0,
            primed: false,
        })
    }

    /// Core update with a single sample value.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        self.count += 1;
        if self.count == 1 {
            self.previous_sample = sample;
            return f64::NAN;
        }

        let delta = sample - self.previous_sample;
        self.previous_sample = sample;

        if !self.primed {
            self.ring_buffer[self.ring_head] = delta;
            self.ring_head = (self.ring_head + 1) % self.length;

            if delta > 0.0 {
                self.gain_sum += delta;
            } else if delta < 0.0 {
                self.loss_sum += -delta;
            }

            if self.count <= self.length {
                return f64::NAN;
            }

            self.primed = true;
        } else {
            let old = self.ring_buffer[self.ring_head];
            if old > 0.0 {
                self.gain_sum -= old;
            } else if old < 0.0 {
                self.loss_sum -= -old;
            }

            self.ring_buffer[self.ring_head] = delta;
            self.ring_head = (self.ring_head + 1) % self.length;

            if delta > 0.0 {
                self.gain_sum += delta;
            } else if delta < 0.0 {
                self.loss_sum += -delta;
            }

            if self.gain_sum < 0.0 {
                self.gain_sum = 0.0;
            }
            if self.loss_sum < 0.0 {
                self.loss_sum = 0.0;
            }
        }

        let den = self.gain_sum + self.loss_sum;
        if den.abs() < EPSILON {
            return 0.0;
        }

        100.0 * (self.gain_sum - self.loss_sum) / den
    }
}

impl Indicator for ChandeMomentumOscillator {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::ChandeMomentumOscillator,
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

    fn test_book_input() -> Vec<f64> {
        vec![
            101.0313, 101.0313, 101.1250, 101.9687, 102.7813,
            103.0000, 102.9687, 103.0625, 102.9375, 102.7188,
            102.7500, 102.9063, 102.9687,
        ]
    }

    fn test_book_expected() -> Vec<f64> {
        vec![
            f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN,
            f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN,
            69.61963786608334, 71.42857142857143, 71.08377992828775,
        ]
    }

    fn create_cmo(length: usize) -> ChandeMomentumOscillator {
        ChandeMomentumOscillator::new(&ChandeMomentumOscillatorParams {
            length,
            ..Default::default()
        }).unwrap()
    }

    #[test]
    fn test_cmo_update_book_length10() {
        let input = test_book_input();
        let expected = test_book_expected();
        let mut cmo = create_cmo(10);

        for i in 0..10 {
            let act = cmo.update(input[i]);
            assert!(act.is_nan(), "[{}] expected NaN, got {}", i, act);
        }

        for i in 10..input.len() {
            let act = cmo.update(input[i]);
            assert!(
                (act - expected[i]).abs() < 1e-13,
                "[{}] expected {}, got {}",
                i, expected[i], act
            );
        }

        assert!(cmo.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_cmo_is_primed() {
        let input = vec![
            91.5, 94.815, 94.375, 95.095, 93.78, 94.625, 92.53, 92.75, 90.315, 92.47,
            96.125,
        ];

        let mut cmo = create_cmo(10);
        assert!(!cmo.is_primed());

        for i in 0..10 {
            cmo.update(input[i]);
            assert!(!cmo.is_primed(), "[{}] expected not primed", i);
        }

        cmo.update(input[10]);
        assert!(cmo.is_primed());
    }

    #[test]
    fn test_cmo_metadata() {
        let cmo = create_cmo(5);
        let meta = cmo.metadata();

        assert_eq!(meta.identifier, Identifier::ChandeMomentumOscillator);
        assert_eq!(meta.mnemonic, "cmo(5)");
        assert_eq!(meta.description, "Chande Momentum Oscillator cmo(5)");
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].kind, ChandeMomentumOscillatorOutput::Value as i32);
        assert_eq!(meta.outputs[0].mnemonic, "cmo(5)");
    }

    #[test]
    fn test_cmo_invalid_params() {
        assert!(ChandeMomentumOscillator::new(&ChandeMomentumOscillatorParams {
            length: 0,
            ..Default::default()
        }).is_err());
    }

    #[test]
    fn test_cmo_bar_component_mnemonic() {
        let cmo = ChandeMomentumOscillator::new(&ChandeMomentumOscillatorParams {
            length: 5,
            bar_component: Some(BarComponent::Median),
            ..Default::default()
        }).unwrap();
        assert_eq!(cmo.line.mnemonic, "cmo(5, hl/2)");
    }
}
