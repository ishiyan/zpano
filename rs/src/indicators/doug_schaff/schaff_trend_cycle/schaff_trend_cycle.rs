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
use crate::indicators::core::metadata::Metadata;

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the Schaff Trend Cycle indicator.
pub struct SchaffTrendCycleParams {
    /// Fast EMA length for the MACD line. Must be > 0. Default 23.
    pub fast: usize,
    /// Slow EMA length for the MACD line; also the warm-up gate. Must be > 0. Default 50.
    pub slow: usize,
    /// Cycle length — the look-back for both stochastics. Must be > 0. Default 10.
    pub tclen: usize,
    /// EMA smoothing alpha for both %D stages. Must be in (0, 1]. Default 0.5.
    pub factor: f64,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for SchaffTrendCycleParams {
    fn default() -> Self {
        Self {
            fast: 23,
            slow: 50,
            tclen: 10,
            factor: 0.5,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Schaff Trend Cycle indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum SchaffTrendCycleOutput {
    /// The STC oscillator value (range [0, 100]).
    Stc = 1,
    /// The gated MACD line (XMAC) value.
    Macd = 2,
    /// The first smoothed %D stage (PF) value.
    Pf = 3,
}

// ---------------------------------------------------------------------------
// Inlined Blau EMA
// ---------------------------------------------------------------------------

/// Stateful streaming EMA: alpha = 2/(period+1), seeds e0 = x0.
///
/// Inlined verbatim from the Blau exponential moving average so the indicator is a
/// standalone porting unit. Do NOT change its numerics.
struct Ema {
    alpha: f64,
    prev: f64,
    primed: bool,
}

impl Ema {
    fn new(period: usize) -> Self {
        Self {
            alpha: 2.0 / (period as f64 + 1.0),
            prev: 0.0,
            primed: false,
        }
    }

    fn update(&mut self, x: f64) -> f64 {
        if !self.primed {
            self.prev = x;
            self.primed = true;
            return self.prev;
        }
        self.prev = self.alpha * x + (1.0 - self.alpha) * self.prev;
        self.prev
    }
}

// ---------------------------------------------------------------------------
// Window
// ---------------------------------------------------------------------------

/// A fixed-capacity ring buffer of the last n values, providing min/max.
struct Window {
    data: Vec<f64>,
    pos: usize,
    count: usize,
}

impl Window {
    fn new(n: usize) -> Self {
        Self {
            data: vec![0.0; n],
            pos: 0,
            count: 0,
        }
    }

    fn push(&mut self, v: f64) {
        self.data[self.pos] = v;
        self.pos = (self.pos + 1) % self.data.len();
        if self.count < self.data.len() {
            self.count += 1;
        }
    }

    fn min_max(&self) -> (f64, f64) {
        let mut min_val = self.data[0];
        let mut max_val = self.data[0];
        for i in 1..self.count {
            let v = self.data[i];
            if v < min_val {
                min_val = v;
            }
            if v > max_val {
                max_val = v;
            }
        }
        (min_val, max_val)
    }
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Doug Schaff's Schaff Trend Cycle (STC).
///
/// STC runs a MACD line through two cascaded stochastics, each followed by an
/// EMA-style smoothing, producing a cyclical oscillator bounded to [0, 100].
///
/// The indicator produces three outputs:
///   - STC: the oscillator, range [0, 100], NaN during warm-up (bars 0..slow);
///   - MACD: the gated MACD line XMAC (0.0 pre-gate), exposed for stage testing;
///   - PF: the first smoothed %D (0.0 pre-gate), exposed for stage testing.
pub struct SchaffTrendCycle {
    ema_fast: Ema,
    ema_slow: Ema,
    slow: usize,
    factor: f64,
    bar: i64,
    macd_win: Window,
    pf_win: Window,
    frac1: f64,
    frac2: f64,
    pf: f64,
    pff: f64,
    primed: bool,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
    mnemonic: String,
}

impl SchaffTrendCycle {
    /// Creates a new Schaff Trend Cycle from the given parameters.
    pub fn new(params: &SchaffTrendCycleParams) -> Result<Self, String> {
        let invalid = "invalid schaff trend cycle parameters";

        let mut fast = params.fast;
        if fast == 0 {
            fast = 23;
        }
        let mut slow = params.slow;
        if slow == 0 {
            slow = 50;
        }
        let mut tclen = params.tclen;
        if tclen == 0 {
            tclen = 10;
        }
        let mut factor = params.factor;
        if factor == 0.0 {
            factor = 0.5;
        }

        if fast < 1 {
            return Err(format!("{}: fast should be greater than 0", invalid));
        }
        if slow < 1 {
            return Err(format!("{}: slow should be greater than 0", invalid));
        }
        if tclen < 1 {
            return Err(format!("{}: tclen should be greater than 0", invalid));
        }
        if factor <= 0.0 || factor > 1.0 {
            return Err(format!("{}: factor should be in (0, 1]", invalid));
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!(
            "stc({},{},{},{:.2}{})",
            fast,
            slow,
            tclen,
            factor,
            component_triple_mnemonic(bc, qc, tc)
        );

        Ok(Self {
            ema_fast: Ema::new(fast),
            ema_slow: Ema::new(slow),
            slow,
            factor,
            bar: -1,
            macd_win: Window::new(tclen),
            pf_win: Window::new(tclen),
            frac1: 0.0,
            frac2: 0.0,
            pf: 0.0,
            pff: 0.0,
            primed: false,
            bar_func,
            quote_func,
            trade_func,
            mnemonic,
        })
    }

    /// Returns true if the indicator has produced at least one valid output.
    pub fn is_primed(&self) -> bool {
        self.primed
    }

    /// Core update returning (stc, macd, pf).
    pub fn update(&mut self, sample: f64) -> (f64, f64, f64) {
        self.bar += 1;
        let k = self.bar;

        // Price EMAs always advance (they accumulate over the full history).
        let ema_fast = self.ema_fast.update(sample);
        let ema_slow = self.ema_slow.update(sample);

        // GATE: XMAC is only assigned while barindex > slow.
        let gate_open = k > self.slow as i64;
        let macd = if gate_open { ema_fast - ema_slow } else { 0.0 };
        self.macd_win.push(macd);

        if !gate_open {
            self.pf_win.push(self.pf);
            return (f64::NAN, macd, self.pf);
        }

        // 1st stochastic of the MACD over tclen (guard on the range).
        let (ll1, hh1) = self.macd_win.min_max();
        let rng1 = hh1 - ll1;
        if rng1 > 0.0 {
            self.frac1 = ((macd - ll1) / rng1) * 100.0;
        }

        // 1st smoothing: PF = EMA(Frac1, alpha=factor), seed 0.
        self.pf += self.factor * (self.frac1 - self.pf);
        self.pf_win.push(self.pf);

        // 2nd stochastic of PF over tclen.
        let (ll2, hh2) = self.pf_win.min_max();
        let rng2 = hh2 - ll2;
        if rng2 > 0.0 {
            self.frac2 = ((self.pf - ll2) / rng2) * 100.0;
        }

        // 2nd smoothing: STC = PFF = EMA(Frac2, alpha=factor), seed 0.
        self.pff += self.factor * (self.frac2 - self.pff);
        self.primed = true;

        (self.pff, macd, self.pf)
    }
}

impl Indicator for SchaffTrendCycle {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        let desc = format!("Schaff Trend Cycle {}", self.mnemonic);
        build_metadata(
            Identifier::SchaffTrendCycle,
            &self.mnemonic,
            &desc,
            &[
                OutputText {
                    mnemonic: format!("{} stc", self.mnemonic),
                    description: format!("{} STC", desc),
                },
                OutputText {
                    mnemonic: format!("{} macd", self.mnemonic),
                    description: format!("{} MACD", desc),
                },
                OutputText {
                    mnemonic: format!("{} pf", self.mnemonic),
                    description: format!("{} PF", desc),
                },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let (stc, macd, pf) = self.update(sample.value);
        vec![
            Box::new(Scalar { time: sample.time, value: stc }),
            Box::new(Scalar { time: sample.time, value: macd }),
            Box::new(Scalar { time: sample.time, value: pf }),
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

    const TOLERANCE: f64 = 1e-9;

    fn check(name: &str, i: usize, exp: f64, act: f64) {
        if exp.is_nan() {
            assert!(act.is_nan(), "{}[{}]: expected NaN, got {}", name, i, act);
            return;
        }
        assert!(
            (act - exp).abs() <= TOLERANCE,
            "{}[{}]: expected {}, got {}",
            name,
            i,
            exp,
            act
        );
    }

    struct Combo {
        fast: usize,
        slow: usize,
        tclen: usize,
        factor: f64,
        stc: Vec<f64>,
        macd: Option<Vec<f64>>,
        pf: Option<Vec<f64>>,
    }

    #[test]
    fn test_reference_data_all_combos() {
        let combos = vec![
            Combo { fast: 23, slow: 50, tclen: 10, factor: 0.5, stc: testdata::expected_stc_f23_s50_t10_c50(), macd: Some(testdata::expected_macd_f23_s50_t10_c50()), pf: Some(testdata::expected_pf_f23_s50_t10_c50()) },
            Combo { fast: 12, slow: 26, tclen: 10, factor: 0.5, stc: testdata::expected_stc_f12_s26_t10_c50(), macd: Some(testdata::expected_macd_f12_s26_t10_c50()), pf: Some(testdata::expected_pf_f12_s26_t10_c50()) },
            Combo { fast: 5, slow: 10, tclen: 5, factor: 0.5, stc: testdata::expected_stc_f5_s10_t5_c50(), macd: Some(testdata::expected_macd_f5_s10_t5_c50()), pf: Some(testdata::expected_pf_f5_s10_t5_c50()) },
            Combo { fast: 3, slow: 7, tclen: 3, factor: 0.5, stc: testdata::expected_stc_f3_s7_t3_c50(), macd: None, pf: None },
            Combo { fast: 8, slow: 21, tclen: 10, factor: 0.5, stc: testdata::expected_stc_f8_s21_t10_c50(), macd: None, pf: None },
            Combo { fast: 10, slow: 30, tclen: 10, factor: 0.5, stc: testdata::expected_stc_f10_s30_t10_c50(), macd: None, pf: None },
            Combo { fast: 15, slow: 40, tclen: 14, factor: 0.5, stc: testdata::expected_stc_f15_s40_t14_c50(), macd: None, pf: None },
            Combo { fast: 6, slow: 13, tclen: 8, factor: 0.6, stc: testdata::expected_stc_f6_s13_t8_c60(), macd: None, pf: None },
            Combo { fast: 23, slow: 50, tclen: 23, factor: 0.5, stc: testdata::expected_stc_f23_s50_t23_c50(), macd: None, pf: None },
            Combo { fast: 23, slow: 50, tclen: 5, factor: 0.5, stc: testdata::expected_stc_f23_s50_t5_c50(), macd: None, pf: None },
            Combo { fast: 12, slow: 26, tclen: 10, factor: 0.25, stc: testdata::expected_stc_f12_s26_t10_c25(), macd: None, pf: None },
            Combo { fast: 12, slow: 26, tclen: 10, factor: 0.8, stc: testdata::expected_stc_f12_s26_t10_c80(), macd: None, pf: None },
            Combo { fast: 12, slow: 26, tclen: 10, factor: 1.0, stc: testdata::expected_stc_f12_s26_t10_c100(), macd: None, pf: None },
            Combo { fast: 20, slow: 40, tclen: 10, factor: 0.5, stc: testdata::expected_stc_f20_s40_t10_c50(), macd: None, pf: None },
        ];

        let input = testdata::test_input();

        for combo in &combos {
            let mut ind = SchaffTrendCycle::new(&SchaffTrendCycleParams {
                fast: combo.fast,
                slow: combo.slow,
                tclen: combo.tclen,
                factor: combo.factor,
                ..Default::default()
            })
            .unwrap();

            for i in 0..input.len() {
                let (stc, macd, pf) = ind.update(input[i]);
                check("stc", i, combo.stc[i], stc);
                if let Some(ref m) = combo.macd {
                    check("macd", i, m[i], macd);
                }
                if let Some(ref p) = combo.pf {
                    check("pf", i, p[i], pf);
                }
            }
        }
    }

    #[test]
    fn test_mnemonic() {
        let ind = SchaffTrendCycle::new(&SchaffTrendCycleParams::default()).unwrap();
        assert_eq!(ind.metadata().mnemonic, "stc(23,50,10,0.50)");

        let ind2 = SchaffTrendCycle::new(&SchaffTrendCycleParams {
            fast: 12,
            slow: 26,
            tclen: 10,
            factor: 0.25,
            ..Default::default()
        })
        .unwrap();
        assert_eq!(ind2.metadata().mnemonic, "stc(12,26,10,0.25)");
    }

    #[test]
    fn test_metadata() {
        let ind = SchaffTrendCycle::new(&SchaffTrendCycleParams::default()).unwrap();
        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::SchaffTrendCycle);
        assert_eq!(meta.outputs.len(), 3);
        assert_eq!(meta.outputs[0].kind, SchaffTrendCycleOutput::Stc as i32);
        assert_eq!(meta.outputs[1].kind, SchaffTrendCycleOutput::Macd as i32);
        assert_eq!(meta.outputs[2].kind, SchaffTrendCycleOutput::Pf as i32);
    }

    #[test]
    fn test_update_scalar_ordering() {
        let input = testdata::test_input();
        let exp_stc = testdata::expected_stc_f23_s50_t10_c50();
        let exp_macd = testdata::expected_macd_f23_s50_t10_c50();
        let exp_pf = testdata::expected_pf_f23_s50_t10_c50();

        let mut ind = SchaffTrendCycle::new(&SchaffTrendCycleParams::default()).unwrap();
        let mut out: Output = vec![];
        for i in 0..input.len() {
            out = ind.update_scalar(&Scalar { time: 0, value: input[i] });
        }
        let last = input.len() - 1;
        let stc = out[0].downcast_ref::<Scalar>().unwrap().value;
        let macd = out[1].downcast_ref::<Scalar>().unwrap().value;
        let pf = out[2].downcast_ref::<Scalar>().unwrap().value;
        check("stc", last, exp_stc[last], stc);
        check("macd", last, exp_macd[last], macd);
        check("pf", last, exp_pf[last], pf);
    }

    #[test]
    fn test_invalid_params() {
        assert!(SchaffTrendCycle::new(&SchaffTrendCycleParams { factor: 1.5, ..Default::default() }).is_err());
        assert!(SchaffTrendCycle::new(&SchaffTrendCycleParams { factor: -0.5, ..Default::default() }).is_err());
        // Note: fast/slow/tclen = 0 are treated as defaults (matching other languages),
        // so they are not invalid; factor out of (0, 1] is the testable invalid case here.
    }

    #[test]
    fn test_nan_warmup() {
        let mut ind = SchaffTrendCycle::new(&SchaffTrendCycleParams::default()).unwrap();
        let (stc, _macd, _pf) = ind.update(f64::NAN);
        assert!(stc.is_nan());
    }
}
