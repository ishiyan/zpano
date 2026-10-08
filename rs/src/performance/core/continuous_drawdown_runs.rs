//! Streaming continuous drawdown runs for Burke-type measures.

use std::collections::VecDeque;

use crate::streaming_kbn::KleinKbnAccumulator;

/// Converts a compounded log return into a percentage drawdown,
/// `expm1(logsum) * 100`.
pub fn dd_percent(logsum: f64) -> f64 {
    logsum.exp_m1() * 100.0
}

/// Python `x ** 2` (C `pow(x, 2.0)`).
#[inline]
fn sq(x: f64) -> f64 {
    x.powf(2.0)
}

/// Streaming 'continuous' drawdown runs for Burke-type measures.
///
/// A continuous drawdown is the compounded loss over a maximal run of
/// consecutive negative returns. Following PerformanceAnalytics
/// `BurkeRatio`, returns are compounded as if they were percentages:
///
/// ```text
/// DD = (prod(1 + r_i * 0.01) - 1) * 100
/// ```
///
/// For decimal returns this is close to the sum of the run's returns,
/// not their compounded return; the quirk is kept to match R.
///
/// The Burke denominator is `sqrt(sum(DD_j^2))`, where the sum is taken
/// over all continuous losing runs in the current window.
///
/// This type is a pure accumulator: the caller owns the rolling window
/// and feeds evicted values to `revert()` and new values to `update()`.
/// Within one step, call `revert(old)` BEFORE `update(new)` so run
/// adjacency stays correct.
///
/// The complexity is O(1) per call.
#[derive(Debug, Clone, Default)]
pub struct ContinuousDrawdownRuns {
    /// Each run: (logsum, count).
    runs: VecDeque<(f64, usize)>,
    /// Sum of squared continuous drawdowns.
    sum_sq: KleinKbnAccumulator,
    /// Whether the most recent return is negative.
    last_was_negative: bool,
}

impl ContinuousDrawdownRuns {
    /// Creates an empty accumulator.
    pub fn new() -> Self {
        Self::default()
    }

    /// Resets to the initial empty state.
    pub fn reset(&mut self) {
        self.runs.clear();
        self.sum_sq.reset();
        self.last_was_negative = false;
    }

    /// Removes the oldest return from the left edge of the window.
    ///
    /// # Panics
    ///
    /// Panics if `old_ret < 0` and there are no runs (Python: `IndexError`).
    pub fn revert(&mut self, old_ret: f64) {
        if old_ret < 0.0 {
            // oldest negative is the front of the left-most run, shrink it
            let run = &mut self.runs[0];
            self.sum_sq.revert(sq(dd_percent(run.0)));
            run.0 -= (old_ret * 0.01).ln_1p();
            run.1 -= 1;
            if run.1 == 0 {
                self.runs.pop_front(); // run fully evicted
            } else {
                let v = sq(dd_percent(run.0));
                self.sum_sq.update(v);
            }
        }
        // old_ret >= 0 is a separator, nothing to update
    }

    /// Adds a new (most-recent) return at the right edge of the window.
    pub fn update(&mut self, ret: f64) {
        if ret < 0.0 {
            let logr = (ret * 0.01).ln_1p();
            if self.last_was_negative && !self.runs.is_empty() {
                // Extend the currently-open (right-most) run.
                let run = self.runs.back_mut().unwrap();
                self.sum_sq.revert(sq(dd_percent(run.0)));
                run.0 += logr;
                run.1 += 1;
                self.sum_sq.update(sq(dd_percent(run.0)));
            } else {
                // Start a new run.
                self.runs.push_back((logr, 1));
                self.sum_sq.update(sq(dd_percent(logr)));
            }
            self.last_was_negative = true;
        } else {
            // Non-negative return closes any open run (already counted) — a separator.
            self.last_was_negative = false;
        }
    }

    /// Continuous drawdowns (negative percentages), one value for each losing run.
    pub fn drawdowns(&self) -> Vec<f64> {
        self.runs.iter().map(|&(logsum, _)| dd_percent(logsum)).collect()
    }

    /// Sum of squared continuous drawdowns (clamped at zero).
    pub fn sum_drawdowns_squared(&self) -> f64 {
        py_max0(self.sum_sq.value())
    }

    /// Square root of the sum of squared continuous drawdowns.
    ///
    /// This is the denominator used by the Burke ratio.
    pub fn sqrt_sum_drawdowns_squared(&self) -> f64 {
        py_max0(self.sum_sq.value()).sqrt()
    }

    /// Number of continuous losing runs in the current window.
    pub fn run_count(&self) -> usize {
        self.runs.len()
    }
}

/// Python `max(v, 0.0)`.
#[inline]
fn py_max0(v: f64) -> f64 {
    if 0.0 > v { 0.0 } else { v }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::performance::core::test_support::{almost_equal, choice2, gauss, prod, Rng};

    /// Compounded drawdown of one continuous losing run.
    fn dd(returns: &[f64]) -> f64 {
        let compounded = prod(returns.iter().map(|r| 1.0 + r * 0.01));
        (compounded - 1.0) * 100.0
    }

    fn assert_state(acc: &ContinuousDrawdownRuns, expected: &[f64], expected_run_count: Option<usize>, places: i32, prefix: &str) {
        let actual = acc.drawdowns();
        assert_eq!(actual.len(), expected.len(), "{prefix}");
        for (i, (&a, &e)) in actual.iter().zip(expected.iter()).enumerate() {
            assert!(almost_equal(a, e, places), "{prefix} step {i} (actual {a}, expected {e})");
        }
        if let Some(n) = expected_run_count {
            assert_eq!(acc.run_count(), n, "{prefix}");
        }
        let expected_sum_sq: f64 = expected.iter().map(|x| x * x).sum();
        assert!(almost_equal(acc.sum_drawdowns_squared(), expected_sum_sq, places), "{prefix}");
        assert!(almost_equal(acc.sqrt_sum_drawdowns_squared(), expected_sum_sq.sqrt(), places), "{prefix}");
    }

    fn st(acc: &ContinuousDrawdownRuns, expected: &[f64], n: usize) {
        assert_state(acc, expected, Some(n), 12, "");
    }

    fn feed(acc: &mut ContinuousDrawdownRuns, returns: &[f64]) {
        for &r in returns {
            acc.update(r);
        }
    }

    // Expanding-window tests

    #[test]
    fn test_expanding_single_losing_run() {
        let mut acc = ContinuousDrawdownRuns::new();
        acc.update(-1.0);
        st(&acc, &[dd(&[-1.0])], 1);
        acc.update(-2.0);
        st(&acc, &[dd(&[-1.0, -2.0])], 1);
        acc.update(-3.0);
        st(&acc, &[dd(&[-1.0, -2.0, -3.0])], 1);
    }

    #[test]
    fn test_expanding_multiple_losing_runs() {
        let mut acc = ContinuousDrawdownRuns::new();
        feed(&mut acc, &[-1.0, -2.0, 1.0, -3.0, -4.0, 2.0, -5.0]);
        st(&acc, &[dd(&[-1.0, -2.0]), dd(&[-3.0, -4.0]), dd(&[-5.0])], 3);
    }

    #[test]
    fn test_expanding_non_negative_returns_are_separators() {
        let mut acc = ContinuousDrawdownRuns::new();
        feed(&mut acc, &[-2.0, 0.0, -3.0, 1.0, -4.0]);
        st(&acc, &[dd(&[-2.0]), dd(&[-3.0]), dd(&[-4.0])], 3);
    }

    #[test]
    fn test_expanding_positive_return_does_not_create_drawdown() {
        let mut acc = ContinuousDrawdownRuns::new();
        feed(&mut acc, &[2.0, 3.0, 0.0, 5.0]);
        st(&acc, &[], 0);
    }

    #[test]
    fn test_reset() {
        let mut acc = ContinuousDrawdownRuns::new();
        feed(&mut acc, &[-1.0, -2.0, 1.0, -3.0]);
        assert!(acc.run_count() > 0);
        acc.reset();
        st(&acc, &[], 0);
        // It must also be possible to use it again after reset.
        acc.update(-4.0);
        st(&acc, &[dd(&[-4.0])], 1);
    }

    // Rolling-window tests

    #[test]
    fn test_rolling_window_eviction_from_front_of_losing_run() {
        let mut acc = ContinuousDrawdownRuns::new();
        feed(&mut acc, &[-1.0, -2.0, -3.0]);
        st(&acc, &[dd(&[-1.0, -2.0, -3.0])], 1);
        acc.revert(-1.0);
        acc.update(-4.0);
        st(&acc, &[dd(&[-2.0, -3.0, -4.0])], 1);
    }

    #[test]
    fn test_rolling_window_eviction_of_entire_losing_run() {
        let mut acc = ContinuousDrawdownRuns::new();
        feed(&mut acc, &[-1.0, -2.0, 1.0]);
        st(&acc, &[dd(&[-1.0, -2.0])], 1);
        acc.revert(-1.0);
        acc.update(-3.0);
        st(&acc, &[dd(&[-2.0]), dd(&[-3.0])], 2);
    }

    #[test]
    fn test_rolling_window_eviction_of_separator() {
        let mut acc = ContinuousDrawdownRuns::new();
        feed(&mut acc, &[-1.0, 1.0, -2.0]);
        st(&acc, &[dd(&[-1.0]), dd(&[-2.0])], 2);
        acc.revert(-1.0);
        acc.update(-3.0);
        st(&acc, &[dd(&[-2.0, -3.0])], 1);
    }

    #[test]
    fn test_rolling_window_multiple_runs() {
        let mut acc = ContinuousDrawdownRuns::new();
        feed(&mut acc, &[-1.0, -2.0, 1.0, -3.0, -4.0]);
        st(&acc, &[dd(&[-1.0, -2.0]), dd(&[-3.0, -4.0])], 2);
        // Slide 1: remove -1%, add +2%
        acc.revert(-1.0);
        acc.update(2.0);
        st(&acc, &[dd(&[-2.0]), dd(&[-3.0, -4.0])], 2);
        // Slide 2: remove -2%, add -5%
        acc.revert(-2.0);
        acc.update(-5.0);
        st(&acc, &[dd(&[-3.0, -4.0]), dd(&[-5.0])], 2);
    }

    #[test]
    fn test_rolling_window_new_return_extends_existing_run() {
        let mut acc = ContinuousDrawdownRuns::new();
        feed(&mut acc, &[1.0, -2.0, -3.0]);
        st(&acc, &[dd(&[-2.0, -3.0])], 1);
        acc.revert(1.0);
        acc.update(-4.0);
        st(&acc, &[dd(&[-2.0, -3.0, -4.0])], 1);
    }

    #[test]
    fn test_rolling_window_new_negative_starts_new_run_after_separator() {
        let mut acc = ContinuousDrawdownRuns::new();
        feed(&mut acc, &[-2.0, 1.0, -3.0]);
        st(&acc, &[dd(&[-2.0]), dd(&[-3.0])], 2);
        acc.revert(-2.0);
        acc.update(-4.0);
        st(&acc, &[dd(&[-3.0, -4.0])], 1);
    }

    #[test]
    fn test_revert_then_update_order_is_required() {
        let mut acc = ContinuousDrawdownRuns::new();
        feed(&mut acc, &[-1.0, -2.0, 2.0]);
        st(&acc, &[dd(&[-1.0, -2.0])], 1);
        // New window: [-2%, +2%, -3%]
        acc.revert(-1.0);
        acc.update(-3.0);
        st(&acc, &[dd(&[-2.0]), dd(&[-3.0])], 2);
    }

    // Numerical consistency

    #[test]
    fn test_sqrt_sum_drawdowns_squared() {
        let mut acc = ContinuousDrawdownRuns::new();
        feed(&mut acc, &[-1.0, -2.0, 1.0, -3.0]);
        let dd1 = dd(&[-1.0, -2.0]);
        let dd2 = dd(&[-3.0]);
        let expected_sum_sq = dd1.powf(2.0) + dd2.powf(2.0);
        let expected_sqrt = expected_sum_sq.sqrt();
        assert!(almost_equal(acc.sum_drawdowns_squared(), expected_sum_sq, 12));
        assert!(almost_equal(acc.sqrt_sum_drawdowns_squared(), expected_sqrt, 12));
    }

    // Brute-force rolling-window test

    #[test]
    fn test_rolling_window_matches_fresh_calculation() {
        fn reference_runs(window: &[f64]) -> Vec<f64> {
            let mut runs = Vec::new();
            let mut current: Vec<f64> = Vec::new();
            for &r in window {
                if r < 0.0 {
                    current.push(r);
                } else if !current.is_empty() {
                    runs.push(dd(&current));
                    current.clear();
                }
            }
            if !current.is_empty() {
                runs.push(dd(&current));
            }
            runs
        }

        let mut rng = Rng::new(42);
        for window_size in [1usize, 2, 3, 5, 12] {
            let returns: Vec<f64> = (0..150)
                .map(|_| {
                    let g = gauss(&mut rng, 0.0, 3.0);
                    choice2(&mut rng, 0.0, g)
                })
                .collect();
            let mut acc = ContinuousDrawdownRuns::new();
            for (i, &ret) in returns.iter().enumerate() {
                if i >= window_size {
                    acc.revert(returns[i - window_size]);
                }
                acc.update(ret);
                let window = &returns[(i + 1).saturating_sub(window_size)..=i];
                let expected = reference_runs(window);
                assert_state(
                    &acc,
                    &expected,
                    Some(expected.len()),
                    12,
                    &format!("window {window_size} step {i}"),
                );
            }
        }
    }
}
