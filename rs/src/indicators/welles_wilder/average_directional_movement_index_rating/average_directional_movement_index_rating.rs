use crate::entities::bar::Bar;
use crate::entities::quote::Quote;
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::metadata::Metadata;
use super::super::average_directional_movement_index::{AverageDirectionalMovementIndex, AverageDirectionalMovementIndexParams};

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters for the Average Directional Movement Index Rating indicator.
pub struct AverageDirectionalMovementIndexRatingParams {
    /// Number of time periods. Must be >= 1. Default is 14.
    pub length: usize,
}

impl Default for AverageDirectionalMovementIndexRatingParams {
    fn default() -> Self {
        Self { length: 14 }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the ADXR indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum AverageDirectionalMovementIndexRatingOutput {
    /// The scalar value of the average directional movement index rating.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

const ADXR_MNEMONIC: &str = "adxr";
const ADXR_DESCRIPTION: &str = "Average Directional Movement Index Rating";

/// Welles Wilder's Average Directional Movement Index Rating (ADXR).
///
/// The average directional movement index rating averages the current ADX value with
/// the ADX value from (length - 1) periods ago. It is calculated as:
///
///   ADXR = (ADX[current] + ADX[current - (length - 1)]) / 2
pub struct AverageDirectionalMovementIndexRating {
    length: usize,
    buffer_size: usize,
    buffer: Vec<f64>,
    buffer_index: usize,
    buffer_count: usize,
    primed: bool,
    value: f64,
    average_directional_movement_index: AverageDirectionalMovementIndex,
}

impl AverageDirectionalMovementIndexRating {
    /// Creates a new AverageDirectionalMovementIndexRating indicator.
    pub fn new(params: &AverageDirectionalMovementIndexRatingParams) -> Result<Self, String> {
        if params.length < 1 {
            return Err(format!("invalid length {}: must be >= 1", params.length));
        }

        let adx = AverageDirectionalMovementIndex::new(&AverageDirectionalMovementIndexParams { length: params.length })?;

        let buffer_size = params.length;

        Ok(Self {
            length: params.length,
            buffer_size,
            buffer: vec![0.0; buffer_size],
            buffer_index: 0,
            buffer_count: 0,
            primed: false,
            value: f64::NAN,
            average_directional_movement_index: adx,
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

        let adx_value = self.average_directional_movement_index.update(close, high, low);

        if !self.average_directional_movement_index.is_primed() {
            return f64::NAN;
        }

        // Store ADX value in circular buffer.
        self.buffer[self.buffer_index] = adx_value;
        self.buffer_index = (self.buffer_index + 1) % self.buffer_size;
        self.buffer_count += 1;

        if self.buffer_count < self.buffer_size {
            return f64::NAN;
        }

        // The oldest value in the buffer is at buffer_index (since we just advanced it).
        let old_adx = self.buffer[self.buffer_index % self.buffer_size];
        self.value = (adx_value + old_adx) / 2.0;
        self.primed = true;

        self.value
    }

    /// Updates using a single sample value as substitute for close, high, and low.
    pub fn update_sample(&mut self, sample: f64) -> f64 {
        self.update(sample, sample, sample)
    }
}

impl Indicator for AverageDirectionalMovementIndexRating {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::AverageDirectionalMovementIndexRating,
            ADXR_MNEMONIC,
            ADXR_DESCRIPTION,
            &[
                OutputText {
                    mnemonic: ADXR_MNEMONIC.to_string(),
                    description: ADXR_DESCRIPTION.to_string(),
                },
                OutputText {
                    mnemonic: "adx".to_string(),
                    description: "Average Directional Movement Index".to_string(),
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
    fn test_adxr_constructor() {
        let adxr = AverageDirectionalMovementIndexRating::new(&AverageDirectionalMovementIndexRatingParams { length: 14 }).unwrap();
        assert_eq!(adxr.length(), 14);
        assert!(!adxr.is_primed());

        assert!(AverageDirectionalMovementIndexRating::new(&AverageDirectionalMovementIndexRatingParams { length: 0 }).is_err());
    }

    #[test]
    fn test_adxr_is_primed() {
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        let close = testdata::test_input_close();

        let mut adxr = AverageDirectionalMovementIndexRating::new(&AverageDirectionalMovementIndexRatingParams { length: 14 }).unwrap();
        for i in 0..40 {
            adxr.update(close[i], high[i], low[i]);
            assert!(!adxr.is_primed(), "[{}] should not be primed yet", i);
        }
        adxr.update(close[40], high[40], low[40]);
        assert!(adxr.is_primed());
    }

    #[test]
    fn test_adxr_update() {
        let tolerance = 1e-8;
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        let close = testdata::test_input_close();
        let expected = testdata::test_expected_adxr14();
        let mut adxr = AverageDirectionalMovementIndexRating::new(&AverageDirectionalMovementIndexRatingParams { length: 14 }).unwrap();

        for i in 0..high.len() {
            let act = adxr.update(close[i], high[i], low[i]);

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
    fn test_adxr_nan_passthrough() {
        let mut adxr = AverageDirectionalMovementIndexRating::new(&AverageDirectionalMovementIndexRatingParams { length: 14 }).unwrap();
        assert!(adxr.update(f64::NAN, 1.0, 1.0).is_nan());
        assert!(adxr.update(1.0, f64::NAN, 1.0).is_nan());
        assert!(adxr.update(1.0, 1.0, f64::NAN).is_nan());
        assert!(adxr.update_sample(f64::NAN).is_nan());
    }

    #[test]
    fn test_adxr_metadata() {
        let adxr = AverageDirectionalMovementIndexRating::new(&AverageDirectionalMovementIndexRatingParams { length: 14 }).unwrap();
        let meta = adxr.metadata();
        assert_eq!(meta.identifier, Identifier::AverageDirectionalMovementIndexRating);
        assert_eq!(meta.mnemonic, "adxr");
        assert_eq!(meta.description, "Average Directional Movement Index Rating");
        assert_eq!(meta.outputs.len(), 9);
        assert_eq!(meta.outputs[0].mnemonic, "adxr");
        assert_eq!(meta.outputs[0].description, "Average Directional Movement Index Rating");
    }
}
