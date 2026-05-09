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

/// Parameters for the Advance-Decline indicator.
/// Advance-Decline requires HLCV bar data and has no configurable parameters.
pub struct AdvanceDeclineParams;

impl Default for AdvanceDeclineParams {
    fn default() -> Self {
        Self
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Advance-Decline indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum AdvanceDeclineOutput {
    /// The scalar value of the advance-decline line.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

const AD_MNEMONIC: &str = "ad";
const AD_DESCRIPTION: &str = "Advance-Decline";

/// Marc Chaikin's Advance-Decline (A/D) Line.
///
/// The Accumulation/Distribution Line is a cumulative indicator that uses volume
/// and price to assess whether a stock is being accumulated or distributed.
///
/// The value is calculated as:
///
///   CLV = ((Close - Low) - (High - Close)) / (High - Low)
///   AD  = AD_previous + CLV × Volume
///
/// When High equals Low, the A/D value is unchanged (no division by zero).
pub struct AdvanceDecline {
    ad: f64,
    value: f64,
    primed: bool,
}

impl AdvanceDecline {
    /// Creates a new Advance-Decline indicator.
    pub fn new(_params: &AdvanceDeclineParams) -> Result<Self, String> {
        Ok(Self {
            ad: 0.0,
            value: f64::NAN,
            primed: false,
        })
    }

    /// Updates the indicator with high, low, close, and volume values.
    pub fn update_hlcv(&mut self, high: f64, low: f64, close: f64, volume: f64) -> f64 {
        if high.is_nan() || low.is_nan() || close.is_nan() || volume.is_nan() {
            return f64::NAN;
        }

        let temp = high - low;
        if temp > 0.0 {
            self.ad += ((close - low) - (high - close)) / temp * volume;
        }

        self.value = self.ad;
        self.primed = true;

        self.value
    }

    /// Updates using a single sample value (H=L=C, so range=0, AD unchanged).
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return f64::NAN;
        }

        self.update_hlcv(sample, sample, sample, 1.0)
    }
}

impl Indicator for AdvanceDecline {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::AdvanceDecline,
            AD_MNEMONIC,
            AD_DESCRIPTION,
            &[OutputText {
                mnemonic: AD_MNEMONIC.to_string(),
                description: AD_DESCRIPTION.to_string(),
            }],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let v = sample.value;
        let result = self.update_hlcv(v, v, v, 1.0);
        vec![Box::new(Scalar::new(sample.time, result))]
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let result = self.update_hlcv(sample.high, sample.low, sample.close, sample.volume);
        vec![Box::new(Scalar::new(sample.time, result))]
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (sample.bid_price + sample.ask_price) / 2.0;
        let result = self.update_hlcv(v, v, v, 1.0);
        vec![Box::new(Scalar::new(sample.time, result))]
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let v = sample.price;
        let result = self.update_hlcv(v, v, v, 1.0);
        vec![Box::new(Scalar::new(sample.time, result))]
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::testdata;
    fn round_to(v: f64, digits: u32) -> f64 {
        let p = 10f64.powi(digits as i32);
        (v * p).round() / p
    }

    #[test]
    fn test_advance_decline_with_volume() {
        let digits = 2;
        let highs = testdata::test_highs();
        let lows = testdata::test_lows();
        let closes = testdata::test_closes();
        let volumes = testdata::test_volumes();
        let expected = testdata::test_expected_ad();
        let count = highs.len();

        let mut ad = AdvanceDecline::new(&AdvanceDeclineParams).unwrap();

        for i in 0..count {
            let v = ad.update_hlcv(highs[i], lows[i], closes[i], volumes[i]);
            assert!(!v.is_nan(), "[{}] expected non-NaN, got NaN", i);
            assert!(ad.is_primed(), "[{}] expected primed", i);

            let got = round_to(v, digits);
            let exp = round_to(expected[i], digits);
            assert_eq!(got, exp, "[{}] expected {}, got {}", i, exp, got);
        }
    }

    #[test]
    fn test_advance_decline_spot_checks() {
        let digits = 2;
        let highs = testdata::test_highs();
        let lows = testdata::test_lows();
        let closes = testdata::test_closes();
        let volumes = testdata::test_volumes();
        let count = highs.len();

        let mut ad = AdvanceDecline::new(&AdvanceDeclineParams).unwrap();

        let mut values = Vec::new();
        for i in 0..count {
            let v = ad.update_hlcv(highs[i], lows[i], closes[i], volumes[i]);
            values.push(v);
        }

        let spot_checks: Vec<(usize, f64)> = vec![
            (0, -1631000.00),
            (1, 2974412.02),
            (250, 8707691.07),
            (251, 8328944.54),
        ];

        for (index, expected) in spot_checks {
            let got = round_to(values[index], digits);
            let exp = round_to(expected, digits);
            assert_eq!(got, exp, "spot check [{}]: expected {}, got {}", index, exp, got);
        }
    }

    #[test]
    fn test_advance_decline_update_bar() {
        let digits = 2;
        let mut ad = AdvanceDecline::new(&AdvanceDeclineParams).unwrap();

        let highs = testdata::test_highs();
        let lows = testdata::test_lows();
        let closes = testdata::test_closes();
        let volumes = testdata::test_volumes();
        let expected = testdata::test_expected_ad();

        for i in 0..10 {
            let bar = Bar {
                time: 1_000_000,
                open: highs[i],
                high: highs[i],
                low: lows[i],
                close: closes[i],
                volume: volumes[i],
            };

            let output = ad.update_bar(&bar);
            let scalar = output[0].downcast_ref::<Scalar>().unwrap();

            let got = round_to(scalar.value, digits);
            let exp = round_to(expected[i], digits);
            assert_eq!(got, exp, "[{}] bar: expected {}, got {}", i, exp, got);
        }
    }

    #[test]
    fn test_advance_decline_scalar_update() {
        let mut ad = AdvanceDecline::new(&AdvanceDeclineParams).unwrap();

        let v = ad.update(100.0);
        assert_eq!(v, 0.0, "expected 0 after scalar update, got {}", v);
        assert!(ad.is_primed(), "expected primed after first update");
    }

    #[test]
    fn test_advance_decline_nan() {
        let mut ad = AdvanceDecline::new(&AdvanceDeclineParams).unwrap();

        assert!(ad.update(f64::NAN).is_nan());
        assert!(ad.update_hlcv(f64::NAN, 1.0, 2.0, 3.0).is_nan());
        assert!(ad.update_hlcv(1.0, f64::NAN, 2.0, 3.0).is_nan());
        assert!(ad.update_hlcv(1.0, 2.0, f64::NAN, 3.0).is_nan());
        assert!(ad.update_hlcv(1.0, 2.0, 3.0, f64::NAN).is_nan());
    }

    #[test]
    fn test_advance_decline_not_primed_initially() {
        let mut ad = AdvanceDecline::new(&AdvanceDeclineParams).unwrap();

        assert!(!ad.is_primed());
        assert!(ad.update(f64::NAN).is_nan());
        assert!(!ad.is_primed());
    }

    #[test]
    fn test_advance_decline_metadata() {
        let ad = AdvanceDecline::new(&AdvanceDeclineParams).unwrap();
        let meta = ad.metadata();
        assert_eq!(meta.identifier, Identifier::AdvanceDecline);
        assert_eq!(meta.mnemonic, "ad");
        assert_eq!(meta.description, "Advance-Decline");
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].mnemonic, "ad");
        assert_eq!(meta.outputs[0].description, "Advance-Decline");
    }
}
