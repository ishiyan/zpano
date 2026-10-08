//! Tests for [`Measures`] (port of Python `py/performance/test_measures.py`).
//!
//! Test input data is the 'portfolio_bacon' dataset from the
//! PerformanceAnalytics R package: Bacon (2008, 2nd ed.) p. 65 portfolio
//! and p. 66 benchmark returns. Expected values are in
//! `performance::reference_data` (generated from the Python package).

use super::measures::Measures;
use super::core::test_support::fsum;

mod benchmark;
mod drawdowns;
mod formulas;
mod general;
mod ratios;
mod risk;

pub(super) const BACON_PORTFOLIO_RETURNS: [f64; 24] = [
    0.003, 0.026, 0.011, -0.010, 0.015, 0.025, 0.016, 0.067, -0.014, 0.040, -0.005, 0.081, 0.040, -0.037,
    -0.061, 0.017, -0.049, -0.022, 0.070, 0.058, -0.065, 0.024, -0.005, -0.009,
];

pub(super) const BACON_BENCHMARK_RETURNS: [f64; 24] = [
    0.002, 0.025, 0.018, -0.011, 0.014, 0.018, 0.014, 0.065, -0.015, 0.042, -0.006, 0.083, 0.039, -0.038,
    -0.062, 0.015, -0.048, 0.021, 0.060, 0.056, -0.067, 0.019, -0.003, 0.000,
];

pub(super) const BACON_PORTFOLIO_LEN: usize = 24;

/// Extended Bacon 2023 (3rd edition) portfolio data.
pub(super) const BACON_2023_PORTFOLIO_RETURNS: [f64; 36] = [
    0.003, 0.026, 0.011, -0.009, 0.014, 0.024, 0.015, 0.066, -0.014, 0.039, -0.005, 0.081, 0.040, -0.037,
    -0.061, 0.014, -0.049, -0.021, 0.062, 0.058, -0.064, 0.017, -0.004, -0.002, -0.021, 0.011, 0.047,
    0.024, 0.033, -0.007, 0.047, 0.006, 0.010, -0.002, 0.034, 0.010,
];

pub(super) const BACON_2023_DRAWDOWN_CONTINUOUS_WITHOUT_ZEROES: [f64; 9] =
    [-0.0090, -0.0140, -0.0050, -0.0960, -0.0690, -0.0640, -0.0270, -0.0070, -0.0020];

pub(super) const BACON_2023_DRAWDOWN_FROM_PEAK: [f64; 36] = [
    0.0, 0.0, 0.0, -0.0090, 0.0, 0.0, 0.0, 0.0, -0.0140, 0.0, -0.0050, 0.0, 0.0, -0.0370, -0.0957,
    -0.0831, -0.1280, -0.1463, -0.0934, -0.0408, -0.1022, -0.0869, -0.0906, -0.0924, -0.1115, -0.1017,
    -0.0595, -0.0369, -0.0051, -0.0121, 0.0, 0.0, 0.0, -0.0020, 0.0, 0.0,
];

pub(super) const BACON_2023_PORTFOLIO_LEN: usize = 36;

// ----------------------------------------------------------------------
// Assertions (Python assertFloatEqual / assertSeriesEqual)
// ----------------------------------------------------------------------

/// Tolerance of a float comparison.
#[derive(Debug, Clone, Copy)]
pub(super) enum Tol {
    /// `assertAlmostEqual(places=p)`.
    Places(i32),
    /// `assertAlmostEqual(delta=d)`.
    Delta(f64),
    /// `math.isclose(rel_tol=rel, abs_tol=abs)`.
    Rel(f64, f64),
}

pub(super) use Tol::{Delta, Places, Rel};

/// Python `assertFloatEqual`: NaN and ±infinity must match exactly,
/// otherwise the actual value must be finite and within tolerance.
pub(super) fn check_float(actual: f64, expected: f64, tol: Tol) -> Result<(), String> {
    if expected.is_nan() {
        return if actual.is_nan() { Ok(()) } else { Err(format!("expected NaN, actual {actual}")) };
    }
    if expected.is_infinite() {
        return if actual == expected {
            Ok(())
        } else {
            Err(format!("expected {expected}, actual {actual}"))
        };
    }
    if actual.is_nan() || actual.is_infinite() {
        return Err(format!("expected {expected}, got {actual}"));
    }
    let ok = actual == expected
        || match tol {
            Places(p) => (actual - expected).abs() <= 0.5 * 10f64.powi(-p),
            Delta(d) => (actual - expected).abs() <= d,
            Rel(rel, abs) => (actual - expected).abs() <= (rel * actual.abs().max(expected.abs())).max(abs),
        };
    if ok {
        Ok(())
    } else {
        Err(format!("expected {expected}, got {actual} ({tol:?})"))
    }
}

/// Panicking [`check_float`].
pub(super) fn assert_float(actual: f64, expected: f64, tol: Tol, prefix: &str) {
    if let Err(e) = check_float(actual, expected, tol) {
        panic!("{prefix}: {e}");
    }
}

/// Python `assertSeriesEqual`: equal lengths, and element-wise
/// [`check_float`] from index `skip` on.
pub(super) fn check_series(actual: &[f64], expected: &[f64], tol: Tol, skip: usize) -> Result<(), String> {
    if actual.len() != expected.len() {
        return Err(format!("series length {} != {}", actual.len(), expected.len()));
    }
    for (i, (&a, &e)) in actual.iter().zip(expected).enumerate() {
        if i >= skip {
            check_float(a, e, tol).map_err(|err| format!("step {i}: {err}"))?;
        }
    }
    Ok(())
}

/// Panicking [`check_series`].
pub(super) fn assert_series(actual: &[f64], expected: &[f64], tol: Tol, skip: usize, prefix: &str) {
    if let Err(e) = check_series(actual, expected, tol, skip) {
        panic!("{prefix}: {e}");
    }
}

// ----------------------------------------------------------------------
// Streaming runners
// ----------------------------------------------------------------------

/// Python `periods_per_annum(daily, monthly)`: 252, 12, or 1.
pub(super) const DAILY: f64 = 252.0;
pub(super) const MONTHLY: f64 = 12.0;
pub(super) const YEARLY: f64 = 1.0;

/// Configuration of a streaming run (Python `run_stream_*` keyword
/// arguments).
#[derive(Debug, Clone, Copy)]
pub(super) struct Cfg<'a> {
    pub ppa: f64,
    pub rf: f64,
    pub mar: f64,
    pub returns: &'a [f64],
    pub bench: &'a [f64],
    pub window: usize,
}

/// Yearly, zero rates, Bacon portfolio and benchmark, unbounded window.
pub(super) fn cfg<'a>() -> Cfg<'a> {
    Cfg {
        ppa: YEARLY,
        rf: 0.0,
        mar: 0.0,
        returns: &BACON_PORTFOLIO_RETURNS,
        bench: &BACON_BENCHMARK_RETURNS,
        window: 0,
    }
}

impl<'a> Cfg<'a> {
    pub fn daily(mut self) -> Self {
        self.ppa = DAILY;
        self
    }
    pub fn monthly(mut self) -> Self {
        self.ppa = MONTHLY;
        self
    }
    pub fn rf(mut self, rf: f64) -> Self {
        self.rf = rf;
        self
    }
    pub fn mar(mut self, mar: f64) -> Self {
        self.mar = mar;
        self
    }
    pub fn returns(mut self, returns: &'a [f64]) -> Self {
        self.returns = returns;
        self
    }
    pub fn bench(mut self, bench: &'a [f64]) -> Self {
        self.bench = bench;
        self
    }
    /// Same series for portfolio and benchmark.
    pub fn data(self, returns: &'a [f64]) -> Self {
        self.returns(returns).bench(returns)
    }
}

/// Python `run_stream_property` / `run_stream_method` /
/// `run_stream_callback`: the value of `f` after each added observation.
pub(super) fn run<T>(c: Cfg<'_>, f: impl Fn(&Measures) -> T) -> Vec<T> {
    let mut measures = Measures::new(c.ppa, c.rf, c.mar, c.window).unwrap();
    measures.reset();
    let mut results = Vec::with_capacity(c.returns.len());
    for i in 0..c.returns.len() {
        measures.add_return(c.returns[i], c.bench[i]);
        results.push(f(&measures));
    }
    results
}

/// Python `make_measures`.
pub(super) fn make_measures(window: usize, annual_rf: f64, annual_mar: f64, ppa: f64) -> Measures {
    let mut m = Measures::new(ppa, annual_rf, annual_mar, window).unwrap();
    m.reset();
    m
}

/// Python `add_bacon` (all observations of the given series).
pub(super) fn add_bacon(m: &mut Measures, returns: &[f64], bench: &[f64]) {
    for i in 0..returns.len() {
        m.add_return(returns[i], bench[i]);
    }
}

/// Python `(1 + rate) ** periods - 1`.
pub(super) fn compound(rate: f64, periods: f64) -> f64 {
    (1.0 + rate).powf(periods) - 1.0
}

/// Python `statistics.fmean`.
pub(super) fn fmean(xs: &[f64]) -> f64 {
    fsum(xs) / xs.len() as f64
}

/// Python `statistics.pstdev` (computed with `fsum`, not exactly).
pub(super) fn pstdev(xs: &[f64]) -> f64 {
    let mean = fmean(xs);
    let d: Vec<f64> = xs.iter().map(|x| (x - mean) * (x - mean)).collect();
    (fsum(&d) / xs.len() as f64).sqrt()
}

/// Python `statistics.stdev` (computed with `fsum`, not exactly).
pub(super) fn stdev(xs: &[f64]) -> f64 {
    let mean = fmean(xs);
    let d: Vec<f64> = xs.iter().map(|x| (x - mean) * (x - mean)).collect();
    (fsum(&d) / (xs.len() - 1) as f64).sqrt()
}

// ----------------------------------------------------------------------
// Public measures (Python `public_measures()` / `evaluate()`)
// ----------------------------------------------------------------------

/// Value of a public measure.
#[derive(Debug, Clone, PartialEq)]
pub(super) enum Value {
    F(f64),
    B(bool),
    L(Vec<f64>),
}

pub(super) type Eval = fn(&Measures) -> Value;

macro_rules! prop {
    ($name:ident) => {
        (stringify!($name), (|m: &Measures| Value::F(m.$name())) as Eval)
    };
}

macro_rules! meth {
    ($name:ident, $($arg:expr),*) => {
        (stringify!($name), (|m: &Measures| Value::F(m.$name($($arg),*))) as Eval)
    };
}

macro_rules! fallible {
    ($name:ident, $($arg:expr),*) => {
        (stringify!($name), (|m: &Measures| Value::F(m.$name($($arg),*).expect(stringify!($name)))) as Eval)
    };
}

/// All public measures of [`Measures`] (properties, and methods called
/// with the Python defaults), except the mutators. 144 entries, the same
/// set as Python's `inspect`-based `public_measures()`.
pub(super) fn public_measures() -> Vec<(&'static str, Eval)> {
    vec![
        prop!(active_premium),
        prop!(adjusted_sharpe_ratio),
        prop!(adjusted_sharpe_ratio_skew_only),
        prop!(appraisal_ratio),
        prop!(autocorrelation_penalty),
        prop!(bernardo_ledoit_ratio),
        fallible!(bias_ratio, 1.0),
        prop!(burke_ratio),
        prop!(burke_ratio_modified),
        prop!(calmar_ratio),
        fallible!(cdar_alpha, 0.95),
        fallible!(cdar_average, 0.95),
        fallible!(cdar_beta, 0.95),
        fallible!(cdar_discrete, 0.95),
        prop!(compound_annual_growth_rate),
        prop!(cumulative_geometric_return),
        prop!(d_ratio),
        prop!(down_number_ratio),
        prop!(down_percentage_ratio),
        meth!(downside_capture_ratio, true),
        prop!(downside_deviation),
        prop!(downside_deviation_subset),
        prop!(downside_frequency),
        prop!(downside_potential),
        prop!(downside_sharpe_ratio),
        prop!(drawdown_average),
        prop!(drawdown_average_length),
        prop!(drawdown_average_peak_to_trough),
        prop!(drawdown_average_recovery),
        prop!(drawdown_deviation),
        ("drawdowns_continuous_runs", (|m: &Measures| Value::L(m.drawdowns_continuous_runs(None))) as Eval),
        ("drawdowns_cumulative", (|m: &Measures| Value::L(m.drawdowns_cumulative())) as Eval),
        ("drawdowns_high_watermark", (|m: &Measures| Value::L(m.drawdowns_high_watermark())) as Eval),
        meth!(es_cornish_fisher, 0.95),
        meth!(es_gaussian, 0.95),
        meth!(es_historical, 0.95),
        prop!(fama_beta),
        fallible!(farinelli_tibiletti_ratio, 2, 2),
        prop!(gain_loss_ratio),
        prop!(gain_to_pain_ratio),
        prop!(geometric_mean_return),
        prop!(hurst_exponent),
        prop!(information_ratio),
        prop!(information_ratio_modified),
        (
            "is_normal_distribution",
            (|m: &Measures| Value::B(m.is_normal_distribution(0.95).expect("is_normal_distribution"))) as Eval,
        ),
        prop!(jarque_bera_normality_test_statistic),
        prop!(jensen_alpha),
        prop!(jensen_alpha_alternative),
        prop!(jensen_alpha_modified),
        prop!(k_ratio),
        prop!(kappa_1_ratio),
        prop!(kappa_2_ratio),
        prop!(kappa_3_ratio),
        prop!(kappa_4_ratio),
        prop!(kelly_ratio),
        prop!(kelly_ratio_full),
        prop!(kurtosis),
        prop!(kurtosis_excess),
        prop!(kurtosis_moment),
        prop!(kurtosis_sample),
        prop!(kurtosis_sample_corrected),
        prop!(kurtosis_sample_excess),
        prop!(loss_rate),
        prop!(m_squared),
        prop!(m_squared_excess),
        prop!(m_squared_sortino),
        prop!(martin_ratio),
        prop!(mean_absolute_deviation_ratio),
        prop!(mean_loss_return),
        prop!(mean_non_zero_return),
        prop!(mean_win_return),
        prop!(min_drawdowns_cumulative),
        prop!(modigliani),
        prop!(omega_excess_return),
        prop!(omega_ratio),
        prop!(omega_sharpe_ratio),
        meth!(overall_capture_ratio, true),
        prop!(pain_index),
        prop!(pain_ratio),
        meth!(probabilistic_sharpe_ratio, 0.0),
        meth!(probabilistic_sharpe_ratio_full, 0.0),
        meth!(probabilistic_sharpe_ratio_gaussian, 0.0),
        meth!(probabilistic_sharpe_ratio_symmetric, 0.0),
        meth!(prospect_ratio, 2.25),
        prop!(prospect_ratio_performance_analytics),
        fallible!(rachev_ratio, 0.1, 0.1),
        meth!(reward_to_conditional_drawdown, 0.95),
        meth!(reward_to_es_ratio_cornish_fisher, 0.95),
        meth!(reward_to_es_ratio_gaussian, 0.95),
        meth!(reward_to_es_ratio_historical, 0.95),
        meth!(reward_to_var_ratio_cornish_fisher, 0.95),
        meth!(reward_to_var_ratio_gaussian, 0.95),
        meth!(reward_to_var_ratio_historical, 0.95),
        prop!(semi_deviation),
        prop!(sfm_alpha),
        prop!(sfm_beta),
        prop!(sfm_beta_bear),
        prop!(sfm_beta_bull),
        prop!(sfm_r2),
        prop!(sfm_risk_premium),
        prop!(sharpe_ratio),
        meth!(sharpe_ratio_es_cornish_fisher, 0.95),
        meth!(sharpe_ratio_es_gaussian, 0.95),
        meth!(sharpe_ratio_es_historical, 0.95),
        meth!(sharpe_ratio_var_cornish_fisher, 0.95),
        meth!(sharpe_ratio_var_gaussian, 0.95),
        meth!(sharpe_ratio_var_historical, 0.95),
        prop!(skewness),
        prop!(skewness_fisher),
        prop!(skewness_kurtosis_ratio),
        prop!(skewness_moment),
        prop!(skewness_sample),
        prop!(sortino_ratio),
        prop!(sortino_ratio_sqrt2),
        prop!(sortino_satchell_ratio),
        prop!(specific_risk),
        meth!(sterling_ratio, 0.1),
        prop!(systematic_risk),
        fallible!(tail_ratio, 0.95),
        prop!(timing_ratio),
        prop!(total_risk),
        prop!(tracking_error),
        prop!(treynor_ratio),
        prop!(treynor_ratio_modified),
        prop!(ulcer_index),
        prop!(up_number_ratio),
        prop!(up_percentage_ratio),
        meth!(upside_capture_ratio, true),
        prop!(upside_frequency),
        prop!(upside_potential),
        prop!(upside_potential_ratio),
        prop!(upside_potential_ratio_subset),
        prop!(upside_potential_subset),
        prop!(upside_risk),
        prop!(upside_risk_subset),
        prop!(upside_variance),
        prop!(upside_variance_subset),
        meth!(var_cornish_fisher, 0.95),
        meth!(var_gaussian, 0.95),
        meth!(var_historical, 0.95),
        prop!(variability_skewness),
        prop!(volatility_skewness),
        prop!(win_rate),
        prop!(worst_drawdowns_cumulative),
    ]
}

#[test]
fn test_public_measures_count() {
    let names: Vec<&str> = public_measures().iter().map(|(n, _)| *n).collect();
    assert_eq!(names.len(), 144);
    let mut sorted = names.clone();
    sorted.sort();
    sorted.dedup();
    assert_eq!(sorted.len(), 144, "duplicate names");
    assert_eq!(sorted, names, "names are sorted like Python inspect.getmembers");
}
