//! Expected shortfall (ES, CVaR): historical, Gaussian and Cornish-Fisher.

use std::borrow::Borrow;

use crate::streaming_kbn::RawMomentsKleinKbn;

use super::norm::norm_pdf;
use super::var::{ppf, var_historical};

/// Historical ES: the negated mean of the excess returns
/// `r - risk_free_rate` that are at or below `-var_historical(...)`.
///
/// `returns` is any cloneable iterable of `f64` or `&f64` (it is traversed
/// twice), e.g. `&VecDeque<f64>` or `&[f64]`; it is not modified. Returns
/// NaN when empty (Python: `None` or empty) or when the tail is empty.
///
/// Python defaults: `risk_free_rate = 0.0`, `confidence = 0.95`.
///
/// # Panics
///
/// Panics with `"q must be between 0 and 1"` if `confidence` is outside
/// [0, 1] (Python raises `ValueError`).
pub fn es_historical<I>(returns: I, risk_free_rate: f64, confidence: f64) -> f64
where
    I: IntoIterator + Clone,
    I::Item: Borrow<f64>,
{
    if returns.clone().into_iter().next().is_none() {
        return f64::NAN;
    }
    let var = var_historical(returns.clone(), risk_free_rate, confidence);
    if var.is_nan() {
        return f64::NAN;
    }

    let mut sum_tail = 0.0;
    let mut count = 0usize;
    for r in returns {
        let excess = *r.borrow() - risk_free_rate;
        if excess <= -var {
            sum_tail += excess;
            count += 1;
        }
    }

    if count != 0 { -sum_tail / count as f64 } else { f64::NAN }
}

/// Gaussian (parametric) ES, `-mean + φ(z)·σ / (1 - confidence)` with
/// z = Φ⁻¹(confidence) and σ the population standard deviation
/// (NaN when unavailable).
///
/// Python default: `confidence = 0.95`.
///
/// # Panics
///
/// Panics if `confidence` is not in (0, 1) (Python raises `ValueError`).
pub fn es_gaussian(returns_kbn: &RawMomentsKleinKbn, confidence: f64) -> f64 {
    let mean = returns_kbn.mean();
    let std = returns_kbn.standard_deviation_ddof_0();
    if std.is_nan() {
        return f64::NAN;
    }
    let z = ppf(confidence);
    let phi_z = norm_pdf(z);
    -mean + phi_z * std / (1.0 - confidence)
}

/// Cornish-Fisher (modified) ES, following PerformanceAnalytics `ES(method="modified")`.
///
/// Falls back to Gaussian ES when skewness or kurtosis is unavailable
/// (very small samples). Python default: `confidence = 0.95`.
///
/// # Panics
///
/// Panics if `1 - confidence` is not in (0, 1) (Python raises `ValueError`).
pub fn es_cornish_fisher(returns_kbn: &RawMomentsKleinKbn, confidence: f64) -> f64 {
    let alpha = 1.0 - confidence;
    let z = ppf(alpha);
    let mean = returns_kbn.mean();
    let sigma = returns_kbn.standard_deviation_ddof_0();
    let skew = returns_kbn.skewness_moment(); // bias=True
    let kurtosis = returns_kbn.kurtosis_excess(); // bias=True, fisher=True
    // Skewness and kurtosis are unavailable for very small samples.
    // Fall back to Gaussian ES.
    if skew.is_nan() || kurtosis.is_nan() {
        return es_gaussian(returns_kbn, confidence);
    }
    let z2 = z * z;
    let z3 = z2 * z;
    let h = z + (z2 - 1.0) * skew / 6.0 + (z3 - 3.0 * z) * kurtosis / 24.0
        - (2.0 * z3 - 5.0 * z) * skew * skew / 36.0;
    let h2 = h * h;
    let h4 = h2 * h2;
    let mes = norm_pdf(h)
        * (1.0
            + h2 * h * skew / 6.0
            + (h4 * h2 - 9.0 * h4 + 9.0 * h2 + 3.0) * skew * skew / 72.0
            + (h4 - 2.0 * h2 - 1.0) * kurtosis / 24.0);
    // Python min(a, b): b if b < a else a.
    let a = -mes / alpha;
    let m = if h < a { h } else { a };
    -mean - sigma * m
}
