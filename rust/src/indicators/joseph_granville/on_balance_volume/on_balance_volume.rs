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

/// Parameters to create an instance of the on-balance volume indicator.
pub struct OnBalanceVolumeParams {
    /// Bar component to extract for price comparison. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for OnBalanceVolumeParams {
    fn default() -> Self {
        Self {
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the on-balance volume indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum OnBalanceVolumeOutput {
    /// The scalar value of the on-balance volume.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Joseph Granville's On-Balance Volume (OBV).
///
/// OBV is a cumulative volume indicator. On each update, if the price is higher
/// than the previous price, the volume is added to the running total; if the price
/// is lower, the volume is subtracted. If the price is unchanged, the total remains
/// the same.
pub struct OnBalanceVolume {
    line: LineIndicator,
    previous_sample: f64,
    value: f64,
    primed: bool,
}

impl OnBalanceVolume {
    /// Creates a new OnBalanceVolume from the given parameters.
    pub fn new(params: &OnBalanceVolumeParams) -> Result<Self, String> {
        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let suffix = component_triple_mnemonic(bc, qc, tc);
        let mnemonic = if suffix.is_empty() {
            "obv".to_string()
        } else {
            format!("obv({})", &suffix[2..]) // strip leading ", "
        };
        let description = format!("On-Balance Volume OBV");

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        Ok(Self {
            line,
            previous_sample: 0.0,
            value: f64::NAN,
            primed: false,
        })
    }

    /// Core update logic with volume = 1 (scalar path).
    pub fn update(&mut self, sample: f64) -> f64 {
        self.update_with_volume(sample, 1.0)
    }

    /// Updates the indicator with the given sample and volume.
    pub fn update_with_volume(&mut self, sample: f64, volume: f64) -> f64 {
        if sample.is_nan() || volume.is_nan() {
            return f64::NAN;
        }

        if !self.primed {
            self.value = volume;
            self.primed = true;
        } else if sample > self.previous_sample {
            self.value += volume;
        } else if sample < self.previous_sample {
            self.value -= volume;
        }

        self.previous_sample = sample;
        self.value
    }
}

impl Indicator for OnBalanceVolume {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::OnBalanceVolume,
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
        let price = (self.line.bar_func)(sample);
        let value = self.update_with_volume(price, sample.volume);
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
    use crate::indicators::core::outputs::shape::Shape;

    fn test_prices() -> Vec<f64> {
        vec![1.0, 2.0, 8.0, 4.0, 9.0, 6.0, 7.0, 13.0, 9.0, 10.0, 3.0, 12.0]
    }

    fn test_volumes() -> Vec<f64> {
        vec![100.0, 90.0, 200.0, 150.0, 500.0, 100.0, 300.0, 150.0, 100.0, 300.0, 200.0, 100.0]
    }

    fn test_expected() -> Vec<f64> {
        vec![100.0, 190.0, 390.0, 240.0, 740.0, 640.0, 940.0, 1090.0, 990.0, 1290.0, 1090.0, 1190.0]
    }

    #[test]
    fn test_with_volume() {
        let prices = test_prices();
        let vol = test_volumes();
        let expected = test_expected();

        let mut obv = OnBalanceVolume::new(&OnBalanceVolumeParams::default()).unwrap();

        for i in 0..prices.len() {
            let v = obv.update_with_volume(prices[i], vol[i]);
            assert!(!v.is_nan(), "[{}] expected non-NaN", i);
            assert!(obv.is_primed(), "[{}] expected primed", i);
            assert!((v - expected[i]).abs() < 1e-10, "[{}] expected {}, got {}", i, expected[i], v);
        }
    }

    #[test]
    fn test_is_primed() {
        let mut obv = OnBalanceVolume::new(&OnBalanceVolumeParams::default()).unwrap();
        assert!(!obv.is_primed());

        obv.update_with_volume(1.0, 100.0);
        assert!(obv.is_primed());

        obv.update_with_volume(2.0, 50.0);
        assert!(obv.is_primed());
    }

    #[test]
    fn test_nan() {
        let mut obv = OnBalanceVolume::new(&OnBalanceVolumeParams::default()).unwrap();

        assert!(obv.update(f64::NAN).is_nan());
        assert!(obv.update_with_volume(1.0, f64::NAN).is_nan());
        assert!(obv.update_with_volume(f64::NAN, f64::NAN).is_nan());
    }

    #[test]
    fn test_equal_prices() {
        let mut obv = OnBalanceVolume::new(&OnBalanceVolumeParams::default()).unwrap();

        let v = obv.update_with_volume(10.0, 100.0);
        assert!((v - 100.0).abs() < 1e-10);

        // Same price: value should not change.
        let v = obv.update_with_volume(10.0, 200.0);
        assert!((v - 100.0).abs() < 1e-10);
    }

    #[test]
    fn test_metadata() {
        let obv = OnBalanceVolume::new(&OnBalanceVolumeParams::default()).unwrap();
        let meta = obv.metadata();

        assert_eq!(meta.identifier, Identifier::OnBalanceVolume);
        assert_eq!(meta.mnemonic, "obv");
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].kind, OnBalanceVolumeOutput::Value as i32);
        assert_eq!(meta.outputs[0].shape, Shape::Scalar);
    }

    #[test]
    fn test_update_scalar() {
        let mut obv = OnBalanceVolume::new(&OnBalanceVolumeParams::default()).unwrap();

        let scalar = Scalar::new(0, 10.0);
        let out = obv.update_scalar(&scalar);
        let sv = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((sv.value - 1.0).abs() < 1e-10, "expected 1.0 (volume=1 on first call), got {}", sv.value);
    }

    #[test]
    fn test_update_bar() {
        let prices = test_prices();
        let vol = test_volumes();
        let expected = test_expected();

        let mut obv = OnBalanceVolume::new(&OnBalanceVolumeParams::default()).unwrap();

        for i in 0..prices.len() {
            let bar = Bar {
                time: 0,
                open: 0.0,
                high: 0.0,
                low: 0.0,
                close: prices[i],
                volume: vol[i],
            };
            let out = obv.update_bar(&bar);
            let sv = out[0].downcast_ref::<Scalar>().unwrap();
            assert!((sv.value - expected[i]).abs() < 1e-10, "[{}] expected {}, got {}", i, expected[i], sv.value);
        }
    }
}
