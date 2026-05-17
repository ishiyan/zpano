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

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the fractal adaptive simple moving average indicator.
pub struct FractalAdaptiveSimpleMovingAverageParams {
    /// The lookback period N for FDI computation. Must be greater than 1.
    pub period: usize,
    /// Base SMA period before fractal adaptation. Must be greater than 0.
    pub normal_speed: usize,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for FractalAdaptiveSimpleMovingAverageParams {
    fn default() -> Self {
        Self {
            period: 30,
            normal_speed: 20,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the fractal adaptive simple moving average indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum FractalAdaptiveSimpleMovingAverageOutput {
    /// The FRASMA value.
    Value = 0,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Computes the Fractal Adaptive Simple Moving Average (FRASMA).
///
/// Uses the Fractal Dimension Index (FDI) formula to adaptively modify an SMA's period.
/// The indicator is not primed during the first `period - 1` updates.
pub struct FractalAdaptiveSimpleMovingAverage {
    line: LineIndicator,
    window: Vec<f64>,
    closes: Vec<f64>,
    period: usize,
    normal_speed: usize,
    window_count: usize,
    primed: bool,
    log_2p: f64,
    ln2: f64,
    inv_p_sq: f64,
}

impl FractalAdaptiveSimpleMovingAverage {
    /// Creates a new FractalAdaptiveSimpleMovingAverage from the given parameters.
    pub fn new(params: &FractalAdaptiveSimpleMovingAverageParams) -> Result<Self, String> {
        if params.period < 2 {
            return Err("invalid fractal adaptive simple moving average parameters: period should be greater than 1".to_string());
        }
        if params.normal_speed < 1 {
            return Err("invalid fractal adaptive simple moving average parameters: normal_speed should be greater than 0".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!("frasma({},{}{})", params.period, params.normal_speed, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Fractal adaptive simple moving average {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);
        let period_f = params.period as f64;

        Ok(Self {
            line,
            window: vec![0.0; params.period],
            closes: Vec::with_capacity(256),
            period: params.period,
            normal_speed: params.normal_speed,
            window_count: 0,
            primed: false,
            log_2p: (2.0 * period_f).ln(),
            ln2: 2.0_f64.ln(),
            inv_p_sq: 1.0 / (period_f * period_f),
        })
    }

    /// Core update logic. Returns the FRASMA value or NaN.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        let period = self.period;

        // Accumulate close history for SMA computation.
        self.closes.push(sample);

        // Fill the FDI window.
        if self.window_count < period {
            self.window[self.window_count] = sample;
            self.window_count += 1;

            if self.window_count < period {
                return f64::NAN;
            }

            self.primed = true;
        } else {
            for i in 0..period - 1 {
                self.window[i] = self.window[i + 1];
            }
            self.window[period - 1] = sample;
        }

        // --- Compute FDI using iliko's original formula (period-2 segments) ---
        let mut price_max = self.window[0];
        let mut price_min = self.window[0];

        for k in 1..period {
            if self.window[k] > price_max { price_max = self.window[k]; }
            if self.window[k] < price_min { price_min = self.window[k]; }
        }

        let price_range = price_max - price_min;
        if price_range < 1e-10 {
            return f64::NAN;
        }

        // iliko skips iteration 0: prior_norm starts at window[1], loop from window[2].
        let mut prior_norm = (self.window[1] - price_min) / price_range;
        let mut length = 0.0;

        for k in 2..period {
            let curr_norm = (self.window[k] - price_min) / price_range;
            let diff = curr_norm - prior_norm;
            length += (diff * diff + self.inv_p_sq).sqrt();
            prior_norm = curr_norm;
        }

        if length <= 0.0 {
            return f64::NAN;
        }

        let fdi = 1.0 + (length.ln() + self.ln2) / self.log_2p;

        // --- Adaptive speed ---
        let denom = 2.0 - fdi;
        if denom.abs() < 1e-10 {
            return f64::NAN;
        }

        let trail_dim = 1.0 / denom;
        let alpha = trail_dim / 2.0;
        let speed = ((self.normal_speed as f64) * alpha).round() as isize;
        let speed = if speed < 1 { 1usize } else { speed as usize };

        // --- SMA of length `speed` ending at current position ---
        let n_closes = self.closes.len();
        if speed > n_closes {
            return f64::NAN;
        }

        let mut sma_sum = 0.0;
        for k in (n_closes - speed)..n_closes {
            sma_sum += self.closes[k];
        }

        sma_sum / speed as f64
    }
}

impl Indicator for FractalAdaptiveSimpleMovingAverage {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::FractalAdaptiveSimpleMovingAverage,
            &self.line.mnemonic,
            &self.line.description,
            &[OutputText {
                mnemonic: self.line.mnemonic.clone(),
                description: self.line.description.clone(),
            }],
        )
    }

    fn update_bar(&mut self, bar: &Bar) -> Output {
        let sample = (self.line.bar_func)(bar);
        let value = self.update(sample);
        vec![Box::new(Scalar::new(bar.time, value))]
    }

    fn update_quote(&mut self, quote: &Quote) -> Output {
        let sample = (self.line.quote_func)(quote);
        let value = self.update(sample);
        vec![Box::new(Scalar::new(quote.time, value))]
    }

    fn update_trade(&mut self, trade: &Trade) -> Output {
        let sample = (self.line.trade_func)(trade);
        let value = self.update(sample);
        vec![Box::new(Scalar::new(trade.time, value))]
    }

    fn update_scalar(&mut self, scalar: &Scalar) -> Output {
        let value = self.update(scalar.value);
        vec![Box::new(Scalar::new(scalar.time, value))]
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;
    use crate::indicators::jean_philippe_poton::fractal_adaptive_simple_moving_average::testdata::testdata;

    fn create(period: usize, normal_speed: usize) -> FractalAdaptiveSimpleMovingAverage {
        FractalAdaptiveSimpleMovingAverage::new(&FractalAdaptiveSimpleMovingAverageParams {
            period,
            normal_speed,
            ..Default::default()
        }).unwrap()
    }

    fn check(expected: &[f64], actual: &[f64]) {
        assert_eq!(expected.len(), actual.len());
        for (i, (e, a)) in expected.iter().zip(actual.iter()).enumerate() {
            if e.is_nan() {
                assert!(a.is_nan(), "[{}] expected NaN, got {}", i, a);
            } else {
                assert!((e - a).abs() < 1e-13, "[{}] expected {}, got {}", i, e, a);
            }
        }
    }

    #[test]
    fn test_frasma_period_5() {
        let input = testdata::test_input();
        let expected = testdata::expected_p5();
        let mut f = create(5, 20);
        let actual: Vec<f64> = input.iter().map(|&v| f.update(v)).collect();
        check(&expected, &actual);
    }

    #[test]
    fn test_frasma_period_10() {
        let input = testdata::test_input();
        let expected = testdata::expected_p10();
        let mut f = create(10, 20);
        let actual: Vec<f64> = input.iter().map(|&v| f.update(v)).collect();
        check(&expected, &actual);
    }

    #[test]
    fn test_frasma_period_15() {
        let input = testdata::test_input();
        let expected = testdata::expected_p15();
        let mut f = create(15, 20);
        let actual: Vec<f64> = input.iter().map(|&v| f.update(v)).collect();
        check(&expected, &actual);
    }

    #[test]
    fn test_frasma_period_20() {
        let input = testdata::test_input();
        let expected = testdata::expected_p20();
        let mut f = create(20, 20);
        let actual: Vec<f64> = input.iter().map(|&v| f.update(v)).collect();
        check(&expected, &actual);
    }

    #[test]
    fn test_frasma_period_30() {
        let input = testdata::test_input();
        let expected = testdata::expected_p30();
        let mut f = create(30, 20);
        let actual: Vec<f64> = input.iter().map(|&v| f.update(v)).collect();
        check(&expected, &actual);
    }

    #[test]
    fn test_frasma_period_50() {
        let input = testdata::test_input();
        let expected = testdata::expected_p50();
        let mut f = create(50, 20);
        let actual: Vec<f64> = input.iter().map(|&v| f.update(v)).collect();
        check(&expected, &actual);
    }

    #[test]
    fn test_frasma_period_80() {
        let input = testdata::test_input();
        let expected = testdata::expected_p80();
        let mut f = create(80, 20);
        let actual: Vec<f64> = input.iter().map(|&v| f.update(v)).collect();
        check(&expected, &actual);
    }

    #[test]
    fn test_frasma_period_120() {
        let input = testdata::test_input();
        let expected = testdata::expected_p120();
        let mut f = create(120, 20);
        let actual: Vec<f64> = input.iter().map(|&v| f.update(v)).collect();
        check(&expected, &actual);
    }

    #[test]
    fn test_frasma_is_primed() {
        let input = testdata::test_input();
        let mut f = create(30, 20);
        for i in 0..29 {
            f.update(input[i]);
            assert!(!f.is_primed(), "should not be primed at index {}", i);
        }
        f.update(input[29]);
        assert!(f.is_primed());
    }

    #[test]
    fn test_frasma_nan_passthrough() {
        let mut f = create(5, 20);
        assert!(f.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_frasma_invalid_period() {
        let result = FractalAdaptiveSimpleMovingAverage::new(&FractalAdaptiveSimpleMovingAverageParams {
            period: 1,
            normal_speed: 20,
            ..Default::default()
        });
        assert!(result.is_err());
    }

    #[test]
    fn test_frasma_invalid_normal_speed() {
        let result = FractalAdaptiveSimpleMovingAverage::new(&FractalAdaptiveSimpleMovingAverageParams {
            period: 5,
            normal_speed: 0,
            ..Default::default()
        });
        assert!(result.is_err());
    }

    #[test]
    fn test_frasma_metadata() {
        let f = create(30, 20);
        let meta = f.metadata();
        assert_eq!(meta.identifier, Identifier::FractalAdaptiveSimpleMovingAverage);
    }
}
