//! Single-factor model (SFM) regression of excess returns.

use crate::streaming_kbn::LinearRegressionKleinKbn;

/// Streaming single-factor model regression of the portfolio excess
/// returns `ret - risk_free_rate` on the benchmark excess returns
/// `benchmark - risk_free_rate`, over all periods (alpha, beta, R²),
/// bull periods (benchmark excess > 0) and bear periods (< 0).
///
/// Only sums are stored, so `revert()` may remove any previously added
/// pair (FIFO rolling windows).
#[derive(Debug, Clone)]
pub struct SfmRegression {
    risk_free_rate: f64,
    full: LinearRegressionKleinKbn,
    bull: LinearRegressionKleinKbn,
    bear: LinearRegressionKleinKbn,
}

impl SfmRegression {
    /// Creates an empty regression with the per-period risk-free rate.
    pub fn new(risk_free_rate: f64) -> Self {
        Self {
            risk_free_rate,
            full: LinearRegressionKleinKbn::new(),
            bull: LinearRegressionKleinKbn::new(),
            bear: LinearRegressionKleinKbn::new(),
        }
    }

    /// Resets to the initial empty state (keeps the risk-free rate).
    pub fn reset(&mut self) {
        self.full.reset();
        self.bull.reset();
        self.bear.reset();
    }

    /// Removes a previously added (portfolio, benchmark) return pair.
    pub fn revert(&mut self, ret: f64, benchmark: f64) {
        let x = benchmark - self.risk_free_rate;
        let y = ret - self.risk_free_rate;

        self.full.revert(x, y);

        if x > 0.0 {
            self.bull.revert(x, y);
        } else if x < 0.0 {
            self.bear.revert(x, y);
        }
    }

    /// Adds a (portfolio, benchmark) return pair.
    pub fn update(&mut self, ret: f64, benchmark: f64) {
        let x = benchmark - self.risk_free_rate;
        let y = ret - self.risk_free_rate;

        self.full.update(x, y);

        if x > 0.0 {
            self.bull.update(x, y);
        } else if x < 0.0 {
            self.bear.update(x, y);
        }
    }

    /// Intercept of the full regression (NaN when undefined).
    pub fn alpha(&self) -> f64 {
        self.full.intercept()
    }

    /// Slope of the full regression (NaN when undefined).
    pub fn beta(&self) -> f64 {
        self.full.slope()
    }

    /// Slope of the bull-period regression (NaN when undefined).
    pub fn beta_bull(&self) -> f64 {
        self.bull.slope()
    }

    /// Slope of the bear-period regression (NaN when undefined).
    pub fn beta_bear(&self) -> f64 {
        self.bear.slope()
    }

    /// Coefficient of determination R² = r² of the full regression
    /// (NaN when undefined).
    pub fn r2(&self) -> f64 {
        let corr = self.full.correlation();
        if !corr.is_nan() { corr * corr } else { f64::NAN }
    }
}

impl Default for SfmRegression {
    /// Risk-free rate 0.
    fn default() -> Self {
        Self::new(0.0)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::performance::core::test_support::almost_equal;

    #[test]
    fn test_full_bull_and_bear_fits_use_excess_returns() {
        let mut regression = SfmRegression::new(0.01);
        for excess_benchmark in [-0.04, -0.02, 0.0, 0.02, 0.05] {
            let benchmark = 0.01 + excess_benchmark;
            let portfolio = 0.01 + 0.005 + 2.0 * excess_benchmark;
            regression.update(portfolio, benchmark);
        }

        assert!(almost_equal(regression.alpha(), 0.005, 14));
        assert!(almost_equal(regression.beta(), 2.0, 14));
        assert!(almost_equal(regression.beta_bull(), 2.0, 14));
        assert!(almost_equal(regression.beta_bear(), 2.0, 14));
        assert!(almost_equal(regression.r2(), 1.0, 14));

        // Removing an older bear observation leaves too few bear points
        // for a slope, while the full and bull fits remain defined.
        regression.revert(0.01 + 0.005 + 2.0 * -0.04, 0.01 - 0.04);
        assert!(almost_equal(regression.beta(), 2.0, 14));
        assert!(almost_equal(regression.beta_bull(), 2.0, 14));
        assert!(regression.beta_bear().is_nan());
    }

    #[test]
    fn test_reset_and_zero_excess_benchmark() {
        let mut regression = SfmRegression::new(0.01);
        regression.update(0.02, 0.01);
        assert!(regression.beta_bull().is_nan());
        assert!(regression.beta_bear().is_nan());
        regression.reset();
        let getters: [(&str, fn(&SfmRegression) -> f64); 5] = [
            ("alpha", SfmRegression::alpha),
            ("beta", SfmRegression::beta),
            ("beta_bull", SfmRegression::beta_bull),
            ("beta_bear", SfmRegression::beta_bear),
            ("r2", SfmRegression::r2),
        ];
        for (name, getter) in getters {
            assert!(getter(&regression).is_nan(), "{name}");
        }
    }
}
