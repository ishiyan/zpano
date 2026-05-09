use crate::entities::bar::Bar;
use crate::entities::quote::Quote;
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::metadata::Metadata;
use super::super::average_true_range::{AverageTrueRange, AverageTrueRangeParams};

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters for the Normalized Average True Range indicator.
pub struct NormalizedAverageTrueRangeParams {
    /// Number of time periods. Must be >= 1. Default is 14.
    pub length: usize,
}

impl Default for NormalizedAverageTrueRangeParams {
    fn default() -> Self {
        Self { length: 14 }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the NATR indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum NormalizedAverageTrueRangeOutput {
    /// The scalar value of the normalized average true range.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

const NATR_MNEMONIC: &str = "natr";
const NATR_DESCRIPTION: &str = "Normalized Average True Range";

/// Welles Wilder's Normalized Average True Range indicator.
///
/// NATR is calculated as (ATR / close) * 100.
/// If close == 0, the result is 0 (not division by zero).
pub struct NormalizedAverageTrueRange {
    length: usize,
    value: f64,
    primed: bool,
    average_true_range: AverageTrueRange,
}

impl NormalizedAverageTrueRange {
    /// Creates a new NormalizedAverageTrueRange indicator.
    pub fn new(params: &NormalizedAverageTrueRangeParams) -> Result<Self, String> {
        if params.length < 1 {
            return Err(format!("invalid length {}: must be >= 1", params.length));
        }

        let atr = AverageTrueRange::new(&AverageTrueRangeParams { length: params.length }).unwrap();

        Ok(Self {
            length: params.length,
            value: f64::NAN,
            primed: false,
            average_true_range: atr,
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

        let atr_value = self.average_true_range.update(close, high, low);

        if self.average_true_range.is_primed() {
            self.primed = true;

            if close == 0.0 {
                self.value = 0.0;
            } else {
                self.value = (atr_value / close) * 100.0;
            }
        }

        if self.primed {
            return self.value;
        }

        f64::NAN
    }

    /// Updates using a single sample value as substitute for high, low, and close.
    pub fn update_sample(&mut self, sample: f64) -> f64 {
        self.update(sample, sample, sample)
    }
}

impl Indicator for NormalizedAverageTrueRange {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::NormalizedAverageTrueRange,
            NATR_MNEMONIC,
            NATR_DESCRIPTION,
            &[OutputText {
                mnemonic: NATR_MNEMONIC.to_string(),
                description: NATR_DESCRIPTION.to_string(),
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
    fn test_natr_constructor() {
        let natr = NormalizedAverageTrueRange::new(&NormalizedAverageTrueRangeParams { length: 14 }).unwrap();
        assert_eq!(natr.length(), 14);
        assert!(!natr.is_primed());

        assert!(NormalizedAverageTrueRange::new(&NormalizedAverageTrueRangeParams { length: 0 }).is_err());
    }

    #[test]
    fn test_natr_is_primed() {
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        let cls = testdata::test_input_close();
        let mut natr = NormalizedAverageTrueRange::new(&NormalizedAverageTrueRangeParams { length: 5 }).unwrap();

        assert!(!natr.is_primed());

        for i in 0..5 {
            natr.update(cls[i], high[i], low[i]);
            assert!(!natr.is_primed(), "[{}] should not be primed yet", i);
        }

        for i in 5..10 {
            natr.update(cls[i], high[i], low[i]);
            assert!(natr.is_primed(), "[{}] should be primed", i);
        }
    }

    #[test]
    fn test_natr_update() {
        let tolerance = 1e-11;
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        let cls = testdata::test_input_close();
        let expected = testdata::test_expected_natr14();
        let mut natr = NormalizedAverageTrueRange::new(&NormalizedAverageTrueRangeParams { length: 14 }).unwrap();

        for i in 0..cls.len() {
            let act = natr.update(cls[i], high[i], low[i]);

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
    fn test_natr_length_1() {
        let tolerance = 1e-11;
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        let cls = testdata::test_input_close();
        let expected = testdata::test_expected_natr1();
        let mut natr = NormalizedAverageTrueRange::new(&NormalizedAverageTrueRangeParams { length: 1 }).unwrap();

        for i in 0..cls.len() {
            let act = natr.update(cls[i], high[i], low[i]);

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
    fn test_natr_close_zero() {
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        let cls = testdata::test_input_close();
        let mut natr = NormalizedAverageTrueRange::new(&NormalizedAverageTrueRangeParams { length: 14 }).unwrap();

        // Prime the indicator.
        for i in 0..15 {
            natr.update(cls[i], high[i], low[i]);
        }

        // close=0 should return 0, not panic or NaN.
        let result = natr.update(0.0, 3.3, 2.2);
        assert_eq!(result, 0.0);
    }

    #[test]
    fn test_natr_nan_passthrough() {
        let mut natr = NormalizedAverageTrueRange::new(&NormalizedAverageTrueRangeParams { length: 14 }).unwrap();
        assert!(natr.update(f64::NAN, 1.0, 1.0).is_nan());
        assert!(natr.update(1.0, f64::NAN, 1.0).is_nan());
        assert!(natr.update(1.0, 1.0, f64::NAN).is_nan());
        assert!(natr.update_sample(f64::NAN).is_nan());
    }

    #[test]
    fn test_natr_metadata() {
        let natr = NormalizedAverageTrueRange::new(&NormalizedAverageTrueRangeParams { length: 14 }).unwrap();
        let meta = natr.metadata();
        assert_eq!(meta.identifier, Identifier::NormalizedAverageTrueRange);
        assert_eq!(meta.mnemonic, "natr");
        assert_eq!(meta.description, "Normalized Average True Range");
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].mnemonic, "natr");
        assert_eq!(meta.outputs[0].description, "Normalized Average True Range");
    }

    #[test]
    fn test_natr_update_bar() {
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        let cls = testdata::test_input_close();
        let mut natr = NormalizedAverageTrueRange::new(&NormalizedAverageTrueRangeParams { length: 14 }).unwrap();

        for i in 0..15 {
            natr.update(cls[i], high[i], low[i]);
        }

        let bar = Bar {
            time: 1_000_000,
            open: 0.0,
            high: high[15],
            low: low[15],
            close: cls[15],
            volume: 0.0,
        };
        let out = natr.update_bar(&bar);
        assert_eq!(out.len(), 1);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s.time, 1_000_000);
        assert!(!s.value.is_nan());
    }
}
