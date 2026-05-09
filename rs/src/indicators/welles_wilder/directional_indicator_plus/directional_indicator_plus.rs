use crate::entities::bar::Bar;
use crate::entities::quote::Quote;
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::metadata::Metadata;
use super::super::average_true_range::{AverageTrueRange, AverageTrueRangeParams};
use super::super::directional_movement_plus::{DirectionalMovementPlus, DirectionalMovementPlusParams};

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters for the Directional Indicator Plus indicator.
pub struct DirectionalIndicatorPlusParams {
    /// Number of time periods. Must be >= 1. Default is 14.
    pub length: usize,
}

impl Default for DirectionalIndicatorPlusParams {
    fn default() -> Self {
        Self { length: 14 }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the +DI indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum DirectionalIndicatorPlusOutput {
    /// The scalar value of the directional indicator plus.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

const DIP_MNEMONIC: &str = "+di";
const DIP_DESCRIPTION: &str = "Directional Indicator Plus";
const EPSILON: f64 = 1e-8;

/// Welles Wilder's Directional Indicator Plus (+DI).
///
/// The directional indicator plus measures the percentage of the average true range
/// that is attributable to upward movement. It is calculated as:
///
///   +DI = 100 * +DM(n) / (ATR * length)
///
/// where +DM(n) is the Wilder-smoothed directional movement plus and ATR is the
/// average true range over the same length.
pub struct DirectionalIndicatorPlus {
    length: usize,
    value: f64,
    average_true_range: AverageTrueRange,
    directional_movement_plus: DirectionalMovementPlus,
}

impl DirectionalIndicatorPlus {
    /// Creates a new DirectionalIndicatorPlus indicator.
    pub fn new(params: &DirectionalIndicatorPlusParams) -> Result<Self, String> {
        if params.length < 1 {
            return Err(format!("invalid length {}: must be >= 1", params.length));
        }

        let atr = AverageTrueRange::new(&AverageTrueRangeParams { length: params.length })?;
        let dmp = DirectionalMovementPlus::new(&DirectionalMovementPlusParams { length: params.length })?;

        Ok(Self {
            length: params.length,
            value: f64::NAN,
            average_true_range: atr,
            directional_movement_plus: dmp,
        })
    }

    /// Returns the length parameter.
    pub fn length(&self) -> usize {
        self.length
    }

    /// Core update with close, high, and low values.
    pub fn update(&mut self, close: f64, high: f64, low: f64) -> f64 {
        if close.is_nan() || high.is_nan() || low.is_nan() {
            return f64::NAN;
        }

        let atr_value = self.average_true_range.update(close, high, low);
        let dmp_value = self.directional_movement_plus.update(high, low);

        if self.average_true_range.is_primed() && self.directional_movement_plus.is_primed() {
            let atr_scaled = atr_value * self.length as f64;

            if atr_scaled.abs() < EPSILON {
                self.value = 0.0;
            } else {
                self.value = 100.0 * dmp_value / atr_scaled;
            }

            return self.value;
        }

        f64::NAN
    }

    /// Updates using a single sample value as substitute for close, high, and low.
    pub fn update_sample(&mut self, sample: f64) -> f64 {
        self.update(sample, sample, sample)
    }
}

impl Indicator for DirectionalIndicatorPlus {
    fn is_primed(&self) -> bool {
        self.average_true_range.is_primed() && self.directional_movement_plus.is_primed()
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::DirectionalIndicatorPlus,
            DIP_MNEMONIC,
            DIP_DESCRIPTION,
            &[
                OutputText {
                    mnemonic: DIP_MNEMONIC.to_string(),
                    description: DIP_DESCRIPTION.to_string(),
                },
                OutputText {
                    mnemonic: "+dm".to_string(),
                    description: "Directional Movement Plus".to_string(),
                },
                OutputText {
                    mnemonic: "atr".to_string(),
                    description: "Average True Range".to_string(),
                },
                OutputText {
                    mnemonic: "tr".to_string(),
                    description: "True Range".to_string(),
                },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let v = sample.value;
        let result = self.update(v, v, v);
        vec![Box::new(Scalar::new(sample.time, result))]
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let result = self.update(sample.close, sample.high, sample.low);
        vec![Box::new(Scalar::new(sample.time, result))]
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (sample.bid_price + sample.ask_price) / 2.0;
        let result = self.update(v, v, v);
        vec![Box::new(Scalar::new(sample.time, result))]
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let v = sample.price;
        let result = self.update(v, v, v);
        vec![Box::new(Scalar::new(sample.time, result))]
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::testdata;
    #[test]
    fn test_dip_constructor() {
        let dip = DirectionalIndicatorPlus::new(&DirectionalIndicatorPlusParams { length: 14 }).unwrap();
        assert_eq!(dip.length(), 14);
        assert!(!dip.is_primed());

        assert!(DirectionalIndicatorPlus::new(&DirectionalIndicatorPlusParams { length: 0 }).is_err());
    }

    #[test]
    fn test_dip_is_primed() {
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        let close = testdata::test_input_close();

        let mut dip = DirectionalIndicatorPlus::new(&DirectionalIndicatorPlusParams { length: 14 }).unwrap();
        for i in 0..14 {
            dip.update(close[i], high[i], low[i]);
            assert!(!dip.is_primed(), "[{}] should not be primed yet", i);
        }
        dip.update(close[14], high[14], low[14]);
        assert!(dip.is_primed());
    }

    #[test]
    fn test_dip_update() {
        let tolerance = 1e-8;
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        let close = testdata::test_input_close();
        let expected = testdata::test_expected_di14();
        let mut dip = DirectionalIndicatorPlus::new(&DirectionalIndicatorPlusParams { length: 14 }).unwrap();

        for i in 0..high.len() {
            let act = dip.update(close[i], high[i], low[i]);

            if expected[i].is_nan() {
                assert!(act.is_nan(), "[{}] expected NaN, got {}", i, act);
                continue;
            }

            assert!(!act.is_nan(), "[{}] expected {}, got NaN", i, expected[i]);
            assert!(
                (act - expected[i]).abs() < tolerance,
                "[{}] expected {}, got {}",
                i, expected[i], act
            );
        }
    }

    #[test]
    fn test_dip_nan_passthrough() {
        let mut dip = DirectionalIndicatorPlus::new(&DirectionalIndicatorPlusParams { length: 14 }).unwrap();
        assert!(dip.update(f64::NAN, 1.0, 1.0).is_nan());
        assert!(dip.update(1.0, f64::NAN, 1.0).is_nan());
        assert!(dip.update(1.0, 1.0, f64::NAN).is_nan());
        assert!(dip.update_sample(f64::NAN).is_nan());
    }

    #[test]
    fn test_dip_metadata() {
        let dip = DirectionalIndicatorPlus::new(&DirectionalIndicatorPlusParams { length: 14 }).unwrap();
        let meta = dip.metadata();
        assert_eq!(meta.identifier, Identifier::DirectionalIndicatorPlus);
        assert_eq!(meta.mnemonic, "+di");
        assert_eq!(meta.description, "Directional Indicator Plus");
        assert_eq!(meta.outputs.len(), 4);
        assert_eq!(meta.outputs[0].mnemonic, "+di");
        assert_eq!(meta.outputs[0].description, "Directional Indicator Plus");
    }
}
