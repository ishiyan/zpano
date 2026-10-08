//! Streaming cumulative (geometric) returns.

use crate::streaming_kbn::KleinKbnAccumulator;

/// Streaming cumulative (geometric) returns.
///
/// Accumulates the sum of log returns, log(1 + r), with compensated
/// summation. Because only a sum is stored, `revert()` may remove any
/// previously added return, so the type works for FIFO rolling windows:
/// the caller owns the window and feeds evicted returns to `revert()`.
#[derive(Debug, Clone, Default)]
pub struct CumulativeReturn {
    cumlogret_sum: KleinKbnAccumulator,
    count: usize,
}

#[inline]
fn logret(ret: f64) -> f64 {
    if ret != 0.0 { ret.ln_1p() } else { 0.0 }
}

impl CumulativeReturn {
    /// Creates an empty accumulator.
    pub fn new() -> Self {
        Self::default()
    }

    /// Resets to the initial empty state.
    pub fn reset(&mut self) {
        self.cumlogret_sum.reset();
        self.count = 0;
    }

    /// Removes a previously added return.
    ///
    /// # Errors
    ///
    /// Returns `"Cannot revert from an empty accumulator"` if there are
    /// no returns.
    pub fn revert(&mut self, ret: f64) -> Result<(), String> {
        if self.count == 0 {
            return Err("Cannot revert from an empty accumulator".to_string());
        }
        self.count -= 1;
        self.cumlogret_sum.revert(logret(ret));
        Ok(())
    }

    /// Adds a return, expressed as a decimal (must be > -1).
    pub fn update(&mut self, ret: f64) {
        self.count += 1;
        self.cumlogret_sum.update(logret(ret));
    }

    /// The number of returns.
    pub fn count(&self) -> usize {
        self.count
    }

    /// Cumulative geometric return, prod(1 + r) - 1 (0.0 when empty).
    pub fn cumulative_geometric_return(&self) -> f64 {
        self.cumlogret_sum.value().exp_m1()
    }

    /// The geometric mean of the returns, prod(1 + r)^(1/n) - 1
    /// (NaN when empty).
    pub fn geometric_mean_return(&self) -> f64 {
        if self.count > 0 {
            (self.cumlogret_sum.value() / self.count as f64).exp_m1()
        } else {
            f64::NAN
        }
    }

    /// The annualized geometric mean, prod(1 + r)^(periods_per_year/n) - 1
    /// (NaN when empty).
    pub fn annualized_geometric_mean_return(&self, periods_per_year: f64) -> f64 {
        if self.count == 0 {
            return f64::NAN;
        }
        (self.cumlogret_sum.value() * periods_per_year / self.count as f64).exp_m1()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::performance::core::test_support::{almost_equal, choice2, gauss, prod, Rng};

    #[test]
    fn test_annualized_return_definition() {
        let returns = [0.10, -0.05, 0.03, 0.08];
        let periods_per_year = 12.0;
        let mut cr = CumulativeReturn::new();
        for r in returns {
            cr.update(r);
        }
        let growth = prod(returns.iter().map(|r| 1.0 + r));
        let expected = growth.powf(periods_per_year / returns.len() as f64) - 1.0;
        let actual = cr.annualized_geometric_mean_return(periods_per_year);
        assert!(almost_equal(actual, expected, 15), "{actual} vs {expected}");
    }

    #[test]
    fn test_one_month_return() {
        // One monthly return of 1%, expected (1.01)^12-1.
        let mut cr = CumulativeReturn::new();
        cr.update(0.01);
        let expected = 1.01f64.powf(12.0) - 1.0;
        let actual = cr.annualized_geometric_mean_return(12.0);
        assert!(almost_equal(actual, expected, 15), "{actual} vs {expected}");
    }

    #[test]
    fn test_yearly_returns() {
        // For yearly observations the annualized geometric mean return
        // equals the geometric mean return.
        let mut cr = CumulativeReturn::new();
        for r in [0.12, -0.04, 0.08] {
            cr.update(r);
        }
        let expected = cr.geometric_mean_return();
        let actual = cr.annualized_geometric_mean_return(1.0);
        assert!(almost_equal(actual, expected, 15), "{actual} vs {expected}");
    }

    #[test]
    fn test_constant_monthly_return() {
        // ((1+r)^n)^(12/n) = (1+r)^12: the number of observations cancels.
        let mut cr = CumulativeReturn::new();
        for _ in 0..60 {
            cr.update(0.01);
        }
        let expected = 1.01f64.powf(12.0) - 1.0;
        let actual = cr.annualized_geometric_mean_return(12.0);
        assert!(almost_equal(actual, expected, 15), "{actual} vs {expected}");
    }

    #[test]
    fn test_empty() {
        let cr = CumulativeReturn::new();
        assert!(cr.annualized_geometric_mean_return(12.0).is_nan());
    }

    #[test]
    fn test_zero_returns() {
        // log1p(0) == 0 and expm1(0) == 0, so the result is exactly zero.
        let mut cr = CumulativeReturn::new();
        for _ in 0..100 {
            cr.update(0.0);
        }
        let actual = cr.annualized_geometric_mean_return(252.0);
        assert!(almost_equal(actual, 0.0, 15), "{actual}");
    }

    #[test]
    fn test_consistency_with_geometric_mean_return() {
        // 1 + annualized = (1 + geometric mean)^p
        let returns = [0.0010, -0.0005, 0.0003, 0.0008];
        let periods_per_year = 252.0;
        let mut cr = CumulativeReturn::new();
        for r in returns {
            cr.update(r);
        }
        let expected = (1.0 + cr.geometric_mean_return()).powf(periods_per_year) - 1.0;
        let actual = cr.annualized_geometric_mean_return(periods_per_year);
        assert!(almost_equal(actual, expected, 13), "{actual} vs {expected}");
    }

    #[test]
    fn test_rolling_window_matches_fresh_calculation() {
        let mut rng = Rng::new(42);
        let returns: Vec<f64> = (0..100)
            .map(|_| {
                let g = gauss(&mut rng, 0.0, 0.03);
                choice2(&mut rng, 0.0, g)
            })
            .collect();
        let window_size = 7;
        let mut cr = CumulativeReturn::new();
        for (i, &r) in returns.iter().enumerate() {
            if i >= window_size {
                cr.revert(returns[i - window_size]).unwrap();
            }
            cr.update(r);
            let window = &returns[(i + 1).saturating_sub(window_size)..=i];
            let growth = prod(window.iter().map(|x| 1.0 + x));
            assert_eq!(cr.count(), window.len());
            assert!(almost_equal(cr.cumulative_geometric_return(), growth - 1.0, 14), "step {i}");
            assert!(
                almost_equal(cr.geometric_mean_return(), growth.powf(1.0 / window.len() as f64) - 1.0, 14),
                "step {i}"
            );
        }
    }

    #[test]
    fn test_revert_empty_raises() {
        let mut cr = CumulativeReturn::new();
        let result = cr.revert(0.01);
        assert!(result.is_err());
        assert_eq!(result, Err("Cannot revert from an empty accumulator".to_string()));
    }

    #[test]
    fn test_reset() {
        let mut cr = CumulativeReturn::new();
        for r in [0.1, -0.2] {
            cr.update(r);
        }
        cr.reset();
        assert_eq!(cr.count(), 0);
        assert_eq!(cr.cumulative_geometric_return(), 0.0);
        assert!(cr.geometric_mean_return().is_nan());
    }
}
