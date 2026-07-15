use super::klein_kbn_accumulator::KleinKbnAccumulator;
use super::raw_moments_klein_kbn::RawMomentsKleinKbn;

/// Streaming simple linear regression (y = slope * x + intercept) with
/// KBN-compensated accumulation.
///
/// Internally uses two `RawMomentsKleinKbn` (ddof=0) for x and y moments,
/// and a `KleinKbnAccumulator` for the cross-product S_xy.
///
/// Supports both LIFO revert and FIFO rolling window via the revert/update cycle.
pub struct LinearRegressionKleinKbn {
    n: usize,
    x_moments: RawMomentsKleinKbn,
    y_moments: RawMomentsKleinKbn,
    s_xy: KleinKbnAccumulator,
}

impl LinearRegressionKleinKbn {
    pub fn new() -> Self {
        Self {
            n: 0,
            x_moments: RawMomentsKleinKbn::new(0, true, true),
            y_moments: RawMomentsKleinKbn::new(0, true, true),
            s_xy: KleinKbnAccumulator::new(),
        }
    }

    /// Clears all accumulated state.
    pub fn reset(&mut self) {
        self.n = 0;
        self.x_moments.reset();
        self.y_moments.reset();
        self.s_xy.reset();
    }

    /// Adds a new (x, y) observation.
    pub fn update(&mut self, x: f64, y: f64) {
        let n_old = self.n;
        self.n += 1;
        let term =
            (self.x_moments.mean() - x) * (self.y_moments.mean() - y) * (n_old as f64) / ((n_old + 1) as f64);
        self.s_xy.update(term);
        self.x_moments.update(x);
        self.y_moments.update(y);
    }

    /// Removes a previously added (x, y) observation.
    pub fn revert(&mut self, x: f64, y: f64) {
        if self.n == 0 {
            return;
        }
        if self.n == 1 {
            self.reset();
            return;
        }
        self.x_moments.revert(x);
        self.y_moments.revert(y);
        let n = self.n - 1;
        let term = (self.x_moments.mean() - x) * (self.y_moments.mean() - y) * (n as f64) / ((n + 1) as f64);
        self.s_xy.revert(term);
        self.n = n;
    }

    /// Returns the current slope coefficient.
    /// Returns NaN if `n < 2` or S_xx == 0.
    pub fn slope(&mut self) -> f64 {
        if self.n < 2 {
            return f64::NAN;
        }
        let sxx = self.x_moments.variance() * (self.n as f64);
        if sxx == 0.0 {
            return f64::NAN;
        }
        self.s_xy.value() / sxx
    }

    /// Returns the current intercept coefficient.
    /// Returns NaN if `n < 2`.
    pub fn intercept(&mut self) -> f64 {
        let s = self.slope();
        self.y_moments.mean() - s * self.x_moments.mean()
    }

    /// Returns the current Pearson correlation coefficient.
    /// Returns NaN if `n < 2` or either standard deviation is zero.
    pub fn correlation(&self) -> f64 {
        if self.n < 2 {
            return f64::NAN;
        }
        let t = self.x_moments.standard_deviation() * self.y_moments.standard_deviation();
        if t == 0.0 {
            return f64::NAN;
        }
        self.s_xy.value() / (t * (self.n as f64))
    }
}

impl Default for LinearRegressionKleinKbn {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn almost_equal(a: f64, b: f64, eps: f64) -> bool {
        (a - b).abs() < eps
    }

    #[test]
    fn test_perfect_fit() {
        let mut r = LinearRegressionKleinKbn::new();
        for i in 0..5 {
            let x = i as f64;
            r.update(x, 2.0 * x + 1.0);
        }
        assert!(almost_equal(r.slope(), 2.0, 1e-13),
            "slope = {}, want 2.0", r.slope());
        assert!(almost_equal(r.intercept(), 1.0, 1e-13),
            "intercept = {}, want 1.0", r.intercept());
        assert!(almost_equal(r.correlation(), 1.0, 1e-13),
            "correlation = {}, want 1.0", r.correlation());
    }

    #[test]
    fn test_zero_correlation() {
        let mut r = LinearRegressionKleinKbn::new();
        for i in 0..5 {
            r.update(i as f64, 0.0);
        }
        assert!(almost_equal(r.slope(), 0.0, 1e-13),
            "slope = {}, want 0.0", r.slope());
        assert!(r.correlation().is_nan(),
            "correlation = {}, want NaN", r.correlation());
    }

    #[test]
    fn test_single_point() {
        let mut r = LinearRegressionKleinKbn::new();
        r.update(1.0, 2.0);
        assert!(r.slope().is_nan(),
            "slope = {}, want NaN", r.slope());
        assert!(r.intercept().is_nan(),
            "intercept = {}, want NaN", r.intercept());
        assert!(r.correlation().is_nan(),
            "correlation = {}, want NaN", r.correlation());
    }

    #[test]
    fn test_two_points() {
        let mut r = LinearRegressionKleinKbn::new();
        r.update(0.0, 1.0);
        r.update(2.0, 5.0);
        assert!(almost_equal(r.slope(), 2.0, 1e-13),
            "slope = {}, want 2.0", r.slope());
        assert!(almost_equal(r.intercept(), 1.0, 1e-13),
            "intercept = {}, want 1.0", r.intercept());
        assert!(almost_equal(r.correlation(), 1.0, 1e-13),
            "correlation = {}, want 1.0", r.correlation());
    }

    #[test]
    fn test_revert_matches_single_update() {
        let mut r = LinearRegressionKleinKbn::new();
        r.update(1.0, 2.0);
        r.update(3.0, 4.0);
        r.revert(3.0, 4.0);

        let mut ref_r = LinearRegressionKleinKbn::new();
        ref_r.update(1.0, 2.0);

        assert_eq!(r.n, ref_r.n, "n = {} != {}", r.n, ref_r.n);
        assert!(r.slope().is_nan(),
            "slope = {}, want NaN", r.slope());
        assert!(ref_r.slope().is_nan(),
            "ref.slope = {}, want NaN", ref_r.slope());
    }

    #[test]
    fn test_revert_to_empty() {
        let mut r = LinearRegressionKleinKbn::new();
        r.update(1.0, 2.0);
        r.revert(1.0, 2.0);
        assert_eq!(r.n, 0, "n = {}, want 0", r.n);
        assert!(r.slope().is_nan(),
            "slope = {}, want NaN", r.slope());
        assert!(r.intercept().is_nan(),
            "intercept = {}, want NaN", r.intercept());
        assert!(r.correlation().is_nan(),
            "correlation = {}, want NaN", r.correlation());
    }

    #[test]
    fn test_rolling_window() {
        let data = [(0.0, 1.0), (1.0, 3.0), (2.0, 5.0), (3.0, 7.0), (4.0, 9.0)];

        let mut r = LinearRegressionKleinKbn::new();
        for &(x, y) in &data {
            r.update(x, y);
        }
        r.revert(data[0].0, data[0].1);
        r.revert(data[1].0, data[1].1);
        r.update(5.0, 11.0);
        r.update(6.0, 13.0);

        let mut ref_r = LinearRegressionKleinKbn::new();
        for &(x, y) in &data[2..] {
            ref_r.update(x, y);
        }
        ref_r.update(5.0, 11.0);
        ref_r.update(6.0, 13.0);

        assert_eq!(r.n, ref_r.n, "n = {} != {}", r.n, ref_r.n);
        assert!(almost_equal(r.slope(), ref_r.slope(), 1e-12),
            "slope = {}, want {}", r.slope(), ref_r.slope());
        assert!(almost_equal(r.intercept(), ref_r.intercept(), 1e-12),
            "intercept = {}, want {}", r.intercept(), ref_r.intercept());
        assert!(almost_equal(r.correlation(), ref_r.correlation(), 1e-12),
            "correlation = {}, want {}", r.correlation(), ref_r.correlation());
    }

    #[test]
    fn test_negative_correlation() {
        let mut r = LinearRegressionKleinKbn::new();
        for i in 0..5 {
            let x = i as f64;
            r.update(x, -2.0 * x + 1.0);
        }
        assert!(almost_equal(r.slope(), -2.0, 1e-13),
            "slope = {}, want -2.0", r.slope());
        assert!(almost_equal(r.intercept(), 1.0, 1e-13),
            "intercept = {}, want 1.0", r.intercept());
        assert!(almost_equal(r.correlation(), -1.0, 1e-13),
            "correlation = {}, want -1.0", r.correlation());
    }

    #[test]
    fn test_reset() {
        let mut r = LinearRegressionKleinKbn::new();
        for i in 0..5 {
            r.update(i as f64, 2.0 * (i as f64) + 1.0);
        }
        r.reset();
        assert_eq!(r.n, 0, "n = {}, want 0", r.n);
        assert!(r.slope().is_nan(),
            "slope = {}, want NaN", r.slope());
        assert!(r.intercept().is_nan(),
            "intercept = {}, want NaN", r.intercept());
        assert!(r.correlation().is_nan(),
            "correlation = {}, want NaN", r.correlation());

        r.update(0.0, 1.0);
        r.update(1.0, 3.0);
        assert!(almost_equal(r.slope(), 2.0, 1e-13),
            "slope = {}, want 2.0", r.slope());
    }
}
