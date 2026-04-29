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

    fn test_input() -> Vec<f64> {
        vec![
            91.500000, 94.815000, 94.375000, 95.095000, 93.780000, 94.625000, 92.530000, 92.750000, 90.315000, 92.470000, 96.125000,
            97.250000, 98.500000, 89.875000, 91.000000, 92.815000, 89.155000, 89.345000, 91.625000, 89.875000, 88.375000, 87.625000,
            84.780000, 83.000000, 83.500000, 81.375000, 84.440000, 89.250000, 86.375000, 86.250000, 85.250000, 87.125000, 85.815000,
            88.970000, 88.470000, 86.875000, 86.815000, 84.875000, 84.190000, 83.875000, 83.375000, 85.500000, 89.190000, 89.440000,
            91.095000, 90.750000, 91.440000, 89.000000, 91.000000, 90.500000, 89.030000, 88.815000, 84.280000, 83.500000, 82.690000,
            84.750000, 85.655000, 86.190000, 88.940000, 89.280000, 88.625000, 88.500000, 91.970000, 91.500000, 93.250000, 93.500000,
            93.155000, 91.720000, 90.000000, 89.690000, 88.875000, 85.190000, 83.375000, 84.875000, 85.940000, 97.250000, 99.875000,
            104.940000, 106.000000, 102.500000, 102.405000, 104.595000, 106.125000, 106.000000, 106.065000, 104.625000, 108.625000,
            109.315000, 110.500000, 112.750000, 123.000000, 119.625000, 118.750000, 119.250000, 117.940000, 116.440000, 115.190000,
            111.875000, 110.595000, 118.125000, 116.000000, 116.000000, 112.000000, 113.750000, 112.940000, 116.000000, 120.500000,
            116.620000, 117.000000, 115.250000, 114.310000, 115.500000, 115.870000, 120.690000, 120.190000, 120.750000, 124.750000,
            123.370000, 122.940000, 122.560000, 123.120000, 122.560000, 124.620000, 129.250000, 131.000000, 132.250000, 131.000000,
            132.810000, 134.000000, 137.380000, 137.810000, 137.880000, 137.250000, 136.310000, 136.250000, 134.630000, 128.250000,
            129.000000, 123.870000, 124.810000, 123.000000, 126.250000, 128.380000, 125.370000, 125.690000, 122.250000, 119.370000,
            118.500000, 123.190000, 123.500000, 122.190000, 119.310000, 123.310000, 121.120000, 123.370000, 127.370000, 128.500000,
            123.870000, 122.940000, 121.750000, 124.440000, 122.000000, 122.370000, 122.940000, 124.000000, 123.190000, 124.560000,
            127.250000, 125.870000, 128.860000, 132.000000, 130.750000, 134.750000, 135.000000, 132.380000, 133.310000, 131.940000,
            130.000000, 125.370000, 130.130000, 127.120000, 125.190000, 122.000000, 125.000000, 123.000000, 123.500000, 120.060000,
            121.000000, 117.750000, 119.870000, 122.000000, 119.190000, 116.370000, 113.500000, 114.250000, 110.000000, 105.060000,
            107.000000, 107.870000, 107.000000, 107.120000, 107.000000, 91.000000, 93.940000, 93.870000, 95.500000, 93.000000,
            94.940000, 98.250000, 96.750000, 94.810000, 94.370000, 91.560000, 90.250000, 93.940000, 93.620000, 97.000000, 95.000000,
            95.870000, 94.060000, 94.620000, 93.750000, 98.000000, 103.940000, 107.870000, 106.060000, 104.500000, 105.000000,
            104.190000, 103.060000, 103.420000, 105.270000, 111.870000, 116.000000, 116.620000, 118.280000, 113.370000, 109.000000,
            109.700000, 109.250000, 107.000000, 109.190000, 110.000000, 109.200000, 110.120000, 108.000000, 108.620000, 109.750000,
            109.810000, 109.000000, 108.750000, 107.870000,
        ]
    }

    #[test]
    fn test_stoch_rsi_14_14_1_sma() {
        let tolerance = 1e-4;
        let input = test_input();

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
        let input = test_input();

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
        let input = test_input();

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
        let input = test_input();
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
