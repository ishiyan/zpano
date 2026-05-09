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

/// Parameters for the True Range indicator.
/// True Range has no configurable parameters.
pub struct TrueRangeParams;

impl Default for TrueRangeParams {
    fn default() -> Self {
        Self
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the True Range indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum TrueRangeOutput {
    /// The scalar value of the true range.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

const TR_MNEMONIC: &str = "tr";
const TR_DESCRIPTION: &str = "True Range";

/// Welles Wilder's True Range indicator.
///
/// The True Range is defined as the largest of:
/// - the distance from today's high to today's low
/// - the distance from yesterday's close to today's high
/// - the distance from yesterday's close to today's low
///
/// The first update stores the close and returns NaN (not primed).
/// The indicator is primed from the second update onward.
pub struct TrueRange {
    previous_close: f64,
    value: f64,
    primed: bool,
}

impl TrueRange {
    /// Creates a new TrueRange indicator.
    pub fn new(_params: &TrueRangeParams) -> Result<Self, String> {
        Ok(Self {
            previous_close: f64::NAN,
            value: f64::NAN,
            primed: false,
        })
    }

    /// Core update with close, high, low values.
    pub fn update(&mut self, close: f64, high: f64, low: f64) -> f64 {
        if close.is_nan() || high.is_nan() || low.is_nan() {
            return f64::NAN;
        }

        if !self.primed {
            if self.previous_close.is_nan() {
                self.previous_close = close;
                return f64::NAN;
            }
            self.primed = true;
        }

        let mut greatest = high - low;

        let temp = (high - self.previous_close).abs();
        if greatest < temp {
            greatest = temp;
        }

        let temp = (low - self.previous_close).abs();
        if greatest < temp {
            greatest = temp;
        }

        self.value = greatest;
        self.previous_close = close;

        self.value
    }

    /// Updates using a single sample value as substitute for high, low, and close.
    pub fn update_sample(&mut self, sample: f64) -> f64 {
        self.update(sample, sample, sample)
    }
}

impl Indicator for TrueRange {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::TrueRange,
            TR_MNEMONIC,
            TR_DESCRIPTION,
            &[OutputText {
                mnemonic: TR_MNEMONIC.to_string(),
                description: TR_DESCRIPTION.to_string(),
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
    fn test_true_range_update() {
        let tolerance = 1e-3;
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        let close = testdata::test_input_close();
        let expected = testdata::test_expected_tr();
        let mut tr = TrueRange::new(&TrueRangeParams).unwrap();

        for i in 0..close.len() {
            let act = tr.update(close[i], high[i], low[i]);

            if i == 0 {
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
    fn test_true_range_nan_passthrough() {
        let mut tr = TrueRange::new(&TrueRangeParams).unwrap();
        assert!(tr.update(f64::NAN, 1.0, 1.0).is_nan());
        assert!(tr.update(1.0, f64::NAN, 1.0).is_nan());
        assert!(tr.update(1.0, 1.0, f64::NAN).is_nan());
    }

    #[test]
    fn test_true_range_is_primed() {
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        let close = testdata::test_input_close();
        let mut tr = TrueRange::new(&TrueRangeParams).unwrap();

        assert!(!tr.is_primed());
        tr.update(close[0], high[0], low[0]);
        assert!(!tr.is_primed());
        tr.update(close[1], high[1], low[1]);
        assert!(tr.is_primed());
        tr.update(close[2], high[2], low[2]);
        assert!(tr.is_primed());
    }

    #[test]
    fn test_true_range_update_sample() {
        let mut tr = TrueRange::new(&TrueRangeParams).unwrap();

        let v = tr.update_sample(100.0);
        assert!(v.is_nan());

        let v = tr.update_sample(105.0);
        assert!((v - 5.0).abs() < 1e-10);

        let v = tr.update_sample(102.0);
        assert!((v - 3.0).abs() < 1e-10);
    }

    #[test]
    fn test_true_range_update_bar() {
        let mut tr = TrueRange::new(&TrueRangeParams).unwrap();
        tr.update(100.0, 105.0, 95.0);

        let bar = Bar {
            time: 1_000_000,
            open: 0.0,
            high: 110.0,
            low: 98.0,
            close: 108.0,
            volume: 0.0,
        };
        let out = tr.update_bar(&bar);
        assert_eq!(out.len(), 1);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s.time, 1_000_000);
        assert!(!s.value.is_nan());
    }

    #[test]
    fn test_true_range_metadata() {
        let tr = TrueRange::new(&TrueRangeParams).unwrap();
        let meta = tr.metadata();
        assert_eq!(meta.identifier, Identifier::TrueRange);
        assert_eq!(meta.mnemonic, "tr");
        assert_eq!(meta.description, "True Range");
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].mnemonic, "tr");
        assert_eq!(meta.outputs[0].description, "True Range");
    }
}
