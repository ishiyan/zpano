//! Value at risk (VaR): historical, Gaussian and Cornish-Fisher.

use std::borrow::Borrow;

use crate::streaming_kbn::RawMomentsKleinKbn;

use super::norm::norm_ppf;
use super::percentile::percentile;

/// Historical VaR: the negated `(1 - confidence)` percentile of the
/// excess returns `r - risk_free_rate` (NumPy `method="linear"`).
///
/// `returns` is any iterable of `f64` or `&f64`, e.g. `&VecDeque<f64>` or
/// `&[f64]`; it is not modified. Returns NaN when empty (Python: `None`
/// or empty).
///
/// Python defaults: `risk_free_rate = 0.0`, `confidence = 0.95`.
///
/// # Panics
///
/// Panics with `"q must be between 0 and 1"` if `confidence` is outside
/// [0, 1] (Python raises `ValueError`).
pub fn var_historical<I>(returns: I, risk_free_rate: f64, confidence: f64) -> f64
where
    I: IntoIterator,
    I::Item: Borrow<f64>,
{
    let mut w = returns.into_iter().peekable();
    if w.peek().is_none() {
        return f64::NAN;
    }
    let q = 1.0 - confidence;
    let result = if risk_free_rate == 0.0 {
        percentile(w, q)
    } else {
        percentile(w.map(|r| *r.borrow() - risk_free_rate), q)
    };
    -result.unwrap_or_else(|e| panic!("{e}"))
}

/// Gaussian (parametric) VaR, `-(mean + z·σ)` with z = Φ⁻¹(1 - confidence)
/// and σ the population standard deviation (NaN when unavailable).
///
/// Python default: `confidence = 0.95`.
///
/// # Panics
///
/// Panics if `1 - confidence` is not in (0, 1) (Python raises `ValueError`).
pub fn var_gaussian(returns_kbn: &RawMomentsKleinKbn, confidence: f64) -> f64 {
    let mean = returns_kbn.mean();
    let std = returns_kbn.standard_deviation_ddof_0();
    if std.is_nan() {
        return f64::NAN;
    }
    let z = ppf(1.0 - confidence);
    -(mean + z * std)
}

/// Cornish-Fisher (modified) VaR: Gaussian VaR with the z-score adjusted
/// by the population skewness and excess kurtosis.
///
/// Falls back to Gaussian VaR when skewness or kurtosis is unavailable
/// (very small samples). Python default: `confidence = 0.95`.
///
/// # Panics
///
/// Panics if `1 - confidence` is not in (0, 1) (Python raises `ValueError`).
pub fn var_cornish_fisher(returns_kbn: &RawMomentsKleinKbn, confidence: f64) -> f64 {
    let mean = returns_kbn.mean();
    let std = returns_kbn.standard_deviation_ddof_0();
    if std.is_nan() {
        return f64::NAN;
    }
    // Cornish-Fisher expansion for z-score adjustment
    let mut z = ppf(1.0 - confidence);
    let skew = returns_kbn.skewness_moment(); // bias=True
    let kurtosis = returns_kbn.kurtosis_excess(); // bias=True, fisher=True
    // Skewness and kurtosis are unavailable for very small samples.
    // Fall back to Gaussian VaR.
    if skew.is_nan() || kurtosis.is_nan() {
        return -(mean + z * std);
    }
    // Cornish-Fisher expansion
    let z2 = z * z;
    let z3 = z2 * z;
    z = z + (z2 - 1.0) * skew / 6.0 + (z3 - 3.0 * z) * kurtosis / 24.0
        - (2.0 * z3 - 5.0 * z) * skew * skew / 36.0;
    -(mean + z * std)
}

/// `norm_ppf` that panics like Python's `ValueError`.
pub(super) fn ppf(p: f64) -> f64 {
    norm_ppf(p).unwrap_or_else(|e| panic!("{e}"))
}
