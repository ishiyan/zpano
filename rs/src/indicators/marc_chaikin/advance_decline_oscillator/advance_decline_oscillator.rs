use std::any::Any;

use crate::entities::bar::Bar;
use crate::entities::quote::Quote;
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::indicators::common::exponential_moving_average::{
    ExponentialMovingAverage, ExponentialMovingAverageLengthParams,
};
use crate::indicators::common::simple_moving_average::{
    SimpleMovingAverage, SimpleMovingAverageParams,
};
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::Indicator;
use crate::indicators::core::metadata::Metadata;

/// Output index for AdvanceDeclineOscillator.
#[repr(u8)]
#[derive(Debug, Clone, Copy, PartialEq)]
pub enum AdvanceDeclineOscillatorOutput {
    Value = 1,
}

/// Moving average type selector.
#[repr(u8)]
#[derive(Debug, Clone, Copy, PartialEq)]
pub enum AdvanceDeclineOscillatorMaType {
    SMA = 0,
    EMA = 1,
}

/// Parameters for AdvanceDeclineOscillator.
#[derive(Debug, Clone)]
pub struct AdvanceDeclineOscillatorParams {
    pub fast_length: i64,
    pub slow_length: i64,
    // Note: SMA uses usize, EMA uses i64. Conversion handled in new().
    pub moving_average_type: AdvanceDeclineOscillatorMaType,
    pub first_is_average: bool,
}

impl Default for AdvanceDeclineOscillatorParams {
    fn default() -> Self {
        Self {
            fast_length: 3,
            slow_length: 10,
            moving_average_type: AdvanceDeclineOscillatorMaType::EMA,
            first_is_average: false,
        }
    }
}

trait LineUpdater {
    fn update(&mut self, sample: f64) -> f64;
    fn is_primed(&self) -> bool;
}

struct SmaUpdater {
    inner: SimpleMovingAverage,
}

impl LineUpdater for SmaUpdater {
    fn update(&mut self, sample: f64) -> f64 {
        self.inner.update(sample)
    }
    fn is_primed(&self) -> bool {
        self.inner.is_primed()
    }
}

struct EmaUpdater {
    inner: ExponentialMovingAverage,
}

impl LineUpdater for EmaUpdater {
    fn update(&mut self, sample: f64) -> f64 {
        self.inner.update(sample)
    }
    fn is_primed(&self) -> bool {
        self.inner.is_primed()
    }
}

/// Marc Chaikin's Advance-Decline Oscillator.
pub struct AdvanceDeclineOscillator {
    ad: f64,
    fast_ma: Box<dyn LineUpdater>,
    slow_ma: Box<dyn LineUpdater>,
    value: f64,
    primed: bool,
}

impl AdvanceDeclineOscillator {
    pub fn new(params: &AdvanceDeclineOscillatorParams) -> Result<Self, String> {
        if params.fast_length < 2 {
            return Err(
                "invalid advance-decline oscillator parameters: fast length should be greater than 1"
                    .to_string(),
            );
        }
        if params.slow_length < 2 {
            return Err(
                "invalid advance-decline oscillator parameters: slow length should be greater than 1"
                    .to_string(),
            );
        }

        let (fast_ma, slow_ma): (Box<dyn LineUpdater>, Box<dyn LineUpdater>) =
            match params.moving_average_type {
                AdvanceDeclineOscillatorMaType::SMA => {
                    let fast = SimpleMovingAverage::new(&SimpleMovingAverageParams {
                        length: params.fast_length as usize,
                        ..Default::default()
                    })?;
                    let slow = SimpleMovingAverage::new(&SimpleMovingAverageParams {
                        length: params.slow_length as usize,
                        ..Default::default()
                    })?;
                    (
                        Box::new(SmaUpdater { inner: fast }),
                        Box::new(SmaUpdater { inner: slow }),
                    )
                }
                AdvanceDeclineOscillatorMaType::EMA => {
                    let fast = ExponentialMovingAverage::new_from_length(
                        &ExponentialMovingAverageLengthParams {
                            length: params.fast_length,
                            first_is_average: params.first_is_average,
                            ..Default::default()
                        },
                    )?;
                    let slow = ExponentialMovingAverage::new_from_length(
                        &ExponentialMovingAverageLengthParams {
                            length: params.slow_length,
                            first_is_average: params.first_is_average,
                            ..Default::default()
                        },
                    )?;
                    (
                        Box::new(EmaUpdater { inner: fast }),
                        Box::new(EmaUpdater { inner: slow }),
                    )
                }
            };

        Ok(Self {
            ad: 0.0,
            fast_ma,
            slow_ma,
            value: f64::NAN,
            primed: false,
        })
    }

    /// Update with high, low, close, volume.
    pub fn update_hlcv(&mut self, high: f64, low: f64, close: f64, volume: f64) -> f64 {
        if high.is_nan() || low.is_nan() || close.is_nan() || volume.is_nan() {
            return f64::NAN;
        }

        let temp = high - low;
        if temp > 0.0 {
            self.ad += ((close - low) - (high - close)) / temp * volume;
        }

        let fast = self.fast_ma.update(self.ad);
        let slow = self.slow_ma.update(self.ad);
        self.primed = self.fast_ma.is_primed() && self.slow_ma.is_primed();

        if fast.is_nan() || slow.is_nan() {
            self.value = f64::NAN;
            return self.value;
        }

        self.value = fast - slow;
        self.value
    }
}

impl Indicator for AdvanceDeclineOscillator {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::AdvanceDeclineOscillator,
            "adosc",
            "Chaikin Advance-Decline Oscillator",
            &[OutputText {
                mnemonic: "adosc".to_string(),
                description: "Chaikin Advance-Decline Oscillator".to_string(),
            }],
        )
    }

    fn update_scalar(&mut self, scalar: &Scalar) -> Vec<Box<dyn Any>> {
        let v = self.update_hlcv(scalar.value, scalar.value, scalar.value, 1.0);
        vec![Box::new(Scalar::new(scalar.time, v))]
    }

    fn update_bar(&mut self, bar: &Bar) -> Vec<Box<dyn Any>> {
        let v = self.update_hlcv(bar.high, bar.low, bar.close, bar.volume);
        vec![Box::new(Scalar::new(bar.time, v))]
    }

    fn update_quote(&mut self, quote: &Quote) -> Vec<Box<dyn Any>> {
        let mid = quote.mid();
        let v = self.update_hlcv(mid, mid, mid, 1.0);
        vec![Box::new(Scalar::new(quote.time, v))]
    }

    fn update_trade(&mut self, trade: &Trade) -> Vec<Box<dyn Any>> {
        let v = self.update_hlcv(trade.price, trade.price, trade.price, 1.0);
        vec![Box::new(Scalar::new(trade.time, v))]
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
    fn test_advance_decline_oscillator_ema() {
        let highs = testdata::test_highs();
        let lows = testdata::test_lows();
        let closes = testdata::test_closes();
        let volumes = testdata::test_volumes();
        let expected = testdata::test_expected_ema();

        let mut adosc = AdvanceDeclineOscillator::new(&AdvanceDeclineOscillatorParams {
            fast_length: 3,
            slow_length: 10,
            moving_average_type: AdvanceDeclineOscillatorMaType::EMA,
            first_is_average: false,
        })
        .unwrap();

        for i in 0..highs.len() {
            let v = adosc.update_hlcv(highs[i], lows[i], closes[i], volumes[i]);

            if i < 9 {
                assert!(v.is_nan(), "[{}] expected NaN, got {}", i, v);
                assert!(!adosc.is_primed(), "[{}] expected not primed", i);
                continue;
            }

            assert!(!v.is_nan(), "[{}] expected non-NaN", i);
            assert!(adosc.is_primed(), "[{}] expected primed", i);

            let got = round_to(v, 2);
            let exp = round_to(expected[i], 2);
            assert_eq!(got, exp, "[{}] expected {}, got {}", i, exp, got);
        }
    }

    #[test]
    fn test_advance_decline_oscillator_sma() {
        let highs = testdata::test_highs();
        let lows = testdata::test_lows();
        let closes = testdata::test_closes();
        let volumes = testdata::test_volumes();
        let expected = testdata::test_expected_sma();

        let mut adosc = AdvanceDeclineOscillator::new(&AdvanceDeclineOscillatorParams {
            fast_length: 3,
            slow_length: 10,
            moving_average_type: AdvanceDeclineOscillatorMaType::SMA,
            first_is_average: false,
        })
        .unwrap();

        for i in 0..highs.len() {
            let v = adosc.update_hlcv(highs[i], lows[i], closes[i], volumes[i]);

            if i < 9 {
                assert!(v.is_nan(), "[{}] expected NaN, got {}", i, v);
                assert!(!adosc.is_primed(), "[{}] expected not primed", i);
                continue;
            }

            assert!(!v.is_nan(), "[{}] expected non-NaN", i);
            assert!(adosc.is_primed(), "[{}] expected primed", i);

            let got = round_to(v, 2);
            let exp = round_to(expected[i], 2);
            assert_eq!(got, exp, "[{}] expected {}, got {}", i, exp, got);
        }
    }

    #[test]
    fn test_advance_decline_oscillator_invalid_params() {
        let r = AdvanceDeclineOscillator::new(&AdvanceDeclineOscillatorParams {
            fast_length: 1,
            slow_length: 10,
            ..Default::default()
        });
        assert!(r.is_err());

        let r = AdvanceDeclineOscillator::new(&AdvanceDeclineOscillatorParams {
            fast_length: 3,
            slow_length: 1,
            ..Default::default()
        });
        assert!(r.is_err());
    }
}
