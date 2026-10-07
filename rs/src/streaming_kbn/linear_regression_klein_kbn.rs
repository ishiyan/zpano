use super::klein_kbn_accumulator::KleinKbnAccumulator;
use super::raw_moments_klein_kbn::RawMomentsKleinKbn;

/// Streaming ordinary least squares (OLS) regression y = a + b·x with
/// Klein KBN (Kahan-Babuška-Neumaier) compensated accumulation.
///
/// Tracks the means and variances of x and y (as `RawMomentsKleinKbn` with
/// ddof=0) and the co-moment S_xy = Σ(x − x̄)(y − ȳ), updated in O(1) per
/// sample (n = count before adding the sample):
///
/// ```text
/// S_xy += (x − x̄)·(y − ȳ)·n / (n + 1)
/// ```
///
/// where x̄ and ȳ are the means before adding the sample.
///
/// Like `RawMomentsKleinKbn`, revert() may remove any previously added
/// (x, y) pair, not only the most recent one, so the type works for FIFO
/// rolling windows.
///
/// Derived quantities, with S_xx = Σ(x − x̄)² and S_yy = Σ(y − ȳ)²:
///
/// ```text
/// slope       b = S_xy / S_xx
/// intercept   a = ȳ − b·x̄
/// correlation r = S_xy / √(S_xx·S_yy)
/// covariance      S_xy / n   (population)
/// ```
#[derive(Debug, Clone)]
pub struct LinearRegressionKleinKbn {
    n: usize,
    x_moments: RawMomentsKleinKbn,
    y_moments: RawMomentsKleinKbn,
    s_xy: KleinKbnAccumulator,
}

impl LinearRegressionKleinKbn {
    /// Creates an empty regression.
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

    /// Adds a sample (x, y).
    pub fn update(&mut self, x: f64, y: f64) {
        let n_old = self.n;
        self.n += 1;
        let term = (self.x_moments.mean() - x) * (self.y_moments.mean() - y) * n_old as f64
            / (n_old + 1) as f64;
        self.s_xy.update(term);
        self.x_moments.update(x);
        self.y_moments.update(y);
    }

    /// Removes a previously added sample (x, y), not necessarily the most
    /// recent one.
    ///
    /// # Panics
    ///
    /// Panics if there are no samples.
    pub fn revert(&mut self, x: f64, y: f64) {
        if self.n == 0 {
            panic!("Cannot revert from an empty regression");
        }
        if self.n == 1 {
            self.reset();
            return;
        }
        self.x_moments.revert(x);
        self.y_moments.revert(y);
        // The means are now those without (x, y), as in update().
        let n = self.n - 1;
        let term = (self.x_moments.mean() - x) * (self.y_moments.mean() - y) * n as f64
            / (n + 1) as f64;
        self.s_xy.revert(term);
        self.n = n;
    }

    /// The number of samples.
    pub fn n(&self) -> usize {
        self.n
    }

    /// The mean of x (0.0 when empty).
    pub fn mean_x(&self) -> f64 {
        self.x_moments.mean()
    }

    /// The mean of y (0.0 when empty).
    pub fn mean_y(&self) -> f64 {
        self.y_moments.mean()
    }

    /// The population variance of x, S_xx / n (NaN when empty).
    pub fn variance_x(&self) -> f64 {
        self.x_moments.variance()
    }

    /// The population variance of y, S_yy / n (NaN when empty).
    pub fn variance_y(&self) -> f64 {
        self.y_moments.variance()
    }

    /// The co-moment S_xy = Σ(x − x̄)(y − ȳ) (0.0 when empty).
    pub fn co_moment(&self) -> f64 {
        self.s_xy.value()
    }

    /// The population covariance S_xy / n (NaN when empty).
    pub fn covariance(&self) -> f64 {
        let n = self.n;
        if n < 1 {
            return f64::NAN;
        }
        self.s_xy.value() / n as f64
    }

    /// The OLS slope b = S_xy / S_xx.
    ///
    /// NaN when n < 2 or all x are equal (S_xx = 0).
    pub fn slope(&self) -> f64 {
        let n = self.n;
        if n < 2 {
            return f64::NAN;
        }
        let s_xx = self.x_moments.variance() * n as f64;
        if s_xx != 0.0 {
            self.s_xy.value() / s_xx
        } else {
            f64::NAN
        }
    }

    /// The OLS intercept a = ȳ − b·x̄ (NaN when the slope is NaN).
    pub fn intercept(&self) -> f64 {
        self.y_moments.mean() - self.slope() * self.x_moments.mean()
    }

    /// The Pearson correlation coefficient r = S_xy / √(S_xx·S_yy),
    /// clamped to [−1, 1] to absorb rounding.
    ///
    /// NaN when n < 2 or either x or y is constant.
    pub fn correlation(&self) -> f64 {
        let n = self.n;
        if n < 2 {
            return f64::NAN;
        }
        let t = self.x_moments.standard_deviation() * self.y_moments.standard_deviation();
        if t == 0.0 {
            return f64::NAN;
        }
        let r = self.s_xy.value() / (t * n as f64);
        // Python's max(-1.0, min(1.0, r)).
        let r = if r < 1.0 { r } else { 1.0 };
        if r > -1.0 { r } else { -1.0 }
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
    use crate::streaming_kbn::test_support::almost_equal;

    type Getter = fn(&LinearRegressionKleinKbn) -> f64;

    // Bacon, Carl R., Practical Portfolio Performance Measurement and
    // Attribution, 2nd ed., Wiley, 2008, p. 65 (portfolio) and p. 66 (benchmark).
    const PORTFOLIO: [f64; 24] = [
        0.003, 0.026, 0.011, -0.010,
        0.015, 0.025, 0.016, 0.067,
        -0.014, 0.040, -0.005, 0.081,
        0.040, -0.037, -0.061, 0.017,
        -0.049, -0.022, 0.070, 0.058,
        -0.065, 0.024, -0.005, -0.009,
    ];
    const BENCHMARK: [f64; 24] = [
        0.002, 0.025, 0.018, -0.011,
        0.014, 0.018, 0.014, 0.065,
        -0.015, 0.042, -0.006, 0.083,
        0.039, -0.038, -0.062, 0.015,
        -0.048, 0.021, 0.060, 0.056,
        -0.067, 0.019, -0.003, 0.000,
    ];

    // Reference values for x = BENCHMARK, y = PORTFOLIO, computed with exact
    // rational arithmetic (fractions.Fraction on the binary float inputs,
    // square roots with 50-digit decimal.Decimal), rounded to the nearest float.
    const SLOPE: f64 = 0.9988502086225746;
    const INTERCEPT: f64 = -0.001030120844918352;
    const CORRELATION: f64 = 0.9693858148753051;
    const CO_MOMENT: f64 = 0.033844;
    const COVARIANCE: f64 = 0.0014101666666666668;
    const VARIANCE_X: f64 = 0.0014117899305555557;
    const VARIANCE_Y: f64 = 0.0014989166666666668;

    fn feed(mut reg: LinearRegressionKleinKbn, xs: &[f64], ys: &[f64]) -> LinearRegressionKleinKbn {
        for (&x, &y) in xs.iter().zip(ys.iter()) {
            reg.update(x, y);
        }
        reg
    }

    fn assert_all_nan(reg: &LinearRegressionKleinKbn) {
        let getters: [(&str, Getter); 3] = [
            ("slope", LinearRegressionKleinKbn::slope),
            ("intercept", LinearRegressionKleinKbn::intercept),
            ("correlation", LinearRegressionKleinKbn::correlation),
        ];
        for (name, getter) in getters {
            assert!(getter(reg).is_nan(), "{} = {}", name, getter(reg));
        }
    }

    #[test]
    fn test_bacon() {
        let reg = feed(LinearRegressionKleinKbn::new(), &BENCHMARK, &PORTFOLIO);
        assert_eq!(reg.n(), PORTFOLIO.len());
        assert!(almost_equal(reg.slope(), SLOPE, 14));
        assert!(almost_equal(reg.intercept(), INTERCEPT, 15));
        assert!(almost_equal(reg.correlation(), CORRELATION, 14));
        assert!(almost_equal(reg.co_moment(), CO_MOMENT, 16));
        assert!(almost_equal(reg.covariance(), COVARIANCE, 16));
        assert!(almost_equal(reg.variance_x(), VARIANCE_X, 16));
        assert!(almost_equal(reg.variance_y(), VARIANCE_Y, 16));
    }

    #[test]
    fn test_perfect_fit() {
        let mut reg = LinearRegressionKleinKbn::new();
        for i in 0..5 {
            let x = i as f64;
            reg.update(x, 2.0 * x + 1.0);
        }
        assert!(almost_equal(reg.slope(), 2.0, 13));
        assert!(almost_equal(reg.intercept(), 1.0, 13));
        assert!(almost_equal(reg.correlation(), 1.0, 15));
        assert!(almost_equal(reg.mean_x(), 2.0, 15));
        assert!(almost_equal(reg.mean_y(), 5.0, 15));
        assert!(almost_equal(reg.variance_x(), 2.0, 15));
        assert!(almost_equal(reg.co_moment(), 20.0, 13));
        assert!(almost_equal(reg.covariance(), 4.0, 13));
    }

    #[test]
    fn test_negative_correlation() {
        let mut reg = LinearRegressionKleinKbn::new();
        for i in 0..5 {
            let x = i as f64;
            reg.update(x, -2.0 * x + 1.0);
        }
        assert!(almost_equal(reg.slope(), -2.0, 13));
        assert!(almost_equal(reg.intercept(), 1.0, 13));
        assert!(almost_equal(reg.correlation(), -1.0, 15));
        assert!(almost_equal(reg.covariance(), -4.0, 13));
    }

    #[test]
    fn test_constant_y() {
        let mut reg = LinearRegressionKleinKbn::new();
        for i in 0..5 {
            reg.update(i as f64, 3.0);
        }
        assert!(almost_equal(reg.slope(), 0.0, 13));
        assert!(almost_equal(reg.intercept(), 3.0, 13));
        assert_eq!(reg.covariance(), 0.0);
        assert!(reg.correlation().is_nan());
    }

    #[test]
    fn test_constant_x() {
        let mut reg = LinearRegressionKleinKbn::new();
        for i in 0..5 {
            reg.update(3.0, i as f64);
        }
        assert_eq!(reg.covariance(), 0.0);
        assert_all_nan(&reg);
    }

    #[test]
    fn test_empty() {
        let reg = LinearRegressionKleinKbn::default();
        assert_eq!(reg.n(), 0);
        assert_eq!(reg.co_moment(), 0.0);
        assert!(reg.covariance().is_nan());
        assert!(reg.variance_x().is_nan());
        assert_all_nan(&reg);
    }

    #[test]
    fn test_single_point() {
        let mut reg = LinearRegressionKleinKbn::new();
        reg.update(1.0, 2.0);
        assert_eq!(reg.covariance(), 0.0);
        assert_all_nan(&reg);
    }

    #[test]
    fn test_two_points() {
        let mut reg = LinearRegressionKleinKbn::new();
        reg.update(0.0, 1.0);
        reg.update(2.0, 5.0);
        assert!(almost_equal(reg.slope(), 2.0, 13));
        assert!(almost_equal(reg.intercept(), 1.0, 13));
        assert!(almost_equal(reg.correlation(), 1.0, 13));
    }

    #[test]
    fn test_revert_most_recent() {
        let mut xs = BENCHMARK.to_vec();
        xs.push(0.5);
        let mut ys = PORTFOLIO.to_vec();
        ys.push(-0.5);
        let mut reg = feed(LinearRegressionKleinKbn::new(), &xs, &ys);
        reg.revert(0.5, -0.5);
        assert_eq!(reg.n(), PORTFOLIO.len());
        assert!(almost_equal(reg.slope(), SLOPE, 13));
        assert!(almost_equal(reg.intercept(), INTERCEPT, 14));
        assert!(almost_equal(reg.correlation(), CORRELATION, 13));
        assert!(almost_equal(reg.covariance(), COVARIANCE, 15));
    }

    #[test]
    fn test_revert_oldest() {
        let mut xs = vec![0.5];
        xs.extend_from_slice(&BENCHMARK);
        let mut ys = vec![-0.5];
        ys.extend_from_slice(&PORTFOLIO);
        let mut reg = feed(LinearRegressionKleinKbn::new(), &xs, &ys);
        reg.revert(0.5, -0.5);
        assert!(almost_equal(reg.slope(), SLOPE, 13));
        assert!(almost_equal(reg.intercept(), INTERCEPT, 14));
        assert!(almost_equal(reg.correlation(), CORRELATION, 13));
        assert!(almost_equal(reg.covariance(), COVARIANCE, 15));
    }

    #[test]
    fn test_revert_to_single() {
        let mut reg = LinearRegressionKleinKbn::new();
        reg.update(1.0, 2.0);
        reg.update(3.0, 4.0);
        reg.revert(3.0, 4.0);
        assert_eq!(reg.n(), 1);
        assert!(almost_equal(reg.mean_x(), 1.0, 15));
        assert!(almost_equal(reg.mean_y(), 2.0, 15));
        assert!(almost_equal(reg.co_moment(), 0.0, 15));
        assert_all_nan(&reg);
    }

    #[test]
    fn test_revert_to_empty() {
        let mut reg = LinearRegressionKleinKbn::new();
        reg.update(1.0, 2.0);
        reg.revert(1.0, 2.0);
        assert_eq!(reg.n(), 0);
        assert_all_nan(&reg);
    }

    #[test]
    #[should_panic(expected = "Cannot revert from an empty regression")]
    fn test_revert_empty_raises() {
        let mut reg = LinearRegressionKleinKbn::new();
        reg.revert(1.0, 2.0);
    }

    #[test]
    fn test_rolling_window() {
        let names: [(&str, Getter); 7] = [
            ("slope", LinearRegressionKleinKbn::slope),
            ("intercept", LinearRegressionKleinKbn::intercept),
            ("correlation", LinearRegressionKleinKbn::correlation),
            ("covariance", LinearRegressionKleinKbn::covariance),
            ("co_moment", LinearRegressionKleinKbn::co_moment),
            ("variance_x", LinearRegressionKleinKbn::variance_x),
            ("variance_y", LinearRegressionKleinKbn::variance_y),
        ];
        let w = 6usize;
        let mut reg = LinearRegressionKleinKbn::new();
        for (i, (&x, &y)) in BENCHMARK.iter().zip(PORTFOLIO.iter()).enumerate() {
            reg.update(x, y);
            if i >= w {
                reg.revert(BENCHMARK[i - w], PORTFOLIO[i - w]);
            }
            let lo = (i + 1).saturating_sub(w);
            let r = feed(LinearRegressionKleinKbn::new(), &BENCHMARK[lo..i + 1], &PORTFOLIO[lo..i + 1]);
            assert_eq!(reg.n(), r.n());
            for (name, getter) in names {
                let (actual, expected) = (getter(&reg), getter(&r));
                if expected.is_nan() {
                    assert!(actual.is_nan(), "step {} {}: {}", i, name, actual);
                } else {
                    assert!(almost_equal(actual, expected, 13), "step {} {}: {} vs {}", i, name, actual, expected);
                }
            }
        }
    }

    #[test]
    fn test_reset() {
        let mut reg = LinearRegressionKleinKbn::new();
        for i in 0..5 {
            let x = i as f64;
            reg.update(x, 2.0 * x + 1.0);
        }
        reg.reset();
        assert_eq!(reg.n(), 0);
        assert_all_nan(&reg);
        reg.update(0.0, 1.0);
        reg.update(1.0, 3.0);
        assert!(almost_equal(reg.slope(), 2.0, 13));
    }
}
