use crate::entities::bar::Bar;
use crate::entities::quote::Quote;
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::metadata::Metadata;

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters for the Williams %R indicator.
pub struct WilliamsPercentRParams {
    /// The number of time periods. Typical values are 5, 9, or 14. Default is 14. Must be >= 2.
    pub length: usize,
}

impl Default for WilliamsPercentRParams {
    fn default() -> Self {
        Self { length: 14 }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Williams %R indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum WilliamsPercentROutput {
    /// The scalar value of Williams %R.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

const WILLR_MNEMONIC: &str = "willr";
const WILLR_DESCRIPTION: &str = "Williams %R";
const MIN_LENGTH: usize = 2;
const DEFAULT_LENGTH: usize = 14;

/// Larry Williams' Williams %R momentum indicator.
///
/// Williams %R reflects the level of the closing price relative to the
/// highest high over a lookback period. The oscillation ranges from 0 to -100.
///
/// The value is calculated as:
///
///   %R = -100 * (HighestHigh - Close) / (HighestHigh - LowestLow)
///
/// where HighestHigh and LowestLow are computed over the last `length` bars.
pub struct WilliamsPercentR {
    length: usize,
    length_min_one: usize,
    circular_index: usize,
    circular_count: usize,
    low_circular: Vec<f64>,
    high_circular: Vec<f64>,
    value: f64,
    primed: bool,
}

impl WilliamsPercentR {
    /// Creates a new Williams %R indicator.
    pub fn new(params: &WilliamsPercentRParams) -> Result<Self, String> {
        let length = if params.length < MIN_LENGTH {
            DEFAULT_LENGTH
        } else {
            params.length
        };

        Ok(Self {
            length,
            length_min_one: length - 1,
            circular_index: 0,
            circular_count: 0,
            low_circular: vec![0.0; length],
            high_circular: vec![0.0; length],
            value: f64::NAN,
            primed: false,
        })
    }

    /// Updates the Williams %R given the next bar's close, high, and low values.
    pub fn update(&mut self, close: f64, high: f64, low: f64) -> f64 {
        if close.is_nan() || high.is_nan() || low.is_nan() {
            return f64::NAN;
        }

        let index = self.circular_index;
        self.low_circular[index] = low;
        self.high_circular[index] = high;

        // Advance circular buffer index.
        self.circular_index += 1;
        if self.circular_index > self.length_min_one {
            self.circular_index = 0;
        }

        if self.length > self.circular_count {
            if self.length_min_one == self.circular_count {
                // We have exactly `length` samples; compute for the first time.
                let mut min_low = self.low_circular[index];
                let mut max_high = self.high_circular[index];
                let mut idx = index;

                for _ in 0..self.length_min_one {
                    idx -= 1; // Safe: index started at length_min_one.

                    let temp = self.low_circular[idx];
                    if min_low > temp {
                        min_low = temp;
                    }

                    let temp = self.high_circular[idx];
                    if max_high < temp {
                        max_high = temp;
                    }
                }

                if (max_high - min_low).abs() < f64::MIN_POSITIVE {
                    self.value = 0.0;
                } else {
                    self.value = -100.0 * (max_high - close) / (max_high - min_low);
                }

                self.primed = true;
            }

            self.circular_count += 1;

            return self.value;
        }

        // Already primed, compute normally with wrapping.
        let mut min_low = self.low_circular[index];
        let mut max_high = self.high_circular[index];
        let mut idx = index;

        for _ in 0..self.length_min_one {
            if idx == 0 {
                idx = self.length_min_one;
            } else {
                idx -= 1;
            }

            let temp = self.low_circular[idx];
            if min_low > temp {
                min_low = temp;
            }

            let temp = self.high_circular[idx];
            if max_high < temp {
                max_high = temp;
            }
        }

        if (max_high - min_low).abs() < f64::MIN_POSITIVE {
            self.value = 0.0;
        } else {
            self.value = -100.0 * (max_high - close) / (max_high - min_low);
        }

        self.value
    }

    /// Updates using a single sample value as substitute for high, low, and close.
    pub fn update_sample(&mut self, sample: f64) -> f64 {
        self.update(sample, sample, sample)
    }
}

impl Indicator for WilliamsPercentR {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::WilliamsPercentR,
            WILLR_MNEMONIC,
            WILLR_DESCRIPTION,
            &[OutputText {
                mnemonic: WILLR_MNEMONIC.to_string(),
                description: WILLR_DESCRIPTION.to_string(),
            }],
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
    fn test_williams_percent_r_update_14() {
        let tolerance = 1e-6;
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        let close = testdata::test_input_close();
        let expected = testdata::test_expected_14();
        let mut w = WilliamsPercentR::new(&WilliamsPercentRParams { length: 14 }).unwrap();

        for i in 0..close.len() {
            let act = w.update(close[i], high[i], low[i]);

            if i < 13 {
                assert!(act.is_nan(), "[{}] expected NaN, got {}", i, act);
                continue;
            }

            assert!(!act.is_nan(), "[{}] expected {}, got NaN", i, expected[i]);
            assert!(
                (act - expected[i]).abs() < tolerance,
                "[{}] expected {}, got {}",
                i,
                expected[i],
                act
            );
        }
    }

    #[test]
    fn test_williams_percent_r_update_2() {
        let tolerance = 1e-6;
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        let close = testdata::test_input_close();
        let expected = testdata::test_expected_2();
        let mut w = WilliamsPercentR::new(&WilliamsPercentRParams { length: 2 }).unwrap();

        for i in 0..close.len() {
            let act = w.update(close[i], high[i], low[i]);

            if i < 1 {
                assert!(act.is_nan(), "[{}] expected NaN, got {}", i, act);
                continue;
            }

            assert!(!act.is_nan(), "[{}] expected {}, got NaN", i, expected[i]);
            assert!(
                (act - expected[i]).abs() < tolerance,
                "[{}] expected {}, got {}",
                i,
                expected[i],
                act
            );
        }
    }

    #[test]
    fn test_williams_percent_r_nan_passthrough() {
        let mut w = WilliamsPercentR::new(&WilliamsPercentRParams { length: 14 }).unwrap();
        assert!(w.update(f64::NAN, 1.0, 1.0).is_nan());
        assert!(w.update(1.0, f64::NAN, 1.0).is_nan());
        assert!(w.update(1.0, 1.0, f64::NAN).is_nan());
    }

    #[test]
    fn test_williams_percent_r_is_primed() {
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        let close = testdata::test_input_close();
        let mut w = WilliamsPercentR::new(&WilliamsPercentRParams { length: 14 }).unwrap();

        assert!(!w.is_primed());

        for i in 0..13 {
            w.update(close[i], high[i], low[i]);
            assert!(!w.is_primed(), "[{}] should not be primed yet", i);
        }

        w.update(close[13], high[13], low[13]);
        assert!(w.is_primed(), "[13] should be primed after 14th update");

        w.update(close[14], high[14], low[14]);
        assert!(w.is_primed(), "[14] should remain primed");
    }

    #[test]
    fn test_williams_percent_r_update_sample() {
        let mut w = WilliamsPercentR::new(&WilliamsPercentRParams { length: 14 }).unwrap();

        for i in 0..13 {
            let v = w.update_sample(9.0);
            assert!(v.is_nan(), "[{}] expected NaN, got {}", i, v);
        }

        let v = w.update_sample(9.0);
        assert_eq!(v, 0.0, "expected 0, got {}", v);
    }

    #[test]
    fn test_williams_percent_r_metadata() {
        let w = WilliamsPercentR::new(&WilliamsPercentRParams { length: 14 }).unwrap();
        let meta = w.metadata();
        assert_eq!(meta.identifier, Identifier::WilliamsPercentR);
        assert_eq!(meta.mnemonic, "willr");
        assert_eq!(meta.description, "Williams %R");
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].mnemonic, "willr");
        assert_eq!(meta.outputs[0].description, "Williams %R");
    }
}
