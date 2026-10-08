//! Streaming calculation of time-series performance and risk measures
//! (port of Python `py/performance/measures.py`).

use std::cmp::Ordering;
use std::collections::VecDeque;

use crate::streaming_kbn::{KleinKbnAccumulator, RawMomentsKleinKbn};

use super::core;

const SQRT2: f64 = 1.4142135623730950488016887242097;

/// Annual returns.
pub const PERIODS_PER_ANNUM_YEAR: f64 = 1.0;
/// Quarterly returns.
pub const PERIODS_PER_ANNUM_QUARTER: f64 = 4.0;
/// Monthly returns.
pub const PERIODS_PER_ANNUM_MONTH: f64 = 12.0;
/// Weekly returns.
pub const PERIODS_PER_ANNUM_WEEK: f64 = 52.0;
/// Daily (trading day) returns.
pub const PERIODS_PER_ANNUM_DAY: f64 = 252.0;
/// 390 regular-session minutes/day by 252 trading days/year.
pub const PERIODS_PER_ANNUM_MINUTE_US_EQUITIES: f64 = 98280.0;
/// 1440 minutes/day by 365 days/year.
pub const PERIODS_PER_ANNUM_MINUTE_CRYPTO: f64 = 525600.0;

/// Returns true when `0 < confidence < 1` (false for NaN).
fn is_open_unit(confidence: f64) -> bool {
    confidence > 0.0 && confidence < 1.0
}

/// True when the historical helpers accept `confidence`: their percentile
/// `q = 1 - confidence` must lie in [0, 1] (NaN rejected), exactly where
/// Python's `percentile` does not raise.
fn historical_ok(confidence: f64) -> bool {
    (0.0..=1.0).contains(&(1.0 - confidence))
}

/// True when `norm_ppf(p)` accepts `p`: Python raises only for `p <= 0`
/// or `p >= 1` (a NaN `p` passes and yields NaN).
fn ppf_ok(p: f64) -> bool {
    !(p <= 0.0 || p >= 1.0)
}

/// Python `sorted()` of floats (stable, ascending).
fn sorted(values: impl IntoIterator<Item = f64>) -> Vec<f64> {
    let mut v: Vec<f64> = values.into_iter().collect();
    v.sort_by(|a, b| a.partial_cmp(b).unwrap_or(Ordering::Equal));
    v
}

/// Jarque-Bera decision rule used by [`Measures::is_normal_distribution`].
///
/// Returns `Ok(false)` when `jb` is NaN (checked before the confidence),
/// otherwise `jb <= -2 ln(1 - confidence)` (the χ²(2) inverse CDF).
///
/// # Errors
///
/// `"confidence must be between 0 and 1"` unless `0 < confidence < 1`.
pub(crate) fn is_normal_from_jb(jb: f64, confidence: f64) -> Result<bool, String> {
    if jb.is_nan() {
        return Ok(false);
    }
    // As in Python, a NaN confidence passes this check (and yields false).
    if confidence <= 0.0 || confidence >= 1.0 {
        return Err("confidence must be between 0 and 1".to_string());
    }
    let critical = -2.0 * (-confidence).ln_1p();
    Ok(jb <= critical)
}

/// Streaming calculation of time-series performance and risk measures.
///
/// `Measures` operates on return observations that have a common,
/// explicitly defined observation period. It does not require timestamps
/// and does not perform time-based resampling.
///
/// `periods_per_annum` defines the annualization convention used by
/// annualized measures: the number of return observations assumed to
/// represent one year (daily equity 252, weekly 52, monthly 12,
/// quarterly 4, annual 1, one-minute US equity session 98 280).
///
/// The annual risk-free rate and the annual target return (MAR) are
/// converted to periodic rates as `(1 + r)^(1/periods_per_annum) - 1`
/// (used as-is when zero or when `periods_per_annum == 1`).
///
/// `rolling_window_size == 0` means an unbounded running window; a
/// positive value keeps only the most recent observations. Both return
/// series are always stored, so memory is O(n) in both modes.
///
/// Measures return NaN where they are undefined. Python properties are
/// `&self` getters; Python methods take their arguments explicitly (the
/// Python defaults are documented on each method).
#[derive(Debug, Clone)]
pub struct Measures {
    periods_per_annum: f64,
    sqrt_periods_per_annum: f64,
    annual_risk_free_rate: f64,
    risk_free_rate: f64,
    target_return: f64,
    rolling_window_size: usize,

    pub(crate) returns: VecDeque<f64>,
    pub(crate) returns_benchmark: VecDeque<f64>,

    win_loss: core::WinLoss,
    capture: core::Capture,

    // All RawMomentsKleinKbn use `ddof=1, bias=true, fisher=true`,
    // matching scipy's default behavior for kurtosis.
    returns_kbn: RawMomentsKleinKbn,
    excess_returns_kbn: RawMomentsKleinKbn,
    benchmark_returns_kbn: RawMomentsKleinKbn,
    benchmark_excess_returns_kbn: RawMomentsKleinKbn,

    sfm_regression: core::SfmRegression,
    active_returns_kbn: RawMomentsKleinKbn,

    target_returns_kbn: RawMomentsKleinKbn,
    target_partial_moments: core::PartialMoments,
    raw_partial_moments: core::RawPartialMoments,
    benchmark_target_partial_moments: core::PartialMoments,

    cumulative_return: core::CumulativeReturn,
    cumulative_excess_return: core::CumulativeReturn,
    benchmark_cumulative_return: core::CumulativeReturn,

    drawdown_continuous_runs: core::ContinuousDrawdownRuns,
    drawdown_high_watermark: core::HighWaterMarkDrawdown,
    drawdown_high_watermark_benchmark: core::HighWaterMarkDrawdown,
    drawdown_episodes: core::DrawdownEpisodes,
    drawdown_episodes_benchmark: core::DrawdownEpisodes,
}

fn moments() -> RawMomentsKleinKbn {
    RawMomentsKleinKbn::new(1, true, true)
}

impl Measures {
    /// Creates a new instance.
    ///
    /// Python defaults: `periods_per_annum = 252.0`,
    /// `annual_risk_free_rate = 0.0`, `annual_target_return = 0.0`,
    /// `rolling_window_size = 0` (unbounded).
    ///
    /// # Errors
    ///
    /// `"periods_per_annum must be positive"` if `periods_per_annum <= 0`.
    pub fn new(
        periods_per_annum: f64,
        annual_risk_free_rate: f64,
        annual_target_return: f64,
        rolling_window_size: usize,
    ) -> Result<Self, String> {
        if periods_per_annum <= 0.0 {
            return Err("periods_per_annum must be positive".to_string());
        }
        let risk_free_rate = if annual_risk_free_rate == 0.0 || periods_per_annum == 1.0 {
            annual_risk_free_rate
        } else {
            (1.0 + annual_risk_free_rate).powf(1.0 / periods_per_annum) - 1.0
        };
        let target_return = if annual_target_return == 0.0 || periods_per_annum == 1.0 {
            annual_target_return
        } else {
            (1.0 + annual_target_return).powf(1.0 / periods_per_annum) - 1.0
        };
        Ok(Self {
            periods_per_annum,
            sqrt_periods_per_annum: periods_per_annum.sqrt(),
            annual_risk_free_rate,
            risk_free_rate,
            target_return,
            rolling_window_size,
            returns: VecDeque::new(),
            returns_benchmark: VecDeque::new(),
            win_loss: core::WinLoss::new(),
            capture: core::Capture::new(),
            returns_kbn: moments(),
            excess_returns_kbn: moments(),
            benchmark_returns_kbn: moments(),
            benchmark_excess_returns_kbn: moments(),
            sfm_regression: core::SfmRegression::new(risk_free_rate),
            active_returns_kbn: moments(),
            target_returns_kbn: moments(),
            target_partial_moments: core::PartialMoments::new(target_return),
            raw_partial_moments: core::RawPartialMoments::new(),
            benchmark_target_partial_moments: core::PartialMoments::new(target_return),
            cumulative_return: core::CumulativeReturn::new(),
            cumulative_excess_return: core::CumulativeReturn::new(),
            benchmark_cumulative_return: core::CumulativeReturn::new(),
            drawdown_continuous_runs: core::ContinuousDrawdownRuns::new(),
            drawdown_high_watermark: core::HighWaterMarkDrawdown::new(rolling_window_size),
            drawdown_high_watermark_benchmark: core::HighWaterMarkDrawdown::new(rolling_window_size),
            drawdown_episodes: core::DrawdownEpisodes::new(),
            drawdown_episodes_benchmark: core::DrawdownEpisodes::new(),
        })
    }

    /// Number of return periods per annum used for annualization.
    pub fn periods_per_annum(&self) -> f64 {
        self.periods_per_annum
    }

    /// Periodic risk-free rate implied by the annual rate.
    pub fn risk_free_rate(&self) -> f64 {
        self.risk_free_rate
    }

    /// Periodic target return (MAR) implied by the annual target return.
    pub fn target_return(&self) -> f64 {
        self.target_return
    }

    /// Resets all accumulated return data and derived streaming state;
    /// configuration is preserved.
    pub fn reset(&mut self) {
        self.returns.clear();
        self.returns_benchmark.clear();

        self.win_loss.reset();
        self.capture.reset();

        self.returns_kbn.reset();
        self.excess_returns_kbn.reset();
        self.benchmark_returns_kbn.reset();
        self.benchmark_excess_returns_kbn.reset();

        self.sfm_regression.reset();
        self.active_returns_kbn.reset();

        self.target_returns_kbn.reset();
        self.target_partial_moments.reset();
        self.raw_partial_moments.reset();
        self.benchmark_target_partial_moments.reset();

        self.cumulative_return.reset();
        self.cumulative_excess_return.reset();
        self.benchmark_cumulative_return.reset();

        self.drawdown_continuous_runs.reset();
        self.drawdown_high_watermark.reset();
        self.drawdown_high_watermark_benchmark.reset();
        self.drawdown_episodes.reset();
        self.drawdown_episodes_benchmark.reset();
    }

    /// Adds one periodic portfolio return and the benchmark return for the
    /// same period, both as decimals. Inputs are not validated
    /// (`ret <= -1` yields NaN/infinite logarithms instead of Python's
    /// `ValueError`).
    pub fn add_return(&mut self, ret: f64, ret_bench: f64) {
        let evicted = self.rolling_window_size > 0 && self.returns.len() == self.rolling_window_size;
        if evicted {
            let ret_old = self.returns.pop_front().expect("window is full");
            let ret_bench_old = self.returns_benchmark.pop_front().expect("window is full");
            self.returns_kbn.revert(ret_old);
            self.excess_returns_kbn.revert(ret_old - self.risk_free_rate);
            self.target_returns_kbn.revert(ret_old - self.target_return);
            self.target_partial_moments.revert(ret_old);
            self.raw_partial_moments.revert(ret_old);
            self.benchmark_target_partial_moments.revert(ret_bench_old);
            self.win_loss.revert(ret_old);
            self.capture.revert(ret_old, ret_bench_old);
            self.cumulative_return.revert(ret_old).expect("window is not empty");
            self.cumulative_excess_return
                .revert(ret_old - self.risk_free_rate)
                .expect("window is not empty");
            self.benchmark_cumulative_return.revert(ret_bench_old).expect("window is not empty");
            self.benchmark_returns_kbn.revert(ret_bench_old);
            self.benchmark_excess_returns_kbn.revert(ret_bench_old - self.risk_free_rate);
            self.active_returns_kbn.revert(ret_old - ret_bench_old);
            self.sfm_regression.revert(ret_old, ret_bench_old);
            // High-water-mark drawdown and drawdown episodes have no
            // revert(); they evict the oldest observation themselves.
            self.drawdown_continuous_runs.revert(ret_old);
        }

        self.returns_kbn.update(ret);
        let ret_excess = ret - self.risk_free_rate;
        self.excess_returns_kbn.update(ret_excess);
        self.target_returns_kbn.update(ret - self.target_return);
        self.target_partial_moments.update(ret);
        self.raw_partial_moments.update(ret);
        self.benchmark_target_partial_moments.update(ret_bench);
        self.win_loss.update(ret);
        self.capture.update(ret, ret_bench);
        self.benchmark_returns_kbn.update(ret_bench);
        let ret_bench_excess = ret_bench - self.risk_free_rate;
        self.benchmark_excess_returns_kbn.update(ret_bench_excess);
        self.active_returns_kbn.update(ret - ret_bench);
        self.sfm_regression.update(ret, ret_bench);

        self.returns.push_back(ret);
        self.returns_benchmark.push_back(ret_bench);

        self.cumulative_return.update(ret);
        self.cumulative_excess_return.update(ret_excess);
        self.benchmark_cumulative_return.update(ret_bench);

        self.drawdown_continuous_runs.update(ret);

        // Episodes are rebuilt from the window's drawdowns when an
        // observation left the window or the drawdowns were recomputed,
        // so that episode indices refer to positions in the window.
        let recalculated = self.drawdown_high_watermark.update(ret);
        if recalculated || evicted {
            self.drawdown_episodes.recalculate(self.drawdown_high_watermark.drawdowns());
        } else {
            self.drawdown_episodes.update(self.drawdown_high_watermark.drawdown());
        }
        let recalculated = self.drawdown_high_watermark_benchmark.update(ret_bench);
        if recalculated || evicted {
            self.drawdown_episodes_benchmark
                .recalculate(self.drawdown_high_watermark_benchmark.drawdowns());
        } else {
            self.drawdown_episodes_benchmark
                .update(self.drawdown_high_watermark_benchmark.drawdown());
        }
    }

    // ------------------------------------------------------------------
    // Autocorrelation, growth, moments and normality
    // ------------------------------------------------------------------

    /// Lo (2002) autocorrelation penalty factor
    /// `sqrt(1 + 2 Σ_{k=1}^{q-1} (1 - k/q) ρ_k)`, with q = min(P, n - 1).
    ///
    /// Divide a Sharpe or Sortino ratio by it to obtain an
    /// autocorrelation-adjusted ratio. Returns 1.0 with fewer than two
    /// observations or zero variance. O(n·q).
    pub fn autocorrelation_penalty(&self) -> f64 {
        let n = self.returns_kbn.n();
        if n < 2 {
            return 1.0;
        }
        let mean = self.returns_kbn.mean();
        let denom = self.returns_kbn.variance_ddof_0() * n as f64;
        if denom.is_nan() || denom == 0.0 {
            return 1.0;
        }
        let q = self.periods_per_annum.min((n - 1) as f64) as usize;
        let w = &self.returns;
        let mut s = 0.0;
        for k in 1..q {
            let mut numer = 0.0;
            for t in k..n {
                numer += (w[t] - mean) * (w[t - k] - mean);
            }
            let rho = numer / denom;
            s += (1.0 - k as f64 / q as f64) * rho;
        }
        let x = 1.0 + 2.0 * s;
        // Python max(0.0, x)
        (if x > 0.0 { x } else { 0.0 }).sqrt()
    }

    /// Cumulative geometric return (0.0 when empty).
    pub fn cumulative_geometric_return(&self) -> f64 {
        self.cumulative_return.cumulative_geometric_return()
    }

    /// Geometric mean return per observation.
    pub fn geometric_mean_return(&self) -> f64 {
        self.cumulative_return.geometric_mean_return()
    }

    /// Compound annual growth rate (annualized geometric mean return).
    pub fn compound_annual_growth_rate(&self) -> f64 {
        self.cumulative_return.annualized_geometric_mean_return(self.periods_per_annum)
    }

    /// Skewness, scipy default (`bias=True`, i.e. 'moment').
    pub fn skewness(&self) -> f64 {
        self.returns_kbn.skewness()
    }

    /// 'moment' skewness g1 = μ3 / μ2^(3/2).
    pub fn skewness_moment(&self) -> f64 {
        self.returns_kbn.skewness_moment()
    }

    /// 'fisher' skewness g1·sqrt(n(n-1))/(n-2).
    pub fn skewness_fisher(&self) -> f64 {
        self.returns_kbn.skewness_fisher()
    }

    /// 'sample' skewness g1·n²/((n-1)(n-2)).
    pub fn skewness_sample(&self) -> f64 {
        self.returns_kbn.skewness_sample()
    }

    /// Kurtosis, scipy default (biased excess kurtosis).
    pub fn kurtosis(&self) -> f64 {
        self.returns_kbn.kurtosis()
    }

    /// Biased excess kurtosis β2 - 3.
    pub fn kurtosis_excess(&self) -> f64 {
        self.returns_kbn.kurtosis_excess()
    }

    /// Biased Pearson kurtosis β2 = μ4 / μ2².
    pub fn kurtosis_moment(&self) -> f64 {
        self.returns_kbn.kurtosis_moment()
    }

    /// Unbiased excess kurtosis ('sample excess').
    pub fn kurtosis_sample_excess(&self) -> f64 {
        self.returns_kbn.kurtosis_sample_excess()
    }

    /// PerformanceAnalytics 'sample' kurtosis (n²-1)β2/((n-2)(n-3)).
    pub fn kurtosis_sample_corrected(&self) -> f64 {
        self.returns_kbn.kurtosis_sample_corrected()
    }

    /// Unbiased Pearson kurtosis (sample excess + 3).
    pub fn kurtosis_sample(&self) -> f64 {
        self.returns_kbn.kurtosis_sample()
    }

    /// Ratio of the 'moment' skewness g1 to the Pearson kurtosis β2.
    pub fn skewness_kurtosis_ratio(&self) -> f64 {
        let s = self.returns_kbn.skewness_moment();
        let k = self.returns_kbn.kurtosis_moment();
        if k != 0.0 { s / k } else { f64::NAN }
    }

    /// Jarque-Bera statistic `n/6 · (g1² + (β2-3)²/4)` (population moments).
    pub fn jarque_bera_normality_test_statistic(&self) -> f64 {
        let s = self.returns_kbn.skewness_moment();
        let k = self.returns_kbn.kurtosis_excess();
        if s.is_nan() || k.is_nan() {
            return f64::NAN;
        }
        let n = self.returns_kbn.n() as f64;
        (n / 6.0) * (s * s + (k * k) / 4.0)
    }

    /// Jarque-Bera normality test: `Ok(true)` if normality cannot be
    /// rejected, `Ok(false)` if it is rejected or there is insufficient
    /// data (NaN statistic, checked before the confidence).
    ///
    /// Python default: `confidence = 0.95`.
    ///
    /// # Errors
    ///
    /// `"confidence must be between 0 and 1"` unless `0 < confidence < 1`.
    pub fn is_normal_distribution(&self, confidence: f64) -> Result<bool, String> {
        is_normal_from_jb(self.jarque_bera_normality_test_statistic(), confidence)
    }

    // ------------------------------------------------------------------
    // VaR, ES and reward-to-risk
    // ------------------------------------------------------------------

    /// Historical VaR `-percentile(returns, 1 - confidence)`.
    ///
    /// Python default: `confidence = 0.95`. Returns NaN where Python raises:
    /// unless `1 - confidence` is in [0, 1] (so 0 and 1 are valid).
    pub fn var_historical(&self, confidence: f64) -> f64 {
        if !historical_ok(confidence) {
            return f64::NAN;
        }
        core::var_historical(&self.returns, 0.0, confidence)
    }

    /// Gaussian VaR `-(mean + z·σ)`.
    ///
    /// Python default: `confidence = 0.95`. Returns NaN where Python raises:
    /// unless `0 < 1 - confidence < 1` (`norm_ppf(1 - confidence)`).
    pub fn var_gaussian(&self, confidence: f64) -> f64 {
        if !ppf_ok(1.0 - confidence) {
            return f64::NAN;
        }
        core::var_gaussian(&self.returns_kbn, confidence)
    }

    /// Cornish-Fisher (modified) VaR.
    ///
    /// Python default: `confidence = 0.95`. Returns NaN where Python raises:
    /// unless `0 < 1 - confidence < 1` (`norm_ppf(1 - confidence)`).
    pub fn var_cornish_fisher(&self, confidence: f64) -> f64 {
        if !ppf_ok(1.0 - confidence) {
            return f64::NAN;
        }
        core::var_cornish_fisher(&self.returns_kbn, confidence)
    }

    /// Historical expected shortfall.
    ///
    /// Python default: `confidence = 0.95`. Returns NaN where Python raises:
    /// unless `1 - confidence` is in [0, 1] (so 0 and 1 are valid).
    pub fn es_historical(&self, confidence: f64) -> f64 {
        if !historical_ok(confidence) {
            return f64::NAN;
        }
        core::es_historical(&self.returns, 0.0, confidence)
    }

    /// Gaussian expected shortfall.
    ///
    /// Python default: `confidence = 0.95`. Returns NaN where Python raises:
    /// unless `0 < confidence < 1` (`norm_ppf(confidence)`).
    pub fn es_gaussian(&self, confidence: f64) -> f64 {
        if !ppf_ok(confidence) {
            return f64::NAN;
        }
        core::es_gaussian(&self.returns_kbn, confidence)
    }

    /// Cornish-Fisher expected shortfall.
    ///
    /// Python default: `confidence = 0.95`. Returns NaN where Python raises:
    /// unless `0 < 1 - confidence < 1` (`norm_ppf(1 - confidence)`).
    pub fn es_cornish_fisher(&self, confidence: f64) -> f64 {
        if !ppf_ok(1.0 - confidence) {
            return f64::NAN;
        }
        core::es_cornish_fisher(&self.returns_kbn, confidence)
    }

    fn reward_to(&self, denom: f64) -> f64 {
        if denom != 0.0 { self.excess_returns_kbn.mean() / denom } else { f64::NAN }
    }

    /// Mean excess return divided by the historical VaR of raw returns.
    ///
    /// Python default: `confidence = 0.95`; NaN where the risk measure is NaN
    /// (including where Python raises for an invalid confidence).
    pub fn reward_to_var_ratio_historical(&self, confidence: f64) -> f64 {
        self.reward_to(self.var_historical(confidence))
    }

    /// Mean excess return divided by the Gaussian VaR of raw returns.
    ///
    /// Python default: `confidence = 0.95`; NaN where the risk measure is NaN
    /// (including where Python raises for an invalid confidence).
    pub fn reward_to_var_ratio_gaussian(&self, confidence: f64) -> f64 {
        self.reward_to(self.var_gaussian(confidence))
    }

    /// Mean excess return divided by the Cornish-Fisher VaR of raw returns.
    ///
    /// Python default: `confidence = 0.95`; NaN where the risk measure is NaN
    /// (including where Python raises for an invalid confidence).
    pub fn reward_to_var_ratio_cornish_fisher(&self, confidence: f64) -> f64 {
        self.reward_to(self.var_cornish_fisher(confidence))
    }

    /// Mean excess return divided by the historical ES of raw returns.
    ///
    /// Python default: `confidence = 0.95`; NaN where the risk measure is NaN
    /// (including where Python raises for an invalid confidence).
    pub fn reward_to_es_ratio_historical(&self, confidence: f64) -> f64 {
        self.reward_to(self.es_historical(confidence))
    }

    /// Mean excess return divided by the Gaussian ES of raw returns.
    ///
    /// Python default: `confidence = 0.95`; NaN where the risk measure is NaN
    /// (including where Python raises for an invalid confidence).
    pub fn reward_to_es_ratio_gaussian(&self, confidence: f64) -> f64 {
        self.reward_to(self.es_gaussian(confidence))
    }

    /// Mean excess return divided by the Cornish-Fisher ES of raw returns.
    ///
    /// Python default: `confidence = 0.95`; NaN where the risk measure is NaN
    /// (including where Python raises for an invalid confidence).
    pub fn reward_to_es_ratio_cornish_fisher(&self, confidence: f64) -> f64 {
        self.reward_to(self.es_cornish_fisher(confidence))
    }

    /// Mean return divided by the mean absolute deviation `Σ|r - mean|/n`.
    /// O(n).
    pub fn mean_absolute_deviation_ratio(&self) -> f64 {
        let n = self.returns_kbn.n();
        if n < 1 {
            return f64::NAN;
        }
        let mean = self.returns_kbn.mean();
        let mut sum = KleinKbnAccumulator::default();
        for &x in &self.returns {
            sum.update((x - mean).abs());
        }
        let mad = sum.value() / n as f64;
        if mad > 0.0 { mean / mad } else { f64::NAN }
    }

    // ------------------------------------------------------------------
    // Upside / downside partial moments (about the target return)
    // ------------------------------------------------------------------

    /// Upside potential ratio HPM1 / sqrt(LPM2).
    pub fn upside_potential_ratio(&self) -> f64 {
        let hpm1 = self.target_partial_moments.higher_partial_moment_1();
        let lpm2 = self.target_partial_moments.lower_partial_moment_2();
        if hpm1.is_nan() || lpm2.is_nan() || lpm2 == 0.0 {
            return f64::NAN;
        }
        hpm1 / lpm2.sqrt()
    }

    /// Subset upside potential ratio (moments averaged over their subsets).
    pub fn upside_potential_ratio_subset(&self) -> f64 {
        let pm = &self.target_partial_moments;
        let n1 = pm.upper_excess_count();
        let n2 = pm.lower_excess_count();
        if n1 == 0 || n2 == 0 {
            return f64::NAN;
        }
        let hpm1 = pm.upper_excess_moment_1_sum() / n1 as f64;
        let lpm2 = pm.lower_excess_moment_2_sum() / n2 as f64;
        if hpm1.is_nan() || lpm2.is_nan() || lpm2 == 0.0 {
            return f64::NAN;
        }
        hpm1 / lpm2.sqrt()
    }

    /// Fraction of returns above the target return.
    pub fn upside_frequency(&self) -> f64 {
        self.target_partial_moments.upside_frequency()
    }

    /// Upside potential HPM1.
    pub fn upside_potential(&self) -> f64 {
        self.target_partial_moments.higher_partial_moment_1()
    }

    /// Subset upside potential (0.0 when no return exceeds the target).
    pub fn upside_potential_subset(&self) -> f64 {
        let n = self.target_partial_moments.upper_excess_count();
        if n > 0 { self.target_partial_moments.upper_excess_moment_1_sum() / n as f64 } else { 0.0 }
    }

    /// Upside variance HPM2.
    pub fn upside_variance(&self) -> f64 {
        self.target_partial_moments.higher_partial_moment_2()
    }

    /// Subset upside variance (0.0 when no return exceeds the target).
    pub fn upside_variance_subset(&self) -> f64 {
        let n = self.target_partial_moments.upper_excess_count();
        if n > 0 { self.target_partial_moments.upper_excess_moment_2_sum() / n as f64 } else { 0.0 }
    }

    /// Upside risk sqrt(HPM2).
    pub fn upside_risk(&self) -> f64 {
        let variance = self.upside_variance();
        if variance.is_nan() { f64::NAN } else { variance.sqrt() }
    }

    /// Subset upside risk.
    pub fn upside_risk_subset(&self) -> f64 {
        let variance = self.upside_variance_subset();
        if variance.is_nan() { f64::NAN } else { variance.sqrt() }
    }

    /// Semi-deviation: downside deviation about the mean, divided by the
    /// full window length. O(n).
    pub fn semi_deviation(&self) -> f64 {
        let n = self.returns_kbn.n();
        if n == 0 {
            return f64::NAN;
        }
        let mean = self.returns_kbn.mean();
        let mut sum_squared = KleinKbnAccumulator::default();
        for &r in &self.returns {
            let deviation = r - mean;
            if deviation < 0.0 {
                sum_squared.update(deviation * deviation);
            }
        }
        (sum_squared.value() / n as f64).sqrt()
    }

    /// Downside deviation sqrt(LPM2) about the target return.
    pub fn downside_deviation(&self) -> f64 {
        let denom = self.target_partial_moments.total_count();
        if denom == 0 {
            return f64::NAN;
        }
        (self.target_partial_moments.lower_excess_moment_2_sum() / denom as f64).sqrt()
    }

    /// Subset downside deviation (0.0 when no return is below the target).
    pub fn downside_deviation_subset(&self) -> f64 {
        let denom = self.target_partial_moments.lower_excess_count();
        if denom == 0 {
            return 0.0;
        }
        (self.target_partial_moments.lower_excess_moment_2_sum() / denom as f64).sqrt()
    }

    /// Fraction of returns below the target return.
    pub fn downside_frequency(&self) -> f64 {
        self.target_partial_moments.downside_frequency()
    }

    /// Downside potential LPM1.
    pub fn downside_potential(&self) -> f64 {
        self.target_partial_moments.downside_potential()
    }

    // ------------------------------------------------------------------
    // Sharpe family
    // ------------------------------------------------------------------

    /// Sharpe ratio: mean excess return over its sample (ddof=1) standard
    /// deviation.
    pub fn sharpe_ratio(&self) -> f64 {
        let std = self.excess_returns_kbn.standard_deviation_ddof_1();
        if std.is_nan() || std == 0.0 {
            return f64::NAN;
        }
        self.excess_returns_kbn.mean() / std
    }

    fn sharpe_over(&self, denom: f64) -> f64 {
        if denom.is_nan() || denom == 0.0 {
            return f64::NAN;
        }
        self.excess_returns_kbn.mean() / denom
    }

    /// Modified Sharpe ratio with historical VaR of excess returns.
    ///
    /// Python default: `confidence = 0.95`; Returns NaN where Python raises:
    /// unless `1 - confidence` is in [0, 1] (so 0 and 1 are valid).
    pub fn sharpe_ratio_var_historical(&self, confidence: f64) -> f64 {
        if self.excess_returns_kbn.n() < 2 || !historical_ok(confidence) {
            return f64::NAN;
        }
        self.sharpe_over(core::var_historical(&self.returns, self.risk_free_rate, confidence))
    }

    /// Modified Sharpe ratio with Gaussian VaR of excess returns.
    ///
    /// Python default: `confidence = 0.95`; Returns NaN where Python raises:
    /// unless `0 < 1 - confidence < 1` (`norm_ppf(1 - confidence)`).
    pub fn sharpe_ratio_var_gaussian(&self, confidence: f64) -> f64 {
        if self.excess_returns_kbn.n() < 2 || !ppf_ok(1.0 - confidence) {
            return f64::NAN;
        }
        self.sharpe_over(core::var_gaussian(&self.excess_returns_kbn, confidence))
    }

    /// Modified Sharpe ratio with Cornish-Fisher VaR of excess returns.
    ///
    /// Python default: `confidence = 0.95`; Returns NaN where Python raises:
    /// unless `0 < 1 - confidence < 1` (`norm_ppf(1 - confidence)`).
    pub fn sharpe_ratio_var_cornish_fisher(&self, confidence: f64) -> f64 {
        if self.excess_returns_kbn.n() < 2 || !ppf_ok(1.0 - confidence) {
            return f64::NAN;
        }
        self.sharpe_over(core::var_cornish_fisher(&self.excess_returns_kbn, confidence))
    }

    /// Modified Sharpe ratio with historical ES of excess returns.
    ///
    /// Python default: `confidence = 0.95`; Returns NaN where Python raises:
    /// unless `1 - confidence` is in [0, 1] (so 0 and 1 are valid).
    pub fn sharpe_ratio_es_historical(&self, confidence: f64) -> f64 {
        if self.excess_returns_kbn.n() < 2 || !historical_ok(confidence) {
            return f64::NAN;
        }
        self.sharpe_over(core::es_historical(&self.returns, self.risk_free_rate, confidence))
    }

    /// Modified Sharpe ratio with Gaussian ES of excess returns.
    ///
    /// Python default: `confidence = 0.95`; Returns NaN where Python raises:
    /// unless `0 < confidence < 1` (`norm_ppf(confidence)`).
    pub fn sharpe_ratio_es_gaussian(&self, confidence: f64) -> f64 {
        if self.excess_returns_kbn.n() < 2 || !ppf_ok(confidence) {
            return f64::NAN;
        }
        self.sharpe_over(core::es_gaussian(&self.excess_returns_kbn, confidence))
    }

    /// Modified Sharpe ratio with Cornish-Fisher ES of excess returns.
    ///
    /// Python default: `confidence = 0.95`; Returns NaN where Python raises:
    /// unless `0 < 1 - confidence < 1` (`norm_ppf(1 - confidence)`).
    pub fn sharpe_ratio_es_cornish_fisher(&self, confidence: f64) -> f64 {
        if self.excess_returns_kbn.n() < 2 || !ppf_ok(1.0 - confidence) {
            return f64::NAN;
        }
        self.sharpe_over(core::es_cornish_fisher(&self.excess_returns_kbn, confidence))
    }

    /// Symmetric downside-risk Sharpe ratio (Ziemba, 2005):
    /// mean excess / (√2 · semi-deviation); ±infinity when the
    /// semi-deviation is zero.
    pub fn downside_sharpe_ratio(&self) -> f64 {
        let semi_dev = self.semi_deviation();
        if semi_dev.is_nan() {
            return f64::NAN;
        }
        if semi_dev == 0.0 {
            return if self.excess_returns_kbn.mean() < 0.0 { f64::NEG_INFINITY } else { f64::INFINITY };
        }
        self.excess_returns_kbn.mean() / (SQRT2 * semi_dev)
    }

    /// Adjusted Sharpe ratio (Pezier & White, 2006)
    /// SR·(1 + S·SR/6 - K·SR²/24).
    pub fn adjusted_sharpe_ratio(&self) -> f64 {
        let skewness = self.returns_kbn.skewness_moment();
        let kurtosis = self.returns_kbn.kurtosis_excess();
        if skewness.is_nan() || kurtosis.is_nan() {
            return f64::NAN;
        }
        let sr = self.sharpe_ratio();
        if sr.is_nan() {
            return f64::NAN;
        }
        sr * (1.0 + skewness * sr / 6.0 - kurtosis * sr * sr / 24.0)
    }

    /// Skewness-only adjusted Sharpe ratio SR·(1 + S·SR/6).
    pub fn adjusted_sharpe_ratio_skew_only(&self) -> f64 {
        let s = self.returns_kbn.skewness_moment();
        if s.is_nan() {
            return f64::NAN;
        }
        let sr = self.sharpe_ratio();
        if sr.is_nan() {
            return f64::NAN;
        }
        sr * (1.0 + (s * sr) / 6.0)
    }

    /// Probabilistic Sharpe ratio with sample skewness and normal kurtosis.
    ///
    /// Python default: `reference_sr = 0.0`.
    pub fn probabilistic_sharpe_ratio(&self, reference_sr: f64) -> f64 {
        core::probabilistic_sharpe_ratio(&self.returns_kbn, self.sharpe_ratio(), reference_sr, false, true)
    }

    /// Probabilistic Sharpe ratio with sample skewness and sample kurtosis.
    ///
    /// Python default: `reference_sr = 0.0`.
    pub fn probabilistic_sharpe_ratio_full(&self, reference_sr: f64) -> f64 {
        core::probabilistic_sharpe_ratio(&self.returns_kbn, self.sharpe_ratio(), reference_sr, false, false)
    }

    /// Probabilistic Sharpe ratio with zero skewness and sample kurtosis.
    ///
    /// Python default: `reference_sr = 0.0`.
    pub fn probabilistic_sharpe_ratio_symmetric(&self, reference_sr: f64) -> f64 {
        core::probabilistic_sharpe_ratio(&self.returns_kbn, self.sharpe_ratio(), reference_sr, true, false)
    }

    /// Probabilistic Sharpe ratio under Gaussian moments.
    ///
    /// Python default: `reference_sr = 0.0`.
    pub fn probabilistic_sharpe_ratio_gaussian(&self, reference_sr: f64) -> f64 {
        core::probabilistic_sharpe_ratio(&self.returns_kbn, self.sharpe_ratio(), reference_sr, true, true)
    }

    // ------------------------------------------------------------------
    // Sortino, Omega, Kappa and friends
    // ------------------------------------------------------------------

    /// Sortino ratio mean(r - T) / sqrt(LPM2).
    pub fn sortino_ratio(&self) -> f64 {
        let lpm2 = self.target_partial_moments.lower_partial_moment_2();
        if lpm2.is_nan() || lpm2 == 0.0 {
            return f64::NAN;
        }
        self.target_returns_kbn.mean() / lpm2.sqrt()
    }

    /// Sortino ratio divided by √2 (Jack Schwager).
    pub fn sortino_ratio_sqrt2(&self) -> f64 {
        self.sortino_ratio() / SQRT2
    }

    /// Sortino-Satchell ratio (same as the Sortino ratio).
    pub fn sortino_satchell_ratio(&self) -> f64 {
        let lpm2 = self.target_partial_moments.lower_partial_moment_2();
        if lpm2.is_nan() || lpm2 == 0.0 {
            return f64::NAN;
        }
        self.target_returns_kbn.mean() / lpm2.sqrt()
    }

    /// Omega ratio mean(r - T) / LPM1 + 1.
    pub fn omega_ratio(&self) -> f64 {
        let lpm1 = self.target_partial_moments.lower_partial_moment_1();
        if lpm1.is_nan() || lpm1 == 0.0 {
            return f64::NAN;
        }
        self.target_returns_kbn.mean() / lpm1 + 1.0
    }

    /// Omega-Sharpe ratio (Omega ratio - 1).
    pub fn omega_sharpe_ratio(&self) -> f64 {
        self.omega_ratio() - 1.0
    }

    /// Annualized Omega excess return
    /// `rp_ann - 3 · σd_ann · σd_ann(benchmark)`.
    /// The benchmark downside deviation is recomputed over the window, O(n).
    pub fn omega_excess_return(&self) -> f64 {
        let benchmark_downside_deviation = || -> f64 {
            if self.returns_benchmark.is_empty() {
                return f64::NAN;
            }
            let mut lower_excess_kbn = moments();
            for &r in &self.returns_benchmark {
                let excess = r - self.target_return;
                if excess < 0.0 {
                    lower_excess_kbn.update(-excess);
                }
            }
            (lower_excess_kbn.x2_sum() / self.returns_benchmark.len() as f64).sqrt()
        };
        let period = self.periods_per_annum;
        let rp = self.cumulative_return.geometric_mean_return();
        if rp.is_nan() {
            return f64::NAN;
        }
        let rp_annual = (1.0 + rp).powf(period) - 1.0;
        let sqrt_period = period.sqrt();
        let sigma_d_ann = self.downside_deviation() * sqrt_period;
        let sigma_d_ann_bench = benchmark_downside_deviation() * sqrt_period;
        rp_annual - 3.0 * sigma_d_ann * sigma_d_ann_bench
    }

    /// Kappa ratio of order 1: mean(r - T) / LPM1.
    pub fn kappa_1_ratio(&self) -> f64 {
        let lpm = self.target_partial_moments.lower_partial_moment_1();
        if lpm.is_nan() || lpm == 0.0 {
            return f64::NAN;
        }
        self.target_returns_kbn.mean() / lpm
    }

    /// Kappa ratio of order 2: mean(r - T) / LPM2^(1/2).
    pub fn kappa_2_ratio(&self) -> f64 {
        let lpm = self.target_partial_moments.lower_partial_moment_2();
        if lpm.is_nan() || lpm == 0.0 {
            return f64::NAN;
        }
        self.target_returns_kbn.mean() / lpm.sqrt()
    }

    /// Kappa ratio of order 3: mean(r - T) / LPM3^(1/3).
    pub fn kappa_3_ratio(&self) -> f64 {
        let lpm = self.target_partial_moments.lower_partial_moment_3();
        if lpm.is_nan() || lpm == 0.0 {
            return f64::NAN;
        }
        self.target_returns_kbn.mean() / lpm.powf(1.0 / 3.0)
    }

    /// Kappa ratio of order 4: mean(r - T) / LPM4^(1/4).
    pub fn kappa_4_ratio(&self) -> f64 {
        let lpm = self.target_partial_moments.lower_partial_moment_4();
        if lpm.is_nan() || lpm == 0.0 {
            return f64::NAN;
        }
        self.target_returns_kbn.mean() / lpm.powf(1.0 / 4.0)
    }

    /// Prospect ratio (Watanabe, as described by Bacon):
    /// `((Σr⁺ + λ·Σr⁻)/n - T) / downside_deviation`.
    ///
    /// Python default: `lambda_loss = 2.25`.
    pub fn prospect_ratio(&self, lambda_loss: f64) -> f64 {
        let ddev = self.downside_deviation();
        if ddev.is_nan() || ddev == 0.0 {
            return f64::NAN;
        }
        let pm = &self.raw_partial_moments;
        let n = pm.count();
        if n == 0 {
            return f64::NAN;
        }
        let prospect_return = (pm.sum_positive() + lambda_loss * pm.sum_negative()) / n as f64;
        (prospect_return - self.target_return) / ddev
    }

    /// PerformanceAnalytics prospect ratio
    /// `(Σr⁺ + 2.25·Σr⁻ - T) / (downside_deviation · n)`.
    pub fn prospect_ratio_performance_analytics(&self) -> f64 {
        let lambda_loss = 2.25;
        let ddev = self.downside_deviation();
        if ddev.is_nan() || ddev == 0.0 {
            return f64::NAN;
        }
        let pm = &self.raw_partial_moments;
        let n = pm.count();
        if n == 0 {
            return f64::NAN;
        }
        (pm.sum_positive() + lambda_loss * pm.sum_negative() - self.target_return) / (ddev * n as f64)
    }

    /// Bernardo-Ledoit ratio Σmax(r,0) / Σmax(-r,0).
    pub fn bernardo_ledoit_ratio(&self) -> f64 {
        let lpm_1 = self.raw_partial_moments.lower_partial_moment_1();
        let hpm_1 = self.raw_partial_moments.higher_partial_moment_1();
        if lpm_1 != 0.0 { hpm_1 / lpm_1 } else { f64::NAN }
    }

    /// D-ratio `(-n_down · Σdown) / (n_up · Σup)`; +infinity when there are
    /// no positive returns (including the empty case), 0.0 when there are
    /// no negative returns.
    pub fn d_ratio(&self) -> f64 {
        let n_up = self.win_loss.winning_returns_count();
        if n_up == 0 {
            return f64::INFINITY;
        }
        let n_down = self.win_loss.losing_returns_count();
        if n_down == 0 {
            return 0.0;
        }
        let sum_up = self.win_loss.winning_returns_sum();
        let sum_down = self.win_loss.losing_returns_sum();
        (-(n_down as f64) * sum_down) / (n_up as f64 * sum_up)
    }

    /// Gain-loss ratio Σwins / |Σlosses|.
    pub fn gain_loss_ratio(&self) -> f64 {
        let sum_losses = self.win_loss.losing_returns_sum().abs();
        if sum_losses != 0.0 { self.win_loss.winning_returns_sum() / sum_losses } else { f64::NAN }
    }

    /// Arithmetic mean of non-zero returns.
    pub fn mean_non_zero_return(&self) -> f64 {
        self.win_loss.non_zero_returns_mean()
    }

    /// Arithmetic mean of winning returns.
    pub fn mean_win_return(&self) -> f64 {
        self.win_loss.winning_returns_mean()
    }

    /// Arithmetic mean of losing returns.
    pub fn mean_loss_return(&self) -> f64 {
        self.win_loss.losing_returns_mean()
    }

    /// Win rate among non-zero returns.
    pub fn win_rate(&self) -> f64 {
        let non_zero_count = self.win_loss.non_zero_returns_count();
        if non_zero_count == 0 {
            return f64::NAN;
        }
        self.win_loss.winning_returns_count() as f64 / non_zero_count as f64
    }

    /// Loss rate among non-zero returns.
    pub fn loss_rate(&self) -> f64 {
        let non_zero_count = self.win_loss.non_zero_returns_count();
        if non_zero_count == 0 {
            return f64::NAN;
        }
        self.win_loss.losing_returns_count() as f64 / non_zero_count as f64
    }

    /// Variability skewness HPM2 / LPM2.
    pub fn variability_skewness(&self) -> f64 {
        let up_moment = self.target_partial_moments.higher_partial_moment_2();
        let down_moment = self.target_partial_moments.lower_partial_moment_2();
        if up_moment.is_nan() || down_moment.is_nan() || down_moment == 0.0 {
            return f64::NAN;
        }
        up_moment / down_moment
    }

    /// Volatility skewness sqrt(HPM2 / LPM2).
    pub fn volatility_skewness(&self) -> f64 {
        let var_skew = self.variability_skewness();
        if !var_skew.is_nan() { var_skew.sqrt() } else { f64::NAN }
    }

    /// Farinelli-Tibiletti ratio HPM_p^(1/p) / LPM_q^(1/q).
    ///
    /// Python defaults: `upper_order = 2`, `lower_order = 2`.
    ///
    /// # Errors
    ///
    /// `"upper_order must be 1, 2, 3, or 4"` / `"lower_order must be 1, 2,
    /// 3, or 4"` (upper checked first).
    pub fn farinelli_tibiletti_ratio(&self, upper_order: u32, lower_order: u32) -> Result<f64, String> {
        if !(1..=4).contains(&upper_order) {
            return Err("upper_order must be 1, 2, 3, or 4".to_string());
        }
        if !(1..=4).contains(&lower_order) {
            return Err("lower_order must be 1, 2, 3, or 4".to_string());
        }
        let pm = &self.target_partial_moments;
        let denom = match lower_order {
            1 => pm.lower_partial_moment_1(),
            2 => pm.lower_partial_moment_2().sqrt(),
            3 => pm.lower_partial_moment_3().powf(1.0 / 3.0),
            _ => pm.lower_partial_moment_4().powf(1.0 / 4.0),
        };
        let num = match upper_order {
            1 => pm.higher_partial_moment_1(),
            2 => pm.higher_partial_moment_2().sqrt(),
            3 => pm.higher_partial_moment_3().powf(1.0 / 3.0),
            _ => pm.higher_partial_moment_4().powf(1.0 / 4.0),
        };
        if num.is_nan() || denom.is_nan() || denom == 0.0 {
            return Ok(f64::NAN);
        }
        Ok(num / denom)
    }

    /// Rachev ratio (PerformanceAnalytics non-parametric):
    /// upper-tail ES at `beta` over lower-tail ES at `alpha`.
    /// O(n log n).
    ///
    /// Python defaults: `alpha = 0.1`, `beta = 0.1`.
    ///
    /// # Errors
    ///
    /// `"alpha must be between 0 and 1"` / `"beta must be between 0 and 1"`
    /// unless in (0, 1); checked only when there are at least two returns
    /// (otherwise `Ok(NaN)`).
    pub fn rachev_ratio(&self, alpha: f64, beta: f64) -> Result<f64, String> {
        let n = self.returns_kbn.n();
        if n < 2 {
            return Ok(f64::NAN);
        }
        if !is_open_unit(alpha) {
            return Err("alpha must be between 0 and 1".to_string());
        }
        if !is_open_unit(beta) {
            return Err("beta must be between 0 and 1".to_string());
        }
        let returns = &self.returns;
        let lower_var = core::percentile(returns, alpha)?;
        let lower_tail: Vec<f64> = returns.iter().copied().filter(|&r| r <= lower_var).collect();
        if lower_tail.is_empty() {
            return Ok(f64::NAN);
        }
        let es_lower = -lower_tail.iter().sum::<f64>() / lower_tail.len() as f64;
        let sorted_returns = sorted(returns.iter().copied());
        // PerformanceAnalytics: n.upper <- floor((1-beta) * n) (1-based).
        let upper_position = ((1.0 - beta) * n as f64).floor();
        let upper_position = if upper_position < 1.0 {
            1
        } else if upper_position > n as f64 {
            n
        } else {
            upper_position as usize
        };
        let upper_var = sorted_returns[upper_position - 1];
        let upper_tail: Vec<f64> = returns.iter().copied().filter(|&r| r >= upper_var).collect();
        if upper_tail.is_empty() || es_lower == 0.0 {
            return Ok(f64::NAN);
        }
        let es_upper = upper_tail.iter().sum::<f64>() / upper_tail.len() as f64;
        Ok(es_upper / es_lower)
    }

    // ------------------------------------------------------------------
    // Drawdowns
    // ------------------------------------------------------------------

    /// Drawdown series of cumulative geometric returns (copy of the
    /// high-water-mark drawdowns, PerformanceAnalytics `Drawdowns()`).
    pub fn drawdowns_cumulative(&self) -> Vec<f64> {
        self.drawdown_high_watermark.drawdowns().iter().copied().collect()
    }

    /// Minimum (most negative) cumulative drawdown (NaN when empty). O(n).
    pub fn min_drawdowns_cumulative(&self) -> f64 {
        self.drawdown_high_watermark.maximum_drawdown()
    }

    /// Magnitude of the worst cumulative drawdown (NaN when empty). O(n).
    pub fn worst_drawdowns_cumulative(&self) -> f64 {
        self.drawdown_high_watermark.maximum_drawdown().abs()
    }

    /// Drawdown series measured from the high-water mark.
    pub fn drawdowns_high_watermark(&self) -> Vec<f64> {
        self.drawdown_high_watermark.drawdowns().iter().copied().collect()
    }

    /// One (negative) drawdown per continuous losing-return run, in R's
    /// percent convention. With `max_runs = Some(k)`, `k > 0`, sorted
    /// worst-first and truncated to `k` runs.
    ///
    /// Python default: `max_runs = None`.
    pub fn drawdowns_continuous_runs(&self, max_runs: Option<usize>) -> Vec<f64> {
        let drawdowns = self.drawdown_continuous_runs.drawdowns();
        if drawdowns.is_empty() {
            return Vec::new();
        }
        match max_runs {
            Some(k) if k > 0 => {
                let mut d = sorted(drawdowns);
                d.truncate(k);
                d
            }
            _ => drawdowns,
        }
    }

    /// Calmar ratio: geometric mean return over the worst drawdown
    /// (not annualized).
    pub fn calmar_ratio(&self) -> f64 {
        let wdd = self.worst_drawdowns_cumulative();
        if wdd == 0.0 {
            return f64::NAN;
        }
        let cagr = self.cumulative_return.geometric_mean_return();
        if cagr.is_nan() {
            return f64::NAN;
        }
        cagr / wdd
    }

    /// Sterling ratio: geometric mean return over (worst drawdown + excess)
    /// (not annualized, PerformanceAnalytics scale = 1).
    ///
    /// Python default: `excess = 0.1`.
    pub fn sterling_ratio(&self, excess: f64) -> f64 {
        let wdd = self.worst_drawdowns_cumulative() + excess;
        if wdd == 0.0 {
            return f64::NAN;
        }
        let cagr = self.cumulative_return.geometric_mean_return();
        if cagr.is_nan() {
            return f64::NAN;
        }
        cagr / wdd
    }

    /// Burke ratio (Gm - rf) / sqrt(Σ DD²) over continuous losing runs.
    pub fn burke_ratio(&self) -> f64 {
        let rate = self.cumulative_return.geometric_mean_return() - self.risk_free_rate;
        if rate.is_nan() {
            return f64::NAN;
        }
        let sqrt_sum_drawdowns_squared = self.drawdown_continuous_runs.sqrt_sum_drawdowns_squared();
        if sqrt_sum_drawdowns_squared == 0.0 {
            return f64::NAN;
        }
        rate / sqrt_sum_drawdowns_squared
    }

    /// Modified Burke ratio (Burke ratio · sqrt(n)).
    pub fn burke_ratio_modified(&self) -> f64 {
        let burke = self.burke_ratio();
        if burke.is_nan() {
            return f64::NAN;
        }
        burke * (self.returns_kbn.n() as f64).sqrt()
    }

    /// Pain index: average depth below the high-water mark.
    pub fn pain_index(&self) -> f64 {
        -self.drawdown_high_watermark.drawdowns_mean()
    }

    /// Pain ratio (Gm - rf) / pain index.
    pub fn pain_ratio(&self) -> f64 {
        let rate = self.cumulative_return.geometric_mean_return() - self.risk_free_rate;
        if rate.is_nan() {
            return f64::NAN;
        }
        let pain_index = self.pain_index();
        if pain_index != 0.0 { rate / pain_index } else { f64::NAN }
    }

    /// Ulcer index: root-mean-square high-water-mark drawdown.
    pub fn ulcer_index(&self) -> f64 {
        self.drawdown_high_watermark.drawdowns_squared_mean().sqrt()
    }

    /// Martin ratio (Gm - rf) / ulcer index.
    pub fn martin_ratio(&self) -> f64 {
        let rate = self.cumulative_return.geometric_mean_return() - self.risk_free_rate;
        if rate.is_nan() {
            return f64::NAN;
        }
        let ulcer_index = self.ulcer_index();
        if ulcer_index != 0.0 { rate / ulcer_index } else { f64::NAN }
    }

    /// Average drawdown episode depth (0.0 when none).
    pub fn drawdown_average(&self) -> f64 {
        self.drawdown_episodes.average_episode_drawdown()
    }

    /// Average drawdown episode length (0.0 when none).
    pub fn drawdown_average_length(&self) -> f64 {
        self.drawdown_episodes.average_episode_length()
    }

    /// Average drawdown peak-to-trough period (0.0 when none).
    pub fn drawdown_average_peak_to_trough(&self) -> f64 {
        self.drawdown_episodes.average_episode_peak_to_trough()
    }

    /// Average drawdown recovery period (0.0 when none).
    pub fn drawdown_average_recovery(&self) -> f64 {
        self.drawdown_episodes.average_episode_recovery()
    }

    /// Drawdown deviation sqrt(Σ depth² / n_obs).
    pub fn drawdown_deviation(&self) -> f64 {
        self.drawdown_episodes.average_episode_drawdown_squared().sqrt()
    }

    /// Conditional drawdown at risk over the continuous drawdown path:
    /// mean magnitude of drawdowns at or below the `1 - confidence`
    /// percentile (0.0 when empty or when the percentile is >= 0).
    ///
    /// Python default: `confidence = 0.95`.
    ///
    /// # Errors
    ///
    /// `"confidence must be between 0 and 1"` unless `0 < confidence < 1`.
    pub fn cdar_average(&self, confidence: f64) -> Result<f64, String> {
        if !is_open_unit(confidence) {
            return Err("confidence must be between 0 and 1".to_string());
        }
        let drawdowns = self.drawdown_high_watermark.drawdowns();
        if drawdowns.is_empty() {
            return Ok(0.0);
        }
        let q = core::percentile(drawdowns, 1.0 - confidence)?;
        if q >= 0.0 {
            return Ok(0.0);
        }
        let mut tail_sum = KleinKbnAccumulator::default();
        let mut tail_len = 0usize;
        for &dd in drawdowns {
            if dd <= q {
                tail_len += 1;
                tail_sum.update(dd);
            }
        }
        Ok(if tail_len > 0 { -tail_sum.value() / tail_len as f64 } else { 0.0 })
    }

    /// Conditional drawdown at risk over discrete episode depths
    /// (PerformanceAnalytics default). 0.0 when there are no episodes.
    ///
    /// Python default: `confidence = 0.95`.
    ///
    /// # Errors
    ///
    /// `"confidence must be between 0 and 1"` unless `0 < confidence < 1`.
    pub fn cdar_discrete(&self, confidence: f64) -> Result<f64, String> {
        if !is_open_unit(confidence) {
            return Err("confidence must be between 0 and 1".to_string());
        }
        let depths = self.drawdown_episodes.depths();
        if depths.is_empty() {
            return Ok(0.0);
        }
        let q = core::percentile(&depths, 1.0 - confidence)?;
        let mut tail_sum = KleinKbnAccumulator::default();
        let mut tail_len = 0usize;
        for &depth in &depths {
            if depth <= q {
                tail_len += 1;
                tail_sum.update(depth);
            }
        }
        Ok(if tail_len > 0 { -tail_sum.value() / tail_len as f64 } else { 0.0 })
    }

    /// Conditional drawdown beta: portfolio peak-to-trough returns over the
    /// worst benchmark drawdown episodes, relative to the benchmark tail
    /// depth. NaN when the benchmark has no drawdown.
    ///
    /// Python default: `confidence = 0.95`.
    ///
    /// # Errors
    ///
    /// `"confidence must be between 0 and 1"` unless `0 < confidence < 1`.
    pub fn cdar_beta(&self, confidence: f64) -> Result<f64, String> {
        if !is_open_unit(confidence) {
            return Err("confidence must be between 0 and 1".to_string());
        }
        let w = &self.returns;
        let episodes = self.drawdown_episodes_benchmark.episodes();
        let depths = self.drawdown_episodes_benchmark.depths();
        if depths.is_empty() {
            return Ok(f64::NAN);
        }
        let tail_count = ((depths.len() as f64 * (1.0 - confidence)).ceil() as usize).max(1);
        let q = sorted(depths.iter().copied())[tail_count - 1];
        if q == 0.0 {
            return Ok(f64::NAN);
        }
        let mut sum_ret = KleinKbnAccumulator::default();
        let mut ret = KleinKbnAccumulator::default();
        let mut tail_len = 0usize;
        for episode in &episodes {
            if episode.depth <= q {
                tail_len += 1;
                ret.reset();
                for i in episode.from_idx..=episode.trough_idx {
                    ret.update(w[i].ln_1p());
                }
                sum_ret.update(ret.value().exp_m1());
            }
        }
        Ok(if tail_len != 0 { sum_ret.value() / (tail_len as f64 * q) } else { f64::NAN })
    }

    /// Conditional drawdown alpha
    /// `((1 + mean r)^P - 1) - β_CDaR · ((1 + mean b)^P - 1)`
    /// (PerformanceAnalytics hard-codes P = 12).
    ///
    /// Python default: `confidence = 0.95`.
    ///
    /// # Errors
    ///
    /// `"confidence must be between 0 and 1"` unless `0 < confidence < 1`.
    pub fn cdar_alpha(&self, confidence: f64) -> Result<f64, String> {
        let beta = self.cdar_beta(confidence)?;
        if beta.is_nan() {
            return Ok(f64::NAN);
        }
        let period = self.periods_per_annum;
        let r_mean = self.returns_kbn.mean();
        let b_mean = self.benchmark_returns_kbn.mean();
        let r_annual = (1.0 + r_mean).powf(period) - 1.0;
        let b_annual = (1.0 + b_mean).powf(period) - 1.0;
        Ok(r_annual - beta * b_annual)
    }

    /// Geometric mean return over the mean magnitude of the worst
    /// `max(1, trunc(n · (1 - confidence)))` drawdowns.
    /// `confidence` is not validated (a non-finite confidence selects one
    /// drawdown, where Python raises).
    ///
    /// Python default: `confidence = 0.95`.
    pub fn reward_to_conditional_drawdown(&self, confidence: f64) -> f64 {
        let cagr = self.cumulative_return.geometric_mean_return();
        if cagr.is_nan() {
            return f64::NAN;
        }
        let dd = self.drawdowns_cumulative();
        if dd.is_empty() {
            return f64::NAN;
        }
        let n_tail = ((dd.len() as f64 * (1.0 - confidence)) as i64).max(1) as usize;
        let sorted_dd = sorted(dd);
        let sorted_tail = &sorted_dd[..n_tail.min(sorted_dd.len())];
        let cdar = -sorted_tail.iter().sum::<f64>() / sorted_tail.len() as f64;
        if cdar != 0.0 { cagr / cdar } else { f64::NAN }
    }

    // ------------------------------------------------------------------
    // Single-factor model and benchmark-relative measures
    // ------------------------------------------------------------------

    /// SFM risk premium: mean excess return.
    pub fn sfm_risk_premium(&self) -> f64 {
        self.excess_returns_kbn.mean()
    }

    /// SFM alpha (intercept of r - rf on b - rf).
    pub fn sfm_alpha(&self) -> f64 {
        self.sfm_regression.alpha()
    }

    /// SFM beta (slope of r - rf on b - rf).
    pub fn sfm_beta(&self) -> f64 {
        self.sfm_regression.beta()
    }

    /// SFM bull beta (benchmark excess > 0).
    pub fn sfm_beta_bull(&self) -> f64 {
        self.sfm_regression.beta_bull()
    }

    /// SFM bear beta (benchmark excess < 0).
    pub fn sfm_beta_bear(&self) -> f64 {
        self.sfm_regression.beta_bear()
    }

    /// Timing ratio β_bull / β_bear.
    pub fn timing_ratio(&self) -> f64 {
        let denom = self.sfm_regression.beta_bear();
        if denom != 0.0 { self.sfm_regression.beta_bull() / denom } else { f64::NAN }
    }

    /// SFM coefficient of determination R².
    pub fn sfm_r2(&self) -> f64 {
        self.sfm_regression.r2()
    }

    /// Annualized Jensen's alpha `Gann_p - (β·Gann_b + (1 - β)·rf_annual)`.
    pub fn jensen_alpha(&self) -> f64 {
        let rf = self.annual_risk_free_rate;
        let mean = self.cumulative_return.annualized_geometric_mean_return(self.periods_per_annum);
        let mean_b = self.benchmark_cumulative_return.annualized_geometric_mean_return(self.periods_per_annum);
        let beta = self.sfm_beta();
        mean - (beta * mean_b + (1.0 - beta) * rf)
    }

    /// Fama beta σ0(r) / σ0(b).
    pub fn fama_beta(&self) -> f64 {
        let sigma = self.returns_kbn.standard_deviation_ddof_0();
        let sigma_b = self.benchmark_returns_kbn.standard_deviation_ddof_0();
        if sigma_b != 0.0 { sigma / sigma_b } else { f64::NAN }
    }

    /// Modigliani-Modigliani measure (periodic) `rf + mean(e)·σ_b/σ_e`.
    pub fn modigliani(&self) -> f64 {
        let sigma = self.excess_returns_kbn.standard_deviation_ddof_0();
        if sigma == 0.0 {
            return f64::NAN;
        }
        let sigma_b = self.benchmark_returns_kbn.standard_deviation_ddof_0();
        self.risk_free_rate + self.excess_returns_kbn.mean() * sigma_b / sigma
    }

    /// Annualized tracking error σ1(r - b)·sqrt(P).
    pub fn tracking_error(&self) -> f64 {
        self.active_returns_kbn.standard_deviation_ddof_1() * self.sqrt_periods_per_annum
    }

    /// Annualized active premium Gann_p - Gann_b.
    pub fn active_premium(&self) -> f64 {
        let mean = self.cumulative_return.annualized_geometric_mean_return(self.periods_per_annum);
        let mean_b = self.benchmark_cumulative_return.annualized_geometric_mean_return(self.periods_per_annum);
        mean - mean_b
    }

    /// Annualized information ratio: active premium / tracking error.
    pub fn information_ratio(&self) -> f64 {
        let te = self.tracking_error();
        if te != 0.0 { self.active_premium() / te } else { f64::NAN }
    }

    /// Modified information ratio (Israelson): the information ratio with
    /// the sign of the arithmetic mean active return.
    pub fn information_ratio_modified(&self) -> f64 {
        let excess = self.active_returns_kbn.mean();
        let ir = self.information_ratio();
        if excess.is_nan() || ir.is_nan() {
            return f64::NAN;
        }
        if excess > 0.0 { ir } else { -ir }
    }

    /// Annualized systematic risk |β|·σ1(b - rf)·sqrt(P).
    pub fn systematic_risk(&self) -> f64 {
        let beta = self.sfm_beta();
        if beta.is_nan() {
            return f64::NAN;
        }
        let benchmark_risk = self.benchmark_excess_returns_kbn.standard_deviation_ddof_1();
        if benchmark_risk.is_nan() {
            return f64::NAN;
        }
        beta.abs() * benchmark_risk * self.sqrt_periods_per_annum
    }

    /// Annualized Treynor ratio: annualized geometric excess return / β.
    pub fn treynor_ratio(&self) -> f64 {
        let beta = self.sfm_beta();
        if beta == 0.0 {
            return f64::NAN;
        }
        self.cumulative_excess_return.annualized_geometric_mean_return(self.periods_per_annum) / beta
    }

    /// Modified Treynor ratio: annualized geometric excess return /
    /// systematic risk.
    pub fn treynor_ratio_modified(&self) -> f64 {
        let sr = self.systematic_risk();
        if sr == 0.0 {
            return f64::NAN;
        }
        self.cumulative_excess_return.annualized_geometric_mean_return(self.periods_per_annum) / sr
    }

    /// Annualized specific (residual) risk σ0(ε)·sqrt(P). O(n).
    pub fn specific_risk(&self) -> f64 {
        let beta = self.sfm_beta();
        if beta.is_nan() {
            return f64::NAN;
        }
        let alpha = self.sfm_alpha();
        if alpha.is_nan() {
            return f64::NAN;
        }
        let mut epsilon_kbn = RawMomentsKleinKbn::new(0, true, true);
        let rf = self.risk_free_rate;
        for (i, &r) in self.returns.iter().enumerate() {
            epsilon_kbn.update(r - rf - alpha - beta * (self.returns_benchmark[i] - rf));
        }
        epsilon_kbn.standard_deviation_ddof_0() * self.sqrt_periods_per_annum
    }

    /// Annualized total risk sqrt(systematic² + specific²).
    pub fn total_risk(&self) -> f64 {
        let syr = self.systematic_risk();
        if syr.is_nan() {
            return f64::NAN;
        }
        let spr = self.specific_risk();
        if spr.is_nan() {
            return f64::NAN;
        }
        (syr * syr + spr * spr).sqrt()
    }

    /// Appraisal ratio: Jensen's alpha / specific risk.
    pub fn appraisal_ratio(&self) -> f64 {
        let alpha = self.jensen_alpha();
        if alpha.is_nan() {
            return f64::NAN;
        }
        let spr = self.specific_risk();
        if spr != 0.0 { alpha / spr } else { f64::NAN }
    }

    /// Modified Jensen's alpha: Jensen's alpha / β.
    pub fn jensen_alpha_modified(&self) -> f64 {
        let alpha = self.jensen_alpha();
        if alpha.is_nan() {
            return f64::NAN;
        }
        let beta = self.sfm_beta();
        if beta != 0.0 { alpha / beta } else { f64::NAN }
    }

    /// Alternative Jensen's alpha: Jensen's alpha / systematic risk.
    pub fn jensen_alpha_alternative(&self) -> f64 {
        let alpha = self.jensen_alpha();
        if alpha.is_nan() {
            return f64::NAN;
        }
        let spr = self.systematic_risk();
        if spr != 0.0 { alpha / spr } else { f64::NAN }
    }

    /// Annualized M² `Gann_p·s + rf_annual·(1 - s)`, s = σ_b/σ_p
    /// (population, annualized).
    pub fn m_squared(&self) -> f64 {
        let p_ret = self.cumulative_return.annualized_geometric_mean_return(self.periods_per_annum);
        if p_ret.is_nan() {
            return f64::NAN;
        }
        let p_std = self.returns_kbn.standard_deviation_ddof_0() * self.sqrt_periods_per_annum;
        if p_std.is_nan() || p_std == 0.0 {
            return f64::NAN;
        }
        let b_std = self.benchmark_returns_kbn.standard_deviation_ddof_0() * self.sqrt_periods_per_annum;
        if b_std.is_nan() {
            return f64::NAN;
        }
        let scale = b_std / p_std;
        p_ret * scale + self.annual_risk_free_rate * (1.0 - scale)
    }

    /// Geometric excess M² `(1 + M²)/(1 + Gann_b) - 1`.
    pub fn m_squared_excess(&self) -> f64 {
        let m_sq = self.m_squared();
        if m_sq.is_nan() {
            return f64::NAN;
        }
        let b_ret = self.benchmark_cumulative_return.annualized_geometric_mean_return(self.periods_per_annum);
        if b_ret.is_nan() {
            return f64::NAN;
        }
        (1.0 + m_sq) / (1.0 + b_ret) - 1.0
    }

    /// M² Sortino `Gann_p + Sortino·sqrt(P)·(DD_b - DD_p)`.
    pub fn m_squared_sortino(&self) -> f64 {
        let sortino = self.sortino_ratio();
        if sortino.is_nan() {
            return f64::NAN;
        }
        let p_ret = self.cumulative_return.annualized_geometric_mean_return(self.periods_per_annum);
        if p_ret.is_nan() {
            return f64::NAN;
        }
        let p_dd = self.downside_deviation();
        if p_dd.is_nan() {
            return f64::NAN;
        }
        let b_count = self.benchmark_target_partial_moments.total_count();
        if b_count == 0 {
            return f64::NAN;
        }
        let b_dd = (self.benchmark_target_partial_moments.lower_excess_moment_2_sum() / b_count as f64).sqrt();
        if b_dd.is_nan() {
            return f64::NAN;
        }
        p_ret + sortino * self.sqrt_periods_per_annum * (b_dd - p_dd)
    }

    // ------------------------------------------------------------------
    // Miscellaneous
    // ------------------------------------------------------------------

    /// Tail ratio: `cutoff` percentile over |`1 - cutoff` percentile|
    /// (not in R). NaN with fewer than two returns.
    ///
    /// Python default: `cutoff = 0.95`.
    ///
    /// # Errors
    ///
    /// `"cutoff must be between 0.5 and 1.0"` unless `0.5 < cutoff < 1`.
    pub fn tail_ratio(&self, cutoff: f64) -> Result<f64, String> {
        if !(cutoff > 0.5 && cutoff < 1.0) {
            return Err("cutoff must be between 0.5 and 1.0".to_string());
        }
        let w = &self.returns;
        if w.len() < 2 {
            return Ok(f64::NAN);
        }
        let right_tail = core::percentile(w, cutoff)?;
        let left_tail = core::percentile(w, 1.0 - cutoff)?;
        Ok(if left_tail != 0.0 { right_tail / left_tail.abs() } else { f64::NAN })
    }

    /// Full Kelly fraction mean(e) / Var1(e).
    pub fn kelly_ratio_full(&self) -> f64 {
        let mean_excess = self.excess_returns_kbn.mean();
        let var_excess = self.excess_returns_kbn.variance();
        if var_excess != 0.0 { mean_excess / var_excess } else { f64::NAN }
    }

    /// Half Kelly fraction.
    pub fn kelly_ratio(&self) -> f64 {
        self.kelly_ratio_full() / 2.0
    }

    /// Hurst exponent from single-scale rescaled range analysis
    /// ln(R/S) / ln(n). O(n).
    pub fn hurst_exponent(&self) -> f64 {
        let n = self.returns_kbn.n();
        if n < 2 {
            return f64::NAN;
        }
        let mean = self.returns_kbn.mean();
        let std = self.returns_kbn.standard_deviation_ddof_1();
        if std == 0.0 {
            return f64::NAN;
        }
        let mut cum_sum = KleinKbnAccumulator::default();
        let mut cum_min = f64::INFINITY;
        let mut cum_max = f64::NEG_INFINITY;
        for &x in &self.returns {
            cum_sum.update(x - mean);
            let val = cum_sum.value();
            if cum_min > val {
                cum_min = val;
            }
            if cum_max < val {
                cum_max = val;
            }
        }
        let delta = cum_max - cum_min;
        let rescaled_range = delta / std;
        if rescaled_range <= 0.0 {
            return f64::NAN;
        }
        rescaled_range.ln() / (n as f64).ln()
    }

    /// Bias ratio: count of returns in [0, k·σ] over 1 + count in [-k·σ, 0).
    /// O(n).
    ///
    /// Python default: `std_dev_multiplier = 1.0`.
    ///
    /// # Errors
    ///
    /// `"std_dev_multiplier must be positive"` if `std_dev_multiplier <= 0`.
    pub fn bias_ratio(&self, std_dev_multiplier: f64) -> Result<f64, String> {
        if std_dev_multiplier <= 0.0 {
            return Err("std_dev_multiplier must be positive".to_string());
        }
        let std = self.returns_kbn.standard_deviation_ddof_1();
        if std.is_nan() || std == 0.0 {
            return Ok(f64::NAN);
        }
        let threshold = std_dev_multiplier * std;
        let mut count_positive = 0usize;
        let mut count_negative = 0usize;
        for &x in &self.returns {
            if 0.0 <= x && x <= threshold {
                count_positive += 1;
            } else if -threshold <= x && x < 0.0 {
                count_negative += 1;
            }
        }
        Ok(count_positive as f64 / (1 + count_negative) as f64)
    }

    /// K-ratio (Lars Kestner): slope of the cumulative log-equity curve
    /// over (its standard error · sqrt(n)). O(n), plain sums.
    pub fn k_ratio(&self) -> f64 {
        let n = self.returns_kbn.n();
        if n < 3 {
            return f64::NAN;
        }
        let nf = n as f64;
        let mut equity = Vec::with_capacity(n);
        let mut cum_sum = 0.0;
        for &x in &self.returns {
            cum_sum += x.ln_1p();
            equity.push(cum_sum);
        }
        let mut sum_t = 0.0;
        let mut sum_t2 = 0.0;
        let mut sum_eq = 0.0;
        let mut sum_te = 0.0;
        for (i, &eq_val) in equity.iter().enumerate() {
            let t_val = i as f64;
            sum_t += t_val;
            sum_t2 += t_val * t_val;
            sum_eq += eq_val;
            sum_te += t_val * eq_val;
        }
        let t_mean = sum_t / nf;
        let equity_mean = sum_eq / nf;
        let s_tt = sum_t2 - nf * t_mean.powf(2.0);
        let s_te = sum_te - nf * t_mean * equity_mean;
        if s_tt == 0.0 {
            return f64::NAN;
        }
        let slope = s_te / s_tt;
        let intercept = equity_mean - slope * t_mean;
        let mut sum_sq_residuals = 0.0;
        for (i, &eq_val) in equity.iter().enumerate() {
            let t_val = i as f64;
            let predicted = intercept + slope * t_val;
            let residual = eq_val - predicted;
            sum_sq_residuals += residual * residual;
        }
        let mut residual_var = sum_sq_residuals / (nf - 2.0);
        if residual_var < 0.0 {
            residual_var = 0.0;
        }
        let se_slope = (residual_var / s_tt).sqrt();
        if se_slope == 0.0 {
            return f64::NAN;
        }
        slope / (se_slope * nf.sqrt())
    }

    /// Jack Schwager's gain-to-pain ratio, as implemented in Python:
    /// mean return divided by the raw **sum** of losses Σmax(-r, 0)
    /// (effectively (Σr / Σlosses) / n, unlike the textbook Σr / Σlosses).
    pub fn gain_to_pain_ratio(&self) -> f64 {
        let lpm1 = self.raw_partial_moments.lower_partial_moment_1();
        if lpm1.is_nan() || lpm1 == 0.0 {
            return f64::NAN;
        }
        self.returns_kbn.mean() / lpm1
    }

    // ------------------------------------------------------------------
    // Capture
    // ------------------------------------------------------------------

    /// Upside capture ratio (benchmark > 0 periods).
    ///
    /// Python default: `geometric = true`.
    pub fn upside_capture_ratio(&self, geometric: bool) -> f64 {
        if geometric {
            self.capture.upside_capture_ratio_geometric()
        } else {
            self.capture.upside_capture_ratio_arithmetic()
        }
    }

    /// Downside capture ratio (benchmark <= 0 periods).
    ///
    /// Python default: `geometric = true`.
    pub fn downside_capture_ratio(&self, geometric: bool) -> f64 {
        if geometric {
            self.capture.downside_capture_ratio_geometric()
        } else {
            self.capture.downside_capture_ratio_arithmetic()
        }
    }

    /// Overall capture ratio: upside capture / downside capture.
    ///
    /// Python default: `geometric = true`.
    pub fn overall_capture_ratio(&self, geometric: bool) -> f64 {
        let up = self.upside_capture_ratio(geometric);
        let down = self.downside_capture_ratio(geometric);
        if up.is_nan() || down.is_nan() || down == 0.0 {
            return f64::NAN;
        }
        up / down
    }

    /// Up number ratio.
    pub fn up_number_ratio(&self) -> f64 {
        self.capture.up_number_ratio()
    }

    /// Down number ratio (down bucket: benchmark <= 0).
    pub fn down_number_ratio(&self) -> f64 {
        self.capture.down_number_ratio()
    }

    /// Up percentage ratio.
    pub fn up_percentage_ratio(&self) -> f64 {
        self.capture.up_percentage_ratio()
    }

    /// Down percentage ratio (strict benchmark < 0).
    pub fn down_percentage_ratio(&self) -> f64 {
        self.capture.down_percentage_ratio()
    }
}
