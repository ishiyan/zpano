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
// Depth sets
// ---------------------------------------------------------------------------

const DEPTH_SET_1: &[usize] = &[2, 3, 4, 6, 8, 12, 16, 24];
const DEPTH_SET_2: &[usize] = &[2, 3, 4, 6, 8, 12, 16, 24, 32, 48];
const DEPTH_SET_3: &[usize] = &[2, 3, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96];
const DEPTH_SET_4: &[usize] = &[2, 3, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96, 128, 192];

const WEIGHTS_EVEN: &[f64] = &[2.0, 3.0, 6.0, 12.0, 24.0, 48.0, 96.0];
const WEIGHTS_ODD: &[f64] = &[4.0, 8.0, 16.0, 32.0, 64.0, 128.0, 256.0];

fn depth_set(fractal_type: usize) -> &'static [usize] {
    match fractal_type {
        1 => DEPTH_SET_1,
        2 => DEPTH_SET_2,
        3 => DEPTH_SET_3,
        4 => DEPTH_SET_4,
        _ => DEPTH_SET_1,
    }
}

// ---------------------------------------------------------------------------
// CfbAux
// ---------------------------------------------------------------------------

#[derive(Debug)]
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
            int_a_idx: 0,
            src: vec![0.0; depth + 2],
            src_idx: 0,
            jrc04: 0.0,
            jrc05: 0.0,
            jrc06: 0.0,
            prev_sample: 0.0,
            first_call: true,
        }
    }

    fn update(&mut self, sample: f64) -> f64 {
        self.bar += 1;

        let src_size = self.depth + 2;
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
        self.int_a_idx = (self.int_a_idx + 1) % self.depth;

        let ref_bar = self.bar - 1;
        if ref_bar < self.depth {
            return 0.0;
        }

        if ref_bar <= self.depth * 2 {
            self.jrc04 = 0.0;
            self.jrc05 = 0.0;
            self.jrc06 = 0.0;

            let cur_int_a_pos = (self.int_a_idx + self.depth - 1) % self.depth;
            let cur_src_pos = (self.src_idx + src_size - 1) % src_size;

            for j in 0..self.depth {
                let int_a_pos = (cur_int_a_pos + self.depth - j) % self.depth;
                let int_a_v = self.int_a[int_a_pos];

                let src_pos = (cur_src_pos + src_size * 2 - j - 1) % src_size;
                let src_v = self.src[src_pos];

                self.jrc04 += int_a_v;
                self.jrc05 += (self.depth - j) as f64 * int_a_v;
                self.jrc06 += src_v;
            }
        } else {
            self.jrc05 = self.jrc05 - self.jrc04 + int_a_val * self.depth as f64;
            self.jrc04 = self.jrc04 - old_int_a + int_a_val;

            let cur_src_pos = (self.src_idx + src_size - 1) % src_size;
            let src_bar_minus1 = (cur_src_pos + src_size - 1) % src_size;
            let src_bar_minus_depth_minus1 = (cur_src_pos + src_size - self.depth - 1) % src_size;

            self.jrc06 =
                self.jrc06 - self.src[src_bar_minus_depth_minus1] + self.src[src_bar_minus1];
        }

        let cur_src_pos = (self.src_idx + src_size - 1) % src_size;
        let jrc08 = (self.depth as f64 * self.src[cur_src_pos] - self.jrc06).abs();

        if self.jrc05 == 0.0 {
            return 0.0;
        }

        jrc08 / self.jrc05
    }
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the Jurik Composite Fractal Behavior Index indicator.
pub struct JurikCompositeFractalBehaviorIndexParams {
    /// Fractal type (1-4). Default 1.
    pub fractal_type: usize,
    /// Smooth period (>=1). Default 10.
    pub smooth: usize,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for JurikCompositeFractalBehaviorIndexParams {
    fn default() -> Self {
        Self {
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

/// Enumerates the outputs of the Jurik Composite Fractal Behavior Index indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum JurikCompositeFractalBehaviorIndexOutput {
    CompositeFractalBehaviorIndex = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Mark Jurik's Composite Fractal Behavior Index (CFB).
#[derive(Debug)]
pub struct JurikCompositeFractalBehaviorIndex {
    primed: bool,
    param_fractal: usize,
    param_smooth: usize,
    num_channels: usize,
    aux_instances: Vec<CfbAux>,
    aux_windows: Vec<Vec<f64>>,
    aux_win_idx: usize,
    aux_win_len: usize,
    er23: Vec<f64>,
    bar: usize,
    er19: f64,
    bar_func: BarFunc,
    quote_func: QuoteFunc,
    trade_func: TradeFunc,
    mnemonic: String,
    description: String,
}

impl JurikCompositeFractalBehaviorIndex {
    pub fn new(params: &JurikCompositeFractalBehaviorIndexParams) -> Result<Self, String> {
        if params.fractal_type < 1 || params.fractal_type > 4 {
            return Err("invalid jurik composite fractal behavior index parameters: fractal type should be between 1 and 4".to_string());
        }
        if params.smooth < 1 {
            return Err("invalid jurik composite fractal behavior index parameters: smooth should be at least 1".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!(
            "jcfb({},{}{})",
            params.fractal_type,
            params.smooth,
            component_triple_mnemonic(bc, qc, tc)
        );
        let description = format!("Jurik composite fractal behavior index {}", mnemonic);

        let depths = depth_set(params.fractal_type);
        let num_channels = depths.len();

        let aux_instances: Vec<CfbAux> = depths.iter().map(|&d| CfbAux::new(d)).collect();
        let aux_windows: Vec<Vec<f64>> = (0..num_channels)
            .map(|_| vec![0.0; params.smooth])
            .collect();
        let er23 = vec![0.0; num_channels];

        Ok(Self {
            primed: false,
            param_fractal: params.fractal_type,
            param_smooth: params.smooth,
            num_channels,
            aux_instances,
            aux_windows,
            aux_win_idx: 0,
            aux_win_len: 0,
            er23,
            bar: 0,
            er19: 20.0,
            bar_func,
            quote_func,
            trade_func,
            mnemonic,
            description,
        })
    }

    /// Core update. Returns the CFB value.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        self.bar += 1;

        // Feed all aux instances.
        let mut aux_values = vec![0.0; self.num_channels];
        for i in 0..self.num_channels {
            aux_values[i] = self.aux_instances[i].update(sample);
        }

        // Bar 1 returns NaN.
        if self.bar == 1 {
            return f64::NAN;
        }

        let ref_bar = self.bar - 1;
        let smooth = self.param_smooth;

        // Update running averages.
        if ref_bar <= smooth {
            let win_pos = self.aux_win_idx;
            for i in 0..self.num_channels {
                self.aux_windows[i][win_pos] = aux_values[i];
            }
            self.aux_win_idx = (self.aux_win_idx + 1) % smooth;
            self.aux_win_len = ref_bar;

            for i in 0..self.num_channels {
                let mut sum = 0.0;
                for j in 0..ref_bar {
                    let pos = (self.aux_win_idx + smooth * 2 - 1 - j) % smooth;
                    sum += self.aux_windows[i][pos];
                }
                self.er23[i] = sum / ref_bar as f64;
            }
        } else {
            let win_pos = self.aux_win_idx;
            for i in 0..self.num_channels {
                let old_val = self.aux_windows[i][win_pos];
                self.aux_windows[i][win_pos] = aux_values[i];
                self.er23[i] += (aux_values[i] - old_val) / smooth as f64;
            }
            self.aux_win_idx = (self.aux_win_idx + 1) % smooth;
        }

        // Compute weighted composite when refBar > 5.
        if ref_bar > 5 {
            let n = self.num_channels;
            let mut er22 = vec![0.0; n];

            // Odd-indexed channels (descending).
            let mut er15 = 1.0;
            let mut idx = n as isize - 1;
            while idx >= 1 {
                let i = idx as usize;
                er22[i] = er15 * self.er23[i];
                er15 *= 1.0 - er22[i];
                idx -= 2;
            }

            // Even-indexed channels (descending).
            let mut er16 = 1.0;
            idx = n as isize - 2;
            while idx >= 0 {
                let i = idx as usize;
                er22[i] = er16 * self.er23[i];
                er16 *= 1.0 - er22[i];
                idx -= 2;
            }

            // Weighted sum.
            let mut er17 = 0.0;
            let mut er18 = 0.0;
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
                self.er19 = 0.0;
            } else {
                self.er19 = er17 / er18;
            }
        }

        if !self.primed && ref_bar > 5 {
            self.primed = true;
        }

        self.er19
    }
}

impl Indicator for JurikCompositeFractalBehaviorIndex {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::JurikCompositeFractalBehaviorIndex,
            &self.mnemonic,
            &self.description,
            &[OutputText {
                mnemonic: self.mnemonic.clone(),
                description: self.description.clone(),
            }],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let val = self.update(sample.value);
        vec![Box::new(Scalar::new(sample.time, val))]
    }

    fn update_bar(&mut self, bar: &Bar) -> Output {
        let sample = (self.bar_func)(bar);
        let val = self.update(sample);
        vec![Box::new(Scalar::new(bar.time, val))]
    }

    fn update_quote(&mut self, quote: &Quote) -> Output {
        let sample = (self.quote_func)(quote);
        let val = self.update(sample);
        vec![Box::new(Scalar::new(quote.time, val))]
    }

    fn update_trade(&mut self, trade: &Trade) -> Output {
        let sample = (self.trade_func)(trade);
        let val = self.update(sample);
        vec![Box::new(Scalar::new(trade.time, val))]
    }
}

// ===========================================================================
// Tests
// ===========================================================================

#[cfg(test)]
mod tests {
    use super::super::testdata::testdata;
    use super::*;
    use crate::indicators::core::indicator::Indicator;
    use crate::indicators::core::outputs::shape::Shape;

    const TOLERANCE: f64 = 1e-13;

    fn almost_equal(a: f64, b: f64) -> bool {
        (a - b).abs() <= TOLERANCE
    }

    fn run_cfb_test(fractal_type: usize, smooth: usize, expected: &[f64]) {
        let params = JurikCompositeFractalBehaviorIndexParams {
            fractal_type,
            smooth,
            ..Default::default()
        };
        let mut cfb = JurikCompositeFractalBehaviorIndex::new(&params).unwrap();
        let input = testdata::test_input();
        // JCFB reference skips the last bar (reference aux loop stops at len-2)
        let count = input.len() - 1;
        for i in 0..count {
            let result = cfb.update(input[i]);
            if expected[i].is_nan() {
                assert!(result.is_nan(), "bar {}: expected NaN, got {}", i, result);
            } else {
                assert!(
                    almost_equal(result, expected[i]),
                    "bar {}: expected {}, got {}, diff {}",
                    i,
                    expected[i],
                    result,
                    (result - expected[i]).abs()
                );
            }
        }
    }

    #[test]
    fn test_cfb_type1_smooth10() {
        run_cfb_test(1, 10, &testdata::expected_type1_smooth10());
    }
    #[test]
    fn test_cfb_type1_smooth2() {
        run_cfb_test(1, 2, &testdata::expected_type1_smooth2());
    }
    #[test]
    fn test_cfb_type1_smooth50() {
        run_cfb_test(1, 50, &testdata::expected_type1_smooth50());
    }
    #[test]
    fn test_cfb_type2_smooth2() {
        run_cfb_test(2, 2, &testdata::expected_type2_smooth2());
    }
    #[test]
    fn test_cfb_type2_smooth10() {
        run_cfb_test(2, 10, &testdata::expected_type2_smooth10());
    }
    #[test]
    fn test_cfb_type2_smooth50() {
        run_cfb_test(2, 50, &testdata::expected_type2_smooth50());
    }
    #[test]
    fn test_cfb_type3_smooth2() {
        run_cfb_test(3, 2, &testdata::expected_type3_smooth2());
    }
    #[test]
    fn test_cfb_type3_smooth10() {
        run_cfb_test(3, 10, &testdata::expected_type3_smooth10());
    }
    #[test]
    fn test_cfb_type3_smooth50() {
        run_cfb_test(3, 50, &testdata::expected_type3_smooth50());
    }
    #[test]
    fn test_cfb_type4_smooth2() {
        run_cfb_test(4, 2, &testdata::expected_type4_smooth2());
    }
    #[test]
    fn test_cfb_type4_smooth10() {
        run_cfb_test(4, 10, &testdata::expected_type4_smooth10());
    }
    #[test]
    fn test_cfb_type4_smooth50() {
        run_cfb_test(4, 50, &testdata::expected_type4_smooth50());
    }

    #[test]
    fn test_cfb_metadata() {
        let params = JurikCompositeFractalBehaviorIndexParams::default();
        let cfb = JurikCompositeFractalBehaviorIndex::new(&params).unwrap();
        let md = cfb.metadata();
        assert_eq!(md.outputs.len(), 1);
        assert_eq!(md.outputs[0].shape, Shape::Scalar);
    }

    #[test]
    fn test_cfb_invalid_fractal_type() {
        let params = JurikCompositeFractalBehaviorIndexParams {
            fractal_type: 0,
            ..Default::default()
        };
        assert!(JurikCompositeFractalBehaviorIndex::new(&params).is_err());
        let params = JurikCompositeFractalBehaviorIndexParams {
            fractal_type: 5,
            ..Default::default()
        };
        assert!(JurikCompositeFractalBehaviorIndex::new(&params).is_err());
    }
}
