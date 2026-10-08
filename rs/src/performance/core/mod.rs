//! Streaming building blocks shared by the performance measures.
//!
//! Every accumulator here is O(1) (or amortized O(1)) per sample and
//! supports FIFO rolling windows: the caller owns the window and feeds
//! evicted samples to `revert()` and new samples to `update()`.
//!
//! - [`SfmRegression`]: single-factor model (alpha, beta, bull/bear beta, R²).
//! - [`percentile`]: NumPy `method="linear"` percentile.
//! - [`norm_cdf`], [`norm_pdf`], [`norm_ppf`]: standard normal distribution.
//! - [`var_historical`], [`var_gaussian`], [`var_cornish_fisher`]: value at risk.
//! - [`es_historical`], [`es_gaussian`], [`es_cornish_fisher`]: expected shortfall.
//! - [`probabilistic_sharpe_ratio`]: probabilistic Sharpe ratio.
//! - [`CumulativeReturn`]: cumulative and geometric mean returns.
//! - [`Capture`]: up/down capture, number and percentage ratios.
//! - [`WinLoss`]: winning/losing return sums, means and counts.
//! - [`ContinuousDrawdownRuns`]: continuous drawdown runs (Burke ratio).
//! - [`HighWaterMarkDrawdown`]: rolling high-water-mark drawdowns.
//! - [`DrawdownEpisode`], [`DrawdownEpisodes`]: drawdown episode tracking.
//! - [`PartialMoments`]: lower/higher partial moments about a threshold.
//! - [`RawPartialMoments`]: raw partial moment sums about zero.

pub mod capture;
pub mod continuous_drawdown_runs;
pub mod cumulative_return;
pub mod drawdown_episodes;
pub mod es;
pub mod high_watermark_drawdown;
pub mod norm;
pub mod partial_moments;
pub mod partial_moments_raw;
pub mod percentile;
pub mod probabilistic_sharpe_ratio;
pub mod sfm_regression;
pub mod var;
pub mod win_loss;

#[cfg(test)]
mod risk_helpers_tests;

pub use capture::Capture;
pub use continuous_drawdown_runs::{dd_percent, ContinuousDrawdownRuns};
pub use cumulative_return::CumulativeReturn;
pub use drawdown_episodes::{DrawdownEpisode, DrawdownEpisodes};
pub use es::{es_cornish_fisher, es_gaussian, es_historical};
pub use high_watermark_drawdown::HighWaterMarkDrawdown;
pub use norm::{norm_cdf, norm_pdf, norm_ppf};
pub use partial_moments::PartialMoments;
pub use partial_moments_raw::RawPartialMoments;
pub use percentile::percentile;
pub use probabilistic_sharpe_ratio::probabilistic_sharpe_ratio;
pub use sfm_regression::SfmRegression;
pub use var::{var_cornish_fisher, var_gaussian, var_historical};
pub use win_loss::WinLoss;

/// Shared test helpers.
#[cfg(test)]
pub(crate) mod test_support {
    pub use crate::streaming_kbn::test_support::{almost_equal, fsum, Rng};

    /// Python `random.gauss(mu, sigma)` (Box-Muller, without the cached
    /// second variate); the sequence differs from CPython, the
    /// distribution is the same.
    pub fn gauss(rng: &mut Rng, mu: f64, sigma: f64) -> f64 {
        let x2pi = rng.random() * std::f64::consts::TAU;
        let g2rad = (-2.0 * (1.0 - rng.random()).ln()).sqrt();
        mu + x2pi.cos() * g2rad * sigma
    }

    /// Python `random.choice([a, b])`.
    pub fn choice2(rng: &mut Rng, a: f64, b: f64) -> f64 {
        rng.choice(&[a, b])
    }

    /// Asserts Python `assertAlmostEqual(actual, expected, places=p)`,
    /// or that both are NaN when `expected` is NaN.
    pub fn assert_close(actual: f64, expected: f64, places: i32, msg: &str) {
        if expected.is_nan() {
            assert!(actual.is_nan(), "{msg}: expected NaN, got {actual}");
        } else {
            assert!(
                almost_equal(actual, expected, places),
                "{msg}: expected {expected}, got {actual} (places {places})"
            );
        }
    }

    /// Naive left-to-right product.
    pub fn prod(values: impl IntoIterator<Item = f64>) -> f64 {
        values.into_iter().fold(1.0, |p, x| p * x)
    }
}
