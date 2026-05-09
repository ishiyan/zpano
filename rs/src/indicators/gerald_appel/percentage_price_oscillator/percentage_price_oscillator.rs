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
    use super::super::testdata::testdata;
    use crate::indicators::core::indicator::Indicator;
    #[test]
    fn test_sma_3_2() {
        let tolerance = 5e-4;
        let input = testdata::test_input();

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
        let input = testdata::test_input();

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
        let input = testdata::test_input();

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
