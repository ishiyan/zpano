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

/// Parameters for the Money Flow Index indicator.
pub struct MoneyFlowIndexParams {
    /// The number of time periods. Default 14.
    pub length: usize,
}

impl Default for MoneyFlowIndexParams {
    fn default() -> Self {
        Self { length: 14 }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Money Flow Index indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum MoneyFlowIndexOutput {
    /// The scalar value of the money flow index.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Gene Quong's Money Flow Index (MFI).
///
/// MFI is a volume-weighted oscillator calculated over ℓ periods, showing money flow
/// on up days as a percentage of the total of up and down days.
///
///   TypicalPrice = (High + Low + Close) / 3
///   MoneyFlow = TypicalPrice × Volume
///   MFI = 100 × PositiveMoneyFlow / (PositiveMoneyFlow + NegativeMoneyFlow)
pub struct MoneyFlowIndex {
    length: usize,
    negative_buffer: Vec<f64>,
    positive_buffer: Vec<f64>,
    negative_sum: f64,
    positive_sum: f64,
    previous_sample: f64,
    buffer_index: usize,
    buffer_low_index: usize,
    buffer_count: usize,
    value: f64,
    primed: bool,
}

impl MoneyFlowIndex {
    /// Creates a new Money Flow Index indicator.
    pub fn new(params: &MoneyFlowIndexParams) -> Result<Self, String> {
        if params.length < 1 {
            return Err("invalid money flow index parameters: length should be greater than 0".to_string());
        }

        Ok(Self {
            length: params.length,
            negative_buffer: vec![0.0; params.length],
            positive_buffer: vec![0.0; params.length],
            negative_sum: 0.0,
            positive_sum: 0.0,
            previous_sample: 0.0,
            buffer_index: 0,
            buffer_low_index: 0,
            buffer_count: 0,
            value: f64::NAN,
            primed: false,
        })
    }

    /// Updates the indicator with the given sample using volume = 1.
    pub fn update(&mut self, sample: f64) -> f64 {
        self.update_with_volume(sample, 1.0)
    }

    /// Updates the indicator with the given sample and volume.
    pub fn update_with_volume(&mut self, sample: f64, volume: f64) -> f64 {
        if sample.is_nan() || volume.is_nan() {
            return f64::NAN;
        }

        let length_min_one = self.length - 1;

        if self.primed {
            self.negative_sum -= self.negative_buffer[self.buffer_low_index];
            self.positive_sum -= self.positive_buffer[self.buffer_low_index];

            let amount = sample * volume;
            let diff = sample - self.previous_sample;

            if diff < 0.0 {
                self.negative_buffer[self.buffer_index] = amount;
                self.positive_buffer[self.buffer_index] = 0.0;
                self.negative_sum += amount;
            } else if diff > 0.0 {
                self.negative_buffer[self.buffer_index] = 0.0;
                self.positive_buffer[self.buffer_index] = amount;
                self.positive_sum += amount;
            } else {
                self.negative_buffer[self.buffer_index] = 0.0;
                self.positive_buffer[self.buffer_index] = 0.0;
            }

            let sum = self.positive_sum + self.negative_sum;
            if sum < 1.0 {
                self.value = 0.0;
            } else {
                self.value = 100.0 * self.positive_sum / sum;
            }

            self.buffer_index += 1;
            if self.buffer_index > length_min_one {
                self.buffer_index = 0;
            }

            self.buffer_low_index += 1;
            if self.buffer_low_index > length_min_one {
                self.buffer_low_index = 0;
            }
        } else if self.buffer_count == 0 {
            self.buffer_count += 1;
        } else {
            let amount = sample * volume;
            let diff = sample - self.previous_sample;

            if diff < 0.0 {
                self.negative_buffer[self.buffer_index] = amount;
                self.positive_buffer[self.buffer_index] = 0.0;
                self.negative_sum += amount;
            } else if diff > 0.0 {
                self.negative_buffer[self.buffer_index] = 0.0;
                self.positive_buffer[self.buffer_index] = amount;
                self.positive_sum += amount;
            } else {
                self.negative_buffer[self.buffer_index] = 0.0;
                self.positive_buffer[self.buffer_index] = 0.0;
            }

            if self.length == self.buffer_count {
                let sum = self.positive_sum + self.negative_sum;
                if sum < 1.0 {
                    self.value = 0.0;
                } else {
                    self.value = 100.0 * self.positive_sum / sum;
                }

                self.primed = true;
            }

            self.buffer_index += 1;
            if self.buffer_index > length_min_one {
                self.buffer_index = 0;
            }

            self.buffer_count += 1;
        }

        self.previous_sample = sample;

        self.value
    }
}

impl Indicator for MoneyFlowIndex {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        let mnemonic = format!("mfi({}, hlc/3)", self.length);
        let description = format!("Money Flow Index {}", mnemonic);
        build_metadata(
            Identifier::MoneyFlowIndex,
            &mnemonic,
            &description,
            &[OutputText {
                mnemonic: mnemonic.clone(),
                description: description.clone(),
            }],
        )
    }

    fn update_scalar(&mut self, scalar: &Scalar) -> Output {
        let v = self.update(scalar.value);
        vec![Box::new(Scalar::new(scalar.time, v))]
    }

    fn update_bar(&mut self, bar: &Bar) -> Output {
        // Shadow LineIndicator to use bar volume.
        let price = (bar.high + bar.low + bar.close) / 3.0;
        let v = self.update_with_volume(price, bar.volume);
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

    fn round_to(v: f64, digits: i32) -> f64 {
        let p = 10f64.powi(digits);
        (v * p).round() / p
    }
    #[test]
    fn test_mfi_with_volume() {
        let tp = testdata::test_typical_prices();
        let vol = testdata::test_volumes();
        let expected = testdata::test_expected_mfi();
        let count = tp.len();
        const DIGITS: i32 = 9;

        let mut mfi = MoneyFlowIndex::new(&MoneyFlowIndexParams { length: 14 }).unwrap();

        for i in 0..14 {
            let v = mfi.update_with_volume(tp[i], vol[i]);
            assert!(v.is_nan(), "[{}] expected NaN, got {}", i, v);
            assert!(!mfi.is_primed(), "[{}] expected not primed", i);
        }

        for i in 14..count {
            let v = mfi.update_with_volume(tp[i], vol[i]);
            assert!(!v.is_nan(), "[{}] expected non-NaN, got NaN", i);
            assert!(mfi.is_primed(), "[{}] expected primed", i);

            let got = round_to(v, DIGITS);
            let exp = round_to(expected[i], DIGITS);
            assert_eq!(got, exp, "[{}] expected {}, got {}", i, exp, got);
        }
    }

    #[test]
    fn test_mfi_volume1() {
        let tp = testdata::test_typical_prices();
        let expected = testdata::test_expected_mfi_volume1();
        let count = tp.len();
        const DIGITS: i32 = 9;

        let mut mfi = MoneyFlowIndex::new(&MoneyFlowIndexParams { length: 14 }).unwrap();

        for i in 0..14 {
            let v = mfi.update(tp[i]);
            assert!(v.is_nan(), "[{}] expected NaN, got {}", i, v);
        }

        for i in 14..count {
            let v = mfi.update(tp[i]);
            assert!(!v.is_nan(), "[{}] expected non-NaN, got NaN", i);

            let got = round_to(v, DIGITS);
            let exp = round_to(expected[i], DIGITS);
            assert_eq!(got, exp, "[{}] expected {}, got {}", i, exp, got);
        }
    }

    #[test]
    fn test_mfi_is_primed() {
        let mut mfi = MoneyFlowIndex::new(&MoneyFlowIndexParams { length: 5 }).unwrap();

        assert!(!mfi.is_primed(), "expected not primed initially");

        for i in 1..=5 {
            mfi.update(i as f64);
            assert!(!mfi.is_primed(), "[{}] expected not primed", i);
        }

        mfi.update(5.0);
        assert!(mfi.is_primed(), "expected primed after length+1 samples");

        mfi.update(6.0);
        assert!(mfi.is_primed(), "expected still primed");
    }

    #[test]
    fn test_mfi_nan() {
        let mut mfi = MoneyFlowIndex::new(&MoneyFlowIndexParams { length: 5 }).unwrap();

        let v = mfi.update(f64::NAN);
        assert!(v.is_nan(), "expected NaN for NaN sample");

        let v = mfi.update_with_volume(1.0, f64::NAN);
        assert!(v.is_nan(), "expected NaN for NaN volume");

        let v = mfi.update_with_volume(f64::NAN, f64::NAN);
        assert!(v.is_nan(), "expected NaN for both NaN");
    }

    #[test]
    fn test_mfi_metadata() {
        let mfi = MoneyFlowIndex::new(&MoneyFlowIndexParams { length: 14 }).unwrap();
        let meta = mfi.metadata();

        assert_eq!(meta.identifier, Identifier::MoneyFlowIndex);
        assert_eq!(meta.mnemonic, "mfi(14, hlc/3)");
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].kind, MoneyFlowIndexOutput::Value as i32);
    }

    #[test]
    fn test_mfi_invalid_params() {
        assert!(MoneyFlowIndex::new(&MoneyFlowIndexParams { length: 0 }).is_err());
    }

    #[test]
    fn test_mfi_small_sum() {
        let mut mfi = MoneyFlowIndex::new(&MoneyFlowIndexParams { length: 2 }).unwrap();

        for _ in 0..10 {
            mfi.update_with_volume(0.001, 0.5);
        }

        assert!(mfi.is_primed(), "expected primed");

        let v = mfi.update_with_volume(0.001, 0.5);
        assert_eq!(v, 0.0, "expected 0 for small sum");
    }

    #[test]
    fn test_mfi_update_bar() {
        const DIGITS: i32 = 9;

        let input_high = vec![
            93.250000, 94.940000, 96.375000, 96.190000, 96.000000, 94.720000, 95.000000, 93.720000, 92.470000, 92.750000, 96.250000,
            99.625000, 99.125000, 92.750000, 91.315000,
        ];
        let input_low = vec![
            90.750000, 91.405000, 94.250000, 93.500000, 92.815000, 93.500000, 92.000000, 89.750000, 89.440000, 90.625000, 92.750000,
            96.315000, 96.030000, 88.815000, 86.750000,
        ];
        let input_close = vec![
            91.500000, 94.815000, 94.375000, 95.095000, 93.780000, 94.625000, 92.530000, 92.750000, 90.315000, 92.470000, 96.125000,
            97.250000, 98.500000, 89.875000, 91.000000,
        ];
        let input_volume = vec![
            4077500.0, 4955900.0, 4775300.0, 4155300.0, 4593100.0, 3631300.0, 3382800.0, 4954200.0, 4500000.0, 3397500.0,
            4204500.0, 6321400.0, 10203600.0, 19043900.0, 11692000.0,
        ];

        let mut mfi = MoneyFlowIndex::new(&MoneyFlowIndexParams { length: 14 }).unwrap();

        for i in 0..14 {
            let bar = Bar {
                time: 0,
                open: 0.0,
                high: input_high[i],
                low: input_low[i],
                close: input_close[i],
                volume: input_volume[i],
            };
            let out = mfi.update_bar(&bar);
            let s = out[0].downcast_ref::<Scalar>().unwrap();
            assert!(s.value.is_nan(), "[{}] expected NaN, got {}", i, s.value);
        }

        let bar = Bar {
            time: 0,
            open: 0.0,
            high: input_high[14],
            low: input_low[14],
            close: input_close[14],
            volume: input_volume[14],
        };
        let out = mfi.update_bar(&bar);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!(!s.value.is_nan(), "[14] expected non-NaN, got NaN");

        let expected = testdata::test_expected_mfi();
        let got = round_to(s.value, DIGITS);
        let exp = round_to(expected[14], DIGITS);
        assert_eq!(got, exp, "[14] expected {}, got {}", exp, got);
    }
}
