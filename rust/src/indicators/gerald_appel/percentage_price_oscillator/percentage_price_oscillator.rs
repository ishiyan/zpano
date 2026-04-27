use crate::entities::bar::Bar;
use crate::entities::bar_component::{component_value as bar_component_value, BarComponent, DEFAULT_BAR_COMPONENT};
use crate::entities::quote::Quote;
use crate::entities::quote_component::{component_value as quote_component_value, QuoteComponent, DEFAULT_QUOTE_COMPONENT};
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::entities::trade_component::{component_value as trade_component_value, TradeComponent, DEFAULT_TRADE_COMPONENT};
use crate::indicators::common::exponential_moving_average::{ExponentialMovingAverage, ExponentialMovingAverageLengthParams};
use crate::indicators::common::simple_moving_average::{SimpleMovingAverage, SimpleMovingAverageParams};
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::component_triple_mnemonic::component_triple_mnemonic;
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::line_indicator::LineIndicator;
use crate::indicators::core::metadata::Metadata;

// ---------------------------------------------------------------------------
// MovingAverageType
// ---------------------------------------------------------------------------

/// Specifies the type of moving average to use in the PPO calculation.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum MovingAverageType {
    /// Simple Moving Average (default).
    Sma,
    /// Exponential Moving Average.
    Ema,
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the PPO indicator.
pub struct PercentagePriceOscillatorParams {
    /// Fast moving average length. Must be > 1.
    pub fast_length: usize,
    /// Slow moving average length. Must be > 1.
    pub slow_length: usize,
    /// MA type (SMA or EMA). Default SMA.
    pub moving_average_type: MovingAverageType,
    /// EMA seeding: true = TA-Lib (SMA seed), false = Metastock. Only relevant for EMA.
    pub first_is_average: bool,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for PercentagePriceOscillatorParams {
    fn default() -> Self {
        Self {
            fast_length: 12,
            slow_length: 26,
            moving_average_type: MovingAverageType::Sma,
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

/// Enumerates the outputs of the PPO indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum PercentagePriceOscillatorOutput {
    /// The scalar value of the percentage price oscillator.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Internal MA abstraction
// ---------------------------------------------------------------------------

enum MaVariant {
    Ema(ExponentialMovingAverage),
    Sma(SimpleMovingAverage),
}

impl MaVariant {
    fn update(&mut self, sample: f64) -> f64 {
        match self {
            MaVariant::Ema(ema) => ema.update(sample),
            MaVariant::Sma(sma) => sma.update(sample),
        }
    }

    fn is_primed(&self) -> bool {
        match self {
            MaVariant::Ema(ema) => Indicator::is_primed(ema),
            MaVariant::Sma(sma) => Indicator::is_primed(sma),
        }
    }
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Gerald Appel's Percentage Price Oscillator (PPO).
///
/// PPO = 100 * (fast_ma - slow_ma) / slow_ma
pub struct PercentagePriceOscillator {
    line: LineIndicator,
    fast_ma: MaVariant,
    slow_ma: MaVariant,
    value: f64,
    primed: bool,
}

impl PercentagePriceOscillator {
    /// Creates a new PPO from the given parameters.
    pub fn new(params: &PercentagePriceOscillatorParams) -> Result<Self, String> {
        let invalid = "invalid percentage price oscillator parameters";

        if params.fast_length < 2 {
            return Err(format!("{}: fast length should be greater than 1", invalid));
        }
        if params.slow_length < 2 {
            return Err(format!("{}: slow length should be greater than 1", invalid));
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let ma_label = match params.moving_average_type {
            MovingAverageType::Ema => "EMA",
            MovingAverageType::Sma => "SMA",
        };

        let (fast_ma, slow_ma) = match params.moving_average_type {
            MovingAverageType::Ema => {
                let fast = ExponentialMovingAverage::new_from_length(&ExponentialMovingAverageLengthParams {
                    length: params.fast_length as i64,
                    first_is_average: params.first_is_average,
                    bar_component: None,
                    quote_component: None,
                    trade_component: None,
                })?;
                let slow = ExponentialMovingAverage::new_from_length(&ExponentialMovingAverageLengthParams {
                    length: params.slow_length as i64,
                    first_is_average: params.first_is_average,
                    bar_component: None,
                    quote_component: None,
                    trade_component: None,
                })?;
                (MaVariant::Ema(fast), MaVariant::Ema(slow))
            }
            MovingAverageType::Sma => {
                let fast = SimpleMovingAverage::new(&SimpleMovingAverageParams {
                    length: params.fast_length,
                    bar_component: None,
                    quote_component: None,
                    trade_component: None,
                })?;
                let slow = SimpleMovingAverage::new(&SimpleMovingAverageParams {
                    length: params.slow_length,
                    bar_component: None,
                    quote_component: None,
                    trade_component: None,
                })?;
                (MaVariant::Sma(fast), MaVariant::Sma(slow))
            }
        };

        let mnemonic = format!(
            "ppo({}{}/{}{}{})",
            ma_label, params.fast_length, ma_label, params.slow_length,
            component_triple_mnemonic(bc, qc, tc)
        );
        let description = format!("Percentage Price Oscillator {}", mnemonic);

        let line = LineIndicator::new(mnemonic.clone(), description.clone(), bar_func, quote_func, trade_func);

        Ok(Self {
            line,
            fast_ma,
            slow_ma,
            value: f64::NAN,
            primed: false,
        })
    }

    /// Core update logic. Returns the PPO value or NaN if not yet primed.
    pub fn update(&mut self, sample: f64) -> f64 {
        const EPSILON: f64 = 1e-8;

        if sample.is_nan() {
            return sample;
        }

        let slow = self.slow_ma.update(sample);
        let fast = self.fast_ma.update(sample);
        self.primed = self.slow_ma.is_primed() && self.fast_ma.is_primed();

        if fast.is_nan() || slow.is_nan() {
            self.value = f64::NAN;
            return self.value;
        }

        if slow.abs() < EPSILON {
            self.value = 0.0;
        } else {
            self.value = 100.0 * (fast - slow) / slow;
        }

        self.value
    }
}

impl Indicator for PercentagePriceOscillator {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::PercentagePriceOscillator,
            &self.line.mnemonic,
            &self.line.description,
            &[OutputText {
                mnemonic: self.line.mnemonic.clone(),
                description: self.line.description.clone(),
            }],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let v = self.update(sample.value);
        vec![Box::new(Scalar { time: sample.time, value: v })]
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let v = (self.line.bar_func)(sample);
        self.update_scalar(&Scalar { time: sample.time, value: v })
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (self.line.quote_func)(sample);
        self.update_scalar(&Scalar { time: sample.time, value: v })
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let v = (self.line.trade_func)(sample);
        self.update_scalar(&Scalar { time: sample.time, value: v })
    }
}

// ===========================================================================
// Tests
// ===========================================================================

#[cfg(test)]
mod tests {
    use super::*;
    use crate::indicators::core::indicator::Indicator;

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
    fn test_sma_3_2() {
        let tolerance = 5e-4;
        let input = test_input();

        let mut ppo = PercentagePriceOscillator::new(&PercentagePriceOscillatorParams {
            fast_length: 2,
            slow_length: 3,
            ..Default::default()
        }).unwrap();

        // First 2 values should be NaN.
        for i in 0..2 {
            let v = ppo.update(input[i]);
            assert!(v.is_nan(), "[{}] expected NaN, got {}", i, v);
        }

        // Index 2: first value.
        let v = ppo.update(input[2]);
        assert!(!v.is_nan(), "[2] expected non-NaN");
        assert!((v - 1.10264).abs() < tolerance, "[2] expected ~1.10264, got {}", v);

        // Index 3.
        let v = ppo.update(input[3]);
        assert!((v - (-0.02813)).abs() < tolerance, "[3] expected ~-0.02813, got {}", v);

        // Feed remaining.
        for i in 4..251 {
            ppo.update(input[i]);
        }

        let v = ppo.update(input[251]);
        assert!((v - (-0.21191)).abs() < tolerance, "[251] expected ~-0.21191, got {}", v);
        assert!(ppo.is_primed());
    }

    #[test]
    fn test_sma_26_12() {
        let tolerance = 5e-4;
        let input = test_input();

        let mut ppo = PercentagePriceOscillator::new(&PercentagePriceOscillatorParams {
            fast_length: 12,
            slow_length: 26,
            ..Default::default()
        }).unwrap();

        for i in 0..25 {
            let v = ppo.update(input[i]);
            assert!(v.is_nan(), "[{}] expected NaN", i);
        }

        let v = ppo.update(input[25]);
        assert!(!v.is_nan());
        assert!((v - (-3.6393)).abs() < tolerance, "[25] expected ~-3.6393, got {}", v);

        let v = ppo.update(input[26]);
        assert!((v - (-3.9534)).abs() < tolerance, "[26] expected ~-3.9534, got {}", v);

        for i in 27..251 {
            ppo.update(input[i]);
        }

        let v = ppo.update(input[251]);
        assert!((v - (-0.15281)).abs() < tolerance, "[251] expected ~-0.15281, got {}", v);
    }

    #[test]
    fn test_ema_26_12() {
        let tolerance = 5e-3;
        let input = test_input();

        let mut ppo = PercentagePriceOscillator::new(&PercentagePriceOscillatorParams {
            fast_length: 12,
            slow_length: 26,
            moving_average_type: MovingAverageType::Ema,
            first_is_average: false,
            ..Default::default()
        }).unwrap();

        for i in 0..25 {
            let v = ppo.update(input[i]);
            assert!(v.is_nan(), "[{}] expected NaN", i);
        }

        let v = ppo.update(input[25]);
        assert!(!v.is_nan());
        assert!((v - (-2.7083)).abs() < tolerance, "[25] expected ~-2.7083, got {}", v);

        let v = ppo.update(input[26]);
        assert!((v - (-2.7390)).abs() < tolerance, "[26] expected ~-2.7390, got {}", v);

        for i in 27..251 {
            ppo.update(input[i]);
        }

        let v = ppo.update(input[251]);
        assert!((v - 0.83644).abs() < tolerance, "[251] expected ~0.83644, got {}", v);
    }

    #[test]
    fn test_is_primed() {
        let mut ppo = PercentagePriceOscillator::new(&PercentagePriceOscillatorParams {
            fast_length: 3,
            slow_length: 5,
            ..Default::default()
        }).unwrap();

        assert!(!ppo.is_primed());

        for i in 1..5 {
            ppo.update(i as f64);
            assert!(!ppo.is_primed(), "[{}] expected not primed", i);
        }

        ppo.update(5.0);
        assert!(ppo.is_primed());

        for i in 6..10 {
            ppo.update(i as f64);
            assert!(ppo.is_primed(), "[{}] expected primed", i);
        }
    }

    #[test]
    fn test_nan() {
        let mut ppo = PercentagePriceOscillator::new(&PercentagePriceOscillatorParams {
            fast_length: 2,
            slow_length: 3,
            ..Default::default()
        }).unwrap();

        let v = ppo.update(f64::NAN);
        assert!(v.is_nan());
    }

    #[test]
    fn test_metadata() {
        let ppo = PercentagePriceOscillator::new(&PercentagePriceOscillatorParams {
            fast_length: 12,
            slow_length: 26,
            ..Default::default()
        }).unwrap();

        let meta = ppo.metadata();
        assert_eq!(meta.identifier, Identifier::PercentagePriceOscillator);
        assert_eq!(meta.mnemonic, "ppo(SMA12/SMA26)");
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].kind, PercentagePriceOscillatorOutput::Value as i32);
    }

    #[test]
    fn test_metadata_ema() {
        let ppo = PercentagePriceOscillator::new(&PercentagePriceOscillatorParams {
            fast_length: 12,
            slow_length: 26,
            moving_average_type: MovingAverageType::Ema,
            ..Default::default()
        }).unwrap();

        let meta = ppo.metadata();
        assert_eq!(meta.mnemonic, "ppo(EMA12/EMA26)");
    }

    #[test]
    fn test_invalid_params() {
        assert!(PercentagePriceOscillator::new(&PercentagePriceOscillatorParams {
            fast_length: 1, slow_length: 26, ..Default::default()
        }).is_err());

        assert!(PercentagePriceOscillator::new(&PercentagePriceOscillatorParams {
            fast_length: 12, slow_length: 1, ..Default::default()
        }).is_err());
    }
}
