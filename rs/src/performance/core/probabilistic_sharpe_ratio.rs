//! Probabilistic Sharpe ratio (Bailey and López de Prado).

use crate::streaming_kbn::RawMomentsKleinKbn;

use super::norm::norm_cdf;

/// Probabilistic Sharpe ratio: the probability that the true Sharpe ratio
/// exceeds `reference_sr`, given the observed (non-annualized) Sharpe
/// ratio `sr` and the sample moments in `returns_kbn`:
///
/// ```text
/// PSR = Φ((SR - SR*) · √(n - 1) / √(1 - γ₃·SR + (γ₄ - 1)/4 · SR²))
/// ```
///
/// * `zero_skewness`: assume γ₃ = 0 instead of the population skewness.
/// * `normal_kurtosis`: assume γ₄ = 3 instead of the population kurtosis.
///
/// Returns NaN when `sr` or a required moment is NaN, or the denominator
/// is zero. Python defaults: `reference_sr = 0.0`, `zero_skewness = false`,
/// `normal_kurtosis = true`.
///
/// Where Python's `math.sqrt` would raise on a negative argument (a
/// negative radicand, or n = 0), this returns NaN.
pub fn probabilistic_sharpe_ratio(
    returns_kbn: &RawMomentsKleinKbn,
    sr: f64,
    reference_sr: f64,
    zero_skewness: bool,
    normal_kurtosis: bool,
) -> f64 {
    if sr.is_nan() {
        return f64::NAN;
    }
    let skewness = if zero_skewness {
        0.0
    } else {
        let s = returns_kbn.skewness_moment(); // or _sample
        if s.is_nan() {
            return f64::NAN;
        }
        s
    };
    let kurtosis = if normal_kurtosis {
        3.0 // excess kurtosis = 0, so K = 3
    } else {
        let k = returns_kbn.kurtosis_excess();
        if k.is_nan() {
            return f64::NAN;
        }
        k + 3.0 // convert to regular kurtosis
    };

    let denom = (1.0 - sr * skewness + (sr * sr) * (kurtosis - 1.0) / 4.0).sqrt();
    if denom == 0.0 {
        return f64::NAN;
    }

    let z = (sr - reference_sr) * (returns_kbn.n() as f64 - 1.0).sqrt() / denom;
    norm_cdf(z)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::performance::core::norm::erf;
    use crate::performance::core::test_support::{almost_equal, fsum};

    #[test]
    fn test_four_moment_assumptions() {
        let values = [-0.03, -0.01, 0.02, 0.04, 0.07];
        let mut moments = RawMomentsKleinKbn::default();
        for &value in &values {
            moments.update(value);
        }
        let n = values.len() as f64;
        let mean = fsum(&values) / n;
        let pw = |k: i32| -> Vec<f64> { values.iter().map(|&x| (x - mean).powi(k)).collect() };
        let mu2 = fsum(&pw(2)) / n;
        let skew = fsum(&pw(3)) / n / mu2.powf(1.5);
        let kurt = fsum(&pw(4)) / n / (mu2 * mu2);
        let sr = 0.75;
        let reference_sr = 0.1;

        for zero_skewness in [false, true] {
            for normal_kurtosis in [false, true] {
                let s = if zero_skewness { 0.0 } else { skew };
                let k = if normal_kurtosis { 3.0 } else { kurt };
                let denominator = (1.0 - sr * s + sr * sr * (k - 1.0) / 4.0).sqrt();
                let z = (sr - reference_sr) * (n - 1.0).sqrt() / denominator;
                let expected = 0.5 * (1.0 + erf(z / 2f64.sqrt()));
                let actual =
                    probabilistic_sharpe_ratio(&moments, sr, reference_sr, zero_skewness, normal_kurtosis);
                assert!(
                    almost_equal(actual, expected, 13),
                    "zero_skewness={zero_skewness} normal_kurtosis={normal_kurtosis}: {actual} vs {expected}"
                );
            }
        }
    }

    #[test]
    fn test_unavailable_sharpe_or_moments() {
        let mut moments = RawMomentsKleinKbn::default();
        moments.update(0.01);
        assert!(probabilistic_sharpe_ratio(&moments, f64::NAN, 0.0, false, true).is_nan());
        assert!(probabilistic_sharpe_ratio(&moments, 0.5, 0.0, false, true).is_nan());
        assert!(probabilistic_sharpe_ratio(&moments, 0.5, 0.0, true, false).is_nan());
    }
}
