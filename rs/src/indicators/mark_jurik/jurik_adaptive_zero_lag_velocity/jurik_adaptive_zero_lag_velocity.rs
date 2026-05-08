use crate::entities::bar::Bar;
use crate::entities::bar_component::{
    component_value as bar_component_value, BarComponent, DEFAULT_BAR_COMPONENT,
};
use crate::entities::quote::Quote;
use crate::entities::quote_component::{
    component_value as quote_component_value, QuoteComponent, DEFAULT_QUOTE_COMPONENT,
};
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::entities::trade_component::{
    component_value as trade_component_value, TradeComponent, DEFAULT_TRADE_COMPONENT,
};
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::component_triple_mnemonic::component_triple_mnemonic;
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::line_indicator::{BarFunc, QuoteFunc, TradeFunc};
use crate::indicators::core::metadata::Metadata;

// ---------------------------------------------------------------------------
// VelSmooth (Stage 2 adaptive smoother)
// ---------------------------------------------------------------------------

struct VelSmooth {
    jrc03: f64,
    jrc06: usize,
    jrc07: usize,
    ema_factor: f64,
    damping: f64,
    eps2: f64,
    buffer_size: usize,
    buffer: Vec<f64>,
    head: usize,
    length: usize,
    bar_count: usize,
    velocity: f64,
    position: f64,
    smoothed_mad: f64,
    initialized: bool,
}

impl VelSmooth {
    fn new(period: f64) -> Self {
        let eps2 = 0.0001;
        let jrc03 = period.max(eps2).min(500.0);
        let jrc06 = (2.0 * period).ceil().max(31.0) as usize;
        let jrc07 = period.ceil().min(30.0) as usize;
        let ema_factor = 1.0 - (-4.0_f64.ln() / (period / 2.0)).exp();
        let damping = 0.86 - 0.55 / jrc03.sqrt();

        Self {
            jrc03,
            jrc06,
            jrc07,
            ema_factor,
            damping,
            eps2,
            buffer_size: 1001,
            buffer: vec![0.0; 1001],
            head: 0,
            length: 0,
            bar_count: 0,
            velocity: 0.0,
            position: 0.0,
            smoothed_mad: 0.0,
            initialized: false,
        }
    }

    fn update(&mut self, value: f64) -> f64 {
        self.bar_count += 1;

        // Store in circular buffer.
        let old_index = self.head % self.buffer_size;
        self.buffer[old_index] = value;
        self.head += 1;

        if self.length < self.jrc06 {
            self.length += 1;
        }

        let length = self.length;

        // First bar: initialize position.
        if length < 2 {
            if !self.initialized {
                self.position = value;
                self.initialized = true;
            }
            return self.position;
        }

        if !self.initialized {
            self.position = value;
            self.initialized = true;
        }

        // Linear regression over buffer.
        let mut sum_values = 0.0_f64;
        let mut sum_weighted = 0.0_f64;

        for k in 0..length {
            let mut idx = (self.head as isize - length as isize + k as isize) % self.buffer_size as isize;
            if idx < 0 {
                idx += self.buffer_size as isize;
            }
            sum_values += self.buffer[idx as usize];
            sum_weighted += self.buffer[idx as usize] * k as f64;
        }

        let midpoint = (length - 1) as f64 / 2.0;
        let sum_x_sq = length as f64 * (length - 1) as f64 * (2 * length - 1) as f64 / 6.0;
        let regression_denom = sum_x_sq - length as f64 * midpoint * midpoint;

        let regression_slope = if regression_denom.abs() >= self.eps2 {
            (sum_weighted - midpoint * sum_values) / regression_denom
        } else {
            0.0
        };

        let intercept = sum_values / length as f64 - regression_slope * midpoint;

        // Compute MAD from regression residuals.
        let mut sum_abs_dev = 0.0_f64;

        for k in 0..length {
            let mut idx = (self.head as isize - length as isize + k as isize) % self.buffer_size as isize;
            if idx < 0 {
                idx += self.buffer_size as isize;
            }
            let predicted = intercept + regression_slope * k as f64;
            sum_abs_dev += (self.buffer[idx as usize] - predicted).abs();
        }

        let mut raw_mad = sum_abs_dev / length as f64;
        let scale = 1.2 * (self.jrc06 as f64 / length as f64).powf(0.25);
        raw_mad *= scale;

        // Smooth MAD with EMA.
        if self.bar_count <= self.jrc07 + 1 {
            self.smoothed_mad = raw_mad;
        } else {
            self.smoothed_mad += self.ema_factor * (raw_mad - self.smoothed_mad);
        }

        // Adaptive velocity/position dynamics.
        let prediction_error = value - self.position;

        let response_factor = if self.smoothed_mad * self.jrc03 < self.eps2 {
            1.0
        } else {
            1.0 - (-prediction_error.abs() / (self.smoothed_mad * self.jrc03)).exp()
        };

        self.velocity = response_factor * prediction_error + self.velocity * self.damping;
        self.position += self.velocity;

        self.position
    }
}

impl std::fmt::Debug for VelSmooth {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("VelSmooth").finish()
    }
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

pub struct JurikAdaptiveZeroLagVelocityParams {
    pub lo_length: usize,
    pub hi_length: usize,
    pub sensitivity: f64,
    pub period: f64,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for JurikAdaptiveZeroLagVelocityParams {
    fn default() -> Self {
        Self {
            lo_length: 5,
            hi_length: 30,
            sensitivity: 1.0,
            period: 3.0,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum JurikAdaptiveZeroLagVelocityOutput {
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

#[derive(Debug)]
pub struct JurikAdaptiveZeroLagVelocity {
    primed: bool,
    lo_length: usize,
    hi_length: usize,
    sensitivity: f64,
    eps: f64,
    prices: Vec<f64>,
    value1: Vec<f64>,
    bar_count: usize,
    smooth: VelSmooth,
    bar_func: BarFunc,
    quote_func: QuoteFunc,
    trade_func: TradeFunc,
    mnemonic: String,
    description: String,
}

impl JurikAdaptiveZeroLagVelocity {
    pub fn new(params: &JurikAdaptiveZeroLagVelocityParams) -> Result<Self, String> {
        if params.lo_length < 2 {
            return Err("invalid jurik adaptive zero lag velocity parameters: lo_length should be at least 2".to_string());
        }
        if params.hi_length < params.lo_length {
            return Err("invalid jurik adaptive zero lag velocity parameters: hi_length should be at least lo_length".to_string());
        }
        if params.period <= 0.0 {
            return Err("invalid jurik adaptive zero lag velocity parameters: period should be positive".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let triple = component_triple_mnemonic(bc, qc, tc);
        let mnemonic = format!("javel({}, {}, {}, {}{})", params.lo_length, params.hi_length, params.sensitivity, params.period, triple);
        let description = format!("Jurik adaptive zero lag velocity {}", mnemonic);

        Ok(Self {
            primed: false,
            lo_length: params.lo_length,
            hi_length: params.hi_length,
            sensitivity: params.sensitivity,
            eps: 0.001,
            prices: Vec::new(),
            value1: Vec::new(),
            bar_count: 0,
            smooth: VelSmooth::new(params.period),
            bar_func,
            quote_func,
            trade_func,
            mnemonic,
            description,
        })
    }

    fn compute_adaptive_depth(&self, bar: usize) -> f64 {
        let mut long_window = bar.min(99);
        long_window += 1;

        let mut short_window = bar.min(9);
        short_window += 1;

        let mut avg1 = 0.0_f64;
        for i in 0..long_window {
            avg1 += self.value1[bar + 1 - long_window + i];
        }
        avg1 /= long_window as f64;

        let mut avg2 = 0.0_f64;
        for i in 0..short_window {
            avg2 += self.value1[bar + 1 - short_window + i];
        }
        avg2 /= short_window as f64;

        let value2 = self.sensitivity * ((self.eps + avg1) / (self.eps + avg2)).ln();
        let value3 = value2 / (1.0 + value2.abs());

        self.lo_length as f64 + (self.hi_length - self.lo_length) as f64 * (1.0 + value3) / 2.0
    }

    fn compute_wls_slope(&self, bar: usize, depth: usize) -> f64 {
        let n = (depth + 1) as f64;
        let s1 = n * (n + 1.0) / 2.0;
        let s2 = s1 * (2.0 * n + 1.0) / 3.0;
        let denom = s1 * s1 * s1 - s2 * s2;

        let mut sum_xw = 0.0_f64;
        let mut sum_xw2 = 0.0_f64;

        for i in 0..=depth {
            let w = n - i as f64;
            let p = self.prices[bar - i];
            sum_xw += p * w;
            sum_xw2 += p * w * w;
        }

        (sum_xw2 * s1 - sum_xw * s2) / denom
    }

    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        let bar = self.bar_count;
        self.bar_count += 1;

        self.prices.push(sample);

        // Compute value1 (abs diff).
        if bar == 0 {
            self.value1.push(0.0);
        } else {
            self.value1.push((sample - self.prices[bar - 1]).abs());
        }

        // Compute adaptive depth.
        let adaptive_depth = self.compute_adaptive_depth(bar);
        let depth = adaptive_depth.ceil() as usize;

        // Check if we have enough prices for WLS.
        if bar < depth {
            return f64::NAN;
        }

        // Stage 1: WLS slope.
        let slope = self.compute_wls_slope(bar, depth);

        // Stage 2: adaptive smoother.
        let result = self.smooth.update(slope);

        if !self.primed {
            self.primed = true;
        }

        result
    }
}

impl Indicator for JurikAdaptiveZeroLagVelocity {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::JurikAdaptiveZeroLagVelocity,
            &self.mnemonic,
            &self.description,
            &[OutputText {
                mnemonic: self.mnemonic.clone(),
                description: self.description.clone(),
            }],
        )
    }

    fn update_bar(&mut self, bar: &Bar) -> Output {
        let val = self.update((self.bar_func)(bar));
        vec![Box::new(Scalar::new(bar.time, val))]
    }

    fn update_quote(&mut self, quote: &Quote) -> Output {
        let val = self.update((self.quote_func)(quote));
        vec![Box::new(Scalar::new(quote.time, val))]
    }

    fn update_trade(&mut self, trade: &Trade) -> Output {
        let val = self.update((self.trade_func)(trade));
        vec![Box::new(Scalar::new(trade.time, val))]
    }

    fn update_scalar(&mut self, scalar: &Scalar) -> Output {
        let val = self.update(scalar.value);
        vec![Box::new(Scalar::new(scalar.time, val))]
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;
    use crate::indicators::mark_jurik::jurik_adaptive_zero_lag_velocity::testdata::testdata;

    const EPSILON: f64 = 1e-13;

    fn almost_equal(a: f64, b: f64) -> bool {
        if a.is_nan() && b.is_nan() {
            return true;
        }
        if a.is_nan() || b.is_nan() {
            return false;
        }
        (a - b).abs() < EPSILON
    }

    fn run_test(lo: usize, hi: usize, sens: f64, period: f64, expected: &[f64]) {
        let params = JurikAdaptiveZeroLagVelocityParams {
            lo_length: lo,
            hi_length: hi,
            sensitivity: sens,
            period,
            ..Default::default()
        };
        let mut ind = JurikAdaptiveZeroLagVelocity::new(&params).unwrap();
        let input = testdata::test_input();
        for (i, &val) in input.iter().enumerate() {
            let result = ind.update(val);
            assert!(
                almost_equal(result, expected[i]),
                "bar {}: expected {}, got {} (lo={}, hi={}, sens={}, period={})",
                i, expected[i], result, lo, hi, sens, period
            );
        }
    }

    #[test] fn test_lo2_hi15() { run_test(2, 15, 1.0, 3.0, &testdata::expected_lo2_hi15()); }
    #[test] fn test_lo2_hi30() { run_test(2, 30, 1.0, 3.0, &testdata::expected_lo2_hi30()); }
    #[test] fn test_lo2_hi60() { run_test(2, 60, 1.0, 3.0, &testdata::expected_lo2_hi60()); }
    #[test] fn test_lo5_hi15() { run_test(5, 15, 1.0, 3.0, &testdata::expected_lo5_hi15()); }
    #[test] fn test_lo5_hi30() { run_test(5, 30, 1.0, 3.0, &testdata::expected_lo5_hi30()); }
    #[test] fn test_lo5_hi60() { run_test(5, 60, 1.0, 3.0, &testdata::expected_lo5_hi60()); }
    #[test] fn test_lo10_hi15() { run_test(10, 15, 1.0, 3.0, &testdata::expected_lo10_hi15()); }
    #[test] fn test_lo10_hi30() { run_test(10, 30, 1.0, 3.0, &testdata::expected_lo10_hi30()); }
    #[test] fn test_lo10_hi60() { run_test(10, 60, 1.0, 3.0, &testdata::expected_lo10_hi60()); }
    #[test] fn test_sens05() { run_test(5, 30, 0.5, 3.0, &testdata::expected_sens05()); }
    #[test] fn test_sens25() { run_test(5, 30, 2.5, 3.0, &testdata::expected_sens25()); }
    #[test] fn test_sens50() { run_test(5, 30, 5.0, 3.0, &testdata::expected_sens50()); }
    #[test] fn test_period15() { run_test(5, 30, 1.0, 1.5, &testdata::expected_period15()); }
    #[test] fn test_period100() { run_test(5, 30, 1.0, 10.0, &testdata::expected_period100()); }
    #[test] fn test_period300() { run_test(5, 30, 1.0, 30.0, &testdata::expected_period300()); }
}
