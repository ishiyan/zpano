use crate::entities::bar::Bar;
use crate::entities::bar_component::{component_value as bar_component_value, BarComponent, DEFAULT_BAR_COMPONENT};
use crate::entities::quote::Quote;
use crate::entities::quote_component::{component_value as quote_component_value, QuoteComponent, DEFAULT_QUOTE_COMPONENT};
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::entities::trade_component::{component_value as trade_component_value, TradeComponent, DEFAULT_TRADE_COMPONENT};
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::component_triple_mnemonic::component_triple_mnemonic;
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::line_indicator::{BarFunc, QuoteFunc, TradeFunc};
use crate::indicators::core::metadata::Metadata;
use crate::indicators::common::simple_moving_average::{SimpleMovingAverage, SimpleMovingAverageParams};
use crate::indicators::common::exponential_moving_average::{ExponentialMovingAverage, ExponentialMovingAverageLengthParams};
use crate::indicators::welles_wilder::relative_strength_index::{RelativeStrengthIndex, RelativeStrengthIndexParams};

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Type of moving average for Fast-D smoothing.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum MovingAverageType {
    SMA,
    EMA,
}

impl Default for MovingAverageType {
    fn default() -> Self {
        Self::SMA
    }
}

/// Parameters for the Stochastic RSI indicator.
pub struct StochasticRelativeStrengthIndexParams {
    /// RSI length. Must be >= 2. Default is 14.
    pub length: usize,
    /// Fast-K stochastic length. Must be >= 1. Default is 5.
    pub fast_k_length: usize,
    /// Fast-D smoothing length. Must be >= 1. Default is 3.
    pub fast_d_length: usize,
    /// Moving average type for Fast-D.
    pub moving_average_type: MovingAverageType,
    /// EMA seeding: true = first value is SMA of first period.
    pub first_is_average: bool,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for StochasticRelativeStrengthIndexParams {
    fn default() -> Self {
        Self {
            length: 14,
            fast_k_length: 5,
            fast_d_length: 3,
            moving_average_type: MovingAverageType::SMA,
            first_is_average: false,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Stochastic RSI indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum StochasticRelativeStrengthIndexOutput {
    /// The Fast-K line.
    FastK = 1,
    /// The Fast-D line (smoothed Fast-K).
    FastD = 2,
}

// ---------------------------------------------------------------------------
// Fast-D smoother abstraction
// ---------------------------------------------------------------------------

enum FastDSmoother {
    Passthrough,
    Sma(SimpleMovingAverage),
    Ema(ExponentialMovingAverage),
}

impl FastDSmoother {
    fn update(&mut self, v: f64) -> f64 {
        match self {
            Self::Passthrough => v,
            Self::Sma(sma) => sma.update(v),
            Self::Ema(ema) => ema.update(v),
        }
    }

    fn is_primed(&self) -> bool {
        match self {
            Self::Passthrough => true,
            Self::Sma(sma) => sma.is_primed(),
            Self::Ema(ema) => ema.is_primed(),
        }
    }
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Tushar Chande's Stochastic RSI.
///
/// Applies the Stochastic oscillator formula to RSI values. Produces
/// Fast-K and Fast-D outputs oscillating between 0 and 100.
pub struct StochasticRelativeStrengthIndex {
    rsi: RelativeStrengthIndex,
    rsi_buf: Vec<f64>,
    rsi_buffer_index: usize,
    rsi_count: usize,
    fast_k_length: usize,
    fast_d_ma: FastDSmoother,
    fast_k: f64,
    fast_d: f64,
    primed: bool,
    bar_func: BarFunc,
    quote_func: QuoteFunc,
    trade_func: TradeFunc,
    mnemonic: String,
}

impl StochasticRelativeStrengthIndex {
    /// Creates a new StochasticRelativeStrengthIndex from the given parameters.
    pub fn new(params: &StochasticRelativeStrengthIndexParams) -> Result<Self, String> {
        let invalid = "invalid stochastic relative strength index parameters";

        if params.length < 2 {
            return Err(format!("{}: length should be greater than 1", invalid));
        }
        if params.fast_k_length < 1 {
            return Err(format!("{}: fast K length should be greater than 0", invalid));
        }
        if params.fast_d_length < 1 {
            return Err(format!("{}: fast D length should be greater than 0", invalid));
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let rsi = RelativeStrengthIndex::new(&RelativeStrengthIndexParams {
            length: params.length,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }).map_err(|e| format!("{}: {}", invalid, e))?;

        let (fast_d_ma, ma_label) = if params.fast_d_length < 2 {
            (FastDSmoother::Passthrough, "SMA")
        } else {
            match params.moving_average_type {
                MovingAverageType::EMA => {
                    let ema = ExponentialMovingAverage::new_from_length(
                        &ExponentialMovingAverageLengthParams {
                            length: params.fast_d_length as i64,
                            first_is_average: params.first_is_average,
                            ..Default::default()
                        },
                    ).map_err(|e| format!("{}: {}", invalid, e))?;
                    (FastDSmoother::Ema(ema), "EMA")
                }
                MovingAverageType::SMA => {
                    let sma = SimpleMovingAverage::new(
                        &SimpleMovingAverageParams {
                            length: params.fast_d_length,
                            ..Default::default()
                        },
                    ).map_err(|e| format!("{}: {}", invalid, e))?;
                    (FastDSmoother::Sma(sma), "SMA")
                }
            }
        };

        let mnemonic = format!(
            "stochrsi({}/{}/{}{}{})",
            params.length,
            params.fast_k_length,
            ma_label,
            params.fast_d_length,
            component_triple_mnemonic(bc, qc, tc),
        );

        Ok(Self {
            rsi,
            rsi_buf: vec![0.0; params.fast_k_length],
            rsi_buffer_index: 0,
            rsi_count: 0,
            fast_k_length: params.fast_k_length,
            fast_d_ma,
            fast_k: f64::NAN,
            fast_d: f64::NAN,
            primed: false,
            bar_func,
            quote_func,
            trade_func,
            mnemonic,
        })
    }

    /// Core update with a single sample value. Returns (fast_k, fast_d).
    pub fn update(&mut self, sample: f64) -> (f64, f64) {
        if sample.is_nan() {
            return (f64::NAN, f64::NAN);
        }

        let rsi_value = self.rsi.update(sample);
        if rsi_value.is_nan() {
            return (self.fast_k, self.fast_d);
        }

        self.rsi_buf[self.rsi_buffer_index] = rsi_value;
        self.rsi_buffer_index = (self.rsi_buffer_index + 1) % self.fast_k_length;
        self.rsi_count += 1;

        if self.rsi_count < self.fast_k_length {
            return (self.fast_k, self.fast_d);
        }

        // Find min and max RSI in window.
        let mut min_rsi = self.rsi_buf[0];
        let mut max_rsi = self.rsi_buf[0];
        for i in 1..self.fast_k_length {
            if self.rsi_buf[i] < min_rsi {
                min_rsi = self.rsi_buf[i];
            }
            if self.rsi_buf[i] > max_rsi {
                max_rsi = self.rsi_buf[i];
            }
        }

        let diff = max_rsi - min_rsi;
        self.fast_k = if diff > 0.0 {
            100.0 * (rsi_value - min_rsi) / diff
        } else {
            0.0
        };

        self.fast_d = self.fast_d_ma.update(self.fast_k);

        if !self.primed && self.fast_d_ma.is_primed() {
            self.primed = true;
        }

        (self.fast_k, self.fast_d)
    }
}

impl Indicator for StochasticRelativeStrengthIndex {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        let desc = format!("Stochastic Relative Strength Index {}", self.mnemonic);
        build_metadata(
            Identifier::StochasticRelativeStrengthIndex,
            &self.mnemonic,
            &desc,
            &[
                OutputText {
                    mnemonic: format!("{} fastK", self.mnemonic),
                    description: format!("{} Fast-K", desc),
                },
                OutputText {
                    mnemonic: format!("{} fastD", self.mnemonic),
                    description: format!("{} Fast-D", desc),
                },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let (fast_k, fast_d) = self.update(sample.value);
        vec![
            Box::new(Scalar::new(sample.time, fast_k)),
            Box::new(Scalar::new(sample.time, fast_d)),
        ]
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let v = (self.bar_func)(sample);
        let (fast_k, fast_d) = self.update(v);
        vec![
            Box::new(Scalar::new(sample.time, fast_k)),
            Box::new(Scalar::new(sample.time, fast_d)),
        ]
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (self.quote_func)(sample);
        let (fast_k, fast_d) = self.update(v);
        vec![
            Box::new(Scalar::new(sample.time, fast_k)),
            Box::new(Scalar::new(sample.time, fast_d)),
        ]
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let v = (self.trade_func)(sample);
        let (fast_k, fast_d) = self.update(v);
        vec![
            Box::new(Scalar::new(sample.time, fast_k)),
            Box::new(Scalar::new(sample.time, fast_d)),
        ]
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::testdata;
    #[test]
    fn test_stoch_rsi_14_14_1_sma() {
        let tolerance = 1e-4;
        let input = testdata::test_input();

        let mut ind = StochasticRelativeStrengthIndex::new(&StochasticRelativeStrengthIndexParams {
            length: 14,
            fast_k_length: 14,
            fast_d_length: 1,
            ..Default::default()
        }).unwrap();

        for i in 0..27 {
            let (fast_k, _) = ind.update(input[i]);
            assert!(fast_k.is_nan(), "[{}] expected NaN FastK, got {}", i, fast_k);
        }

        let (fast_k, fast_d) = ind.update(input[27]);
        assert!(!fast_k.is_nan(), "[27] expected non-NaN FastK");
        assert!((fast_k - 94.156709).abs() < tolerance, "[27] FastK: expected ~94.156709, got {}", fast_k);
        assert!((fast_d - 94.156709).abs() < tolerance, "[27] FastD: expected ~94.156709, got {}", fast_d);

        for i in 28..251 {
            ind.update(input[i]);
        }

        let (fast_k, fast_d) = ind.update(input[251]);
        assert!((fast_k - 0.0).abs() < tolerance, "[251] FastK: expected ~0.0, got {}", fast_k);
        assert!((fast_d - 0.0).abs() < tolerance, "[251] FastD: expected ~0.0, got {}", fast_d);
    }

    #[test]
    fn test_stoch_rsi_14_45_1_sma() {
        let tolerance = 1e-4;
        let input = testdata::test_input();

        let mut ind = StochasticRelativeStrengthIndex::new(&StochasticRelativeStrengthIndexParams {
            length: 14,
            fast_k_length: 45,
            fast_d_length: 1,
            ..Default::default()
        }).unwrap();

        for i in 0..58 {
            let (fast_k, _) = ind.update(input[i]);
            assert!(fast_k.is_nan(), "[{}] expected NaN FastK, got {}", i, fast_k);
        }

        let (fast_k, fast_d) = ind.update(input[58]);
        assert!(!fast_k.is_nan(), "[58] expected non-NaN FastK");
        assert!((fast_k - 79.729186).abs() < tolerance, "[58] FastK: expected ~79.729186, got {}", fast_k);
        assert!((fast_d - 79.729186).abs() < tolerance, "[58] FastD: expected ~79.729186, got {}", fast_d);

        for i in 59..251 {
            ind.update(input[i]);
        }

        let (fast_k, fast_d) = ind.update(input[251]);
        assert!((fast_k - 48.1550743).abs() < tolerance, "[251] FastK: expected ~48.1550743, got {}", fast_k);
        assert!((fast_d - 48.1550743).abs() < tolerance, "[251] FastD: expected ~48.1550743, got {}", fast_d);
    }

    #[test]
    fn test_stoch_rsi_11_13_16_sma() {
        let tolerance = 1e-3;
        let input = testdata::test_input();

        let mut ind = StochasticRelativeStrengthIndex::new(&StochasticRelativeStrengthIndexParams {
            length: 11,
            fast_k_length: 13,
            fast_d_length: 16,
            ..Default::default()
        }).unwrap();

        for i in 0..38 {
            ind.update(input[i]);
        }

        let (fast_k, fast_d) = ind.update(input[38]);
        assert!((fast_k - 5.25947).abs() < tolerance, "[38] FastK: expected ~5.25947, got {}", fast_k);
        assert!((fast_d - 57.1711).abs() < tolerance, "[38] FastD: expected ~57.1711, got {}", fast_d);
        assert!(ind.is_primed());

        for i in 39..251 {
            ind.update(input[i]);
        }

        let (fast_k, fast_d) = ind.update(input[251]);
        assert!((fast_k - 0.0).abs() < tolerance, "[251] FastK: expected ~0.0, got {}", fast_k);
        assert!((fast_d - 15.7303).abs() < tolerance, "[251] FastD: expected ~15.7303, got {}", fast_d);
    }

    #[test]
    fn test_stoch_rsi_is_primed() {
        let input = testdata::test_input();
        let mut ind = StochasticRelativeStrengthIndex::new(&StochasticRelativeStrengthIndexParams {
            length: 14,
            fast_k_length: 14,
            fast_d_length: 1,
            ..Default::default()
        }).unwrap();

        assert!(!ind.is_primed());

        for i in 0..27 {
            ind.update(input[i]);
            assert!(!ind.is_primed(), "[{}] expected not primed", i);
        }

        ind.update(input[27]);
        assert!(ind.is_primed());
    }

    #[test]
    fn test_stoch_rsi_nan() {
        let mut ind = StochasticRelativeStrengthIndex::new(&StochasticRelativeStrengthIndexParams {
            length: 14,
            fast_k_length: 14,
            fast_d_length: 1,
            ..Default::default()
        }).unwrap();

        let (fast_k, fast_d) = ind.update(f64::NAN);
        assert!(fast_k.is_nan());
        assert!(fast_d.is_nan());
    }

    #[test]
    fn test_stoch_rsi_metadata() {
        let ind = StochasticRelativeStrengthIndex::new(&StochasticRelativeStrengthIndexParams {
            length: 14,
            fast_k_length: 14,
            fast_d_length: 3,
            ..Default::default()
        }).unwrap();

        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::StochasticRelativeStrengthIndex);
        assert_eq!(meta.mnemonic, "stochrsi(14/14/SMA3)");
        assert_eq!(meta.outputs.len(), 2);
        assert_eq!(meta.outputs[0].kind, StochasticRelativeStrengthIndexOutput::FastK as i32);
        assert_eq!(meta.outputs[1].kind, StochasticRelativeStrengthIndexOutput::FastD as i32);
    }

    #[test]
    fn test_stoch_rsi_invalid_params() {
        assert!(StochasticRelativeStrengthIndex::new(&StochasticRelativeStrengthIndexParams { length: 1, fast_k_length: 14, fast_d_length: 3, ..Default::default() }).is_err());
        assert!(StochasticRelativeStrengthIndex::new(&StochasticRelativeStrengthIndexParams { length: 14, fast_k_length: 0, fast_d_length: 3, ..Default::default() }).is_err());
        assert!(StochasticRelativeStrengthIndex::new(&StochasticRelativeStrengthIndexParams { length: 14, fast_k_length: 14, fast_d_length: 0, ..Default::default() }).is_err());
    }
}
