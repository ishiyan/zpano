//! Rolling high-water-mark drawdown.

use std::collections::VecDeque;

use crate::streaming_kbn::KleinKbnAccumulator;

/// Rolling high-water-mark drawdown.
///
/// Drawdown at each observation is measured from the high-water mark, the
/// highest equity value reached up to that observation within the current
/// rolling window, including the equity at the start of the window:
///
/// ```text
/// drawdown_t = equity_t / max(equity_start, equity_1, ..., equity_t) - 1
/// ```
///
/// This matches R PerformanceAnalytics `Drawdowns()`, which uses
/// `cummax(c(1, cumprod(1 + R)))`: a first negative return already
/// produces a drawdown. For a rolling window, `equity_start` is the equity
/// just before the first observation in the window, so the result equals a
/// fresh calculation over the window's returns.
///
/// Drawdowns are expressed as decimals and are non-positive.
///
/// Cumulative log-equity is maintained internally so that returns can be
/// accumulated accurately. Running sums of drawdowns and squared drawdowns
/// are maintained using compensated floating-point accumulation.
///
/// When an observation leaves the window, the window's starting equity
/// changes. The drawdowns in the window are recomputed only when this
/// changes their high-water marks; otherwise, the update is O(1).
///
/// A window size of zero means an expanding (unbounded) window.
#[derive(Debug, Clone)]
pub struct HighWaterMarkDrawdown {
    window_size: usize,
    /// Cumulative log-equity at each observation.
    cumlog: VecDeque<f64>,
    /// Drawdown at each observation, as a decimal (<= 0).
    dd: VecDeque<f64>,
    /// Cumulative log return.
    c: KleinKbnAccumulator,
    /// Log-equity just before the first observation in the window.
    base: f64,
    /// Current high-water mark in log-equity space.
    peak: f64,
    /// Running drawdown aggregates.
    sum_dd: KleinKbnAccumulator,
    sum_dd2: KleinKbnAccumulator,
}

impl HighWaterMarkDrawdown {
    /// Creates an empty accumulator; `window_size == 0` means an expanding
    /// (unbounded) window.
    pub fn new(window_size: usize) -> Self {
        Self {
            window_size,
            cumlog: VecDeque::with_capacity(window_size),
            dd: VecDeque::with_capacity(window_size),
            c: KleinKbnAccumulator::new(),
            base: 0.0,
            peak: 0.0,
            sum_dd: KleinKbnAccumulator::new(),
            sum_dd2: KleinKbnAccumulator::new(),
        }
    }

    /// The window size (0 for an expanding window).
    pub fn window_size(&self) -> usize {
        self.window_size
    }

    /// Resets the accumulator to its initial empty state.
    pub fn reset(&mut self) {
        self.cumlog.clear();
        self.dd.clear();
        self.sum_dd.reset();
        self.sum_dd2.reset();
        self.c.reset();
        self.base = 0.0;
        self.peak = 0.0;
    }

    /// Recomputes all drawdowns from the cumulative log-equity values.
    ///
    /// This is required when an observation leaving the rolling window
    /// changes the high-water marks of the remaining observations.
    fn recompute(&mut self) {
        self.dd.clear();
        self.sum_dd.reset();
        self.sum_dd2.reset();
        let mut peak = self.base;
        for &c in &self.cumlog {
            let dd = if c >= peak {
                peak = c;
                0.0
            } else {
                (c - peak).exp_m1()
            };
            self.dd.push_back(dd);
            self.sum_dd.update(dd);
            self.sum_dd2.update(dd * dd);
        }
        self.peak = peak;
    }

    /// Adds a return observation, expressed as a decimal (for example
    /// `0.02` for a 2% return and `-0.015` for a -1.5% return).
    ///
    /// If the rolling window is full, the oldest observation is removed
    /// before the new observation is added.
    ///
    /// Returns true if the rolling window required a drawdown
    /// recomputation, otherwise false.
    pub fn update(&mut self, ret: f64) -> bool {
        let mut old_base: Option<f64> = None;
        if self.window_size > 0 && self.cumlog.len() == self.window_size {
            let old_c = self.cumlog.pop_front().unwrap();
            let old_dd = self.dd.pop_front().unwrap();

            self.sum_dd.revert(old_dd);
            self.sum_dd2.revert(old_dd * old_dd);

            // The evicted observation's equity is the new starting equity.
            // High-water marks of the remaining observations can only
            // change if the old starting equity was above the evicted one.
            if old_c < self.base {
                old_base = Some(self.base);
            }
            self.base = old_c;
        }

        // Global cumulative log-equity.
        self.c.update(ret.ln_1p());
        let c = self.c.value();
        self.cumlog.push_back(c);

        // Peaks of all remaining observations were max(old_base, c0, ..., cj);
        // without old_base they are max(c0, c1, ..., cj). They differ only
        // if the new first observation is also below old_base.
        if let Some(ob) = old_base
            && self.cumlog[0] < ob
        {
            self.recompute();
            return true;
        }

        let dd = if c >= self.peak {
            self.peak = c;
            0.0
        } else {
            (c - self.peak).exp_m1()
        };

        self.dd.push_back(dd);
        self.sum_dd.update(dd);
        self.sum_dd2.update(dd * dd);
        false
    }

    /// Drawdowns for observations currently in the window (oldest first).
    pub fn drawdowns(&self) -> &VecDeque<f64> {
        &self.dd
    }

    /// The most recent drawdown in the current window (NaN when empty).
    pub fn drawdown(&self) -> f64 {
        self.dd.back().copied().unwrap_or(f64::NAN)
    }

    /// Maximum drawdown in the current window (NaN when empty).
    ///
    /// Drawdowns are non-positive, the largest loss is the minimum
    /// drawdown value.
    pub fn maximum_drawdown(&self) -> f64 {
        let mut it = self.dd.iter().copied();
        match it.next() {
            None => f64::NAN,
            // Python min(): keeps the first of equal values.
            Some(first) => it.fold(first, |m, x| if x < m { x } else { m }),
        }
    }

    /// Arithmetic mean of drawdowns in the current window (NaN when empty).
    pub fn drawdowns_mean(&self) -> f64 {
        let n = self.dd.len();
        if n > 0 { self.sum_dd.value() / n as f64 } else { f64::NAN }
    }

    /// Mean squared drawdown in the current window (NaN when empty).
    pub fn drawdowns_squared_mean(&self) -> f64 {
        let n = self.dd.len();
        if n > 0 { self.sum_dd2.value() / n as f64 } else { f64::NAN }
    }

    /// Number of observations in the current window.
    pub fn drawdowns_count(&self) -> usize {
        self.dd.len()
    }
}

impl Default for HighWaterMarkDrawdown {
    /// An expanding (unbounded) window.
    fn default() -> Self {
        Self::new(0)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::performance::core::test_support::{almost_equal, gauss, Rng};

    /// Independent reference implementation of chronological
    /// high-water-mark drawdowns (R `Drawdowns()`): the high-water mark
    /// starts at the initial equity 1.
    fn expected_drawdowns(returns: &[f64]) -> Vec<f64> {
        let mut equity = 1.0;
        let mut peak = 1.0;
        let mut result = Vec::new();
        for &ret in returns {
            equity *= 1.0 + ret;
            let dd = if equity >= peak {
                peak = equity;
                0.0
            } else {
                equity / peak - 1.0
            };
            result.push(dd);
        }
        result
    }

    /// Drawdowns of the rolling window ending at index i.
    fn expected_rolling_drawdowns(returns: &[f64], window_size: usize, i: usize) -> Vec<f64> {
        let lo = if window_size == 0 { 0 } else { (i + 1).saturating_sub(window_size) };
        expected_drawdowns(&returns[lo..=i])
    }

    fn assert_state_p(acc: &HighWaterMarkDrawdown, expected: &[f64], places: i32) {
        let actual = acc.drawdowns();
        assert_eq!(actual.len(), expected.len());
        for (&a, &e) in actual.iter().zip(expected.iter()) {
            assert!(almost_equal(a, e, places), "{a} vs {e}");
        }
        assert_eq!(acc.drawdowns_count(), expected.len());
        if !expected.is_empty() {
            assert!(almost_equal(acc.drawdown(), expected[expected.len() - 1], places));
            let min = expected.iter().copied().fold(f64::INFINITY, f64::min);
            assert!(almost_equal(acc.maximum_drawdown(), min, places));
            let n = expected.len() as f64;
            let expected_mean = expected.iter().sum::<f64>() / n;
            let expected_squared_mean = expected.iter().map(|x| x * x).sum::<f64>() / n;
            assert!(almost_equal(acc.drawdowns_mean(), expected_mean, places));
            assert!(almost_equal(acc.drawdowns_squared_mean(), expected_squared_mean, places));
        } else {
            assert!(acc.drawdown().is_nan());
            assert!(acc.maximum_drawdown().is_nan());
            assert!(acc.drawdowns_mean().is_nan());
            assert!(acc.drawdowns_squared_mean().is_nan());
        }
    }

    fn assert_state(acc: &HighWaterMarkDrawdown, expected: &[f64]) {
        assert_state_p(acc, expected, 14);
    }

    fn feed(mut acc: HighWaterMarkDrawdown, returns: &[f64]) -> HighWaterMarkDrawdown {
        for &r in returns {
            acc.update(r);
        }
        acc
    }

    // Expanding-window tests

    #[test]
    fn test_expanding_empty() {
        assert_state(&HighWaterMarkDrawdown::new(0), &[]);
    }

    #[test]
    fn test_expanding_all_positive_returns() {
        let acc = feed(HighWaterMarkDrawdown::new(0), &[0.10, 0.05, 0.20]);
        assert_state(&acc, &[0.0, 0.0, 0.0]);
    }

    #[test]
    fn test_expanding_first_negative_return() {
        let acc = feed(HighWaterMarkDrawdown::new(0), &[-0.05, -0.02, 0.10]);
        assert_state(&acc, &[-0.05, -0.069, 0.0]);
    }

    #[test]
    fn test_expanding_simple_drawdown_and_recovery() {
        let acc = feed(HighWaterMarkDrawdown::new(0), &[0.10, -0.05, 0.10]);
        assert_state(&acc, &[0.0, -0.05, 0.0]);
    }

    #[test]
    fn test_expanding_compounded_drawdown() {
        let acc = feed(HighWaterMarkDrawdown::new(0), &[0.10, -0.10, -0.10]);
        assert_state(&acc, &[0.0, -0.10, -0.19]);
    }

    #[test]
    fn test_expanding_new_high_water_mark_resets_drawdown() {
        let acc = feed(HighWaterMarkDrawdown::new(0), &[0.10, -0.05, 0.06, -0.02]);
        assert_state(&acc, &[0.0, -0.05, 0.0, -0.02]);
    }

    #[test]
    fn test_expanding_reset() {
        let mut acc = feed(HighWaterMarkDrawdown::new(0), &[0.10, -0.05, -0.02]);
        assert!(acc.drawdowns_count() > 0);
        acc.reset();
        assert_state(&acc, &[]);
        // The accumulator can be reused, starting from equity 1.0 again.
        acc.update(-0.05);
        assert_state(&acc, &[-0.05]);
    }

    #[test]
    fn test_expanding_matches_reference() {
        let mut rng = Rng::new(42);
        let returns: Vec<f64> = (0..200).map(|_| gauss(&mut rng, 0.0, 0.03)).collect();
        let acc = feed(HighWaterMarkDrawdown::new(0), &returns);
        assert_state_p(&acc, &expected_drawdowns(&returns), 13);
    }

    #[test]
    fn test_zero_size_means_expanding() {
        let returns = [0.10, -0.05, -0.02, 0.05];
        let acc = feed(HighWaterMarkDrawdown::new(0), &returns);
        assert_state(&acc, &expected_drawdowns(&returns));
    }

    #[test]
    fn test_negative_size_means_expanding() {
        // Python normalizes non-positive window sizes to zero; a negative
        // size is not representable with `usize`, so the default
        // (expanding) constructor is exercised instead.
        let returns = [0.10, -0.05, -0.02];
        let acc = feed(HighWaterMarkDrawdown::default(), &returns);
        assert_eq!(acc.window_size(), 0);
        assert_state(&acc, &expected_drawdowns(&returns));
    }

    // Rolling-window tests: the window equals a fresh calculation over its
    // returns, starting from the equity just before the window.

    #[test]
    fn test_rolling_window_peak_eviction() {
        let mut acc = feed(HighWaterMarkDrawdown::new(3), &[0.10, -0.05, -0.02]);
        assert_state(&acc, &[0.0, -0.05, -0.069]);
        acc.update(0.03);
        assert_state(&acc, &[-0.05, -0.069, -0.04107]);
    }

    #[test]
    fn test_rolling_window_evicted_peak_followed_by_new_peak() {
        let mut acc = feed(HighWaterMarkDrawdown::new(3), &[0.05, -0.02, 0.10]);
        assert_state(&acc, &[0.0, -0.02, 0.0]);
        acc.update(-0.03);
        assert_state(&acc, &[-0.02, 0.0, -0.03]);
    }

    #[test]
    fn test_rolling_window_peak_eviction_recomputes_drawdowns() {
        let mut acc = feed(HighWaterMarkDrawdown::new(3), &[0.10, -0.05, -0.05]);
        assert_state(&acc, &[0.0, -0.05, -0.0975]);
        acc.update(0.01);
        assert_state(&acc, &[-0.05, -0.0975, -0.088475]);
    }

    #[test]
    fn test_rolling_window_all_negative_returns() {
        let returns = [-0.01, -0.02, -0.03, -0.04];
        let mut acc = HighWaterMarkDrawdown::new(3);
        for (i, &ret) in returns.iter().enumerate() {
            acc.update(ret);
            assert_state(&acc, &expected_rolling_drawdowns(&returns, 3, i));
        }
    }

    #[test]
    fn test_rolling_window_size_one() {
        let returns = [0.10, -0.05, -0.02, 0.03, -0.04];
        let mut acc = HighWaterMarkDrawdown::new(1);
        for &ret in &returns {
            acc.update(ret);
            assert_state(&acc, &[ret.min(0.0)]);
        }
    }

    #[test]
    fn test_rolling_window_matches_fresh_calculation() {
        let mut rng = Rng::new(7);
        for window_size in [2usize, 3, 4, 7, 20] {
            for _ in 0..10 {
                let returns: Vec<f64> = (0..80).map(|_| gauss(&mut rng, 0.0, 0.03)).collect();
                let mut acc = HighWaterMarkDrawdown::new(window_size);
                for i in 0..returns.len() {
                    acc.update(returns[i]);
                    assert_state_p(&acc, &expected_rolling_drawdowns(&returns, window_size, i), 13);
                }
            }
        }
    }

    #[test]
    fn test_rolling_window_recompute_flag() {
        let mut acc = HighWaterMarkDrawdown::new(2);
        assert!(!acc.update(0.10));
        assert!(!acc.update(-0.05));
        // Evicts +10%: the remaining peaks don't change.
        assert!(!acc.update(-0.02));
        // Evicts -5% and the new first observation (-2%) is also below
        // the old starting equity: recompute.
        assert!(acc.update(0.01));
        assert_state(&acc, &[-0.02, -0.0102]);
        // Evicts -2%; the new first observation (equity 1.0343) is still
        // below the old starting equity (1.045): recompute.
        assert!(acc.update(0.03));
        assert_state(&acc, &[0.0, 0.0]);
        // Evicts +1%: its equity is above the window's starting equity,
        // so the remaining peaks don't change.
        assert!(!acc.update(-0.01));
        assert_state(&acc, &[0.0, -0.01]);
    }
}
