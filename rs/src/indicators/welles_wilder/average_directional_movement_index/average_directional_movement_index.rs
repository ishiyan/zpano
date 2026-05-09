use crate::entities::bar::Bar;
use crate::entities::quote::Quote;
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::metadata::Metadata;
use super::super::directional_movement_index::{DirectionalMovementIndex, DirectionalMovementIndexParams};

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters for the Average Directional Movement Index indicator.
pub struct AverageDirectionalMovementIndexParams {
    /// Number of time periods. Must be >= 1. Default is 14.
    pub length: usize,
}

impl Default for AverageDirectionalMovementIndexParams {
    fn default() -> Self {
        Self { length: 14 }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the ADX indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum AverageDirectionalMovementIndexOutput {
    /// The scalar value of the average directional movement index.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

const ADX_MNEMONIC: &str = "adx";
const ADX_DESCRIPTION: &str = "Average Directional Movement Index";

/// Welles Wilder's Average Directional Movement Index (ADX).
///
/// The average directional movement index smooths the directional movement index (DX)
/// using Wilder's smoothing technique. It is calculated as:
///
///   Initial ADX = SMA of first `length` DX values
///   Subsequent ADX = (previousADX * (length-1) + DX) / length
pub struct AverageDirectionalMovementIndex {
    length: usize,
    length_minus_one: f64,
    count: usize,
    sum: f64,
    primed: bool,
    value: f64,
    directional_movement_index: DirectionalMovementIndex,
}

impl AverageDirectionalMovementIndex {
    /// Creates a new AverageDirectionalMovementIndex indicator.
    pub fn new(params: &AverageDirectionalMovementIndexParams) -> Result<Self, String> {
        if params.length < 1 {
            return Err(format!("invalid length {}: must be >= 1", params.length));
        }

        let dx = DirectionalMovementIndex::new(&DirectionalMovementIndexParams { length: params.length })?;

        Ok(Self {
            length: params.length,
            length_minus_one: (params.length as f64) - 1.0,
            count: 0,
            sum: 0.0,
            primed: false,
            value: f64::NAN,
            directional_movement_index: dx,
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

        let dx_value = self.directional_movement_index.update(close, high, low);

        if !self.directional_movement_index.is_primed() {
            return f64::NAN;
        }

        if self.primed {
            self.value = (self.value * self.length_minus_one + dx_value) / self.length as f64;
            return self.value;
        }

        self.count += 1;
        self.sum += dx_value;

        if self.count == self.length {
            self.value = self.sum / self.length as f64;
            self.primed = true;
            return self.value;
        }

        f64::NAN
    }

    /// Updates using a single sample value as substitute for close, high, and low.
    pub fn update_sample(&mut self, sample: f64) -> f64 {
        self.update(sample, sample, sample)
    }
}

impl Indicator for AverageDirectionalMovementIndex {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::AverageDirectionalMovementIndex,
            ADX_MNEMONIC,
            ADX_DESCRIPTION,
            &[
                OutputText {
                    mnemonic: ADX_MNEMONIC.to_string(),
                    description: ADX_DESCRIPTION.to_string(),
                },
                OutputText {
                    mnemonic: "dx".to_string(),
                    description: "Directional Movement Index".to_string(),
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
    fn test_adx_constructor() {
        let adx = AverageDirectionalMovementIndex::new(&AverageDirectionalMovementIndexParams { length: 14 }).unwrap();
        assert_eq!(adx.length(), 14);
        assert!(!adx.is_primed());

        assert!(AverageDirectionalMovementIndex::new(&AverageDirectionalMovementIndexParams { length: 0 }).is_err());
    }

    #[test]
    fn test_adx_is_primed() {
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        let close = testdata::test_input_close();

        let mut adx = AverageDirectionalMovementIndex::new(&AverageDirectionalMovementIndexParams { length: 14 }).unwrap();
        for i in 0..27 {
            adx.update(close[i], high[i], low[i]);
            assert!(!adx.is_primed(), "[{}] should not be primed yet", i);
        }
        adx.update(close[27], high[27], low[27]);
        assert!(adx.is_primed());
    }

    #[test]
    fn test_adx_update() {
        let tolerance = 1e-8;
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        let close = testdata::test_input_close();
        let expected = testdata::test_expected_adx14();
        let mut adx = AverageDirectionalMovementIndex::new(&AverageDirectionalMovementIndexParams { length: 14 }).unwrap();

        for i in 0..high.len() {
            let act = adx.update(close[i], high[i], low[i]);

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
    fn test_adx_nan_passthrough() {
        let mut adx = AverageDirectionalMovementIndex::new(&AverageDirectionalMovementIndexParams { length: 14 }).unwrap();
        assert!(adx.update(f64::NAN, 1.0, 1.0).is_nan());
        assert!(adx.update(1.0, f64::NAN, 1.0).is_nan());
        assert!(adx.update(1.0, 1.0, f64::NAN).is_nan());
        assert!(adx.update_sample(f64::NAN).is_nan());
    }

    #[test]
    fn test_adx_metadata() {
        let adx = AverageDirectionalMovementIndex::new(&AverageDirectionalMovementIndexParams { length: 14 }).unwrap();
        let meta = adx.metadata();
        assert_eq!(meta.identifier, Identifier::AverageDirectionalMovementIndex);
        assert_eq!(meta.mnemonic, "adx");
        assert_eq!(meta.description, "Average Directional Movement Index");
        assert_eq!(meta.outputs.len(), 8);
        assert_eq!(meta.outputs[0].mnemonic, "adx");
        assert_eq!(meta.outputs[0].description, "Average Directional Movement Index");
    }
}
