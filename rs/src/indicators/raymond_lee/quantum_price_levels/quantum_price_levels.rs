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

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the Quantum Price Levels indicator.
pub struct QuantumPriceLevelsParams {
    /// Number of price-return ratios in the sliding window. >= 2. Default 2048.
    pub lookback: usize,
    /// Number of quantum energy levels (n = 0..num_levels-1). >= 1. Default 21.
    pub num_levels: usize,
    /// Number of histogram bins for the wavefunction distribution. >= 2. Default 100.
    pub num_bins: usize,
    /// Empirical scaling constant in the NQPR formula. > 0. Default 0.21.
    pub scale_factor: f64,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for QuantumPriceLevelsParams {
    fn default() -> Self {
        Self {
            lookback: 2048,
            num_levels: 21,
            num_bins: 100,
            scale_factor: 0.21,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Quantum Price Levels indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum QuantumPriceLevelsOutput {
    /// The anharmonic coefficient (lambda) of the quantum potential well.
    Lambda = 1,
    /// The population standard deviation of the price-return ratios in the window.
    ReturnStdDev = 2,
    /// The normalized QPR multipliers (1 + scale_factor*sigma*QPR(n)), one per level.
    NormalizedMultipliers = 3,
    /// The resistance price levels above the current price (price * NQPR(n)).
    Resistances = 4,
    /// The support price levels below the current price (price / NQPR(n)).
    Supports = 5,
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/// Signed real cube root via powf (matches the reference implementation).
fn cbrt(x: f64) -> f64 {
    if x >= 0.0 {
        x.powf(1.0 / 3.0)
    } else {
        -(-x).powf(1.0 / 3.0)
    }
}

/// K0 constant for energy level n (Dasgupta et al. 2007).
fn compute_k0(n: usize) -> f64 {
    let fln = n as f64;
    let numerator = 1.1924 + 33.2383 * fln + 56.2169 * fln * fln;
    let denominator = 1.0 + 43.6106 * fln;
    (numerator / denominator).powf(1.0 / 3.0)
}

/// A computed QPL result set.
#[derive(Default)]
struct QplResult {
    lambda: f64,
    sigma: f64,
    nqpr: Vec<f64>,
    resistances: Vec<f64>,
    supports: Vec<f64>,
    valid: bool,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Raymond Lee's Quantum Price Levels (QPL).
///
/// Computes discrete support/resistance price levels from a quantum-finance analogy:
/// the market is modelled as a quantum anharmonic oscillator, and the discrete energy
/// eigenvalues of the system map to price levels above and below the current price.
///
/// Reference: Lee, R. S. T. (2021). Quantum Finance Forecast System with Quantum
/// Anharmonic Oscillator Model for Quantum Price Level Modeling. IAJER, 4(02), 1-21.
pub struct QuantumPriceLevels {
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
    mnemonic: String,
    description: String,

    lookback: usize,
    num_levels: usize,
    num_bins: usize,
    scale_factor: f64,
    k: Vec<f64>,

    returns: Vec<f64>,
    buf_pos: usize,
    count: usize,
    prev_price: f64,
    have_prev: bool,
    primed: bool,

    last: QplResult,
}

impl QuantumPriceLevels {
    /// Creates a new Quantum Price Levels from the given parameters.
    pub fn new(params: &QuantumPriceLevelsParams) -> Result<Self, String> {
        let invalid = "invalid quantum price levels parameters";

        let lookback = params.lookback;
        let num_levels = params.num_levels;
        let num_bins = params.num_bins;
        let scale_factor = params.scale_factor;

        if lookback < 2 {
            return Err(format!("{}: lookback should be >= 2", invalid));
        }
        if num_levels < 1 {
            return Err(format!("{}: num levels should be >= 1", invalid));
        }
        if num_bins < 2 {
            return Err(format!("{}: num bins should be >= 2", invalid));
        }
        if scale_factor <= 0.0 {
            return Err(format!("{}: scale factor should be > 0", invalid));
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let mnemonic = format!(
            "qpl({},{},{},{}{})",
            lookback, num_levels, num_bins, format_scale(scale_factor), component_triple_mnemonic(bc, qc, tc)
        );
        let description = format!("Quantum price levels {}", mnemonic);

        let k: Vec<f64> = (0..num_levels).map(compute_k0).collect();

        Ok(Self {
            bar_func: bar_component_value(bc),
            quote_func: quote_component_value(qc),
            trade_func: trade_component_value(tc),
            mnemonic,
            description,
            lookback,
            num_levels,
            num_bins,
            scale_factor,
            k,
            returns: vec![0.0; lookback],
            buf_pos: 0,
            count: 0,
            prev_price: 0.0,
            have_prev: false,
            primed: false,
            last: QplResult::default(),
        })
    }

    /// Returns true if the indicator has received enough data to be primed.
    pub fn is_primed(&self) -> bool {
        self.primed
    }

    /// Computes the QPL set for the given price; stores it in `self.last`.
    pub fn update_values(&mut self, sample: f64) {
        self.last = QplResult::default();

        if !self.have_prev {
            self.prev_price = sample;
            self.have_prev = true;
            self.primed = false;
            return;
        }

        let new_return = if sample > 0.0 { self.prev_price / sample } else { 1.0 };
        self.prev_price = sample;

        if self.count < self.lookback {
            self.returns[self.count] = new_return;
            self.count += 1;
        } else {
            self.returns[self.buf_pos] = new_return;
            self.buf_pos = (self.buf_pos + 1) % self.lookback;
        }

        if self.count < self.lookback {
            self.primed = false;
            return;
        }

        self.primed = true;

        let lookback = self.lookback;
        let num_bins = self.num_bins;
        let num_levels = self.num_levels;
        let scale_factor = self.scale_factor;

        // Statistics (population mu, sigma).
        let mut sum_r = 0.0;
        for i in 0..lookback {
            sum_r += self.returns[i];
        }
        let mu = sum_r / lookback as f64;

        let mut sum_var = 0.0;
        for i in 0..lookback {
            let diff = self.returns[i] - mu;
            sum_var += diff * diff;
        }
        let sigma = (sum_var / lookback as f64).sqrt();
        if sigma == 0.0 {
            return;
        }

        // Histogram centred at r = 1.
        let half_bins = num_bins / 2;
        let dr = 3.0 * sigma / half_bins as f64;
        let left_boundary = 1.0 - half_bins as f64 * dr;

        let mut q = vec![0usize; num_bins];
        let mut total_count = 0usize;
        for i in 0..lookback {
            let r = self.returns[i];
            let idx_f = (r - left_boundary) / dr;
            if idx_f >= 0.0 {
                let bin_index = idx_f as usize;
                if bin_index < num_bins {
                    q[bin_index] += 1;
                    total_count += 1;
                }
            }
        }

        if total_count == 0 {
            return;
        }
        let total_f = total_count as f64;

        // Ground state (peak bin).
        let mut max_q = 0.0;
        let mut max_qno = 0usize;
        for k in 0..num_bins {
            let nq = q[k] as f64 / total_f;
            if nq > max_q {
                max_q = nq;
                max_qno = k;
            }
        }

        if max_qno == 0 || max_qno == num_bins - 1 {
            return;
        }

        // lambda via FDM.
        let phi_plus1 = q[max_qno + 1] as f64 / total_f;
        let phi_minus1 = q[max_qno - 1] as f64 / total_f;

        let r_peak = left_boundary + max_qno as f64 * dr;
        let r0 = r_peak - dr / 2.0;
        let r_plus1 = r0 + dr;
        let r_minus1 = r0 - dr;

        let l_up = (r_minus1 * r_minus1) * phi_minus1 - (r_plus1 * r_plus1) * phi_plus1;
        let l_dw = (r_plus1 * r_plus1 * r_plus1 * r_plus1) * phi_plus1
            - (r_minus1 * r_minus1 * r_minus1 * r_minus1) * phi_minus1;

        if l_dw == 0.0 {
            return;
        }

        let lambda = (l_up / l_dw).abs();

        // Energy levels via Cardano.
        let mut qfel = vec![0.0; num_levels];
        for n in 0..num_levels {
            let two_n_plus_1 = (2 * n + 1) as f64;
            let p = -(two_n_plus_1 * two_n_plus_1);
            let q_coef = -lambda * (two_n_plus_1 * two_n_plus_1 * two_n_plus_1) * (self.k[n] * self.k[n] * self.k[n]);
            let discriminant = (q_coef * q_coef) / 4.0 + (p * p * p) / 27.0;
            if discriminant < 0.0 {
                return;
            }
            let sqrt_d = discriminant.sqrt();
            let u = cbrt(-q_coef / 2.0 + sqrt_d);
            let v = cbrt(-q_coef / 2.0 - sqrt_d);
            qfel[n] = u + v;
        }

        if qfel[0] == 0.0 {
            return;
        }

        // NQPR and projection from the current price.
        let mut nqpr = vec![0.0; num_levels];
        let mut resistances = vec![0.0; num_levels];
        let mut supports = vec![0.0; num_levels];
        for n in 0..num_levels {
            let qpr = qfel[n] / qfel[0];
            nqpr[n] = 1.0 + scale_factor * sigma * qpr;
            resistances[n] = sample * nqpr[n];
            supports[n] = sample / nqpr[n];
        }

        self.last = QplResult {
            lambda,
            sigma,
            nqpr,
            resistances,
            supports,
            valid: true,
        };
    }

    fn levels_of(time: i64, values: &[f64]) -> Levels {
        if values.is_empty() {
            return Levels::empty(time);
        }
        Levels::new(time, values.iter().map(|&v| Level::value_only(v)).collect())
    }

    fn wrap(&self, time: i64) -> Output {
        let (lambda, sigma) = if self.last.valid {
            (self.last.lambda, self.last.sigma)
        } else {
            (f64::NAN, f64::NAN)
        };
        vec![
            Box::new(Scalar { time, value: lambda }),
            Box::new(Scalar { time, value: sigma }),
            Box::new(Self::levels_of(time, &self.last.nqpr)),
            Box::new(Self::levels_of(time, &self.last.resistances)),
            Box::new(Self::levels_of(time, &self.last.supports)),
        ]
    }
}

/// Formats the scale factor compactly (e.g. 0.21, 0.1, 0.42).
fn format_scale(v: f64) -> String {
    let s = format!("{}", v);
    s
}

impl Indicator for QuantumPriceLevels {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        let d = &self.description;
        let m = &self.mnemonic;
        build_metadata(
            Identifier::QuantumPriceLevels,
            m,
            d,
            &[
                OutputText { mnemonic: format!("{} lambda", m), description: format!("{} anharmonic coefficient", d) },
                OutputText { mnemonic: format!("{} stddev", m), description: format!("{} return standard deviation", d) },
                OutputText { mnemonic: format!("{} nqpr", m), description: format!("{} normalized multipliers", d) },
                OutputText { mnemonic: format!("{} resistances", m), description: format!("{} resistance levels", d) },
                OutputText { mnemonic: format!("{} supports", m), description: format!("{} support levels", d) },
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

    fn run_last(inputs: &[f64], lookback: usize, num_levels: usize, num_bins: usize, scale_factor: f64) -> QuantumPriceLevels {
        let lb = if lookback == 0 { inputs.len() - 1 } else { lookback };
        let mut ind = QuantumPriceLevels::new(&QuantumPriceLevelsParams {
            lookback: lb, num_levels, num_bins, scale_factor, ..Default::default()
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

    fn check_combo(name: &str, inputs: &[f64], lookback: usize, num_levels: usize, num_bins: usize, scale_factor: f64,
                   exp_nqpr: &[f64], exp_up: &[f64], exp_lo: &[f64]) {
        let ind = run_last(inputs, lookback, num_levels, num_bins, scale_factor);
        assert!(ind.last.valid, "{}: no valid output", name);
        check_series(&format!("{} NQPR", name), &ind.last.nqpr, exp_nqpr);
        check_series(&format!("{} UPPER", name), &ind.last.resistances, exp_up);
        check_series(&format!("{} LOWER", name), &ind.last.supports, exp_lo);
    }

    #[test]
    fn test_batch_combos() {
        let input = testdata::test_input();
        check_combo("default", &input, 0, 21, 100, 0.21, &testdata::expected_nqpr(), &testdata::expected_upper(), &testdata::expected_lower());
        check_combo("F0_10", &input, 0, 21, 100, 0.10, &testdata::expected_nqpr_f0_10(), &testdata::expected_upper_f0_10(), &testdata::expected_lower_f0_10());
        check_combo("F0_42", &input, 0, 21, 100, 0.42, &testdata::expected_nqpr_f0_42(), &testdata::expected_upper_f0_42(), &testdata::expected_lower_f0_42());
        check_combo("B50", &input, 0, 21, 50, 0.21, &testdata::expected_nqpr_b50(), &testdata::expected_upper_b50(), &testdata::expected_lower_b50());
        check_combo("B50_F0_10", &input, 0, 21, 50, 0.10, &testdata::expected_nqpr_b50_f0_10(), &testdata::expected_upper_b50_f0_10(), &testdata::expected_lower_b50_f0_10());
        check_combo("B50_F0_42", &input, 0, 21, 50, 0.42, &testdata::expected_nqpr_b50_f0_42(), &testdata::expected_upper_b50_f0_42(), &testdata::expected_lower_b50_f0_42());
        check_combo("L5", &input, 0, 5, 100, 0.21, &testdata::expected_nqpr_l5(), &testdata::expected_upper_l5(), &testdata::expected_lower_l5());
        check_combo("L10", &input, 0, 10, 100, 0.21, &testdata::expected_nqpr_l10(), &testdata::expected_upper_l10(), &testdata::expected_lower_l10());
        check_combo("L10_B50_F0_42", &input, 0, 10, 50, 0.42, &testdata::expected_nqpr_l10_b50_f0_42(), &testdata::expected_upper_l10_b50_f0_42(), &testdata::expected_lower_l10_b50_f0_42());
    }

    #[test]
    fn test_long_2k() {
        let input = testdata::test_input_2k();
        check_combo("2K", &input, 0, 21, 100, 0.21, &testdata::expected_nqpr_2k(), &testdata::expected_upper_2k(), &testdata::expected_lower_2k());
    }

    #[test]
    fn test_streaming_combos() {
        let input = testdata::test_input();
        check_combo("S100", &input, 100, 21, 100, 0.21, &testdata::expected_nqpr_s100(), &testdata::expected_upper_s100(), &testdata::expected_lower_s100());
        check_combo("S150_B50", &input, 150, 21, 50, 0.21, &testdata::expected_nqpr_s150_b50(), &testdata::expected_upper_s150_b50(), &testdata::expected_lower_s150_b50());
        check_combo("S200_F0_42", &input, 200, 21, 100, 0.42, &testdata::expected_nqpr_s200_f0_42(), &testdata::expected_upper_s200_f0_42(), &testdata::expected_lower_s200_f0_42());
    }

    #[test]
    fn test_reference_projection() {
        let input = testdata::test_input();
        let ind = run_last(&input, 0, 21, 100, 0.21);
        assert!(ind.last.valid);
        let nqpr = &ind.last.nqpr;

        check_series("R50_0 NQPR", nqpr, &testdata::expected_nqpr_r50_0());
        let up: Vec<f64> = nqpr.iter().map(|&m| 50.0 * m).collect();
        let lo: Vec<f64> = nqpr.iter().map(|&m| 50.0 / m).collect();
        check_series("R50_0 UPPER", &up, &testdata::expected_upper_r50_0());
        check_series("R50_0 LOWER", &lo, &testdata::expected_lower_r50_0());

        check_series("R1000_0 NQPR", nqpr, &testdata::expected_nqpr_r1000_0());
        let up: Vec<f64> = nqpr.iter().map(|&m| 1000.0 * m).collect();
        let lo: Vec<f64> = nqpr.iter().map(|&m| 1000.0 / m).collect();
        check_series("R1000_0 UPPER", &up, &testdata::expected_upper_r1000_0());
        check_series("R1000_0 LOWER", &lo, &testdata::expected_lower_r1000_0());

        check_series("R1_2345 NQPR", nqpr, &testdata::expected_nqpr_r1_2345());
        let up: Vec<f64> = nqpr.iter().map(|&m| 1.2345 * m).collect();
        let lo: Vec<f64> = nqpr.iter().map(|&m| 1.2345 / m).collect();
        check_series("R1_2345 UPPER", &up, &testdata::expected_upper_r1_2345());
        check_series("R1_2345 LOWER", &lo, &testdata::expected_lower_r1_2345());
    }

    #[test]
    fn test_scalars() {
        let input = testdata::test_input();
        let ind = run_last(&input, 0, 21, 100, 0.21);
        assert!((ind.last.lambda - 9.739608012591481e-01).abs() <= 1e-9);
        assert!((ind.last.sigma - 2.662021797593086e-02).abs() <= 1e-9);
    }

    #[test]
    fn test_mnemonic_and_metadata() {
        let ind = QuantumPriceLevels::new(&QuantumPriceLevelsParams::default()).unwrap();
        assert_eq!(ind.metadata().mnemonic, "qpl(2048,21,100,0.21)");
        assert_eq!(ind.metadata().outputs.len(), 5);
        assert_eq!(ind.metadata().identifier, Identifier::QuantumPriceLevels);
    }

    #[test]
    fn test_update_scalar_outputs() {
        let mut ind = QuantumPriceLevels::new(&QuantumPriceLevelsParams { lookback: 100, ..Default::default() }).unwrap();
        let mut out: Output = vec![];
        for &p in &testdata::test_input() {
            out = ind.update_scalar(&Scalar { time: 0, value: p });
        }
        assert_eq!(out.len(), 5);
        let lvls = out[3].downcast_ref::<Levels>().unwrap();
        assert_eq!(lvls.levels.len(), 21);
    }

    #[test]
    fn test_invalid_params() {
        assert!(QuantumPriceLevels::new(&QuantumPriceLevelsParams { lookback: 1, ..Default::default() }).is_err());
        assert!(QuantumPriceLevels::new(&QuantumPriceLevelsParams { num_levels: 0, ..Default::default() }).is_err());
        assert!(QuantumPriceLevels::new(&QuantumPriceLevelsParams { num_bins: 1, ..Default::default() }).is_err());
        assert!(QuantumPriceLevels::new(&QuantumPriceLevelsParams { scale_factor: 0.0, ..Default::default() }).is_err());
    }
}
