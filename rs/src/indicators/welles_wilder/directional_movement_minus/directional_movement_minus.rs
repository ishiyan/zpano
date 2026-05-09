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

/// Parameters for the Directional Movement Minus indicator.
pub struct DirectionalMovementMinusParams {
    /// Number of time periods. Must be >= 1. Default is 14.
    pub length: usize,
}

impl Default for DirectionalMovementMinusParams {
    fn default() -> Self {
        Self { length: 14 }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the DM- indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum DirectionalMovementMinusOutput {
    /// The scalar value of the directional movement minus.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

const DMM_MNEMONIC: &str = "-dm";
const DMM_DESCRIPTION: &str = "Directional Movement Minus";

/// Welles Wilder's Directional Movement Minus indicator.
///
/// -DM measures downward price movement. When the length is greater than 1,
/// Wilder's smoothing method is applied.
pub struct DirectionalMovementMinus {
    length: usize,
    no_smoothing: bool,
    count: usize,
    previous_high: f64,
    previous_low: f64,
    value: f64,
    accumulator: f64,
    primed: bool,
}

impl DirectionalMovementMinus {
    /// Creates a new DirectionalMovementMinus indicator.
    pub fn new(params: &DirectionalMovementMinusParams) -> Result<Self, String> {
        if params.length < 1 {
            return Err(format!("invalid length {}: must be >= 1", params.length));
        }

        Ok(Self {
            length: params.length,
            no_smoothing: params.length == 1,
            count: 0,
            previous_high: 0.0,
            previous_low: 0.0,
            value: f64::NAN,
            accumulator: 0.0,
            primed: false,
        })
    }

    /// Returns the length parameter.
    pub fn length(&self) -> usize {
        self.length
    }

    /// Core update with high and low values.
    pub fn update(&mut self, mut high: f64, mut low: f64) -> f64 {
        if high.is_nan() || low.is_nan() {
            return f64::NAN;
        }

        if high < low {
            std::mem::swap(&mut high, &mut low);
        }

        if self.no_smoothing {
            if self.primed {
                let delta_minus = self.previous_low - low;
                let delta_plus = high - self.previous_high;

                if delta_minus > 0.0 && delta_plus < delta_minus {
                    self.value = delta_minus;
                } else {
                    self.value = 0.0;
                }
            } else {
                if self.count > 0 {
                    let delta_minus = self.previous_low - low;
                    let delta_plus = high - self.previous_high;

                    if delta_minus > 0.0 && delta_plus < delta_minus {
                        self.value = delta_minus;
                    } else {
                        self.value = 0.0;
                    }

                    self.primed = true;
                }

                self.count += 1;
            }
        } else if self.primed {
            let delta_minus = self.previous_low - low;
            let delta_plus = high - self.previous_high;

            if delta_minus > 0.0 && delta_plus < delta_minus {
                self.accumulator += -self.accumulator / self.length as f64 + delta_minus;
            } else {
                self.accumulator += -self.accumulator / self.length as f64;
            }

            self.value = self.accumulator;
        } else {
            if self.count > 0 && self.length >= self.count {
                let delta_minus = self.previous_low - low;
                let delta_plus = high - self.previous_high;

                if self.length > self.count {
                    if delta_minus > 0.0 && delta_plus < delta_minus {
                        self.accumulator += delta_minus;
                    }
                } else {
                    if delta_minus > 0.0 && delta_plus < delta_minus {
                        self.accumulator += -self.accumulator / self.length as f64 + delta_minus;
                    } else {
                        self.accumulator += -self.accumulator / self.length as f64;
                    }

                    self.value = self.accumulator;
                    self.primed = true;
                }
            }

            self.count += 1;
        }

        self.previous_low = low;
        self.previous_high = high;

        self.value
    }

    /// Updates using a single sample value as substitute for high and low.
    pub fn update_sample(&mut self, sample: f64) -> f64 {
        self.update(sample, sample)
    }
}

impl Indicator for DirectionalMovementMinus {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::DirectionalMovementMinus,
            DMM_MNEMONIC,
            DMM_DESCRIPTION,
            &[OutputText {
                mnemonic: DMM_MNEMONIC.to_string(),
                description: DMM_DESCRIPTION.to_string(),
            }],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let v = sample.value;
        let result = self.update(v, v);
        vec![Box::new(Scalar::new(sample.time, result))]
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let result = self.update(sample.high, sample.low);
        vec![Box::new(Scalar::new(sample.time, result))]
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (sample.bid_price + sample.ask_price) / 2.0;
        let result = self.update(v, v);
        vec![Box::new(Scalar::new(sample.time, result))]
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let v = sample.price;
        let result = self.update(v, v);
        vec![Box::new(Scalar::new(sample.time, result))]
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::testdata;
    #[test]
    fn test_dmm_constructor() {
        let dmm = DirectionalMovementMinus::new(&DirectionalMovementMinusParams { length: 14 }).unwrap();
        assert_eq!(dmm.length(), 14);
        assert!(!dmm.is_primed());

        assert!(DirectionalMovementMinus::new(&DirectionalMovementMinusParams { length: 0 }).is_err());
    }

    #[test]
    fn test_dmm_is_primed() {
        let mut dmm = DirectionalMovementMinus::new(&DirectionalMovementMinusParams { length: 1 }).unwrap();
        assert!(!dmm.is_primed());
        dmm.update(10.0, 5.0);
        assert!(!dmm.is_primed());
        dmm.update(12.0, 6.0);
        assert!(dmm.is_primed());

        let mut dmm = DirectionalMovementMinus::new(&DirectionalMovementMinusParams { length: 14 }).unwrap();
        for i in 0..14 {
            dmm.update((i + 10) as f64, i as f64);
            assert!(!dmm.is_primed());
        }
        dmm.update(24.0, 14.0);
        assert!(dmm.is_primed());
    }

    #[test]
    fn test_dmm_update() {
        let tolerance = 1e-8;
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        let expected = testdata::test_expected_dmm14();
        let mut dmm = DirectionalMovementMinus::new(&DirectionalMovementMinusParams { length: 14 }).unwrap();

        for i in 0..high.len() {
            let act = dmm.update(high[i], low[i]);

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
    fn test_dmp_length_1() {
        let tolerance = 1e-8;
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        let expected = testdata::test_expected_dmm1();
        let mut dmm = DirectionalMovementMinus::new(&DirectionalMovementMinusParams { length: 1 }).unwrap();

        for i in 0..high.len() {
            let act = dmm.update(high[i], low[i]);

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
    fn test_dmm_nan_passthrough() {
        let mut dmm = DirectionalMovementMinus::new(&DirectionalMovementMinusParams { length: 14 }).unwrap();
        assert!(dmm.update(f64::NAN, 1.0).is_nan());
        assert!(dmm.update(1.0, f64::NAN).is_nan());
        assert!(dmm.update(f64::NAN, f64::NAN).is_nan());
        assert!(dmm.update_sample(f64::NAN).is_nan());
    }

    #[test]
    fn test_dmm_metadata() {
        let dmm = DirectionalMovementMinus::new(&DirectionalMovementMinusParams { length: 14 }).unwrap();
        let meta = dmm.metadata();
        assert_eq!(meta.identifier, Identifier::DirectionalMovementMinus);
        assert_eq!(meta.mnemonic, "-dm");
        assert_eq!(meta.description, "Directional Movement Minus");
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].mnemonic, "-dm");
        assert_eq!(meta.outputs[0].description, "Directional Movement Minus");
    }
}
