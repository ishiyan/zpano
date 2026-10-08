//! Streaming winning/losing return averages and counts.

use crate::streaming_kbn::KleinKbnSummator;

/// Streaming winning/loosing return averages and counts.
///
/// Only sums and counts are stored, so `revert()` may remove any
/// previously added return (FIFO rolling windows).
#[derive(Debug, Clone, Default)]
pub struct WinLoss {
    non_zero_sum: KleinKbnSummator,
    win_sum: KleinKbnSummator,
    loss_sum: KleinKbnSummator,
}

impl WinLoss {
    /// Creates an empty accumulator.
    pub fn new() -> Self {
        Self::default()
    }

    /// Resets to the initial empty state.
    pub fn reset(&mut self) {
        self.non_zero_sum.reset();
        self.win_sum.reset();
        self.loss_sum.reset();
    }

    /// Removes a previously added return.
    pub fn revert(&mut self, ret: f64) {
        if ret != 0.0 {
            self.non_zero_sum.revert(ret);
        }
        if ret > 0.0 {
            self.win_sum.revert(ret);
        }
        if ret < 0.0 {
            self.loss_sum.revert(ret);
        }
    }

    /// Adds a return.
    pub fn update(&mut self, ret: f64) {
        if ret != 0.0 {
            self.non_zero_sum.update(ret);
        }
        if ret > 0.0 {
            self.win_sum.update(ret);
        }
        if ret < 0.0 {
            self.loss_sum.update(ret);
        }
    }

    /// Arithmetic mean (average) of non-zero returns (NaN when none).
    pub fn non_zero_returns_mean(&self) -> f64 {
        self.non_zero_sum.mean()
    }

    /// The number of non-zero returns.
    pub fn non_zero_returns_count(&self) -> usize {
        self.non_zero_sum.n()
    }

    /// Sum of winning (positive) returns.
    pub fn winning_returns_sum(&self) -> f64 {
        self.win_sum.value()
    }

    /// Arithmetic mean (average) of winning (positive) returns (NaN when none).
    pub fn winning_returns_mean(&self) -> f64 {
        self.win_sum.mean()
    }

    /// The number of winning (positive) returns.
    pub fn winning_returns_count(&self) -> usize {
        self.win_sum.n()
    }

    /// Sum of losing (negative) returns.
    pub fn losing_returns_sum(&self) -> f64 {
        self.loss_sum.value()
    }

    /// Arithmetic mean (average) of losing (negative) returns (NaN when none).
    pub fn losing_returns_mean(&self) -> f64 {
        self.loss_sum.mean()
    }

    /// The number of losing (negative) returns.
    pub fn losing_returns_count(&self) -> usize {
        self.loss_sum.n()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::performance::core::test_support::{almost_equal, assert_close, choice2, gauss, Rng};

    fn mean(v: &[f64]) -> f64 {
        if v.is_empty() { f64::NAN } else { v.iter().sum::<f64>() / v.len() as f64 }
    }

    fn assert_matches(wl: &WinLoss, returns: &[f64], places: i32, msg: &str) {
        let wins: Vec<f64> = returns.iter().copied().filter(|&r| r > 0.0).collect();
        let losses: Vec<f64> = returns.iter().copied().filter(|&r| r < 0.0).collect();
        let non_zero: Vec<f64> = returns.iter().copied().filter(|&r| r != 0.0).collect();
        assert_eq!(wl.non_zero_returns_count(), non_zero.len(), "{msg} non_zero_returns_count");
        assert_close(wl.non_zero_returns_mean(), mean(&non_zero), places, &format!("{msg} non_zero_returns_mean"));
        assert_eq!(wl.winning_returns_count(), wins.len(), "{msg} winning_returns_count");
        assert_close(wl.winning_returns_sum(), wins.iter().sum(), places, &format!("{msg} winning_returns_sum"));
        assert_close(wl.winning_returns_mean(), mean(&wins), places, &format!("{msg} winning_returns_mean"));
        assert_eq!(wl.losing_returns_count(), losses.len(), "{msg} losing_returns_count");
        assert_close(wl.losing_returns_sum(), losses.iter().sum(), places, &format!("{msg} losing_returns_sum"));
        assert_close(wl.losing_returns_mean(), mean(&losses), places, &format!("{msg} losing_returns_mean"));
    }

    #[test]
    fn test_empty() {
        assert_matches(&WinLoss::new(), &[], 15, "");
    }

    #[test]
    fn test_hand_computed() {
        let mut wl = WinLoss::new();
        for r in [0.02, 0.0, -0.01, 0.04, 0.0, -0.03] {
            wl.update(r);
        }
        assert_eq!(wl.non_zero_returns_count(), 4);
        assert!(almost_equal(wl.non_zero_returns_mean(), 0.02 / 4.0, 16));
        assert_eq!(wl.winning_returns_count(), 2);
        assert!(almost_equal(wl.winning_returns_mean(), 0.03, 16));
        assert_eq!(wl.losing_returns_count(), 2);
        assert!(almost_equal(wl.losing_returns_mean(), -0.02, 16));
    }

    #[test]
    fn test_rolling_window_matches_reference() {
        let mut rng = Rng::new(42);
        let returns: Vec<f64> = (0..120)
            .map(|_| {
                let g = gauss(&mut rng, 0.0, 0.03);
                choice2(&mut rng, 0.0, g)
            })
            .collect();
        let w = 6;
        let mut wl = WinLoss::new();
        for (i, &r) in returns.iter().enumerate() {
            if i >= w {
                wl.revert(returns[i - w]);
            }
            wl.update(r);
            let lo = (i + 1).saturating_sub(w);
            assert_matches(&wl, &returns[lo..=i], 15, &format!("step {i}"));
        }
    }

    #[test]
    fn test_reset() {
        let mut wl = WinLoss::new();
        for r in [0.01, -0.02] {
            wl.update(r);
        }
        wl.reset();
        assert_matches(&wl, &[], 15, "");
    }
}
