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
use crate::indicators::core::line_indicator::LineIndicator;
use crate::indicators::core::metadata::Metadata;

use crate::indicators::common::exponential_moving_average::{
    ExponentialMovingAverage, ExponentialMovingAverageLengthParams,
};
use crate::indicators::common::simple_moving_average::{
    SimpleMovingAverage, SimpleMovingAverageParams,
};

// ---------------------------------------------------------------------------
// MovingAverageType
// ---------------------------------------------------------------------------

/// Specifies the type of moving average to use in the APO calculation.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum MovingAverageType {
    /// Simple Moving Average.
    SMA = 0,
    /// Exponential Moving Average.
    EMA = 1,
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the Absolute Price Oscillator.
pub struct AbsolutePriceOscillatorParams {
    /// The number of periods for the fast moving average. Must be > 1.
    pub fast_length: i64,
    /// The number of periods for the slow moving average. Must be > 1.
    pub slow_length: i64,
    /// The type of moving average (SMA or EMA). Defaults to SMA.
    pub moving_average_type: MovingAverageType,
    /// Controls the EMA seeding algorithm (only relevant when using EMA).
    /// When true, the first EMA value is the simple average of the first period values.
    pub first_is_average: bool,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for AbsolutePriceOscillatorParams {
    fn default() -> Self {
        Self {
            fast_length: 12,
            slow_length: 26,
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

/// Enumerates the outputs of the absolute price oscillator indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum AbsolutePriceOscillatorOutput {
    /// The scalar value of the absolute price oscillator.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Internal trait for polymorphic MA
// ---------------------------------------------------------------------------

trait LineUpdater {
    fn update(&mut self, sample: f64) -> f64;
    fn is_primed(&self) -> bool;
}

impl LineUpdater for SimpleMovingAverage {
    fn update(&mut self, sample: f64) -> f64 {
        self.update(sample)
    }
    fn is_primed(&self) -> bool {
        Indicator::is_primed(self)
    }
}

impl LineUpdater for ExponentialMovingAverage {
    fn update(&mut self, sample: f64) -> f64 {
        self.update(sample)
    }
    fn is_primed(&self) -> bool {
        Indicator::is_primed(self)
    }
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Computes the Absolute Price Oscillator (APO).
///
/// APO = fast MA − slow MA.
pub struct AbsolutePriceOscillator {
    line: LineIndicator,
    fast_ma: Box<dyn LineUpdater>,
    slow_ma: Box<dyn LineUpdater>,
    value: f64,
    primed: bool,
}

impl AbsolutePriceOscillator {
    /// Creates a new Absolute Price Oscillator from the given parameters.
    pub fn new(params: &AbsolutePriceOscillatorParams) -> Result<Self, String> {
        const INVALID: &str = "invalid absolute price oscillator parameters";
        const MIN_LENGTH: i64 = 2;

        if params.fast_length < MIN_LENGTH {
            return Err(format!("{}: fast length should be greater than 1", INVALID));
        }
        if params.slow_length < MIN_LENGTH {
            return Err(format!("{}: slow length should be greater than 1", INVALID));
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let (fast_ma, slow_ma, ma_label): (Box<dyn LineUpdater>, Box<dyn LineUpdater>, &str) =
            match params.moving_average_type {
                MovingAverageType::EMA => {
                    let fast = ExponentialMovingAverage::new_from_length(
                        &ExponentialMovingAverageLengthParams {
                            length: params.fast_length,
                            first_is_average: params.first_is_average,
                            bar_component: None,
                            quote_component: None,
                            trade_component: None,
                        },
                    )
                    .map_err(|e| format!("{}: {}", INVALID, e))?;

                    let slow = ExponentialMovingAverage::new_from_length(
                        &ExponentialMovingAverageLengthParams {
                            length: params.slow_length,
                            first_is_average: params.first_is_average,
                            bar_component: None,
                            quote_component: None,
                            trade_component: None,
                        },
                    )
                    .map_err(|e| format!("{}: {}", INVALID, e))?;

                    (Box::new(fast), Box::new(slow), "EMA")
                }
                MovingAverageType::SMA => {
                    let fast = SimpleMovingAverage::new(&SimpleMovingAverageParams {
                        length: params.fast_length as usize,
                        bar_component: None,
                        quote_component: None,
                        trade_component: None,
                    })
                    .map_err(|e| format!("{}: {}", INVALID, e))?;

                    let slow = SimpleMovingAverage::new(&SimpleMovingAverageParams {
                        length: params.slow_length as usize,
                        bar_component: None,
                        quote_component: None,
                        trade_component: None,
                    })
                    .map_err(|e| format!("{}: {}", INVALID, e))?;

                    (Box::new(fast), Box::new(slow), "SMA")
                }
            };

        let mnemonic = format!(
            "apo({}{}/{}{}{})",
            ma_label,
            params.fast_length,
            ma_label,
            params.slow_length,
            component_triple_mnemonic(bc, qc, tc)
        );
        let description = format!("Absolute Price Oscillator {}", mnemonic);
        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        Ok(Self {
            line,
            fast_ma,
            slow_ma,
            value: f64::NAN,
            primed: false,
        })
    }

    /// Core update logic. Returns the APO value or NaN if not yet primed.
    pub fn update(&mut self, sample: f64) -> f64 {
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

        self.value = fast - slow;
        self.value
    }
}

impl Indicator for AbsolutePriceOscillator {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::AbsolutePriceOscillator,
            &self.line.mnemonic,
            &self.line.description,
            &[OutputText {
                mnemonic: self.line.mnemonic.clone(),
                description: self.line.description.clone(),
            }],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let value = self.update(sample.value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let sample_value = (self.line.bar_func)(sample);
        let value = self.update(sample_value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let sample_value = (self.line.quote_func)(sample);
        let value = self.update(sample_value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let sample_value = (self.line.trade_func)(sample);
        let value = self.update(sample_value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::testdata;
    use crate::indicators::core::outputs::shape::Shape;
    #[test]
    fn test_sma_12_26() {
        let input = testdata::test_input();
        let mut apo = AbsolutePriceOscillator::new(&AbsolutePriceOscillatorParams {
            fast_length: 12,
            slow_length: 26,
            ..Default::default()
        })
        .unwrap();

        const TOLERANCE: f64 = 5e-4;

        // First 25 values should be NaN.
        for i in 0..25 {
            let v = apo.update(input[i]);
            assert!(v.is_nan(), "[{}] expected NaN, got {}", i, v);
        }

        // Index 25: first value.
        let v = apo.update(input[25]);
        assert!(!v.is_nan(), "[25] expected non-NaN");
        assert!(
            (v - (-3.3124)).abs() < TOLERANCE,
            "[25] expected ~-3.3124, got {}",
            v
        );

        // Index 26.
        let v = apo.update(input[26]);
        assert!(
            (v - (-3.5876)).abs() < TOLERANCE,
            "[26] expected ~-3.5876, got {}",
            v
        );

        // Feed remaining and check last.
        for i in 27..251 {
            apo.update(input[i]);
        }
        let v = apo.update(input[251]);
        assert!(
            (v - (-0.1667)).abs() < TOLERANCE,
            "[251] expected ~-0.1667, got {}",
            v
        );

        assert!(apo.is_primed());
    }

    #[test]
    fn test_ema_12_26() {
        let input = testdata::test_input();
        let mut apo = AbsolutePriceOscillator::new(&AbsolutePriceOscillatorParams {
            fast_length: 12,
            slow_length: 26,
            moving_average_type: MovingAverageType::EMA,
            first_is_average: false,
            ..Default::default()
        })
        .unwrap();

        const TOLERANCE: f64 = 5e-4;

        // First 25 values should be NaN.
        for i in 0..25 {
            let v = apo.update(input[i]);
            assert!(v.is_nan(), "[{}] expected NaN, got {}", i, v);
        }

        // Index 25: first value.
        let v = apo.update(input[25]);
        assert!(!v.is_nan(), "[25] expected non-NaN");
        assert!(
            (v - (-2.4193)).abs() < TOLERANCE,
            "[25] expected ~-2.4193, got {}",
            v
        );

        // Index 26.
        let v = apo.update(input[26]);
        assert!(
            (v - (-2.4367)).abs() < TOLERANCE,
            "[26] expected ~-2.4367, got {}",
            v
        );

        // Feed remaining and check last.
        for i in 27..251 {
            apo.update(input[i]);
        }
        let v = apo.update(input[251]);
        assert!(
            (v - 0.90401).abs() < TOLERANCE,
            "[251] expected ~0.90401, got {}",
            v
        );
    }

    #[test]
    fn test_is_primed() {
        let mut apo = AbsolutePriceOscillator::new(&AbsolutePriceOscillatorParams {
            fast_length: 3,
            slow_length: 5,
            ..Default::default()
        })
        .unwrap();

        assert!(!apo.is_primed());

        for i in 1..5 {
            apo.update(i as f64);
            assert!(!apo.is_primed(), "[{}] expected not primed", i);
        }

        apo.update(5.0);
        assert!(apo.is_primed(), "expected primed after 5 samples");

        for i in 6..10 {
            apo.update(i as f64);
            assert!(apo.is_primed(), "[{}] expected primed", i);
        }
    }

    #[test]
    fn test_nan_passthrough() {
        let mut apo = AbsolutePriceOscillator::new(&AbsolutePriceOscillatorParams {
            fast_length: 2,
            slow_length: 3,
            ..Default::default()
        })
        .unwrap();

        let v = apo.update(f64::NAN);
        assert!(v.is_nan());
    }

    #[test]
    fn test_metadata_sma() {
        let apo = AbsolutePriceOscillator::new(&AbsolutePriceOscillatorParams {
            fast_length: 12,
            slow_length: 26,
            ..Default::default()
        })
        .unwrap();

        let meta = apo.metadata();
        assert_eq!(meta.identifier, Identifier::AbsolutePriceOscillator);
        assert_eq!(meta.mnemonic, "apo(SMA12/SMA26)");
        assert_eq!(
            meta.description,
            "Absolute Price Oscillator apo(SMA12/SMA26)"
        );
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(
            meta.outputs[0].kind,
            AbsolutePriceOscillatorOutput::Value as i32
        );
        assert_eq!(meta.outputs[0].shape, Shape::Scalar);
    }

    #[test]
    fn test_metadata_ema() {
        let apo = AbsolutePriceOscillator::new(&AbsolutePriceOscillatorParams {
            fast_length: 12,
            slow_length: 26,
            moving_average_type: MovingAverageType::EMA,
            ..Default::default()
        })
        .unwrap();

        let meta = apo.metadata();
        assert_eq!(meta.mnemonic, "apo(EMA12/EMA26)");
    }

    #[test]
    fn test_update_entity() {
        let input = testdata::test_input();
        let mut apo = AbsolutePriceOscillator::new(&AbsolutePriceOscillatorParams {
            fast_length: 2,
            slow_length: 3,
            ..Default::default()
        })
        .unwrap();

        let time = 1617235200_i64;

        for i in 0..2 {
            let scalar = Scalar::new(time, input[i]);
            let out = apo.update_scalar(&scalar);
            let s = out[0].downcast_ref::<Scalar>().unwrap();
            assert!(s.value.is_nan(), "[{}] expected NaN", i);
        }

        let scalar = Scalar::new(time, input[2]);
        let out = apo.update_scalar(&scalar);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!(!s.value.is_nan(), "[2] expected non-NaN");
    }

    #[test]
    fn test_invalid_params() {
        let cases = vec![
            ("fast too small", 1, 26),
            ("slow too small", 12, 1),
            ("fast negative", -8, 12),
            ("slow negative", 26, -7),
        ];

        for (name, fast, slow) in cases {
            let r = AbsolutePriceOscillator::new(&AbsolutePriceOscillatorParams {
                fast_length: fast,
                slow_length: slow,
                ..Default::default()
            });
            assert!(r.is_err(), "{}: expected error", name);
        }
    }
}
