use crate::entities::bar::Bar;
use crate::entities::quote::Quote;
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::metadata::Metadata;
use super::super::true_range::{TrueRange, TrueRangeParams};

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters for the Average True Range indicator.
pub struct AverageTrueRangeParams {
    /// Number of time periods. Must be >= 1. Default is 14.
    pub length: usize,
}

impl Default for AverageTrueRangeParams {
    fn default() -> Self {
        Self { length: 14 }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the ATR indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum AverageTrueRangeOutput {
    /// The scalar value of the average true range.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

const ATR_MNEMONIC: &str = "atr";
const ATR_DESCRIPTION: &str = "Average True Range";

/// Welles Wilder's Average True Range indicator.
///
/// ATR averages True Range (TR) values over the specified length using the Wilder method:
/// - multiply the previous value by (length - 1)
/// - add the current TR value
/// - divide by length
///
/// The initial ATR value is a simple average of the first length TR values.
pub struct AverageTrueRange {
    length: usize,
    last_index: usize,
    stage: usize,
    window_count: usize,
    window: Vec<f64>,
    window_sum: f64,
    value: f64,
    primed: bool,
    true_range: TrueRange,
}

impl AverageTrueRange {
    /// Creates a new AverageTrueRange indicator.
    pub fn new(params: &AverageTrueRangeParams) -> Result<Self, String> {
        if params.length < 1 {
            return Err(format!("invalid length {}: must be >= 1", params.length));
        }

        let last_index = params.length - 1;
        let window = if last_index > 0 {
            vec![0.0; params.length]
        } else {
            vec![]
        };

        Ok(Self {
            length: params.length,
            last_index,
            stage: 0,
            window_count: 0,
            window,
            window_sum: 0.0,
            value: f64::NAN,
            primed: false,
            true_range: TrueRange::new(&TrueRangeParams).unwrap(),
        })
    }

    /// Returns the length parameter.
    pub fn length(&self) -> usize {
        self.length
    }

    /// Core update with close, high, low values.
    pub fn update(&mut self, close: f64, high: f64, low: f64) -> f64 {
        if close.is_nan() || high.is_nan() || low.is_nan() {
            return f64::NAN;
        }

        let true_range_value = self.true_range.update(close, high, low);

        if self.last_index == 0 {
            self.value = true_range_value;

            if self.stage == 0 {
                self.stage += 1;
            } else if self.stage == 1 {
                self.stage += 1;
                self.primed = true;
            }

            return self.value;
        }

        if self.stage > 1 {
            // Wilder smoothing method.
            self.value *= self.last_index as f64;
            self.value += true_range_value;
            self.value /= self.length as f64;

            return self.value;
        }

        if self.stage == 1 {
            self.window_sum += true_range_value;
            self.window[self.window_count] = true_range_value;
            self.window_count += 1;

            if self.window_count == self.length {
                self.stage += 1;
                self.primed = true;
                self.value = self.window_sum / self.length as f64;
            }

            if self.primed {
                return self.value;
            }

            return f64::NAN;
        }

        // The very first sample is used by the True Range.
        self.stage += 1;

        f64::NAN
    }

    /// Updates using a single sample value as substitute for high, low, and close.
    pub fn update_sample(&mut self, sample: f64) -> f64 {
        self.update(sample, sample, sample)
    }
}

impl Indicator for AverageTrueRange {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::AverageTrueRange,
            ATR_MNEMONIC,
            ATR_DESCRIPTION,
            &[OutputText {
                mnemonic: ATR_MNEMONIC.to_string(),
                description: ATR_DESCRIPTION.to_string(),
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
    fn test_atr_constructor() {
        let atr = AverageTrueRange::new(&AverageTrueRangeParams { length: 14 }).unwrap();
        assert_eq!(atr.length(), 14);
        assert!(!atr.is_primed());

        assert!(AverageTrueRange::new(&AverageTrueRangeParams { length: 0 }).is_err());
    }

    #[test]
    fn test_atr_is_primed() {
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        let cls = testdata::test_input_close();
        let mut atr = AverageTrueRange::new(&AverageTrueRangeParams { length: 5 }).unwrap();

        assert!(!atr.is_primed());

        for i in 0..5 {
            atr.update(cls[i], high[i], low[i]);
            assert!(!atr.is_primed(), "[{}] should not be primed yet", i);
        }

        for i in 5..10 {
            atr.update(cls[i], high[i], low[i]);
            assert!(atr.is_primed(), "[{}] should be primed", i);
        }
    }

    #[test]
    fn test_atr_update() {
        let tolerance = 1e-12;
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        let cls = testdata::test_input_close();
        let expected = testdata::test_expected_atr();
        let mut atr = AverageTrueRange::new(&AverageTrueRangeParams { length: 14 }).unwrap();

        for i in 0..cls.len() {
            let act = atr.update(cls[i], high[i], low[i]);

            if expected[i].is_nan() {
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
    fn test_atr_length_1() {
        let tolerance = 1e-3;
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        let cls = testdata::test_input_close();
        let mut atr = AverageTrueRange::new(&AverageTrueRangeParams { length: 1 }).unwrap();

        let expected_tr: Vec<f64> = vec![
            f64::NAN, 3.535, 2.125, 2.69, 3.185, 1.22, 3.0, 3.97, 3.31, 2.435,
            3.78, 3.5, 3.095, 9.685, 4.565, 2.31, 4.5, 1.875, 2.72, 2.5,
            2.845, 1.97, 3.625, 3.22, 2.875, 3.875, 3.19, 5.34, 3.655, 3.155,
            2.75, 2.155, 1.875, 3.44, 2.125, 3.28, 2.315, 3.565, 2.31, 2.03,
            1.94, 5.125, 3.97, 1.47, 3.16, 1.315, 2.22, 2.72, 2.59, 1.655,
        ];

        for i in 0..50 {
            let act = atr.update(cls[i], high[i], low[i]);

            if expected_tr[i].is_nan() {
                assert!(act.is_nan(), "[{}] expected NaN, got {}", i, act);
                continue;
            }

            assert!(!act.is_nan(), "[{}] expected {}, got NaN", i, expected_tr[i]);
            assert!(
                (act - expected_tr[i]).abs() < tolerance,
                "[{}] expected {}, got {}",
                i,
                expected_tr[i],
                act
            );
        }
    }

    #[test]
    fn test_atr_nan_passthrough() {
        let mut atr = AverageTrueRange::new(&AverageTrueRangeParams { length: 14 }).unwrap();
        assert!(atr.update(f64::NAN, 1.0, 1.0).is_nan());
        assert!(atr.update(1.0, f64::NAN, 1.0).is_nan());
        assert!(atr.update(1.0, 1.0, f64::NAN).is_nan());
        assert!(atr.update_sample(f64::NAN).is_nan());
    }

    #[test]
    fn test_atr_metadata() {
        let atr = AverageTrueRange::new(&AverageTrueRangeParams { length: 14 }).unwrap();
        let meta = atr.metadata();
        assert_eq!(meta.identifier, Identifier::AverageTrueRange);
        assert_eq!(meta.mnemonic, "atr");
        assert_eq!(meta.description, "Average True Range");
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].mnemonic, "atr");
        assert_eq!(meta.outputs[0].description, "Average True Range");
    }

    #[test]
    fn test_atr_update_bar() {
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        let cls = testdata::test_input_close();
        let mut atr = AverageTrueRange::new(&AverageTrueRangeParams { length: 14 }).unwrap();

        for i in 0..14 {
            atr.update(cls[i], high[i], low[i]);
        }

        let bar = Bar {
            time: 1_000_000,
            open: 0.0,
            high: high[14],
            low: low[14],
            close: cls[14],
            volume: 0.0,
        };
        let out = atr.update_bar(&bar);
        assert_eq!(out.len(), 1);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s.time, 1_000_000);
        assert!(!s.value.is_nan());
    }
}
