//! Streaming lower/higher partial moments about a threshold.

use crate::streaming_kbn::RawMomentsKleinKbn;

/// Streaming partial moments.
///
/// For each return r and threshold τ (target return or minimum
/// acceptable return, MAR):
///
/// * lower partial moments use max(τ - r, 0) over all returns;
/// * higher partial moments use max(r - τ, 0) over all returns;
/// * lower excess moments use τ - r over returns strictly below τ;
/// * upper excess moments use r - τ over returns strictly above τ.
///
/// The k-th moments are raw means (Σxᵏ / n). Only power sums are stored,
/// so `revert()` may remove any previously added return (FIFO rolling
/// windows).
#[derive(Debug, Clone)]
pub struct PartialMoments {
    /// Target return or minimum acceptable return (MAR), in the same
    /// periodicity as the returns.
    pub threshold: f64,
    count_total: usize,
    upper_excess_kbn: RawMomentsKleinKbn,
    lower_excess_kbn: RawMomentsKleinKbn,
    lpm_kbn: RawMomentsKleinKbn,
    hpm_kbn: RawMomentsKleinKbn,
}

/// ddof=1, bias=True, fisher=True: matches scipy's default kurtosis.
fn moments() -> RawMomentsKleinKbn {
    RawMomentsKleinKbn::new(1, true, true)
}

impl PartialMoments {
    /// Streaming low/high partial moments.
    ///
    /// `threshold`: target return or minimum acceptable return (MAR) in
    /// the same periodicity as the returns.
    pub fn new(threshold: f64) -> Self {
        Self {
            threshold,
            count_total: 0,
            upper_excess_kbn: moments(),
            lower_excess_kbn: moments(),
            lpm_kbn: moments(),
            hpm_kbn: moments(),
        }
    }

    /// Resets to the initial empty state (keeps the threshold).
    pub fn reset(&mut self) {
        self.count_total = 0;
        self.upper_excess_kbn.reset();
        self.lower_excess_kbn.reset();
        self.lpm_kbn.reset();
        self.hpm_kbn.reset();
    }

    /// Removes a previously added return.
    pub fn revert(&mut self, ret: f64) {
        self.count_total -= 1;
        // Lower partial moments for the raw returns less target return
        let mut pm = self.threshold - ret;
        if pm < 0.0 {
            self.upper_excess_kbn.revert(-pm);
            pm = 0.0;
        }
        self.lpm_kbn.revert(pm);

        // Higher partial moments for the raw returns less required return
        let mut pm = ret - self.threshold;
        if pm < 0.0 {
            self.lower_excess_kbn.revert(-pm);
            pm = 0.0;
        }
        self.hpm_kbn.revert(pm);
    }

    /// Adds a return.
    pub fn update(&mut self, ret: f64) {
        self.count_total += 1;
        // Lower partial moments for the raw returns less target return
        let mut pm = self.threshold - ret;
        if pm < 0.0 {
            self.upper_excess_kbn.update(-pm);
            pm = 0.0;
        }
        self.lpm_kbn.update(pm);

        // Higher partial moments for the raw returns less required return
        let mut pm = ret - self.threshold;
        if pm < 0.0 {
            self.lower_excess_kbn.update(-pm);
            pm = 0.0;
        }
        self.hpm_kbn.update(pm);
    }

    /// First lower partial moment, mean of max(τ - r, 0).
    pub fn lower_partial_moment_1(&self) -> f64 {
        self.lpm_kbn.x1()
    }

    /// Second lower partial moment, mean of max(τ - r, 0)².
    pub fn lower_partial_moment_2(&self) -> f64 {
        self.lpm_kbn.x2()
    }

    /// Third lower partial moment, mean of max(τ - r, 0)³.
    pub fn lower_partial_moment_3(&self) -> f64 {
        self.lpm_kbn.x3()
    }

    /// Fourth lower partial moment, mean of max(τ - r, 0)⁴.
    pub fn lower_partial_moment_4(&self) -> f64 {
        self.lpm_kbn.x4()
    }

    /// First higher partial moment, mean of max(r - τ, 0).
    pub fn higher_partial_moment_1(&self) -> f64 {
        self.hpm_kbn.x1()
    }

    /// Second higher partial moment, mean of max(r - τ, 0)².
    pub fn higher_partial_moment_2(&self) -> f64 {
        self.hpm_kbn.x2()
    }

    /// Third higher partial moment, mean of max(r - τ, 0)³.
    pub fn higher_partial_moment_3(&self) -> f64 {
        self.hpm_kbn.x3()
    }

    /// Fourth higher partial moment, mean of max(r - τ, 0)⁴.
    pub fn higher_partial_moment_4(&self) -> f64 {
        self.hpm_kbn.x4()
    }

    /// Proportion of returns below threshold (NaN when empty).
    pub fn downside_frequency(&self) -> f64 {
        let total = self.count_total;
        if total == 0 {
            return f64::NAN;
        }
        self.lower_excess_kbn.n() as f64 / total as f64
    }

    /// Proportion of returns above threshold (NaN when empty).
    pub fn upside_frequency(&self) -> f64 {
        let total = self.count_total;
        if total == 0 {
            return f64::NAN;
        }
        self.upper_excess_kbn.n() as f64 / total as f64
    }

    /// Mean of lower partial moments (also called shortfall).
    pub fn downside_potential(&self) -> f64 {
        self.lpm_kbn.mean()
    }

    /// The number of returns.
    pub fn total_count(&self) -> usize {
        self.count_total
    }

    /// The number of returns strictly above the threshold.
    pub fn upper_excess_count(&self) -> usize {
        self.upper_excess_kbn.n()
    }

    /// The number of returns strictly below the threshold.
    pub fn lower_excess_count(&self) -> usize {
        self.lower_excess_kbn.n()
    }

    /// Mean of (r - τ) over returns above the threshold.
    pub fn upper_excess_moment_1(&self) -> f64 {
        self.upper_excess_kbn.x1()
    }

    /// Sum of (r - τ) over returns above the threshold.
    pub fn upper_excess_moment_1_sum(&self) -> f64 {
        self.upper_excess_kbn.x1_sum()
    }

    /// Mean of (r - τ)² over returns above the threshold.
    pub fn upper_excess_moment_2(&self) -> f64 {
        self.upper_excess_kbn.x2()
    }

    /// Sum of (r - τ)² over returns above the threshold.
    pub fn upper_excess_moment_2_sum(&self) -> f64 {
        self.upper_excess_kbn.x2_sum()
    }

    /// Mean of (r - τ)³ over returns above the threshold.
    pub fn upper_excess_moment_3(&self) -> f64 {
        self.upper_excess_kbn.x3()
    }

    /// Mean of (r - τ)⁴ over returns above the threshold.
    pub fn upper_excess_moment_4(&self) -> f64 {
        self.upper_excess_kbn.x4()
    }

    /// Mean of (τ - r) over returns below the threshold.
    pub fn lower_excess_moment_1(&self) -> f64 {
        self.lower_excess_kbn.x1()
    }

    /// Mean of (τ - r)² over returns below the threshold.
    pub fn lower_excess_moment_2(&self) -> f64 {
        self.lower_excess_kbn.x2()
    }

    /// Sum of (τ - r)² over returns below the threshold.
    pub fn lower_excess_moment_2_sum(&self) -> f64 {
        self.lower_excess_kbn.x2_sum()
    }

    /// Mean of (τ - r)³ over returns below the threshold.
    pub fn lower_excess_moment_3(&self) -> f64 {
        self.lower_excess_kbn.x3()
    }

    /// Mean of (τ - r)⁴ over returns below the threshold.
    pub fn lower_excess_moment_4(&self) -> f64 {
        self.lower_excess_kbn.x4()
    }
}

impl Default for PartialMoments {
    /// Threshold 0.
    fn default() -> Self {
        Self::new(0.0)
    }
}

#[cfg(test)]
pub(super) mod tests {
    use super::*;
    use crate::performance::core::test_support::{almost_equal, assert_close, gauss, Rng};

    pub fn mean(values: &[f64]) -> f64 {
        if values.is_empty() { f64::NAN } else { values.iter().sum::<f64>() / values.len() as f64 }
    }

    /// Mix in values exactly equal to the threshold and to zero.
    pub fn random_returns(seed: u64, n: usize, threshold: f64) -> Vec<f64> {
        let mut rng = Rng::new(seed);
        (0..n)
            .map(|_| {
                let g1 = gauss(&mut rng, 0.0, 0.03);
                let g2 = gauss(&mut rng, 0.0, 0.03);
                rng.choice(&[threshold, 0.0, g1, g2])
            })
            .collect()
    }

    /// Compares against naive partial moments about the threshold.
    fn assert_matches(pm: &PartialMoments, returns: &[f64], threshold: f64, places: i32, msg: &str) {
        let lower: Vec<f64> = returns.iter().filter(|&&r| r < threshold).map(|&r| threshold - r).collect();
        let upper: Vec<f64> = returns.iter().filter(|&&r| r > threshold).map(|&r| r - threshold).collect();
        let n = returns.len();
        let close = |a: f64, e: f64, name: &str| assert_close(a, e, places, &format!("{msg} {name}"));

        assert_eq!(pm.total_count(), n, "{msg} total_count");
        assert_eq!(pm.lower_excess_count(), lower.len(), "{msg} lower_excess_count");
        assert_eq!(pm.upper_excess_count(), upper.len(), "{msg} upper_excess_count");
        let freq = |k: usize| if n > 0 { k as f64 / n as f64 } else { f64::NAN };
        close(pm.downside_frequency(), freq(lower.len()), "downside_frequency");
        close(pm.upside_frequency(), freq(upper.len()), "upside_frequency");
        let lpm = |k: i32| -> Vec<f64> { returns.iter().map(|&r| (threshold - r).max(0.0).powi(k)).collect() };
        let hpm = |k: i32| -> Vec<f64> { returns.iter().map(|&r| (r - threshold).max(0.0).powi(k)).collect() };
        let pw = |v: &[f64], k: i32| -> Vec<f64> { v.iter().map(|x| x.powi(k)).collect() };
        close(pm.downside_potential(), mean(&lpm(1)), "downside_potential");
        close(pm.upper_excess_moment_1_sum(), upper.iter().sum(), "upper_excess_moment_1_sum");
        close(pm.upper_excess_moment_2_sum(), pw(&upper, 2).iter().sum(), "upper_excess_moment_2_sum");
        close(pm.lower_excess_moment_2_sum(), pw(&lower, 2).iter().sum(), "lower_excess_moment_2_sum");

        let lpms = [
            pm.lower_partial_moment_1(),
            pm.lower_partial_moment_2(),
            pm.lower_partial_moment_3(),
            pm.lower_partial_moment_4(),
        ];
        let hpms = [
            pm.higher_partial_moment_1(),
            pm.higher_partial_moment_2(),
            pm.higher_partial_moment_3(),
            pm.higher_partial_moment_4(),
        ];
        let uems = [
            pm.upper_excess_moment_1(),
            pm.upper_excess_moment_2(),
            pm.upper_excess_moment_3(),
            pm.upper_excess_moment_4(),
        ];
        let lems = [
            pm.lower_excess_moment_1(),
            pm.lower_excess_moment_2(),
            pm.lower_excess_moment_3(),
            pm.lower_excess_moment_4(),
        ];
        for k in 1..=4 {
            let i = (k - 1) as usize;
            close(lpms[i], mean(&lpm(k)), &format!("lower_partial_moment_{k}"));
            close(hpms[i], mean(&hpm(k)), &format!("higher_partial_moment_{k}"));
            close(uems[i], mean(&pw(&upper, k)), &format!("upper_excess_moment_{k}"));
            close(lems[i], mean(&pw(&lower, k)), &format!("lower_excess_moment_{k}"));
        }
    }

    #[test]
    fn test_empty() {
        let pm = PartialMoments::new(0.01);
        assert_eq!(pm.total_count(), 0);
        assert!(pm.downside_frequency().is_nan());
        assert!(pm.upside_frequency().is_nan());
        assert!(pm.lower_partial_moment_2().is_nan());
    }

    #[test]
    fn test_hand_computed() {
        let mut pm = PartialMoments::new(0.01);
        for r in [0.03, -0.01, 0.01, 0.00] {
            pm.update(r);
        }
        // Shortfalls below 1%: 0.02, 0.01; excesses above: 0.02.
        assert!(almost_equal(pm.lower_partial_moment_1(), (0.02 + 0.01) / 4.0, 16));
        assert!(almost_equal(pm.lower_partial_moment_2(), (0.0004 + 0.0001) / 4.0, 16));
        assert!(almost_equal(pm.higher_partial_moment_1(), 0.02 / 4.0, 16));
        assert_eq!(pm.lower_excess_count(), 2);
        assert_eq!(pm.upper_excess_count(), 1);
        assert_eq!(pm.downside_frequency(), 0.5);
        assert_eq!(pm.upside_frequency(), 0.25);
    }

    #[test]
    fn test_matches_reference() {
        for threshold in [0.0, 0.005] {
            let returns = random_returns(42, 200, threshold);
            let mut pm = PartialMoments::new(threshold);
            for &r in &returns {
                pm.update(r);
            }
            assert_matches(&pm, &returns, threshold, 14, &format!("threshold {threshold}"));
        }
    }

    #[test]
    fn test_rolling_window_matches_reference() {
        let threshold = 0.005;
        let returns = random_returns(7, 120, threshold);
        let w = 8;
        let mut pm = PartialMoments::new(threshold);
        for (i, &r) in returns.iter().enumerate() {
            if i >= w {
                pm.revert(returns[i - w]);
            }
            pm.update(r);
            let window = &returns[(i + 1).saturating_sub(w)..=i];
            assert_matches(&pm, window, threshold, 13, &format!("step {i}"));
        }
    }

    #[test]
    fn test_reset() {
        let mut pm = PartialMoments::new(0.0);
        for r in [0.01, -0.02] {
            pm.update(r);
        }
        pm.reset();
        assert_eq!(pm.total_count(), 0);
        assert_eq!(pm.lower_excess_count(), 0);
        assert_eq!(pm.upper_excess_count(), 0);
    }
}
