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

#[repr(u8)]
#[derive(Debug, Clone, Copy, PartialEq)]
pub enum StochasticOutput {
    FastK = 1,
    SlowK = 2,
    SlowD = 3,
}

#[repr(u8)]
#[derive(Debug, Clone, Copy, PartialEq)]
pub enum StochasticMaType {
    SMA = 0,
    EMA = 1,
}

#[derive(Debug, Clone)]
pub struct StochasticParams {
    pub fast_k_length: usize,
    pub slow_k_length: usize,
    pub slow_d_length: usize,
    pub slow_k_ma_type: StochasticMaType,
    pub slow_d_ma_type: StochasticMaType,
    pub first_is_average: bool,
}

impl Default for StochasticParams {
    fn default() -> Self {
        Self {
            fast_k_length: 5,
            slow_k_length: 3,
            slow_d_length: 3,
            slow_k_ma_type: StochasticMaType::SMA,
            slow_d_ma_type: StochasticMaType::SMA,
            first_is_average: false,
        }
    }
}

trait LineUpdater {
    fn update(&mut self, sample: f64) -> f64;
    fn is_primed(&self) -> bool;
}

struct Passthrough;
impl LineUpdater for Passthrough {
    fn update(&mut self, sample: f64) -> f64 { sample }
    fn is_primed(&self) -> bool { true }
}

struct SmaUpdater { inner: SimpleMovingAverage }
impl LineUpdater for SmaUpdater {
    fn update(&mut self, sample: f64) -> f64 { self.inner.update(sample) }
    fn is_primed(&self) -> bool { self.inner.is_primed() }
}

struct EmaUpdater { inner: ExponentialMovingAverage }
impl LineUpdater for EmaUpdater {
    fn update(&mut self, sample: f64) -> f64 { self.inner.update(sample) }
    fn is_primed(&self) -> bool { self.inner.is_primed() }
}

fn create_ma(
    ma_type: StochasticMaType,
    length: usize,
    first_is_average: bool,
) -> Result<Box<dyn LineUpdater>, String> {
    if length < 2 {
        return Ok(Box::new(Passthrough));
    }
    match ma_type {
        StochasticMaType::EMA => {
            let ema = ExponentialMovingAverage::new_from_length(
                &ExponentialMovingAverageLengthParams {
                    length: length as i64,
                    first_is_average,
                    ..Default::default()
                },
            )?;
            Ok(Box::new(EmaUpdater { inner: ema }))
        }
        StochasticMaType::SMA => {
            let sma = SimpleMovingAverage::new(&SimpleMovingAverageParams {
                length,
                ..Default::default()
            })?;
            Ok(Box::new(SmaUpdater { inner: sma }))
        }
    }
}

pub struct Stochastic {
    fast_k_length: usize,
    high_buf: Vec<f64>,
    low_buf: Vec<f64>,
    buffer_index: usize,
    count: usize,
    slow_k_ma: Box<dyn LineUpdater>,
    slow_d_ma: Box<dyn LineUpdater>,
    fast_k: f64,
    slow_k: f64,
    slow_d: f64,
    primed: bool,
}

impl Stochastic {
    pub fn new(params: &StochasticParams) -> Result<Self, String> {
        if params.fast_k_length < 1 {
            return Err("invalid stochastic parameters: fast K length should be greater than 0".to_string());
        }
        if params.slow_k_length < 1 {
            return Err("invalid stochastic parameters: slow K length should be greater than 0".to_string());
        }
        if params.slow_d_length < 1 {
            return Err("invalid stochastic parameters: slow D length should be greater than 0".to_string());
        }

        let slow_k_ma = create_ma(params.slow_k_ma_type, params.slow_k_length, params.first_is_average)?;
        let slow_d_ma = create_ma(params.slow_d_ma_type, params.slow_d_length, params.first_is_average)?;

        Ok(Self {
            fast_k_length: params.fast_k_length,
            high_buf: vec![0.0; params.fast_k_length],
            low_buf: vec![0.0; params.fast_k_length],
            buffer_index: 0,
            count: 0,
            slow_k_ma,
            slow_d_ma,
            fast_k: f64::NAN,
            slow_k: f64::NAN,
            slow_d: f64::NAN,
            primed: false,
        })
    }

    pub fn update_hlc(&mut self, close: f64, high: f64, low: f64) -> (f64, f64, f64) {
        if close.is_nan() || high.is_nan() || low.is_nan() {
            return (f64::NAN, f64::NAN, f64::NAN);
        }

        self.high_buf[self.buffer_index] = high;
        self.low_buf[self.buffer_index] = low;
        self.buffer_index = (self.buffer_index + 1) % self.fast_k_length;
        self.count += 1;

        if self.count < self.fast_k_length {
            return (self.fast_k, self.slow_k, self.slow_d);
        }

        // Find highest high and lowest low.
        let mut hh = self.high_buf[0];
        let mut ll = self.low_buf[0];
        for i in 1..self.fast_k_length {
            if self.high_buf[i] > hh { hh = self.high_buf[i]; }
            if self.low_buf[i] < ll { ll = self.low_buf[i]; }
        }

        let diff = hh - ll;
        self.fast_k = if diff > 0.0 { 100.0 * (close - ll) / diff } else { 0.0 };

        self.slow_k = self.slow_k_ma.update(self.fast_k);

        if self.slow_k_ma.is_primed() {
            self.slow_d = self.slow_d_ma.update(self.slow_k);
            if !self.primed && self.slow_d_ma.is_primed() {
                self.primed = true;
            }
        }

        (self.fast_k, self.slow_k, self.slow_d)
    }
}

impl Indicator for Stochastic {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::Stochastic,
            "stoch",
            "Stochastic Oscillator",
            &[
                OutputText { mnemonic: "fastK".to_string(), description: "Fast-K".to_string() },
                OutputText { mnemonic: "slowK".to_string(), description: "Slow-K".to_string() },
                OutputText { mnemonic: "slowD".to_string(), description: "Slow-D".to_string() },
            ],
        )
    }

    fn update_scalar(&mut self, scalar: &Scalar) -> Vec<Box<dyn Any>> {
        let v = scalar.value;
        let (fk, sk, sd) = self.update_hlc(v, v, v);
        vec![
            Box::new(Scalar::new(scalar.time, fk)),
            Box::new(Scalar::new(scalar.time, sk)),
            Box::new(Scalar::new(scalar.time, sd)),
        ]
    }

    fn update_bar(&mut self, bar: &Bar) -> Vec<Box<dyn Any>> {
        let (fk, sk, sd) = self.update_hlc(bar.close, bar.high, bar.low);
        vec![
            Box::new(Scalar::new(bar.time, fk)),
            Box::new(Scalar::new(bar.time, sk)),
            Box::new(Scalar::new(bar.time, sd)),
        ]
    }

    fn update_quote(&mut self, quote: &Quote) -> Vec<Box<dyn Any>> {
        let mid = quote.mid();
        let (fk, sk, sd) = self.update_hlc(mid, mid, mid);
        vec![
            Box::new(Scalar::new(quote.time, fk)),
            Box::new(Scalar::new(quote.time, sk)),
            Box::new(Scalar::new(quote.time, sd)),
        ]
    }

    fn update_trade(&mut self, trade: &Trade) -> Vec<Box<dyn Any>> {
        let (fk, sk, sd) = self.update_hlc(trade.price, trade.price, trade.price);
        vec![
            Box::new(Scalar::new(trade.time, fk)),
            Box::new(Scalar::new(trade.time, sk)),
            Box::new(Scalar::new(trade.time, sd)),
        ]
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::testdata;
    #[test]
    fn test_stochastic_5_sma3_sma4_single() {
        let high = testdata::test_high();
        let low = testdata::test_low();
        let close = testdata::test_close();

        let mut ind = Stochastic::new(&StochasticParams {
            fast_k_length: 5,
            slow_k_length: 3,
            slow_d_length: 4,
            ..Default::default()
        }).unwrap();

        for i in 0..9 {
            ind.update_hlc(close[i], high[i], low[i]);
        }

        let (_, slow_k, slow_d) = ind.update_hlc(close[9], high[9], low[9]);
        assert!((slow_k - 38.139).abs() < 1e-2, "SlowK: expected ~38.139, got {}", slow_k);
        assert!((slow_d - 36.725).abs() < 1e-2, "SlowD: expected ~36.725, got {}", slow_d);
        assert!(ind.is_primed());
    }

    #[test]
    fn test_stochastic_5_sma3_sma3_first() {
        let high = testdata::test_high();
        let low = testdata::test_low();
        let close = testdata::test_close();

        let mut ind = Stochastic::new(&StochasticParams {
            fast_k_length: 5,
            slow_k_length: 3,
            slow_d_length: 3,
            ..Default::default()
        }).unwrap();

        for i in 0..8 {
            ind.update_hlc(close[i], high[i], low[i]);
        }

        let (_, slow_k, slow_d) = ind.update_hlc(close[8], high[8], low[8]);
        assert!((slow_k - 24.0128).abs() < 1e-2, "SlowK: expected ~24.0128, got {}", slow_k);
        assert!((slow_d - 36.254).abs() < 1e-2, "SlowD: expected ~36.254, got {}", slow_d);
        assert!(ind.is_primed());
    }

    #[test]
    fn test_stochastic_5_sma3_sma3_last() {
        let high = testdata::test_high();
        let low = testdata::test_low();
        let close = testdata::test_close();

        let mut ind = Stochastic::new(&StochasticParams {
            fast_k_length: 5,
            slow_k_length: 3,
            slow_d_length: 3,
            ..Default::default()
        }).unwrap();

        let mut slow_k = 0.0;
        let mut slow_d = 0.0;
        for i in 0..252 {
            let (_, sk, sd) = ind.update_hlc(close[i], high[i], low[i]);
            slow_k = sk;
            slow_d = sd;
        }

        assert!((slow_k - 30.194).abs() < 1e-2, "SlowK: expected ~30.194, got {}", slow_k);
        assert!((slow_d - 43.69).abs() < 1e-2, "SlowD: expected ~43.69, got {}", slow_d);
    }

    #[test]
    fn test_stochastic_5_sma3_sma4_last() {
        let high = testdata::test_high();
        let low = testdata::test_low();
        let close = testdata::test_close();

        let mut ind = Stochastic::new(&StochasticParams {
            fast_k_length: 5,
            slow_k_length: 3,
            slow_d_length: 4,
            ..Default::default()
        }).unwrap();

        let mut slow_k = 0.0;
        let mut slow_d = 0.0;
        for i in 0..252 {
            let (_, sk, sd) = ind.update_hlc(close[i], high[i], low[i]);
            slow_k = sk;
            slow_d = sd;
        }

        assert!((slow_k - 30.194).abs() < 1e-2, "SlowK: expected ~30.194, got {}", slow_k);
        assert!((slow_d - 46.641).abs() < 1e-2, "SlowD: expected ~46.641, got {}", slow_d);
    }

    #[test]
    fn test_stochastic_is_primed() {
        let high = testdata::test_high();
        let low = testdata::test_low();
        let close = testdata::test_close();

        let mut ind = Stochastic::new(&StochasticParams {
            fast_k_length: 5,
            slow_k_length: 3,
            slow_d_length: 3,
            ..Default::default()
        }).unwrap();

        assert!(!ind.is_primed());

        for i in 0..8 {
            ind.update_hlc(close[i], high[i], low[i]);
            assert!(!ind.is_primed(), "[{}] expected not primed", i);
        }

        ind.update_hlc(close[8], high[8], low[8]);
        assert!(ind.is_primed(), "expected primed after index 8");
    }

    #[test]
    fn test_stochastic_invalid_params() {
        assert!(Stochastic::new(&StochasticParams { fast_k_length: 0, slow_k_length: 3, slow_d_length: 3, ..Default::default() }).is_err());
        assert!(Stochastic::new(&StochasticParams { fast_k_length: 5, slow_k_length: 0, slow_d_length: 3, ..Default::default() }).is_err());
        assert!(Stochastic::new(&StochasticParams { fast_k_length: 5, slow_k_length: 3, slow_d_length: 0, ..Default::default() }).is_err());
    }
}
