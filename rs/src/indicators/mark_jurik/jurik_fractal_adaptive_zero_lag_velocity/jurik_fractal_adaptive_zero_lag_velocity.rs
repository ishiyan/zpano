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
// Scale sets for different fractal types
// ---------------------------------------------------------------------------

fn scale_set(fractal_type: usize) -> &'static [usize] {
    match fractal_type {
        1 => &[2, 3, 4, 6, 8, 12, 16, 24],
        2 => &[2, 3, 4, 6, 8, 12, 16, 24, 32, 48],
        3 => &[2, 3, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96],
        4 => &[2, 3, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96, 128, 192],
        _ => &[2, 3, 4, 6, 8, 12, 16, 24],
    }
}

const WEIGHTS_EVEN: &[f64] = &[2.0, 3.0, 6.0, 12.0, 24.0, 48.0, 96.0];
const WEIGHTS_ODD: &[f64] = &[4.0, 8.0, 16.0, 32.0, 64.0, 128.0, 256.0];

// ---------------------------------------------------------------------------
// CfbAux
// ---------------------------------------------------------------------------

struct CfbAux {
    depth: usize,
    bar: usize,
    int_a: Vec<f64>,
    int_a_idx: usize,
    src: Vec<f64>,
    src_idx: usize,
    jrc04: f64,
    jrc05: f64,
    jrc06: f64,
    prev_sample: f64,
    first_call: bool,
}

impl CfbAux {
    fn new(depth: usize) -> Self {
        Self {
            depth,
            bar: 0,
            int_a: vec![0.0; depth],
            src: vec![0.0; depth + 2],
            src_idx: 0,
            int_a_idx: 0,
            jrc04: 0.0,
            jrc05: 0.0,
            jrc06: 0.0,
            prev_sample: 0.0,
            first_call: true,
        }
    }

    fn update(&mut self, sample: f64) -> f64 {
        self.bar += 1;
        let depth = self.depth;
        let src_size = depth + 2;

        self.src[self.src_idx] = sample;
        self.src_idx = (self.src_idx + 1) % src_size;

        if self.first_call {
            self.first_call = false;
            self.prev_sample = sample;
            return 0.0;
        }

        let int_a_val = (sample - self.prev_sample).abs();
        self.prev_sample = sample;

        let old_int_a = self.int_a[self.int_a_idx];
        self.int_a[self.int_a_idx] = int_a_val;
        self.int_a_idx = (self.int_a_idx + 1) % depth;

        let ref_bar = self.bar - 1;
        if ref_bar < depth {
            return 0.0;
        }

        if ref_bar <= depth * 2 {
            // Recompute from scratch.
            self.jrc04 = 0.0;
            self.jrc05 = 0.0;
            self.jrc06 = 0.0;

            let cur_int_a_pos = (self.int_a_idx + depth - 1) % depth;
            let cur_src_pos = (self.src_idx + src_size - 1) % src_size;

            for j in 0..depth {
                let int_a_pos = (cur_int_a_pos + depth - j) % depth;
                let src_pos = (cur_src_pos + src_size * 2 - j - 1) % src_size;

                self.jrc04 += self.int_a[int_a_pos];
                self.jrc05 += (depth - j) as f64 * self.int_a[int_a_pos];
                self.jrc06 += self.src[src_pos];
            }
        } else {
            // Incremental update.
            self.jrc05 = self.jrc05 - self.jrc04 + int_a_val * depth as f64;
            self.jrc04 = self.jrc04 - old_int_a + int_a_val;

            let cur_src_pos = (self.src_idx + src_size - 1) % src_size;
            let src_bar_minus1 = (cur_src_pos + src_size - 1) % src_size;
            let src_bar_minus_depth_minus1 = (cur_src_pos + src_size * 2 - depth - 1) % src_size;

            self.jrc06 = self.jrc06 - self.src[src_bar_minus_depth_minus1] + self.src[src_bar_minus1];
        }

        let cur_src_pos = (self.src_idx + src_size - 1) % src_size;
        let jrc08 = (depth as f64 * self.src[cur_src_pos] - self.jrc06).abs();

        if self.jrc05 == 0.0 {
            return 0.0;
        }

        jrc08 / self.jrc05
    }
}

// ---------------------------------------------------------------------------
// Cfb
// ---------------------------------------------------------------------------

struct Cfb {
    num_channels: usize,
    auxs: Vec<CfbAux>,
    aux_windows: Vec<Vec<f64>>,
    aux_win_idx: usize,
    er23: Vec<f64>,
    smooth: usize,
    bar: usize,
    cfb_value: f64,
}

impl Cfb {
    fn new(fractal_type: usize, smooth: usize) -> Self {
        let scales = scale_set(fractal_type);
        let n = scales.len();
        let auxs: Vec<CfbAux> = scales.iter().map(|&d| CfbAux::new(d)).collect();
        let aux_windows: Vec<Vec<f64>> = (0..n).map(|_| vec![0.0; smooth]).collect();

        Self {
            num_channels: n,
            auxs,
            aux_windows,
            aux_win_idx: 0,
            er23: vec![0.0; n],
            smooth,
            bar: 0,
            cfb_value: 0.0,
        }
    }

    fn update(&mut self, sample: f64) -> f64 {
        self.bar += 1;
        let ref_bar = self.bar - 1;

        let mut aux_values = vec![0.0_f64; self.num_channels];
        for i in 0..self.num_channels {
            aux_values[i] = self.auxs[i].update(sample);
        }

        if ref_bar == 0 {
            return 0.0;
        }

        let smooth = self.smooth;
        let n = self.num_channels;

        if ref_bar <= smooth {
            let win_pos = self.aux_win_idx;
            for i in 0..n {
                self.aux_windows[i][win_pos] = aux_values[i];
            }
            self.aux_win_idx = (self.aux_win_idx + 1) % smooth;

            for i in 0..n {
                let mut s = 0.0;
                for j in 0..ref_bar {
                    let pos = (self.aux_win_idx + smooth * 2 - 1 - j) % smooth;
                    s += self.aux_windows[i][pos];
                }
                self.er23[i] = s / ref_bar as f64;
            }
        } else {
            let win_pos = self.aux_win_idx;
            for i in 0..n {
                let old_val = self.aux_windows[i][win_pos];
                self.aux_windows[i][win_pos] = aux_values[i];
                self.er23[i] += (aux_values[i] - old_val) / smooth as f64;
            }
            self.aux_win_idx = (self.aux_win_idx + 1) % smooth;
        }

        if ref_bar > 5 {
            let mut er22 = vec![0.0_f64; n];

            // Odd-indexed channels (descending).
            let mut er15 = 1.0_f64;
            let mut idx = n as isize - 1;
            while idx >= 1 {
                er22[idx as usize] = er15 * self.er23[idx as usize];
                er15 *= 1.0 - er22[idx as usize];
                idx -= 2;
            }

            // Even-indexed channels (descending).
            let mut er16 = 1.0_f64;
            idx = n as isize - 2;
            while idx >= 0 {
                er22[idx as usize] = er16 * self.er23[idx as usize];
                er16 *= 1.0 - er22[idx as usize];
                idx -= 2;
            }

            // Weighted sum.
            let mut er17 = 0.0_f64;
            let mut er18 = 0.0_f64;
            for i in 0..n {
                let sq = er22[i] * er22[i];
                er18 += sq;
                if i % 2 == 0 {
                    er17 += sq * WEIGHTS_EVEN[i / 2];
                } else {
                    er17 += sq * WEIGHTS_ODD[i / 2];
                }
            }

            if er18 == 0.0 {
                self.cfb_value = 0.0;
            } else {
                self.cfb_value = er17 / er18;
            }
        }

        self.cfb_value
    }
}

// ---------------------------------------------------------------------------
// VelSmooth (Stage 2)
// ---------------------------------------------------------------------------

struct VelSmooth2 {
    jrc03: f64,
    jrc06: usize,
    _jrc07: usize,
    ema_factor: f64,
    damping: f64,
    eps2: f64,
    buffer_size: usize,
    buffer: Vec<f64>,
    idx: usize,
    length: usize,
    velocity: f64,
    position: f64,
    smoothed_mad: f64,
    mad_init: bool,
    initialized: bool,
}

impl VelSmooth2 {
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
            _jrc07: jrc07,
            ema_factor,
            damping,
            eps2,
            buffer_size: 1001,
            buffer: vec![0.0; 1001],
            idx: 0,
            length: 0,
            velocity: 0.0,
            position: 0.0,
            smoothed_mad: 0.0,
            mad_init: false,
            initialized: false,
        }
    }

    fn update(&mut self, value: f64) -> f64 {
        self.buffer[self.idx] = value;
        self.idx = (self.idx + 1) % self.buffer_size;
        self.length += 1;

        if self.length > self.buffer_size {
            self.length = self.buffer_size;
        }

        let length = self.length;

        if !self.initialized {
            self.initialized = true;
            self.position = value;
            self.velocity = 0.0;
            self.smoothed_mad = 0.0;
            return self.position;
        }

        // Linear regression over capped window.
        let n = length.min(self.jrc06);

        let mut sx = 0.0_f64;
        let mut sy = 0.0_f64;
        let mut sxy = 0.0_f64;
        let mut sx2 = 0.0_f64;

        for i in 0..n {
            let buf_idx = (self.idx + self.buffer_size - 1 - i) % self.buffer_size;
            let x = i as f64;
            let y = self.buffer[buf_idx];
            sx += x;
            sy += y;
            sxy += x * y;
            sx2 += x * x;
        }

        let fn_val = n as f64;
        let slope = if n > 1 {
            (fn_val * sxy - sx * sy) / (fn_val * sx2 - sx * sx)
        } else {
            0.0
        };

        let intercept = (sy - slope * sx) / fn_val;

        // MAD from regression residuals.
        let mut mad = 0.0_f64;
        for i in 0..n {
            let buf_idx = (self.idx + self.buffer_size - 1 - i) % self.buffer_size;
            let predicted = intercept + slope * i as f64;
            mad += (self.buffer[buf_idx] - predicted).abs();
        }
        mad /= fn_val;

        // Scale MAD.
        let scaled_mad = mad * 1.2 * (self.jrc06 as f64 / fn_val).powf(0.25);

        // Smooth MAD with EMA.
        if !self.mad_init {
            self.smoothed_mad = scaled_mad;
            if scaled_mad > 0.0 {
                self.mad_init = true;
            }
        } else {
            self.smoothed_mad += (scaled_mad - self.smoothed_mad) * self.ema_factor;
        }

        let smoothed_mad = self.smoothed_mad.max(self.eps2);

        // Adaptive velocity/position dynamics.
        let prediction_error = value - self.position;
        let response_factor = 1.0 - (-prediction_error.abs() / (smoothed_mad * self.jrc03)).exp();
        self.velocity = response_factor * prediction_error + self.velocity * self.damping;
        self.position += self.velocity;

        self.position
    }
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

pub struct JurikFractalAdaptiveZeroLagVelocityParams {
    pub lo_depth: usize,
    pub hi_depth: usize,
    pub fractal_type: usize,
    pub smooth: usize,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for JurikFractalAdaptiveZeroLagVelocityParams {
    fn default() -> Self {
        Self {
            lo_depth: 5,
            hi_depth: 30,
            fractal_type: 1,
            smooth: 10,
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
pub enum JurikFractalAdaptiveZeroLagVelocityOutput {
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

#[derive(Debug)]
pub struct JurikFractalAdaptiveZeroLagVelocity {
    primed: bool,
    lo_depth: usize,
    hi_depth: usize,
    prices: Vec<f64>,
    bar_count: usize,
    cfb_inst: Cfb,
    cfb_min: Option<f64>,
    cfb_max: Option<f64>,
    smooth: VelSmooth2,
    bar_func: BarFunc,
    quote_func: QuoteFunc,
    trade_func: TradeFunc,
    mnemonic: String,
    description: String,
}

impl std::fmt::Debug for Cfb {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("Cfb").finish()
    }
}

impl std::fmt::Debug for VelSmooth2 {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("VelSmooth2").finish()
    }
}

impl JurikFractalAdaptiveZeroLagVelocity {
    pub fn new(params: &JurikFractalAdaptiveZeroLagVelocityParams) -> Result<Self, String> {
        if params.lo_depth < 2 {
            return Err("invalid jurik fractal adaptive zero lag velocity parameters: lo_depth should be at least 2".to_string());
        }
        if params.hi_depth < params.lo_depth {
            return Err("invalid jurik fractal adaptive zero lag velocity parameters: hi_depth should be at least lo_depth".to_string());
        }
        if params.fractal_type < 1 || params.fractal_type > 4 {
            return Err("invalid jurik fractal adaptive zero lag velocity parameters: fractal_type should be 1-4".to_string());
        }
        if params.smooth < 1 {
            return Err("invalid jurik fractal adaptive zero lag velocity parameters: smooth should be at least 1".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let triple = component_triple_mnemonic(bc, qc, tc);
        let mnemonic = format!("jvelcfb({}, {}, {}, {}{})", params.lo_depth, params.hi_depth, params.fractal_type, params.smooth, triple);
        let description = format!("Jurik fractal adaptive zero lag velocity {}", mnemonic);

        Ok(Self {
            primed: false,
            lo_depth: params.lo_depth,
            hi_depth: params.hi_depth,
            prices: Vec::new(),
            bar_count: 0,
            cfb_inst: Cfb::new(params.fractal_type, params.smooth),
            cfb_min: None,
            cfb_max: None,
            smooth: VelSmooth2::new(3.0),
            bar_func,
            quote_func,
            trade_func,
            mnemonic,
            description,
        })
    }

    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        let bar = self.bar_count;
        self.bar_count += 1;

        self.prices.push(sample);

        // CFB computation.
        let cfb_val = self.cfb_inst.update(sample);

        if bar == 0 {
            return f64::NAN;
        }

        // Stochastic normalization.
        match (self.cfb_min, self.cfb_max) {
            (None, _) => {
                self.cfb_min = Some(cfb_val);
                self.cfb_max = Some(cfb_val);
            }
            (Some(min), Some(max)) => {
                if cfb_val < min {
                    self.cfb_min = Some(cfb_val);
                }
                if cfb_val > max {
                    self.cfb_max = Some(cfb_val);
                }
            }
            _ => {}
        }

        let cfb_min = self.cfb_min.unwrap();
        let cfb_max = self.cfb_max.unwrap();
        let cfb_range = cfb_max - cfb_min;

        let sr = if cfb_range != 0.0 {
            (cfb_val - cfb_min) / cfb_range
        } else {
            0.5
        };

        let depth_f = self.lo_depth as f64 + sr * (self.hi_depth - self.lo_depth) as f64;
        let depth = depth_f.round() as usize;

        // Stage 1: WLS slope.
        if bar < depth {
            return f64::NAN;
        }

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

        let slope = (sum_xw2 * s1 - sum_xw * s2) / denom;

        // Stage 2: adaptive smoother.
        let result = self.smooth.update(slope);

        if !self.primed {
            self.primed = true;
        }

        result
    }
}

impl Indicator for JurikFractalAdaptiveZeroLagVelocity {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::JurikFractalAdaptiveZeroLagVelocity,
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
    use crate::indicators::mark_jurik::jurik_fractal_adaptive_zero_lag_velocity::testdata::testdata;

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

    fn run_test(lo: usize, hi: usize, ftype: usize, smooth: usize, expected: &[f64]) {
        let params = JurikFractalAdaptiveZeroLagVelocityParams {
            lo_depth: lo,
            hi_depth: hi,
            fractal_type: ftype,
            smooth,
            ..Default::default()
        };
        let mut ind = JurikFractalAdaptiveZeroLagVelocity::new(&params).unwrap();
        let input = testdata::test_input();
        for (i, &val) in input.iter().enumerate() {
            let result = ind.update(val);
            assert!(
                almost_equal(result, expected[i]),
                "bar {}: expected {}, got {} (lo={}, hi={}, ftype={}, smooth={})",
                i, expected[i], result, lo, hi, ftype, smooth
            );
        }
    }

    #[test] fn test_lo2_hi15() { run_test(2, 15, 1, 10, &testdata::expected_lo2_hi15()); }
    #[test] fn test_lo2_hi30() { run_test(2, 30, 1, 10, &testdata::expected_lo2_hi30()); }
    #[test] fn test_lo2_hi60() { run_test(2, 60, 1, 10, &testdata::expected_lo2_hi60()); }
    #[test] fn test_lo5_hi15() { run_test(5, 15, 1, 10, &testdata::expected_lo5_hi15()); }
    #[test] fn test_lo5_hi30() { run_test(5, 30, 1, 10, &testdata::expected_lo5_hi30()); }
    #[test] fn test_lo5_hi60() { run_test(5, 60, 1, 10, &testdata::expected_lo5_hi60()); }
    #[test] fn test_lo10_hi15() { run_test(10, 15, 1, 10, &testdata::expected_lo10_hi15()); }
    #[test] fn test_lo10_hi30() { run_test(10, 30, 1, 10, &testdata::expected_lo10_hi30()); }
    #[test] fn test_lo10_hi60() { run_test(10, 60, 1, 10, &testdata::expected_lo10_hi60()); }
    #[test] fn test_ftype2() { run_test(5, 30, 2, 10, &testdata::expected_ftype2()); }
    #[test] fn test_ftype3() { run_test(5, 30, 3, 10, &testdata::expected_ftype3()); }
    #[test] fn test_ftype4() { run_test(5, 30, 4, 10, &testdata::expected_ftype4()); }
    #[test] fn test_smooth5() { run_test(5, 30, 1, 5, &testdata::expected_smooth5()); }
    #[test] fn test_smooth20() { run_test(5, 30, 1, 20, &testdata::expected_smooth20()); }
    #[test] fn test_smooth40() { run_test(5, 30, 1, 40, &testdata::expected_smooth40()); }
}
