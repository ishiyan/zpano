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

/// Parameters to create an instance of the RS fractal adaptive simple moving average indicator.
pub struct RescaledFractalAdaptiveSimpleMovingAverageParams {
    /// The lookback window for R/S analysis. Must be a power of 2, >= 4.
    pub period: usize,
    /// Base SMA period before fractal adaptation. Must be >= 1.
    pub normal_speed: usize,
    /// Multiplier applied to prices before R/S calculation. Default is 1.0.
    pub price_scale: f64,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for RescaledFractalAdaptiveSimpleMovingAverageParams {
    fn default() -> Self {
        Self {
            period: 64,
            normal_speed: 30,
            price_scale: 1.0,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the RS fractal adaptive simple moving average indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum RescaledFractalAdaptiveSimpleMovingAverageOutput {
    /// The RSFRASMA value.
    Value = 0,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Computes the RS Fractal Adaptive Simple Moving Average (RSFRASMA).
///
/// Uses Rescaled Range (R/S) analysis to estimate the Hurst exponent,
/// then adapts the SMA period accordingly.
/// The indicator is not primed during the first `period` updates.
pub struct RescaledFractalAdaptiveSimpleMovingAverage {
    line: LineIndicator,
    closes: Vec<f64>,
    period: usize,
    normal_speed: usize,
    price_scale: f64,
    n_iter: usize,
    block_sizes: Vec<usize>,
    block_counts: Vec<usize>,
    primed: bool,
}

impl RescaledFractalAdaptiveSimpleMovingAverage {
    /// Creates a new RescaledFractalAdaptiveSimpleMovingAverage from the given parameters.
    pub fn new(params: &RescaledFractalAdaptiveSimpleMovingAverageParams) -> Result<Self, String> {
        if params.period < 4 {
            return Err("invalid RS fractal adaptive simple moving average parameters: period should be greater than 3".to_string());
        }
        if params.period & (params.period - 1) != 0 {
            return Err("invalid RS fractal adaptive simple moving average parameters: period must be a power of 2".to_string());
        }
        if params.normal_speed < 1 {
            return Err("invalid RS fractal adaptive simple moving average parameters: normal_speed should be greater than 0".to_string());
        }

        let price_scale = if params.price_scale == 0.0 { 1.0 } else { params.price_scale };

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!("rsfrasma({},{},{:.1}{})", params.period, params.normal_speed, price_scale, component_triple_mnemonic(bc, qc, tc));
        let description = format!("RS fractal adaptive simple moving average {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        // Precompute R/S parameters.
        let k0 = params.period / 4;
        let n_iter = if k0 >= 2 {
            ((k0 as f64).ln() / 2.0_f64.ln()).floor() as usize
        } else {
            0
        };

        let mut block_sizes = vec![0usize; n_iter + 1];
        let mut block_counts = vec![0usize; n_iter + 1];

        for u in 1..=n_iter {
            block_sizes[u] = 1 << (u + 1);
            block_counts[u] = params.period / block_sizes[u];
        }

        Ok(Self {
            line,
            closes: Vec::with_capacity(256),
            period: params.period,
            normal_speed: params.normal_speed,
            price_scale,
            n_iter,
            block_sizes,
            block_counts,
            primed: false,
        })
    }

    /// Core update logic. Returns the RSFRASMA value or NaN.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        let period = self.period;
        let price_scale = self.price_scale;

        self.closes.push(sample);
        let n_closes = self.closes.len();

        if n_closes <= period {
            return f64::NAN;
        }

        if !self.primed {
            self.primed = true;
        }

        let pos = n_closes - 1;

        // R/S analysis.
        let n_iter = self.n_iter;
        let mut sumx = 0.0;
        let mut sumy = 0.0;
        let mut sumx2 = 0.0;
        let mut sumxy = 0.0;
        let mut valid_scales = 0usize;

        for u in 1..=n_iter {
            let block_size = self.block_sizes[u];
            let n_blocks_u = self.block_counts[u];

            if n_blocks_u < 1 {
                continue;
            }

            let mut rs_sum = 0.0;
            let mut t = 0usize;
            let mut block_count = 0usize;

            while t <= period - block_size {
                // Block mean.
                let mut mu = 0.0;
                for j in 1..=block_size {
                    mu += price_scale * self.closes[pos - (t + j)];
                }
                mu /= block_size as f64;

                // Population std.
                let mut sum_sq = 0.0;
                for j in 1..=block_size {
                    let diff = price_scale * self.closes[pos - (t + j)] - mu;
                    sum_sq += diff * diff;
                }
                let mut std_val = (sum_sq / block_size as f64).sqrt();
                if std_val <= 0.0 {
                    std_val = 0.1;
                }

                // Cumulative deviations and range.
                let mut cum_dev = 0.0;
                let mut w_max = 0.0;
                let mut w_min = 9999999999.0;

                for k in 1..=block_size {
                    cum_dev += price_scale * self.closes[pos - (t + k)] - mu;
                    if cum_dev > w_max { w_max = cum_dev; }
                    if cum_dev < w_min { w_min = cum_dev; }
                }

                if w_max < 0.0 { w_max = 0.0; }
                if w_min > 0.0 { w_min = 0.0; }

                let r_val = w_max - w_min;
                rs_sum += r_val / std_val;
                t += block_size;
                block_count += 1;
            }

            // Average R/S for this scale.
            let mut rs_avg = 1.0;
            if block_count > 0 {
                rs_avg = rs_sum / block_count as f64;
            }
            if rs_avg <= 0.0 {
                rs_avg = 1e-10;
            }

            let log2_d = (block_size as f64).ln() / 2.0_f64.ln();
            let log2_rs = rs_avg.ln() / 2.0_f64.ln();

            sumx += log2_d;
            sumy += log2_rs;
            sumx2 += log2_d * log2_d;
            sumxy += log2_d * log2_rs;
            valid_scales += 1;
        }

        // Linear regression slope = Hurst exponent.
        let mut h = 0.5;
        if valid_scales >= 2 {
            let vs = valid_scales as f64;
            let h1 = vs * sumxy - sumx * sumy;
            let mut h2 = vs * sumx2 - sumx * sumx;
            if h2 <= 0.0 {
                h2 = 0.1;
            }
            h = h1 / h2;
        }

        // Guard H.
        if 2.0 * h <= 0.0 {
            h = 0.001;
        }

        let alpha = 1.0 / (2.0 * h);
        let spd_raw = (self.normal_speed as f64 * alpha).round() as isize;
        let spd = if spd_raw < 1 { 1usize } else { spd_raw as usize };

        // Compute SMA with adapted speed.
        let sma_start = if pos + 1 >= spd { pos + 1 - spd } else { 0 };
        let count = pos - sma_start + 1;

        let mut total = 0.0;
        for i in sma_start..=pos {
            total += self.closes[i];
        }

        total / count as f64
    }
}

impl Indicator for RescaledFractalAdaptiveSimpleMovingAverage {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::RescaledFractalAdaptiveSimpleMovingAverage,
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
    use crate::indicators::jean_philippe_poton::rescaled_fractal_adaptive_simple_moving_average::testdata::testdata;

    fn create(period: usize, normal_speed: usize, price_scale: f64) -> RescaledFractalAdaptiveSimpleMovingAverage {
        RescaledFractalAdaptiveSimpleMovingAverage::new(&RescaledFractalAdaptiveSimpleMovingAverageParams {
            period,
            normal_speed,
            price_scale,
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

    fn run_test(period: usize, normal_speed: usize, price_scale: f64, expected: Vec<f64>) {
        let input = testdata::test_input();
        let mut f = create(period, normal_speed, price_scale);
        let actual: Vec<f64> = input.iter().map(|&v| f.update(v)).collect();
        check(&expected, &actual);
    }

    #[test]
    fn test_rsfrasma_p4_s1() { run_test(4, 30, 1.0, testdata::expected_p4_s1()); }

    #[test]
    fn test_rsfrasma_p8_s1() { run_test(8, 30, 1.0, testdata::expected_p8_s1()); }

    #[test]
    fn test_rsfrasma_p16_s1() { run_test(16, 30, 1.0, testdata::expected_p16_s1()); }

    #[test]
    fn test_rsfrasma_p32_s1() { run_test(32, 30, 1.0, testdata::expected_p32_s1()); }

    #[test]
    fn test_rsfrasma_p64_s1() { run_test(64, 30, 1.0, testdata::expected_p64_s1()); }

    #[test]
    fn test_rsfrasma_p128_s1() { run_test(128, 30, 1.0, testdata::expected_p128_s1()); }

    #[test]
    fn test_rsfrasma_p32_s100() { run_test(32, 30, 100.0, testdata::expected_p32_s100()); }

    #[test]
    fn test_rsfrasma_p32_s10000() { run_test(32, 30, 10000.0, testdata::expected_p32_s10000()); }

    #[test]
    fn test_rsfrasma_is_primed() {
        let input = testdata::test_input();
        let mut f = create(64, 30, 1.0);
        for i in 0..64 {
            f.update(input[i]);
            assert!(!f.is_primed(), "should not be primed at index {}", i);
        }
        f.update(input[64]);
        assert!(f.is_primed());
    }

    #[test]
    fn test_rsfrasma_nan_passthrough() {
        let mut f = create(4, 30, 1.0);
        assert!(f.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_rsfrasma_invalid_period() {
        assert!(RescaledFractalAdaptiveSimpleMovingAverage::new(&RescaledFractalAdaptiveSimpleMovingAverageParams {
            period: 2, normal_speed: 30, price_scale: 1.0, ..Default::default()
        }).is_err());
    }

    #[test]
    fn test_rsfrasma_invalid_period_not_power_of_2() {
        assert!(RescaledFractalAdaptiveSimpleMovingAverage::new(&RescaledFractalAdaptiveSimpleMovingAverageParams {
            period: 6, normal_speed: 30, price_scale: 1.0, ..Default::default()
        }).is_err());
    }

    #[test]
    fn test_rsfrasma_invalid_normal_speed() {
        assert!(RescaledFractalAdaptiveSimpleMovingAverage::new(&RescaledFractalAdaptiveSimpleMovingAverageParams {
            period: 4, normal_speed: 0, price_scale: 1.0, ..Default::default()
        }).is_err());
    }

    #[test]
    fn test_rsfrasma_metadata() {
        let f = create(64, 30, 1.0);
        let meta = f.metadata();
        assert_eq!(meta.identifier, Identifier::RescaledFractalAdaptiveSimpleMovingAverage);
    }
}
