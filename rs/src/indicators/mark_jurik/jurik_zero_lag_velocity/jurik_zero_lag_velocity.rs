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
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the Jurik Zero Lag Velocity indicator.
pub struct JurikZeroLagVelocityParams {
    /// Depth controls the linear regression window (window = depth+1). Must be >= 2. Default 10.
    pub depth: usize,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for JurikZeroLagVelocityParams {
    fn default() -> Self {
        Self {
            depth: 10,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Jurik Zero Lag Velocity indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum JurikZeroLagVelocityOutput {
    ZeroLagVelocity = 1,
}

// ---------------------------------------------------------------------------
// VelAux1 — linear regression slope
// ---------------------------------------------------------------------------

#[derive(Debug)]
struct VelAux1 {
    depth: usize,
    win: Vec<f64>,
    idx: usize,
    bar: usize,
    jrc04: f64,
    jrc05: f64,
    jrc06: f64,
    jrc07: f64,
}

impl VelAux1 {
    fn new(depth: usize) -> Self {
        let size = depth + 1;
        let jrc04 = size as f64;
        let jrc05 = jrc04 * (jrc04 + 1.0) / 2.0;
        let jrc06 = jrc05 * (2.0 * jrc04 + 1.0) / 3.0;
        let jrc07 = jrc05 * jrc05 * jrc05 - jrc06 * jrc06;

        Self {
            depth,
            win: vec![0.0; size],
            idx: 0,
            bar: 0,
            jrc04,
            jrc05,
            jrc06,
            jrc07,
        }
    }

    fn update(&mut self, sample: f64) -> f64 {
        let size = self.depth + 1;
        self.win[self.idx] = sample;
        self.idx = (self.idx + 1) % size;
        self.bar += 1;

        if self.bar <= self.depth {
            return 0.0;
        }

        let mut jrc08 = 0.0_f64;
        let mut jrc09 = 0.0_f64;

        for j in 0..=self.depth {
            let pos = (self.idx + size * 2 - 1 - j) % size;
            let w = self.jrc04 - j as f64;
            jrc08 += self.win[pos] * w;
            jrc09 += self.win[pos] * w * w;
        }

        (jrc09 * self.jrc05 - jrc08 * self.jrc06) / self.jrc07
    }
}

// ---------------------------------------------------------------------------
// VelAux3State — adaptive smoother
// ---------------------------------------------------------------------------

#[derive(Debug)]
struct VelAux3State {
    length: usize,
    eps: f64,
    decay: usize,
    beta: f64,
    alpha: f64,
    max_win: usize,

    src_ring: [f64; 100],
    dev_ring: [f64; 100],
    src_idx: usize,
    dev_idx: usize,

    jr08: f64,
    jr09: f64,
    jr10: f64,
    jr11: usize,
    jr12: f64,
    jr13: f64,
    jr14: f64,
    jr19: f64,
    jr20: f64,
    jr21: f64,
    jr21a: f64,
    jr21b: f64,
    jr22: f64,
    jr23: f64,

    bar: usize,
    init_done: bool,
    history: Vec<f64>,
}

impl VelAux3State {
    fn new() -> Self {
        let length = 30;
        let decay = 3;
        Self {
            length,
            eps: 0.0001,
            decay,
            beta: 0.86 - 0.55 / (decay as f64).sqrt(),
            alpha: 1.0 - (-((4.0_f64).ln()) / decay as f64 / 2.0).exp(),
            max_win: length + 1,
            src_ring: [0.0; 100],
            dev_ring: [0.0; 100],
            src_idx: 0,
            dev_idx: 0,
            jr08: 0.0,
            jr09: 0.0,
            jr10: 0.0,
            jr11: 0,
            jr12: 0.0,
            jr13: 0.0,
            jr14: 0.0,
            jr19: 0.0,
            jr20: 0.0,
            jr21: 0.0,
            jr21a: 0.0,
            jr21b: 0.0,
            jr22: 0.0,
            jr23: 0.0,
            bar: 0,
            init_done: false,
            history: Vec::with_capacity(length),
        }
    }

    fn feed(&mut self, sample: f64, bar_idx: usize) -> f64 {
        if bar_idx < self.length {
            self.history.push(sample);
            return 0.0;
        }

        self.bar += 1;

        if !self.init_done {
            self.init_done = true;

            // Count consecutive equal values.
            let mut jr28 = 0.0_f64;
            let hist_len = self.history.len();
            for j in 1..self.length {
                if self.history[hist_len - j] == self.history[hist_len - j - 1] {
                    jr28 += 1.0;
                }
            }

            let jr26 = if jr28 < (self.length - 1) as f64 {
                bar_idx - self.length
            } else {
                bar_idx
            };

            self.jr11 = ((1 + bar_idx - jr26) as f64)
                .min(self.max_win as f64)
                .trunc() as usize;

            // jr21 = history[last-1] (i.e. second to last)
            self.jr21 = self.history[hist_len - 1];

            // jr08 = (sample - history[last-3]) / 3
            let jr07 = 3;
            self.jr08 = (sample - self.history[hist_len - jr07]) / jr07 as f64;

            // Fill source ring with historical values.
            for jr15 in (1..self.jr11).rev() {
                if self.src_idx == 0 {
                    self.src_idx = 100;
                }
                self.src_idx -= 1;
                self.src_ring[self.src_idx] = self.history[hist_len - jr15];
            }

            self.history = Vec::new(); // free memory
        }

        // Push current value to source ring.
        if self.src_idx == 0 {
            self.src_idx = 100;
        }
        self.src_idx -= 1;
        self.src_ring[self.src_idx] = sample;

        if self.jr11 <= self.length {
            // Growing phase.
            if self.bar == 1 {
                self.jr21 = sample;
            } else {
                self.jr21 = self.alpha.sqrt() * sample + (1.0 - self.alpha.sqrt()) * self.jr21a;
            }

            if self.bar > 2 {
                self.jr08 = (self.jr21 - self.jr21b) / 2.0;
            } else {
                self.jr08 = 0.0;
            }

            self.jr11 += 1;
        } else if self.jr11 <= self.max_win {
            // Transition phase: recompute from scratch.
            self.jr12 = (self.jr11 * (self.jr11 + 1) * (self.jr11 - 1)) as f64 / 12.0;
            self.jr13 = (self.jr11 + 1) as f64 / 2.0;
            self.jr14 = (self.jr11 - 1) as f64 / 2.0;

            self.jr09 = 0.0;
            self.jr10 = 0.0;

            for jr15 in (0..self.jr11).rev() {
                let jr24 = (self.src_idx + jr15) % 100;
                self.jr09 += self.src_ring[jr24];
                self.jr10 += self.src_ring[jr24] * (self.jr14 - jr15 as f64);
            }

            let jr16 = self.jr10 / self.jr12;
            let mut jr17 = (self.jr09 / self.jr11 as f64) - (jr16 * self.jr13);

            self.jr19 = 0.0;
            for jr15 in (0..self.jr11).rev() {
                jr17 += jr16;
                let jr24 = (self.src_idx + jr15) % 100;
                self.jr19 += (self.src_ring[jr24] - jr17).abs();
            }

            self.jr20 = (self.jr19 / self.jr11 as f64)
                * (self.max_win as f64 / self.jr11 as f64).powf(0.25);
            self.jr11 += 1;

            // Adaptive step.
            self.jr20 = self.jr20.max(self.eps);
            self.jr22 = sample - (self.jr21 + self.jr08 * self.beta);
            self.jr23 = 1.0 - (-self.jr22.abs() / self.jr20 / self.decay as f64).exp();
            self.jr08 = self.jr23 * self.jr22 + self.jr08 * self.beta;
            self.jr21 += self.jr08;
        } else {
            // Steady state.
            let jr24out = (self.src_idx + self.max_win) % 100;
            self.jr10 =
                self.jr10 - self.jr09 + self.src_ring[jr24out] * self.jr13 + sample * self.jr14;
            self.jr09 = self.jr09 - self.src_ring[jr24out] + sample;

            // Deviation ring update.
            if self.dev_idx == 0 {
                self.dev_idx = self.max_win;
            }
            self.dev_idx -= 1;
            self.jr19 -= self.dev_ring[self.dev_idx];

            let jr16 = self.jr10 / self.jr12;
            let jr17 = (self.jr09 / self.max_win as f64) + (jr16 * self.jr14);
            self.dev_ring[self.dev_idx] = (sample - jr17).abs();
            self.jr19 = (self.jr19 + self.dev_ring[self.dev_idx]).max(self.eps);
            self.jr20 += ((self.jr19 / self.max_win as f64) - self.jr20) * self.alpha;

            // Adaptive step.
            self.jr20 = self.jr20.max(self.eps);
            self.jr22 = sample - (self.jr21 + self.jr08 * self.beta);
            self.jr23 = 1.0 - (-self.jr22.abs() / self.jr20 / self.decay as f64).exp();
            self.jr08 = self.jr23 * self.jr22 + self.jr08 * self.beta;
            self.jr21 += self.jr08;
        }

        self.jr21b = self.jr21a;
        self.jr21a = self.jr21;

        self.jr21
    }
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Mark Jurik's Zero Lag Velocity (VEL) indicator.
#[derive(Debug)]
pub struct JurikZeroLagVelocity {
    primed: bool,
    param_depth: usize,
    aux1: VelAux1,
    aux3: VelAux3State,
    bar: usize,
    bar_func: BarFunc,
    quote_func: QuoteFunc,
    trade_func: TradeFunc,
    mnemonic: String,
    description: String,
}

impl JurikZeroLagVelocity {
    pub fn new(params: &JurikZeroLagVelocityParams) -> Result<Self, String> {
        if params.depth < 2 {
            return Err(
                "invalid jurik zero lag velocity parameters: depth should be at least 2"
                    .to_string(),
            );
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!(
            "jvel({}{})",
            params.depth,
            component_triple_mnemonic(bc, qc, tc)
        );
        let description = format!("Jurik zero lag velocity {}", mnemonic);

        Ok(Self {
            primed: false,
            param_depth: params.depth,
            aux1: VelAux1::new(params.depth),
            aux3: VelAux3State::new(),
            bar: 0,
            bar_func,
            quote_func,
            trade_func,
            mnemonic,
            description,
        })
    }

    /// Core update. Returns the VEL value.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        // Stage 1: compute linear regression slope.
        let aux1_val = self.aux1.update(sample);

        // Stage 2: feed into adaptive smoother.
        let bar_idx = self.bar;
        self.bar += 1;

        let result = self.aux3.feed(aux1_val, bar_idx);

        // Output NaN during warmup.
        if bar_idx < self.aux3.length {
            return f64::NAN;
        }

        if !self.primed {
            self.primed = true;
        }

        result
    }
}

impl Indicator for JurikZeroLagVelocity {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::JurikZeroLagVelocity,
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
    use super::*;
    use super::super::testdata::testdata;
    use crate::indicators::core::indicator::Indicator;
    use crate::indicators::core::outputs::shape::Shape;

    const TOLERANCE: f64 = 1e-13;

    fn almost_equal(a: f64, b: f64) -> bool {
        (a - b).abs() <= TOLERANCE
    }

    fn run_vel_test(depth: usize, expected: &[f64]) {
        let params = JurikZeroLagVelocityParams { depth, ..Default::default() };
        let mut vel = JurikZeroLagVelocity::new(&params).unwrap();
        let input = testdata::test_input();
        for (i, &val) in input.iter().enumerate() {
            let result = vel.update(val);
            if expected[i].is_nan() {
                assert!(result.is_nan(), "bar {}: expected NaN, got {}", i, result);
            } else {
                assert!(
                    almost_equal(result, expected[i]),
                    "bar {}: expected {}, got {}, diff {}",
                    i, expected[i], result, (result - expected[i]).abs()
                );
            }
        }
    }

    #[test] fn test_vel_depth_2() { run_vel_test(2, &testdata::expected_depth2()); }
    #[test] fn test_vel_depth_3() { run_vel_test(3, &testdata::expected_depth3()); }
    #[test] fn test_vel_depth_4() { run_vel_test(4, &testdata::expected_depth4()); }
    #[test] fn test_vel_depth_5() { run_vel_test(5, &testdata::expected_depth5()); }
    #[test] fn test_vel_depth_6() { run_vel_test(6, &testdata::expected_depth6()); }
    #[test] fn test_vel_depth_7() { run_vel_test(7, &testdata::expected_depth7()); }
    #[test] fn test_vel_depth_8() { run_vel_test(8, &testdata::expected_depth8()); }
    #[test] fn test_vel_depth_9() { run_vel_test(9, &testdata::expected_depth9()); }
    #[test] fn test_vel_depth_10() { run_vel_test(10, &testdata::expected_depth10()); }
    #[test] fn test_vel_depth_11() { run_vel_test(11, &testdata::expected_depth11()); }
    #[test] fn test_vel_depth_12() { run_vel_test(12, &testdata::expected_depth12()); }
    #[test] fn test_vel_depth_13() { run_vel_test(13, &testdata::expected_depth13()); }
    #[test] fn test_vel_depth_14() { run_vel_test(14, &testdata::expected_depth14()); }
    #[test] fn test_vel_depth_15() { run_vel_test(15, &testdata::expected_depth15()); }

    #[test]
    fn test_vel_metadata() {
        let params = JurikZeroLagVelocityParams::default();
        let vel = JurikZeroLagVelocity::new(&params).unwrap();
        let md = vel.metadata();
        assert_eq!(md.outputs.len(), 1);
        assert_eq!(md.outputs[0].shape, Shape::Scalar);
    }

    #[test]
    fn test_vel_invalid_params() {
        let params = JurikZeroLagVelocityParams { depth: 1, ..Default::default() };
        assert!(JurikZeroLagVelocity::new(&params).is_err());
    }
}
