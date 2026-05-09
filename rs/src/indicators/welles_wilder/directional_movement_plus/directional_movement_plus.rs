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

/// Parameters for the Directional Movement Plus indicator.
pub struct DirectionalMovementPlusParams {
    /// Number of time periods. Must be >= 1. Default is 14.
    pub length: usize,
}

impl Default for DirectionalMovementPlusParams {
    fn default() -> Self {
        Self { length: 14 }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the DM+ indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum DirectionalMovementPlusOutput {
    /// The scalar value of the directional movement plus.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

const DMP_MNEMONIC: &str = "+dm";
const DMP_DESCRIPTION: &str = "Directional Movement Plus";

/// Welles Wilder's Directional Movement Plus indicator.
///
/// +DM measures upward price movement. When the length is greater than 1,
/// Wilder's smoothing method is applied.
pub struct DirectionalMovementPlus {
    length: usize,
    no_smoothing: bool,
    count: usize,
    previous_high: f64,
    previous_low: f64,
    value: f64,
    accumulator: f64,
    primed: bool,
}

impl DirectionalMovementPlus {
    /// Creates a new DirectionalMovementPlus indicator.
    pub fn new(params: &DirectionalMovementPlusParams) -> Result<Self, String> {
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
                let delta_plus = high - self.previous_high;
                let delta_minus = self.previous_low - low;

                if delta_plus > 0.0 && delta_plus > delta_minus {
                    self.value = delta_plus;
                } else {
                    self.value = 0.0;
                }
            } else {
                if self.count > 0 {
                    let delta_plus = high - self.previous_high;
                    let delta_minus = self.previous_low - low;

                    if delta_plus > 0.0 && delta_plus > delta_minus {
                        self.value = delta_plus;
                    } else {
                        self.value = 0.0;
                    }

                    self.primed = true;
                }

                self.count += 1;
            }
        } else if self.primed {
            let delta_plus = high - self.previous_high;
            let delta_minus = self.previous_low - low;

            if delta_plus > 0.0 && delta_plus > delta_minus {
                self.accumulator += -self.accumulator / self.length as f64 + delta_plus;
            } else {
                self.accumulator += -self.accumulator / self.length as f64;
            }

            self.value = self.accumulator;
        } else {
            if self.count > 0 && self.length >= self.count {
                let delta_plus = high - self.previous_high;
                let delta_minus = self.previous_low - low;

                if self.length > self.count {
                    if delta_plus > 0.0 && delta_plus > delta_minus {
                        self.accumulator += delta_plus;
                    }
                } else {
                    if delta_plus > 0.0 && delta_plus > delta_minus {
                        self.accumulator += -self.accumulator / self.length as f64 + delta_plus;
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

impl Indicator for DirectionalMovementPlus {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::DirectionalMovementPlus,
            DMP_MNEMONIC,
            DMP_DESCRIPTION,
            &[OutputText {
                mnemonic: DMP_MNEMONIC.to_string(),
                description: DMP_DESCRIPTION.to_string(),
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
    fn test_dmp_constructor() {
        let dmp = DirectionalMovementPlus::new(&DirectionalMovementPlusParams { length: 14 }).unwrap();
        assert_eq!(dmp.length(), 14);
        assert!(!dmp.is_primed());

        assert!(DirectionalMovementPlus::new(&DirectionalMovementPlusParams { length: 0 }).is_err());
    }

    #[test]
    fn test_dmp_is_primed() {
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();

        // length=1
        let mut dmp = DirectionalMovementPlus::new(&DirectionalMovementPlusParams { length: 1 }).unwrap();
        assert!(!dmp.is_primed());
        dmp.update(high[0], low[0]);
        assert!(!dmp.is_primed());
        dmp.update(high[1], low[1]);
        assert!(dmp.is_primed());

        // length=14
        let mut dmp = DirectionalMovementPlus::new(&DirectionalMovementPlusParams { length: 14 }).unwrap();
        for i in 0..14 {
            dmp.update(high[i], low[i]);
            assert!(!dmp.is_primed(), "[{}] should not be primed yet", i);
        }
        dmp.update(high[14], low[14]);
        assert!(dmp.is_primed());
    }

    #[test]
    fn test_dmp_update() {
        let tolerance = 1e-8;
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        let expected = testdata::test_expected_dmp14();
        let mut dmp = DirectionalMovementPlus::new(&DirectionalMovementPlusParams { length: 14 }).unwrap();

        for i in 0..high.len() {
            let act = dmp.update(high[i], low[i]);

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
        let expected = testdata::test_expected_dmp1();
        let mut dmp = DirectionalMovementPlus::new(&DirectionalMovementPlusParams { length: 1 }).unwrap();

        for i in 0..high.len() {
            let act = dmp.update(high[i], low[i]);

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
    fn test_dmp_nan_passthrough() {
        let mut dmp = DirectionalMovementPlus::new(&DirectionalMovementPlusParams { length: 14 }).unwrap();
        assert!(dmp.update(f64::NAN, 1.0).is_nan());
        assert!(dmp.update(1.0, f64::NAN).is_nan());
        assert!(dmp.update(f64::NAN, f64::NAN).is_nan());
        assert!(dmp.update_sample(f64::NAN).is_nan());
    }

    #[test]
    fn test_dmp_high_low_swap() {
        let mut dmp1 = DirectionalMovementPlus::new(&DirectionalMovementPlusParams { length: 1 }).unwrap();
        let mut dmp2 = DirectionalMovementPlus::new(&DirectionalMovementPlusParams { length: 1 }).unwrap();

        dmp1.update(10.0, 5.0);
        dmp2.update(5.0, 10.0);

        let v1 = dmp1.update(12.0, 6.0);
        let v2 = dmp2.update(6.0, 12.0);

        assert_eq!(v1, v2);
    }

    #[test]
    fn test_dmp_zero_inputs() {
        let mut dmp = DirectionalMovementPlus::new(&DirectionalMovementPlusParams { length: 10 }).unwrap();

        for _ in 0..20 {
            dmp.update_sample(0.0);
        }

        assert!(dmp.is_primed());
    }

    #[test]
    fn test_dmp_metadata() {
        let dmp = DirectionalMovementPlus::new(&DirectionalMovementPlusParams { length: 14 }).unwrap();
        let meta = dmp.metadata();
        assert_eq!(meta.identifier, Identifier::DirectionalMovementPlus);
        assert_eq!(meta.mnemonic, "+dm");
        assert_eq!(meta.description, "Directional Movement Plus");
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].mnemonic, "+dm");
        assert_eq!(meta.outputs[0].description, "Directional Movement Plus");
    }
}
