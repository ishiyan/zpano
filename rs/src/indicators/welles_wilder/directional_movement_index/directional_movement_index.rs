use crate::entities::bar::Bar;
use crate::entities::quote::Quote;
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::metadata::Metadata;
use super::super::directional_indicator_plus::{DirectionalIndicatorPlus, DirectionalIndicatorPlusParams};
use super::super::directional_indicator_minus::{DirectionalIndicatorMinus, DirectionalIndicatorMinusParams};

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters for the Directional Movement Index indicator.
pub struct DirectionalMovementIndexParams {
    /// Number of time periods. Must be >= 1. Default is 14.
    pub length: usize,
}

impl Default for DirectionalMovementIndexParams {
    fn default() -> Self {
        Self { length: 14 }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the DX indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum DirectionalMovementIndexOutput {
    /// The scalar value of the directional movement index.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

const DX_MNEMONIC: &str = "dx";
const DX_DESCRIPTION: &str = "Directional Movement Index";
const EPSILON: f64 = 1e-8;

/// Welles Wilder's Directional Movement Index (DX).
///
/// The directional movement index measures the strength of a trend by comparing
/// the positive and negative directional indicators. It is calculated as:
///
///   DX = 100 * |+DI - -DI| / (+DI + -DI)
pub struct DirectionalMovementIndex {
    length: usize,
    value: f64,
    directional_indicator_plus: DirectionalIndicatorPlus,
    directional_indicator_minus: DirectionalIndicatorMinus,
}

impl DirectionalMovementIndex {
    /// Creates a new DirectionalMovementIndex indicator.
    pub fn new(params: &DirectionalMovementIndexParams) -> Result<Self, String> {
        if params.length < 1 {
            return Err(format!("invalid length {}: must be >= 1", params.length));
        }

        let dip = DirectionalIndicatorPlus::new(&DirectionalIndicatorPlusParams { length: params.length })?;
        let dim = DirectionalIndicatorMinus::new(&DirectionalIndicatorMinusParams { length: params.length })?;

        Ok(Self {
            length: params.length,
            value: f64::NAN,
            directional_indicator_plus: dip,
            directional_indicator_minus: dim,
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

        let dip_value = self.directional_indicator_plus.update(close, high, low);
        let dim_value = self.directional_indicator_minus.update(close, high, low);

        if self.directional_indicator_plus.is_primed() && self.directional_indicator_minus.is_primed() {
            let sum = dip_value + dim_value;

            if sum.abs() < EPSILON {
                self.value = 0.0;
            } else {
                self.value = 100.0 * (dip_value - dim_value).abs() / sum;
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

impl Indicator for DirectionalMovementIndex {
    fn is_primed(&self) -> bool {
        self.directional_indicator_plus.is_primed() && self.directional_indicator_minus.is_primed()
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::DirectionalMovementIndex,
            DX_MNEMONIC,
            DX_DESCRIPTION,
            &[
                OutputText {
                    mnemonic: DX_MNEMONIC.to_string(),
                    description: DX_DESCRIPTION.to_string(),
                },
                OutputText {
                    mnemonic: "+di".to_string(),
                    description: "Directional Indicator Plus".to_string(),
                },
                OutputText {
                    mnemonic: "-di".to_string(),
                    description: "Directional Indicator Minus".to_string(),
                },
                OutputText {
                    mnemonic: "+dm".to_string(),
                    description: "Directional Movement Plus".to_string(),
                },
                OutputText {
                    mnemonic: "-dm".to_string(),
                    description: "Directional Movement Minus".to_string(),
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
    fn test_dx_constructor() {
        let dx = DirectionalMovementIndex::new(&DirectionalMovementIndexParams { length: 14 }).unwrap();
        assert_eq!(dx.length(), 14);
        assert!(!dx.is_primed());

        assert!(DirectionalMovementIndex::new(&DirectionalMovementIndexParams { length: 0 }).is_err());
    }

    #[test]
    fn test_dx_is_primed() {
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        let close = testdata::test_input_close();

        let mut dx = DirectionalMovementIndex::new(&DirectionalMovementIndexParams { length: 14 }).unwrap();
        for i in 0..14 {
            dx.update(close[i], high[i], low[i]);
            assert!(!dx.is_primed(), "[{}] should not be primed yet", i);
        }
        dx.update(close[14], high[14], low[14]);
        assert!(dx.is_primed());
    }

    #[test]
    fn test_dx_update() {
        let tolerance = 1e-8;
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        let close = testdata::test_input_close();
        let expected = testdata::test_expected_dx14();
        let mut dx = DirectionalMovementIndex::new(&DirectionalMovementIndexParams { length: 14 }).unwrap();

        for i in 0..high.len() {
            let act = dx.update(close[i], high[i], low[i]);

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
    fn test_dx_nan_passthrough() {
        let mut dx = DirectionalMovementIndex::new(&DirectionalMovementIndexParams { length: 14 }).unwrap();
        assert!(dx.update(f64::NAN, 1.0, 1.0).is_nan());
        assert!(dx.update(1.0, f64::NAN, 1.0).is_nan());
        assert!(dx.update(1.0, 1.0, f64::NAN).is_nan());
        assert!(dx.update_sample(f64::NAN).is_nan());
    }

    #[test]
    fn test_dx_metadata() {
        let dx = DirectionalMovementIndex::new(&DirectionalMovementIndexParams { length: 14 }).unwrap();
        let meta = dx.metadata();
        assert_eq!(meta.identifier, Identifier::DirectionalMovementIndex);
        assert_eq!(meta.mnemonic, "dx");
        assert_eq!(meta.description, "Directional Movement Index");
        assert_eq!(meta.outputs.len(), 7);
        assert_eq!(meta.outputs[0].mnemonic, "dx");
        assert_eq!(meta.outputs[0].description, "Directional Movement Index");
    }
}
