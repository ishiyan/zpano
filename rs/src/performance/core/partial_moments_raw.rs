//! Streaming raw partial moment sums about zero.

use crate::streaming_kbn::KleinKbnAccumulator;

/// Streaming raw partial moments.
///
/// Sums (not means) about a zero threshold: Σmax(-r, 0), Σmax(r, 0),
/// and the counts and sums of strictly negative and strictly positive
/// returns. Only sums are stored, so `revert()` may remove any previously
/// added return (FIFO rolling windows).
#[derive(Debug, Clone, Default)]
pub struct RawPartialMoments {
    count: usize,
    count_pos: usize,
    count_neg: usize,
    lpm_kbn: KleinKbnAccumulator,
    hpm_kbn: KleinKbnAccumulator,
    pos_kbn: KleinKbnAccumulator,
    neg_kbn: KleinKbnAccumulator,
}

impl RawPartialMoments {
    /// Streaming raw low/high partial moments.
    pub fn new() -> Self {
        Self::default()
    }

    /// Resets to the initial empty state.
    pub fn reset(&mut self) {
        self.count = 0;
        self.count_pos = 0;
        self.count_neg = 0;
        self.lpm_kbn.reset();
        self.hpm_kbn.reset();
        self.pos_kbn.reset();
        self.neg_kbn.reset();
    }

    /// Removes a previously added return.
    pub fn revert(&mut self, ret: f64) {
        self.count -= 1;
        // Lower partial moment
        let mut pm = -ret;
        if pm < 0.0 {
            pm = 0.0;
        }
        self.lpm_kbn.revert(pm);

        // Higher partial moment
        let mut pm = ret;
        if pm < 0.0 {
            pm = 0.0;
        }
        self.hpm_kbn.revert(pm);

        if ret > 0.0 {
            self.count_pos -= 1;
            self.pos_kbn.revert(ret);
        } else if ret < 0.0 {
            self.count_neg -= 1;
            self.neg_kbn.revert(ret);
        }
    }

    /// Adds a return.
    pub fn update(&mut self, ret: f64) {
        self.count += 1;
        // Lower partial moment
        let mut pm = -ret;
        if pm < 0.0 {
            pm = 0.0;
        }
        self.lpm_kbn.update(pm);

        // Higher partial moment
        let mut pm = ret;
        if pm < 0.0 {
            pm = 0.0;
        }
        self.hpm_kbn.update(pm);

        if ret > 0.0 {
            self.count_pos += 1;
            self.pos_kbn.update(ret);
        } else if ret < 0.0 {
            self.count_neg += 1;
            self.neg_kbn.update(ret);
        }
    }

    /// The number of returns.
    pub fn count(&self) -> usize {
        self.count
    }

    /// Σmax(-r, 0).
    pub fn lower_partial_moment_1(&self) -> f64 {
        self.lpm_kbn.value()
    }

    /// Σmax(r, 0).
    pub fn higher_partial_moment_1(&self) -> f64 {
        self.hpm_kbn.value()
    }

    /// The number of strictly negative returns.
    pub fn count_negative(&self) -> usize {
        self.count_neg
    }

    /// Sum of strictly negative returns.
    pub fn sum_negative(&self) -> f64 {
        self.neg_kbn.value()
    }

    /// The number of strictly positive returns.
    pub fn count_positive(&self) -> usize {
        self.count_pos
    }

    /// Sum of strictly positive returns.
    pub fn sum_positive(&self) -> f64 {
        self.pos_kbn.value()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::performance::core::partial_moments::tests::random_returns;
    use crate::performance::core::test_support::assert_close;

    /// Compares against naive raw partial moments (sums, threshold 0).
    fn assert_matches(pm: &RawPartialMoments, returns: &[f64], places: i32, msg: &str) {
        let close = |a: f64, e: f64, name: &str| assert_close(a, e, places, &format!("{msg} {name}"));
        assert_eq!(pm.count(), returns.len(), "{msg} count");
        close(pm.lower_partial_moment_1(), returns.iter().map(|&r| (-r).max(0.0)).sum(), "lower_partial_moment_1");
        close(pm.higher_partial_moment_1(), returns.iter().map(|&r| r.max(0.0)).sum(), "higher_partial_moment_1");
        assert_eq!(pm.count_negative(), returns.iter().filter(|&&r| r < 0.0).count(), "{msg} count_negative");
        close(pm.sum_negative(), returns.iter().filter(|&&r| r < 0.0).sum(), "sum_negative");
        assert_eq!(pm.count_positive(), returns.iter().filter(|&&r| r > 0.0).count(), "{msg} count_positive");
        close(pm.sum_positive(), returns.iter().filter(|&&r| r > 0.0).sum(), "sum_positive");
    }

    #[test]
    fn test_empty() {
        assert_matches(&RawPartialMoments::new(), &[], 14, "");
    }

    #[test]
    fn test_matches_reference() {
        let returns = random_returns(42, 200, 0.0);
        let mut pm = RawPartialMoments::new();
        for &r in &returns {
            pm.update(r);
        }
        assert_matches(&pm, &returns, 14, "");
    }

    #[test]
    fn test_rolling_window_matches_reference() {
        let returns = random_returns(7, 120, 0.0);
        let w = 8;
        let mut pm = RawPartialMoments::new();
        for (i, &r) in returns.iter().enumerate() {
            if i >= w {
                pm.revert(returns[i - w]);
            }
            pm.update(r);
            assert_matches(&pm, &returns[(i + 1).saturating_sub(w)..=i], 14, &format!("step {i}"));
        }
    }

    #[test]
    fn test_reset() {
        let mut pm = RawPartialMoments::new();
        for r in [0.01, -0.02] {
            pm.update(r);
        }
        pm.reset();
        assert_matches(&pm, &[], 14, "");
    }
}
