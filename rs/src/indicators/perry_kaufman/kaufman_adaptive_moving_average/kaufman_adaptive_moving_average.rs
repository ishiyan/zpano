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

/// Parameters for creating KAMA from lengths.
pub struct KaufmanAdaptiveMovingAverageLengthParams {
    pub efficiency_ratio_length: usize,
    pub fastest_length: usize,
    pub slowest_length: usize,
}

impl Default for KaufmanAdaptiveMovingAverageLengthParams {
    fn default() -> Self {
        Self {
            efficiency_ratio_length: 10,
            fastest_length: 2,
            slowest_length: 30,
        }
    }
}

/// Parameters for creating KAMA from smoothing factors.
pub struct KaufmanAdaptiveMovingAverageSmoothingFactorParams {
    pub efficiency_ratio_length: usize,
    pub fastest_smoothing_factor: f64,
    pub slowest_smoothing_factor: f64,
}

impl Default for KaufmanAdaptiveMovingAverageSmoothingFactorParams {
    fn default() -> Self {
        Self {
            efficiency_ratio_length: 10,
            fastest_smoothing_factor: 2.0 / 3.0,
            slowest_smoothing_factor: 2.0 / 31.0,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum KaufmanAdaptiveMovingAverageOutput {
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Perry Kaufman's Adaptive Moving Average (KAMA).
pub struct KaufmanAdaptiveMovingAverage {
    efficiency_ratio_length: usize,
    window_count: usize,
    window: Vec<f64>,
    absolute_delta: Vec<f64>,
    absolute_delta_sum: f64,
    alpha_fastest: f64,
    alpha_slowest: f64,
    alpha_diff: f64,
    value: f64,
    efficiency_ratio: f64,
    primed: bool,
    mnemonic: String,
    description: String,
}

impl KaufmanAdaptiveMovingAverage {
    /// Creates KAMA from length parameters.
    pub fn new_from_lengths(params: &KaufmanAdaptiveMovingAverageLengthParams) -> Result<Self, String> {
        if params.efficiency_ratio_length < 2 {
            return Err("invalid Kaufman adaptive moving average parameters: efficiency ratio length should be larger than 1".to_string());
        }
        if params.fastest_length < 2 {
            return Err("invalid Kaufman adaptive moving average parameters: fastest smoothing length should be larger than 1".to_string());
        }
        if params.slowest_length < 2 {
            return Err("invalid Kaufman adaptive moving average parameters: slowest smoothing length should be larger than 1".to_string());
        }

        let af = 2.0 / (1 + params.fastest_length) as f64;
        let a_s = 2.0 / (1 + params.slowest_length) as f64;
        let mnemonic = format!("kama({}, {}, {})", params.efficiency_ratio_length, params.fastest_length, params.slowest_length);
        let description = format!("Kaufman adaptive moving average {}", mnemonic);

        Ok(Self {
            efficiency_ratio_length: params.efficiency_ratio_length,
            window_count: 0,
            window: vec![0.0; params.efficiency_ratio_length + 1],
            absolute_delta: vec![0.0; params.efficiency_ratio_length + 1],
            absolute_delta_sum: 0.0,
            alpha_fastest: af,
            alpha_slowest: a_s,
            alpha_diff: af - a_s,
            value: f64::NAN,
            efficiency_ratio: f64::NAN,
            primed: false,
            mnemonic,
            description,
        })
    }

    /// Creates KAMA from smoothing factor parameters.
    pub fn new_from_smoothing_factors(params: &KaufmanAdaptiveMovingAverageSmoothingFactorParams) -> Result<Self, String> {
        if params.efficiency_ratio_length < 2 {
            return Err("invalid Kaufman adaptive moving average parameters: efficiency ratio length should be larger than 1".to_string());
        }
        if params.fastest_smoothing_factor < 0.0 || params.fastest_smoothing_factor > 1.0 {
            return Err("invalid Kaufman adaptive moving average parameters: fastest smoothing factor should be in range [0, 1]".to_string());
        }
        if params.slowest_smoothing_factor < 0.0 || params.slowest_smoothing_factor > 1.0 {
            return Err("invalid Kaufman adaptive moving average parameters: slowest smoothing factor should be in range [0, 1]".to_string());
        }

        const EPSILON: f64 = 0.00000001;
        let af = if params.fastest_smoothing_factor < EPSILON { EPSILON } else { params.fastest_smoothing_factor };
        let a_s = if params.slowest_smoothing_factor < EPSILON { EPSILON } else { params.slowest_smoothing_factor };

        let mnemonic = format!("kama({}, {:.4}, {:.4})", params.efficiency_ratio_length, af, a_s);
        let description = format!("Kaufman adaptive moving average {}", mnemonic);

        Ok(Self {
            efficiency_ratio_length: params.efficiency_ratio_length,
            window_count: 0,
            window: vec![0.0; params.efficiency_ratio_length + 1],
            absolute_delta: vec![0.0; params.efficiency_ratio_length + 1],
            absolute_delta_sum: 0.0,
            alpha_fastest: af,
            alpha_slowest: a_s,
            alpha_diff: af - a_s,
            value: f64::NAN,
            efficiency_ratio: f64::NAN,
            primed: false,
            mnemonic,
            description,
        })
    }

    /// Updates the indicator with the next sample value.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        const EPSILON: f64 = 0.00000001;

        if self.primed {
            let temp = (sample - self.window[self.efficiency_ratio_length]).abs();
            self.absolute_delta_sum += temp - self.absolute_delta[1];

            for i in 0..self.efficiency_ratio_length {
                let j = i + 1;
                self.window[i] = self.window[j];
                self.absolute_delta[i] = self.absolute_delta[j];
            }

            self.window[self.efficiency_ratio_length] = sample;
            self.absolute_delta[self.efficiency_ratio_length] = temp;
            let delta = (sample - self.window[0]).abs();

            let er = if self.absolute_delta_sum <= delta || self.absolute_delta_sum < EPSILON {
                1.0
            } else {
                delta / self.absolute_delta_sum
            };

            self.efficiency_ratio = er;
            let sc = self.alpha_slowest + er * self.alpha_diff;
            self.value += (sample - self.value) * sc * sc;

            self.value
        } else {
            self.window[self.window_count] = sample;
            if self.window_count > 0 {
                let temp = (sample - self.window[self.window_count - 1]).abs();
                self.absolute_delta[self.window_count] = temp;
                self.absolute_delta_sum += temp;
            }

            if self.efficiency_ratio_length == self.window_count {
                self.primed = true;
                let delta = (sample - self.window[0]).abs();

                let er = if self.absolute_delta_sum <= delta || self.absolute_delta_sum < EPSILON {
                    1.0
                } else {
                    delta / self.absolute_delta_sum
                };

                self.efficiency_ratio = er;
                let sc = self.alpha_slowest + er * self.alpha_diff;
                self.value = self.window[self.efficiency_ratio_length - 1];
                self.value += (sample - self.value) * sc * sc;

                self.value
            } else {
                self.window_count += 1;
                f64::NAN
            }
        }
    }

    /// Returns the current efficiency ratio.
    pub fn efficiency_ratio(&self) -> f64 {
        self.efficiency_ratio
    }
}

impl Indicator for KaufmanAdaptiveMovingAverage {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::KaufmanAdaptiveMovingAverage,
            &self.mnemonic,
            &self.description,
            &[OutputText {
                mnemonic: self.mnemonic.clone(),
                description: self.description.clone(),
            }],
        )
    }

    fn update_scalar(&mut self, scalar: &Scalar) -> Output {
        let v = self.update(scalar.value);
        vec![Box::new(Scalar::new(scalar.time, v))]
    }

    fn update_bar(&mut self, bar: &Bar) -> Output {
        let v = self.update(bar.close);
        vec![Box::new(Scalar::new(bar.time, v))]
    }

    fn update_quote(&mut self, quote: &Quote) -> Output {
        let v = (quote.bid_price + quote.ask_price) / 2.0;
        let result = self.update(v);
        vec![Box::new(Scalar::new(quote.time, result))]
    }

    fn update_trade(&mut self, trade: &Trade) -> Output {
        let result = self.update(trade.price);
        vec![Box::new(Scalar::new(trade.time, result))]
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::testdata;
    #[test]
    fn test_kama_update_value() {
        let input = testdata::test_input();
        let expected = testdata::test_expected();

        let mut kama = KaufmanAdaptiveMovingAverage::new_from_lengths(
            &KaufmanAdaptiveMovingAverageLengthParams {
                efficiency_ratio_length: 10,
                fastest_length: 2,
                slowest_length: 30,
            },
        ).unwrap();

        for i in 0..10 {
            let v = kama.update(input[i]);
            assert!(v.is_nan(), "[{}] expected NaN, got {}", i, v);
        }

        for i in 10..input.len() {
            let v = kama.update(input[i]);
            assert!((v - expected[i]).abs() < 1e-8, "[{}] expected {}, got {}", i, expected[i], v);
        }

        // NaN passthrough
        assert!(kama.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_kama_efficiency_ratio() {
        let input = testdata::test_input();
        let expected_er = testdata::test_expected_er();

        let mut kama = KaufmanAdaptiveMovingAverage::new_from_lengths(
            &KaufmanAdaptiveMovingAverageLengthParams {
                efficiency_ratio_length: 10,
                fastest_length: 2,
                slowest_length: 30,
            },
        ).unwrap();

        for i in 0..10 {
            kama.update(input[i]);
        }

        for i in 10..input.len() {
            kama.update(input[i]);
            let er = kama.efficiency_ratio();
            assert!((er - expected_er[i]).abs() < 1e-8, "[{}] ER expected {}, got {}", i, expected_er[i], er);
        }
    }

    #[test]
    fn test_kama_is_primed() {
        let input = testdata::test_input();

        let mut kama = KaufmanAdaptiveMovingAverage::new_from_lengths(
            &KaufmanAdaptiveMovingAverageLengthParams {
                efficiency_ratio_length: 10,
                fastest_length: 2,
                slowest_length: 30,
            },
        ).unwrap();

        assert!(!kama.is_primed());

        for i in 0..10 {
            kama.update(input[i]);
            assert!(!kama.is_primed(), "[{}] expected not primed", i);
        }

        for i in 10..input.len() {
            kama.update(input[i]);
            assert!(kama.is_primed(), "[{}] expected primed", i);
        }
    }

    #[test]
    fn test_kama_metadata_length() {
        let kama = KaufmanAdaptiveMovingAverage::new_from_lengths(
            &KaufmanAdaptiveMovingAverageLengthParams {
                efficiency_ratio_length: 10,
                fastest_length: 2,
                slowest_length: 30,
            },
        ).unwrap();

        let meta = kama.metadata();
        assert_eq!(meta.identifier, Identifier::KaufmanAdaptiveMovingAverage);
        assert_eq!(meta.mnemonic, "kama(10, 2, 30)");
        assert_eq!(meta.description, "Kaufman adaptive moving average kama(10, 2, 30)");
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].kind, KaufmanAdaptiveMovingAverageOutput::Value as i32);
    }

    #[test]
    fn test_kama_metadata_smoothing_factor() {
        let kama = KaufmanAdaptiveMovingAverage::new_from_smoothing_factors(
            &KaufmanAdaptiveMovingAverageSmoothingFactorParams {
                efficiency_ratio_length: 10,
                fastest_smoothing_factor: 0.666666666,
                slowest_smoothing_factor: 0.064516129,
            },
        ).unwrap();

        let meta = kama.metadata();
        assert_eq!(meta.identifier, Identifier::KaufmanAdaptiveMovingAverage);
        assert_eq!(meta.mnemonic, "kama(10, 0.6667, 0.0645)");
    }

    #[test]
    fn test_kama_invalid_params() {
        // ER length < 2
        assert!(KaufmanAdaptiveMovingAverage::new_from_lengths(
            &KaufmanAdaptiveMovingAverageLengthParams { efficiency_ratio_length: 1, fastest_length: 2, slowest_length: 30 },
        ).is_err());

        // fastest length < 2
        assert!(KaufmanAdaptiveMovingAverage::new_from_lengths(
            &KaufmanAdaptiveMovingAverageLengthParams { efficiency_ratio_length: 10, fastest_length: 1, slowest_length: 30 },
        ).is_err());

        // slowest length < 2
        assert!(KaufmanAdaptiveMovingAverage::new_from_lengths(
            &KaufmanAdaptiveMovingAverageLengthParams { efficiency_ratio_length: 10, fastest_length: 2, slowest_length: 1 },
        ).is_err());

        // fastest alpha out of range
        assert!(KaufmanAdaptiveMovingAverage::new_from_smoothing_factors(
            &KaufmanAdaptiveMovingAverageSmoothingFactorParams { efficiency_ratio_length: 10, fastest_smoothing_factor: -0.00000001, slowest_smoothing_factor: 0.33 },
        ).is_err());

        assert!(KaufmanAdaptiveMovingAverage::new_from_smoothing_factors(
            &KaufmanAdaptiveMovingAverageSmoothingFactorParams { efficiency_ratio_length: 10, fastest_smoothing_factor: 1.00000001, slowest_smoothing_factor: 0.33 },
        ).is_err());

        // slowest alpha out of range
        assert!(KaufmanAdaptiveMovingAverage::new_from_smoothing_factors(
            &KaufmanAdaptiveMovingAverageSmoothingFactorParams { efficiency_ratio_length: 10, fastest_smoothing_factor: 0.66, slowest_smoothing_factor: -0.00000001 },
        ).is_err());

        assert!(KaufmanAdaptiveMovingAverage::new_from_smoothing_factors(
            &KaufmanAdaptiveMovingAverageSmoothingFactorParams { efficiency_ratio_length: 10, fastest_smoothing_factor: 0.66, slowest_smoothing_factor: 1.00000001 },
        ).is_err());
    }
}
