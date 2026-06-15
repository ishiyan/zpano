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
use crate::indicators::core::outputs::levels::{Level, Levels};
use crate::indicators::core::outputs::polyline::{Point, Polyline};

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the Moving Mini-Max indicator.
pub struct MovingMiniMaxParams {
    /// Smoothing window width controlling the quantum tunnelling ability. >= 1. Default 5.
    pub m: usize,
    /// Lookback window size: number of bars analysed. > 2*m. Default 50.
    pub n: usize,
    /// Number of distinct support/resistance levels to detect. >= 1. Default 3.
    pub num_extrema: usize,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for MovingMiniMaxParams {
    fn default() -> Self {
        Self {
            m: 5,
            n: 50,
            num_extrema: 3,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Moving Mini-Max indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum MovingMiniMaxOutput {
    /// The up mini-max value at the most recent bar (emphasizes local maxima).
    Up = 1,
    /// The down mini-max value at the most recent bar (emphasizes local minima).
    Down = 2,
    /// The detected resistance levels, sorted by strength (strongest first).
    Resistances = 3,
    /// The detected support levels, sorted by strength (strongest first).
    Supports = 4,
    /// The full up mini-max probability distribution over the window.
    UpDistribution = 5,
    /// The full down mini-max probability distribution over the window.
    DownDistribution = 6,
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/// A detected peak as a (strength, index) pair.
#[derive(Clone, Copy)]
struct Peak {
    strength: f64,
    index: usize,
}

/// A detected support/resistance level.
#[derive(Clone, Copy)]
struct MiniMaxLevel {
    price: f64,
    offset: i32,
    strength: f64,
}

/// Computes Q_{i,i+1} and Q_{i,i-1} for each position i = 0..n-1.
fn calc_q_values(window: &[f64], n: usize, m: usize, negate: bool) -> (Vec<f64>, Vec<f64>) {
    let sign = if negate { -1.0 } else { 1.0 };
    let mut q_plus = vec![0.0; n];
    let mut q_minus = vec![0.0; n];

    for i in 0..n {
        let s_i = window[i];
        let mut sum_plus = 0.0;
        let mut sum_minus = 0.0;

        for k in 1..=m {
            let s_forward = if i + k < n { window[i + k] } else { window[n - 1] };
            let s_backward = if i >= k { window[i - k] } else { window[0] };

            let denom_plus = s_forward + s_i;
            let arg_plus = if denom_plus == 0.0 { 0.0 } else { sign * 2.0 * (s_forward - s_i) / denom_plus };

            let denom_minus = s_backward + s_i;
            let arg_minus = if denom_minus == 0.0 { 0.0 } else { sign * 2.0 * (s_backward - s_i) / denom_minus };

            sum_plus += arg_plus.exp();
            sum_minus += arg_minus.exp();
        }

        q_plus[i] = sum_plus;
        q_minus[i] = sum_minus;
    }

    (q_plus, q_minus)
}

/// Computes transition probabilities P_{i,i+1} and P_{i,i-1} from Q-values.
fn calc_p_values(q_plus: &[f64], q_minus: &[f64], n: usize) -> (Vec<f64>, Vec<f64>) {
    let mut p_plus = vec![0.0; n];
    let mut p_minus = vec![0.0; n];

    for i in 0..n {
        let denom = q_plus[i] + q_minus[i];
        if denom == 0.0 {
            p_plus[i] = 0.5;
            p_minus[i] = 0.5;
        } else {
            p_plus[i] = q_plus[i] / denom;
            p_minus[i] = q_minus[i] / denom;
        }
    }

    (p_plus, p_minus)
}

/// Computes the normalized mini-max series from transition probabilities.
fn calc_minimax(p_plus: &[f64], p_minus: &[f64], n: usize) -> Vec<f64> {
    let mut u = vec![0.0; n];
    u[0] = 1.0;

    for i in 1..n {
        let p_prev_to_i = p_plus[i - 1];
        let p_i_to_prev = p_minus[i];
        if p_i_to_prev == 0.0 {
            u[i] = u[i - 1] * 1e10;
        } else {
            u[i] = (p_prev_to_i / p_i_to_prev) * u[i - 1];
        }
    }

    let total: f64 = u.iter().sum();
    if total == 0.0 {
        return vec![1.0 / n as f64; n];
    }
    u.iter().map(|&value| value / total).collect()
}

/// Finds distinct local peaks, returned sorted by strength descending.
fn find_peaks(values: &[f64], num_peaks: usize, min_separation: usize) -> Vec<Peak> {
    let n = values.len();
    let mut candidates: Vec<Peak> = Vec::new();

    for i in 0..n {
        let is_peak = if i == 0 {
            n <= 1 || values[i] >= values[i + 1]
        } else if i == n - 1 {
            values[i] >= values[i - 1]
        } else {
            values[i] >= values[i - 1] && values[i] >= values[i + 1]
        };
        if is_peak {
            candidates.push(Peak { strength: values[i], index: i });
        }
    }

    // Sort by strength descending; ties break on the larger index first (matches the
    // reference, which sorts (value, index) tuples in reverse).
    candidates.sort_by(|a, b| {
        b.strength
            .partial_cmp(&a.strength)
            .unwrap_or(std::cmp::Ordering::Equal)
            .then(b.index.cmp(&a.index))
    });

    let mut selected: Vec<Peak> = Vec::new();
    for c in candidates {
        if selected.len() >= num_peaks {
            break;
        }
        let too_close = selected.iter().any(|sel| {
            let diff = if c.index > sel.index { c.index - sel.index } else { sel.index - c.index };
            diff < min_separation
        });
        if !too_close {
            selected.push(c);
        }
    }

    selected
}

/// A computed MMM result set.
#[derive(Default)]
struct MiniMaxResult {
    up: f64,
    down: f64,
    resistances: Vec<MiniMaxLevel>,
    supports: Vec<MiniMaxLevel>,
    up_dist: Vec<f64>,
    down_dist: Vec<f64>,
    valid: bool,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Zurab Silagadze's Moving Mini-Max (MMM).
///
/// A nonlinear indicator for technical analysis that emphasizes local maximums and minimums
/// in a price series with inherent smoothing. The algorithm is borrowed from gamma-ray
/// spectroscopy peak finding and models price exploration as a quantum particle that can
/// tunnel through small noise barriers but is stopped by genuine trend reversals.
///
/// Reference: Silagadze, Z. K. (2011). Moving Mini-Max -- a new indicator for technical
/// analysis. IFTA Journal 11, 46-49. arXiv:0802.0984v2.
pub struct MovingMiniMax {
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
    mnemonic: String,
    description: String,

    m: usize,
    n: usize,
    num_extrema: usize,

    window: Vec<f64>,
    buf_pos: usize,
    count: usize,
    primed: bool,

    last: MiniMaxResult,
}

impl MovingMiniMax {
    /// Creates a new Moving Mini-Max from the given parameters.
    pub fn new(params: &MovingMiniMaxParams) -> Result<Self, String> {
        let invalid = "invalid moving mini-max parameters";

        let m = params.m;
        let n = params.n;
        let num_extrema = params.num_extrema;

        if m < 1 {
            return Err(format!("{}: m should be >= 1", invalid));
        }
        if n <= 2 * m {
            return Err(format!("{}: n should be > 2*m", invalid));
        }
        if num_extrema < 1 {
            return Err(format!("{}: num extrema should be >= 1", invalid));
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let mnemonic = format!(
            "mmm({},{},{}{})",
            m, n, num_extrema, component_triple_mnemonic(bc, qc, tc)
        );
        let description = format!("Moving mini-max {}", mnemonic);

        Ok(Self {
            bar_func: bar_component_value(bc),
            quote_func: quote_component_value(qc),
            trade_func: trade_component_value(tc),
            mnemonic,
            description,
            m,
            n,
            num_extrema,
            window: vec![0.0; n],
            buf_pos: 0,
            count: 0,
            primed: false,
            last: MiniMaxResult::default(),
        })
    }

    /// Returns true if the indicator has received enough data to be primed.
    pub fn is_primed(&self) -> bool {
        self.primed
    }

    /// Computes the MMM set for the given price; stores it in `self.last`.
    pub fn update_values(&mut self, sample: f64) {
        self.last = MiniMaxResult::default();

        if self.count < self.n {
            self.window[self.count] = sample;
            self.count += 1;
        } else {
            self.window[self.buf_pos] = sample;
            self.buf_pos = (self.buf_pos + 1) % self.n;
        }

        if self.count < self.n {
            self.primed = false;
            return;
        }

        self.primed = true;

        let n = self.n;
        let m = self.m;

        // Reconstruct the window in chronological order (oldest -> newest).
        let window: Vec<f64> = (0..n).map(|i| self.window[(self.buf_pos + i) % n]).collect();

        let (q_up_plus, q_up_minus) = calc_q_values(&window, n, m, false);
        let (q_dn_plus, q_dn_minus) = calc_q_values(&window, n, m, true);

        let (p_up_plus, p_up_minus) = calc_p_values(&q_up_plus, &q_up_minus, n);
        let (p_dn_plus, p_dn_minus) = calc_p_values(&q_dn_plus, &q_dn_minus, n);

        let up_dist = calc_minimax(&p_up_plus, &p_up_minus, n);
        let dn_dist = calc_minimax(&p_dn_plus, &p_dn_minus, n);

        let min_sep = m.max(2);

        let u_peaks = find_peaks(&up_dist, self.num_extrema, min_sep);
        let d_peaks = find_peaks(&dn_dist, self.num_extrema, min_sep);

        let resistances: Vec<MiniMaxLevel> = u_peaks
            .iter()
            .map(|pk| MiniMaxLevel { price: window[pk.index], offset: ((n - 1) - pk.index) as i32, strength: pk.strength })
            .collect();
        let supports: Vec<MiniMaxLevel> = d_peaks
            .iter()
            .map(|pk| MiniMaxLevel { price: window[pk.index], offset: ((n - 1) - pk.index) as i32, strength: pk.strength })
            .collect();

        self.last = MiniMaxResult {
            up: up_dist[n - 1],
            down: dn_dist[n - 1],
            resistances,
            supports,
            up_dist,
            down_dist: dn_dist,
            valid: true,
        };
    }

    fn levels_of(time: i64, levels: &[MiniMaxLevel]) -> Levels {
        if levels.is_empty() {
            return Levels::empty(time);
        }
        Levels::new(time, levels.iter().map(|lv| Level::new(lv.price, lv.offset, lv.strength)).collect())
    }

    fn polyline_of(time: i64, values: &[f64]) -> Polyline {
        if values.is_empty() {
            return Polyline::empty(time);
        }
        Polyline::new(
            time,
            values.iter().enumerate().map(|(i, &v)| Point { offset: i as i32, value: v }).collect(),
        )
    }

    fn wrap(&self, time: i64) -> Output {
        let (up, down) = if self.last.valid {
            (self.last.up, self.last.down)
        } else {
            (f64::NAN, f64::NAN)
        };
        vec![
            Box::new(Scalar { time, value: up }),
            Box::new(Scalar { time, value: down }),
            Box::new(Self::levels_of(time, &self.last.resistances)),
            Box::new(Self::levels_of(time, &self.last.supports)),
            Box::new(Self::polyline_of(time, &self.last.up_dist)),
            Box::new(Self::polyline_of(time, &self.last.down_dist)),
        ]
    }
}

impl Indicator for MovingMiniMax {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        let d = &self.description;
        let m = &self.mnemonic;
        build_metadata(
            Identifier::MovingMiniMax,
            m,
            d,
            &[
                OutputText { mnemonic: format!("{} up", m), description: format!("{} up value", d) },
                OutputText { mnemonic: format!("{} down", m), description: format!("{} down value", d) },
                OutputText { mnemonic: format!("{} resistances", m), description: format!("{} resistances", d) },
                OutputText { mnemonic: format!("{} supports", m), description: format!("{} supports", d) },
                OutputText { mnemonic: format!("{} up dist", m), description: format!("{} up distribution", d) },
                OutputText { mnemonic: format!("{} down dist", m), description: format!("{} down distribution", d) },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        self.update_values(sample.value);
        self.wrap(sample.time)
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let v = (self.bar_func)(sample);
        self.update_values(v);
        self.wrap(sample.time)
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (self.quote_func)(sample);
        self.update_values(v);
        self.wrap(sample.time)
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let v = (self.trade_func)(sample);
        self.update_values(v);
        self.wrap(sample.time)
    }
}

// ===========================================================================
// Tests
// ===========================================================================

#[cfg(test)]
mod tests {
    use super::*;

    include!("testdata.rs");

    const TOLERANCE: f64 = 1e-9;

    fn run_last(inputs: &[f64], m: usize, n: usize, num_extrema: usize) -> MovingMiniMax {
        let mut ind = MovingMiniMax::new(&MovingMiniMaxParams {
            m, n, num_extrema, ..Default::default()
        })
        .unwrap();
        for &p in inputs {
            ind.update_values(p);
        }
        ind
    }

    fn check_series(name: &str, actual: &[f64], expected: &[f64]) {
        assert_eq!(actual.len(), expected.len(), "{}: length mismatch", name);
        for i in 0..expected.len() {
            let delta = TOLERANCE * expected[i].abs().max(1.0);
            assert!(
                (actual[i] - expected[i]).abs() <= delta,
                "{}[{}]: expected {}, got {}",
                name, i, expected[i], actual[i]
            );
        }
    }

    fn check_levels(name: &str, actual: &[MiniMaxLevel], expected: &[testdata::Extremum]) {
        assert_eq!(actual.len(), expected.len(), "{}: length mismatch", name);
        for i in 0..expected.len() {
            assert!(
                (actual[i].price - expected[i].price).abs() <= TOLERANCE * expected[i].price.abs().max(1.0),
                "{}[{}].price", name, i
            );
            assert_eq!(actual[i].offset as usize, expected[i].offset, "{}[{}].offset", name, i);
            assert!(
                (actual[i].strength - expected[i].strength).abs() <= TOLERANCE * expected[i].strength.abs().max(1.0),
                "{}[{}].strength", name, i
            );
        }
    }

    #[allow(clippy::too_many_arguments)]
    fn check_combo(
        name: &str, m: usize, n: usize, e: usize,
        exp_up: &[f64], exp_down: &[f64], exp_res: &[testdata::Extremum], exp_sup: &[testdata::Extremum],
    ) {
        let ind = run_last(&testdata::test_input(), m, n, e);
        assert!(ind.last.valid, "{}: no valid output", name);
        check_series(&format!("{} UP", name), &ind.last.up_dist, exp_up);
        check_series(&format!("{} DOWN", name), &ind.last.down_dist, exp_down);
        check_levels(&format!("{} RES", name), &ind.last.resistances, exp_res);
        check_levels(&format!("{} SUP", name), &ind.last.supports, exp_sup);
    }

    #[test]
    fn test_m3_combos() {
        check_combo("m3_n50_e1", 3, 50, 1, &testdata::expected_m3_n50_e1_up(), &testdata::expected_m3_n50_e1_down(), &testdata::expected_m3_n50_e1_resistances(), &testdata::expected_m3_n50_e1_supports());
        check_combo("m3_n50_e3", 3, 50, 3, &testdata::expected_m3_n50_e3_up(), &testdata::expected_m3_n50_e3_down(), &testdata::expected_m3_n50_e3_resistances(), &testdata::expected_m3_n50_e3_supports());
        check_combo("m3_n100_e1", 3, 100, 1, &testdata::expected_m3_n100_e1_up(), &testdata::expected_m3_n100_e1_down(), &testdata::expected_m3_n100_e1_resistances(), &testdata::expected_m3_n100_e1_supports());
        check_combo("m3_n100_e3", 3, 100, 3, &testdata::expected_m3_n100_e3_up(), &testdata::expected_m3_n100_e3_down(), &testdata::expected_m3_n100_e3_resistances(), &testdata::expected_m3_n100_e3_supports());
        check_combo("m3_n252_e1", 3, 252, 1, &testdata::expected_m3_n252_e1_up(), &testdata::expected_m3_n252_e1_down(), &testdata::expected_m3_n252_e1_resistances(), &testdata::expected_m3_n252_e1_supports());
        check_combo("m3_n252_e3", 3, 252, 3, &testdata::expected_m3_n252_e3_up(), &testdata::expected_m3_n252_e3_down(), &testdata::expected_m3_n252_e3_resistances(), &testdata::expected_m3_n252_e3_supports());
    }

    #[test]
    fn test_m5_combos() {
        check_combo("m5_n50_e1", 5, 50, 1, &testdata::expected_m5_n50_e1_up(), &testdata::expected_m5_n50_e1_down(), &testdata::expected_m5_n50_e1_resistances(), &testdata::expected_m5_n50_e1_supports());
        check_combo("m5_n50_e3", 5, 50, 3, &testdata::expected_m5_n50_e3_up(), &testdata::expected_m5_n50_e3_down(), &testdata::expected_m5_n50_e3_resistances(), &testdata::expected_m5_n50_e3_supports());
        check_combo("m5_n100_e1", 5, 100, 1, &testdata::expected_m5_n100_e1_up(), &testdata::expected_m5_n100_e1_down(), &testdata::expected_m5_n100_e1_resistances(), &testdata::expected_m5_n100_e1_supports());
        check_combo("m5_n100_e3", 5, 100, 3, &testdata::expected_m5_n100_e3_up(), &testdata::expected_m5_n100_e3_down(), &testdata::expected_m5_n100_e3_resistances(), &testdata::expected_m5_n100_e3_supports());
        check_combo("m5_n252_e1", 5, 252, 1, &testdata::expected_m5_n252_e1_up(), &testdata::expected_m5_n252_e1_down(), &testdata::expected_m5_n252_e1_resistances(), &testdata::expected_m5_n252_e1_supports());
        check_combo("m5_n252_e3", 5, 252, 3, &testdata::expected_m5_n252_e3_up(), &testdata::expected_m5_n252_e3_down(), &testdata::expected_m5_n252_e3_resistances(), &testdata::expected_m5_n252_e3_supports());
    }

    #[test]
    fn test_m10_combos() {
        check_combo("m10_n50_e1", 10, 50, 1, &testdata::expected_m10_n50_e1_up(), &testdata::expected_m10_n50_e1_down(), &testdata::expected_m10_n50_e1_resistances(), &testdata::expected_m10_n50_e1_supports());
        check_combo("m10_n50_e3", 10, 50, 3, &testdata::expected_m10_n50_e3_up(), &testdata::expected_m10_n50_e3_down(), &testdata::expected_m10_n50_e3_resistances(), &testdata::expected_m10_n50_e3_supports());
        check_combo("m10_n100_e1", 10, 100, 1, &testdata::expected_m10_n100_e1_up(), &testdata::expected_m10_n100_e1_down(), &testdata::expected_m10_n100_e1_resistances(), &testdata::expected_m10_n100_e1_supports());
        check_combo("m10_n100_e3", 10, 100, 3, &testdata::expected_m10_n100_e3_up(), &testdata::expected_m10_n100_e3_down(), &testdata::expected_m10_n100_e3_resistances(), &testdata::expected_m10_n100_e3_supports());
        check_combo("m10_n252_e1", 10, 252, 1, &testdata::expected_m10_n252_e1_up(), &testdata::expected_m10_n252_e1_down(), &testdata::expected_m10_n252_e1_resistances(), &testdata::expected_m10_n252_e1_supports());
        check_combo("m10_n252_e3", 10, 252, 3, &testdata::expected_m10_n252_e3_up(), &testdata::expected_m10_n252_e3_down(), &testdata::expected_m10_n252_e3_resistances(), &testdata::expected_m10_n252_e3_supports());
    }

    #[test]
    fn test_m20_combos() {
        check_combo("m20_n50_e1", 20, 50, 1, &testdata::expected_m20_n50_e1_up(), &testdata::expected_m20_n50_e1_down(), &testdata::expected_m20_n50_e1_resistances(), &testdata::expected_m20_n50_e1_supports());
        check_combo("m20_n50_e3", 20, 50, 3, &testdata::expected_m20_n50_e3_up(), &testdata::expected_m20_n50_e3_down(), &testdata::expected_m20_n50_e3_resistances(), &testdata::expected_m20_n50_e3_supports());
        check_combo("m20_n100_e1", 20, 100, 1, &testdata::expected_m20_n100_e1_up(), &testdata::expected_m20_n100_e1_down(), &testdata::expected_m20_n100_e1_resistances(), &testdata::expected_m20_n100_e1_supports());
        check_combo("m20_n100_e3", 20, 100, 3, &testdata::expected_m20_n100_e3_up(), &testdata::expected_m20_n100_e3_down(), &testdata::expected_m20_n100_e3_resistances(), &testdata::expected_m20_n100_e3_supports());
        check_combo("m20_n252_e1", 20, 252, 1, &testdata::expected_m20_n252_e1_up(), &testdata::expected_m20_n252_e1_down(), &testdata::expected_m20_n252_e1_resistances(), &testdata::expected_m20_n252_e1_supports());
        check_combo("m20_n252_e3", 20, 252, 3, &testdata::expected_m20_n252_e3_up(), &testdata::expected_m20_n252_e3_down(), &testdata::expected_m20_n252_e3_resistances(), &testdata::expected_m20_n252_e3_supports());
    }

    #[test]
    fn test_latest_scalars_equal_distribution_tails() {
        let ind = run_last(&testdata::test_input(), 3, 50, 1);
        assert!((ind.last.up - ind.last.up_dist[ind.last.up_dist.len() - 1]).abs() <= 1e-12);
        assert!((ind.last.down - ind.last.down_dist[ind.last.down_dist.len() - 1]).abs() <= 1e-12);
    }

    #[test]
    fn test_mnemonic_and_metadata() {
        let ind = MovingMiniMax::new(&MovingMiniMaxParams::default()).unwrap();
        assert_eq!(ind.metadata().mnemonic, "mmm(5,50,3)");
        assert_eq!(ind.metadata().outputs.len(), 6);
        assert_eq!(ind.metadata().identifier, Identifier::MovingMiniMax);
    }

    #[test]
    fn test_update_scalar_outputs() {
        let mut ind = MovingMiniMax::new(&MovingMiniMaxParams { m: 5, n: 50, num_extrema: 3, ..Default::default() }).unwrap();
        let mut out: Output = vec![];
        for &p in &testdata::test_input() {
            out = ind.update_scalar(&Scalar { time: 0, value: p });
        }
        assert_eq!(out.len(), 6);
        let poly = out[4].downcast_ref::<Polyline>().unwrap();
        assert_eq!(poly.points.len(), 50);
    }

    #[test]
    fn test_invalid_params() {
        assert!(MovingMiniMax::new(&MovingMiniMaxParams { m: 0, ..Default::default() }).is_err());
        assert!(MovingMiniMax::new(&MovingMiniMaxParams { m: 5, n: 10, ..Default::default() }).is_err());
        assert!(MovingMiniMax::new(&MovingMiniMaxParams { num_extrema: 0, ..Default::default() }).is_err());
    }
}
