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
use crate::indicators::core::metadata::Metadata;

// ---------------------------------------------------------------------------
// MovingAverageType
// ---------------------------------------------------------------------------

/// Specifies the type of moving average to use.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum MovingAverageType {
    /// Exponential Moving Average (default for classic MACD).
    Ema,
    /// Simple Moving Average.
    Sma,
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the MACD indicator.
pub struct MovingAverageConvergenceDivergenceParams {
    /// Fast moving average length. Must be > 1. Default 12.
    pub fast_length: usize,
    /// Slow moving average length. Must be > 1. Default 26.
    pub slow_length: usize,
    /// Signal line moving average length. Must be > 0. Default 9.
    pub signal_length: usize,
    /// MA type for fast and slow lines. Default Ema.
    pub moving_average_type: MovingAverageType,
    /// MA type for the signal line. Default Ema.
    pub signal_moving_average_type: MovingAverageType,
    /// EMA seeding: true = TA-Lib (SMA seed), false = Metastock (first value).
    /// Default true.
    pub first_is_average: bool,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for MovingAverageConvergenceDivergenceParams {
    fn default() -> Self {
        Self {
            fast_length: 12,
            slow_length: 26,
            signal_length: 9,
            moving_average_type: MovingAverageType::Ema,
            signal_moving_average_type: MovingAverageType::Ema,
            first_is_average: true,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the MACD indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum MovingAverageConvergenceDivergenceOutput {
    /// The MACD line value (fast MA - slow MA).
    Macd = 1,
    /// The signal line value (MA of MACD line).
    Signal = 2,
    /// The histogram value (MACD - signal).
    Histogram = 3,
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

fn new_ma(ma_type: MovingAverageType, length: usize, first_is_average: bool) -> Result<MaVariant, String> {
    match ma_type {
        MovingAverageType::Sma => {
            let sma = SimpleMovingAverage::new(&SimpleMovingAverageParams {
                length,
                bar_component: None,
                quote_component: None,
                trade_component: None,
            })?;
            Ok(MaVariant::Sma(sma))
        }
        MovingAverageType::Ema => {
            let ema = ExponentialMovingAverage::new_from_length(&ExponentialMovingAverageLengthParams {
                length: length as i64,
                first_is_average,
                bar_component: None,
                quote_component: None,
                trade_component: None,
            })?;
            Ok(MaVariant::Ema(ema))
        }
    }
}

fn ma_label(ma_type: MovingAverageType) -> &'static str {
    match ma_type {
        MovingAverageType::Sma => "SMA",
        MovingAverageType::Ema => "EMA",
    }
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Gerald Appel's Moving Average Convergence Divergence (MACD).
///
/// MACD is calculated by subtracting the slow moving average from the fast moving average.
/// A signal line (moving average of MACD) and histogram (MACD minus signal) are also produced.
///
/// The indicator produces three outputs:
///   - MACD: fast MA - slow MA
///   - Signal: MA of the MACD line
///   - Histogram: MACD - Signal
pub struct MovingAverageConvergenceDivergence {
    fast_ma: MaVariant,
    slow_ma: MaVariant,
    signal_ma: MaVariant,
    macd_value: f64,
    signal_value: f64,
    histogram_value: f64,
    primed: bool,
    /// Number of initial samples to skip before feeding the fast MA.
    fast_delay: usize,
    fast_count: usize,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
    mnemonic: String,
}

impl MovingAverageConvergenceDivergence {
    /// Creates a new MACD from the given parameters.
    pub fn new(params: &MovingAverageConvergenceDivergenceParams) -> Result<Self, String> {
        let invalid = "invalid moving average convergence divergence parameters";

        let mut fast_length = params.fast_length;
        if fast_length == 0 { fast_length = 12; }
        let mut slow_length = params.slow_length;
        if slow_length == 0 { slow_length = 26; }
        let mut signal_length = params.signal_length;
        if signal_length == 0 { signal_length = 9; }

        if fast_length < 2 {
            return Err(format!("{}: fast length should be greater than 1", invalid));
        }
        if slow_length < 2 {
            return Err(format!("{}: slow length should be greater than 1", invalid));
        }
        if signal_length < 1 {
            return Err(format!("{}: signal length should be greater than 0", invalid));
        }

        // Auto-swap fast/slow if needed (matches TaLib behavior).
        if slow_length < fast_length {
            std::mem::swap(&mut fast_length, &mut slow_length);
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let fast_ma = new_ma(params.moving_average_type, fast_length, params.first_is_average)?;
        let slow_ma = new_ma(params.moving_average_type, slow_length, params.first_is_average)?;
        let signal_ma = new_ma(params.signal_moving_average_type, signal_length, params.first_is_average)?;

        let suffix = if params.moving_average_type != MovingAverageType::Ema
            || params.signal_moving_average_type != MovingAverageType::Ema
        {
            format!(",{},{}", ma_label(params.moving_average_type), ma_label(params.signal_moving_average_type))
        } else {
            String::new()
        };

        let mnemonic = format!(
            "macd({},{},{}{}{})",
            fast_length, slow_length, signal_length, suffix,
            component_triple_mnemonic(bc, qc, tc)
        );

        Ok(Self {
            fast_ma,
            slow_ma,
            signal_ma,
            macd_value: f64::NAN,
            signal_value: f64::NAN,
            histogram_value: f64::NAN,
            primed: false,
            fast_delay: slow_length - fast_length,
            fast_count: 0,
            bar_func,
            quote_func,
            trade_func,
            mnemonic,
        })
    }

    /// Returns true if the indicator has produced at least one valid complete output.
    pub fn is_primed(&self) -> bool {
        self.primed
    }

    /// Core update returning (macd, signal, histogram).
    pub fn update(&mut self, sample: f64) -> (f64, f64, f64) {
        let nan = f64::NAN;

        if sample.is_nan() {
            return (nan, nan, nan);
        }

        // Feed the slow MA every sample.
        let slow = self.slow_ma.update(sample);

        // Delay the fast MA to align SMA seed windows (matches TaLib batch algorithm).
        let fast = if self.fast_count < self.fast_delay {
            self.fast_count += 1;
            nan
        } else {
            self.fast_ma.update(sample)
        };

        if fast.is_nan() || slow.is_nan() {
            self.macd_value = nan;
            self.signal_value = nan;
            self.histogram_value = nan;
            return (nan, nan, nan);
        }

        let macd = fast - slow;
        self.macd_value = macd;

        let signal = self.signal_ma.update(macd);

        if signal.is_nan() {
            self.signal_value = nan;
            self.histogram_value = nan;
            return (macd, nan, nan);
        }

        self.signal_value = signal;
        let histogram = macd - signal;
        self.histogram_value = histogram;
        self.primed = self.fast_ma.is_primed() && self.slow_ma.is_primed() && self.signal_ma.is_primed();

        (macd, signal, histogram)
    }
}

impl Indicator for MovingAverageConvergenceDivergence {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        let desc = format!("Moving Average Convergence Divergence {}", self.mnemonic);
        build_metadata(
            Identifier::MovingAverageConvergenceDivergence,
            &self.mnemonic,
            &desc,
            &[
                OutputText {
                    mnemonic: format!("{} macd", self.mnemonic),
                    description: format!("{} MACD", desc),
                },
                OutputText {
                    mnemonic: format!("{} signal", self.mnemonic),
                    description: format!("{} Signal", desc),
                },
                OutputText {
                    mnemonic: format!("{} histogram", self.mnemonic),
                    description: format!("{} Histogram", desc),
                },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let (macd, signal, histogram) = self.update(sample.value);
        vec![
            Box::new(Scalar { time: sample.time, value: macd }),
            Box::new(Scalar { time: sample.time, value: signal }),
            Box::new(Scalar { time: sample.time, value: histogram }),
        ]
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let v = (self.bar_func)(sample);
        self.update_scalar(&Scalar { time: sample.time, value: v })
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (self.quote_func)(sample);
        self.update_scalar(&Scalar { time: sample.time, value: v })
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let v = (self.trade_func)(sample);
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
    fn test_default_params() {
        let tolerance = 1e-8;
        let input = testdata::test_input();
        let exp_macd = testdata::test_macd_expected();
        let exp_signal = testdata::test_signal_expected();
        let exp_histogram = testdata::test_histogram_expected();

        let mut ind = MovingAverageConvergenceDivergence::new(
            &MovingAverageConvergenceDivergenceParams::default(),
        ).unwrap();

        for i in 0..252 {
            let (macd, signal, histogram) = ind.update(input[i]);

            if exp_macd[i].is_nan() {
                assert!(macd.is_nan(), "[{}] macd: expected NaN, got {}", i, macd);
                assert!(signal.is_nan(), "[{}] signal: expected NaN, got {}", i, signal);
                assert!(histogram.is_nan(), "[{}] histogram: expected NaN, got {}", i, histogram);
                continue;
            }

            assert!(
                (macd - exp_macd[i]).abs() < tolerance,
                "[{}] macd: expected {}, got {}", i, exp_macd[i], macd
            );

            if exp_signal[i].is_nan() {
                assert!(signal.is_nan(), "[{}] signal: expected NaN, got {}", i, signal);
                assert!(histogram.is_nan(), "[{}] histogram: expected NaN, got {}", i, histogram);
                continue;
            }

            assert!(
                (signal - exp_signal[i]).abs() < tolerance,
                "[{}] signal: expected {}, got {}", i, exp_signal[i], signal
            );

            assert!(
                (histogram - exp_histogram[i]).abs() < tolerance,
                "[{}] histogram: expected {}, got {}", i, exp_histogram[i], histogram
            );
        }
    }

    #[test]
    fn test_talib_spot_check() {
        let tolerance = 5e-4;
        let input = testdata::test_input();

        let mut ind = MovingAverageConvergenceDivergence::new(
            &MovingAverageConvergenceDivergenceParams::default(),
        ).unwrap();

        for i in 0..=33 {
            ind.update(input[i]);
        }

        // Re-run to get spot check at index 33.
        let mut ind2 = MovingAverageConvergenceDivergence::new(
            &MovingAverageConvergenceDivergenceParams::default(),
        ).unwrap();

        let mut macd = f64::NAN;
        let mut signal = f64::NAN;
        let mut histogram = f64::NAN;

        for i in 0..=33 {
            let r = ind2.update(input[i]);
            macd = r.0;
            signal = r.1;
            histogram = r.2;
        }

        assert!((macd - (-1.9738)).abs() < tolerance, "MACD[33] = {}, want -1.9738", macd);
        assert!((signal - (-2.7071)).abs() < tolerance, "Signal[33] = {}, want -2.7071", signal);
        let expected_histogram = (-1.9738) - (-2.7071);
        assert!((histogram - expected_histogram).abs() < tolerance, "Histogram[33] = {}, want {}", histogram, expected_histogram);
    }

    #[test]
    fn test_period_inversion() {
        let tolerance = 5e-4;
        let input = testdata::test_input();

        // Passing fast=26, slow=12 should auto-swap.
        let mut ind = MovingAverageConvergenceDivergence::new(
            &MovingAverageConvergenceDivergenceParams {
                fast_length: 26,
                slow_length: 12,
                ..Default::default()
            },
        ).unwrap();

        let mut macd = f64::NAN;
        let mut signal = f64::NAN;

        for i in 0..=33 {
            let r = ind.update(input[i]);
            macd = r.0;
            signal = r.1;
        }

        assert!((macd - (-1.9738)).abs() < tolerance, "MACD[33] = {}, want -1.9738", macd);
        assert!((signal - (-2.7071)).abs() < tolerance, "Signal[33] = {}, want -2.7071", signal);
    }

    #[test]
    fn test_is_primed() {
        let mut ind = MovingAverageConvergenceDivergence::new(
            &MovingAverageConvergenceDivergenceParams {
                fast_length: 3,
                slow_length: 5,
                signal_length: 2,
                ..Default::default()
            },
        ).unwrap();

        assert!(!ind.is_primed());

        for i in 0..6 {
            ind.update((i + 1) as f64);
            if i < 5 {
                assert!(!ind.is_primed(), "[{}] expected not primed", i);
            }
        }

        assert!(ind.is_primed(), "expected primed after 6 samples");
    }

    #[test]
    fn test_nan() {
        let mut ind = MovingAverageConvergenceDivergence::new(
            &MovingAverageConvergenceDivergenceParams::default(),
        ).unwrap();

        let (macd, signal, histogram) = ind.update(f64::NAN);
        assert!(macd.is_nan());
        assert!(signal.is_nan());
        assert!(histogram.is_nan());
    }

    #[test]
    fn test_metadata() {
        let ind = MovingAverageConvergenceDivergence::new(
            &MovingAverageConvergenceDivergenceParams::default(),
        ).unwrap();

        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::MovingAverageConvergenceDivergence);
        assert_eq!(meta.mnemonic, "macd(12,26,9)");
        assert_eq!(meta.outputs.len(), 3);
        assert_eq!(meta.outputs[0].kind, MovingAverageConvergenceDivergenceOutput::Macd as i32);
        assert_eq!(meta.outputs[1].kind, MovingAverageConvergenceDivergenceOutput::Signal as i32);
        assert_eq!(meta.outputs[2].kind, MovingAverageConvergenceDivergenceOutput::Histogram as i32);
    }

    #[test]
    fn test_metadata_sma() {
        let ind = MovingAverageConvergenceDivergence::new(
            &MovingAverageConvergenceDivergenceParams {
                moving_average_type: MovingAverageType::Sma,
                ..Default::default()
            },
        ).unwrap();

        let meta = ind.metadata();
        assert_eq!(meta.mnemonic, "macd(12,26,9,SMA,EMA)");
    }

    #[test]
    fn test_update_scalar() {
        let tolerance = 5e-4;
        let input = testdata::test_input();

        let mut ind = MovingAverageConvergenceDivergence::new(
            &MovingAverageConvergenceDivergenceParams::default(),
        ).unwrap();

        // Feed first 33 values.
        for i in 0..33 {
            let scalar = Scalar { time: 0, value: input[i] };
            let out = ind.update_scalar(&scalar);
            let s = out[1].downcast_ref::<Scalar>().unwrap().value;
            if i < 25 {
                let m = out[0].downcast_ref::<Scalar>().unwrap().value;
                assert!(m.is_nan(), "[{}] expected NaN macd", i);
            }
            if i < 33 {
                assert!(s.is_nan(), "[{}] expected NaN signal", i);
            }
        }

        // Index 33: first complete output.
        let scalar = Scalar { time: 0, value: input[33] };
        let out = ind.update_scalar(&scalar);

        let macd = out[0].downcast_ref::<Scalar>().unwrap().value;
        let signal = out[1].downcast_ref::<Scalar>().unwrap().value;
        let histogram = out[2].downcast_ref::<Scalar>().unwrap().value;

        assert!((macd - (-1.9738)).abs() < tolerance, "MACD[33] = {}, want -1.9738", macd);
        assert!((signal - (-2.7071)).abs() < tolerance, "Signal[33] = {}, want -2.7071", signal);
        let expected_histogram = (-1.9738) - (-2.7071);
        assert!((histogram - expected_histogram).abs() < tolerance, "Histogram[33] = {}, want {}", histogram, expected_histogram);
    }

    #[test]
    fn test_invalid_params() {
        let cases = vec![
            ("fast too small", 1, 26, 9),
            ("slow too small", 12, 1, 9),
            ("signal negative", 12, 26, 0), // 0 treated as default, but signal_length=0 maps to default 9 - use explicit test
        ];

        // fast=1
        assert!(MovingAverageConvergenceDivergence::new(
            &MovingAverageConvergenceDivergenceParams { fast_length: 1, slow_length: 26, signal_length: 9, ..Default::default() }
        ).is_err());

        // slow=1
        assert!(MovingAverageConvergenceDivergence::new(
            &MovingAverageConvergenceDivergenceParams { fast_length: 12, slow_length: 1, signal_length: 9, ..Default::default() }
        ).is_err());

        // Note: signal_length=0 defaults to 9 in our impl (matching Go), so it's not invalid.
        // We don't have a way to pass negative values since usize is unsigned.
        let _ = cases; // suppress warning
    }
}
