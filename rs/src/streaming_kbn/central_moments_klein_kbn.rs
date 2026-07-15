use super::klein_kbn_accumulator::KleinKbnAccumulator;

/// Streaming mean, variance, skewness, and kurtosis via Pébay's central moment
/// update with KBN double-compensated accumulation.
///
/// Maintains running sums of central moments m2, m3, m4 (as `KleinKbnAccumulator`s)
/// updated in O(1) per sample. Preferred over `RawMomentsKleinKbn` for forward-only
/// computation (no revert) because it avoids the numerical cancellation inherent
/// in converting raw power sums to central moments.
///
/// Only the most recent sample can be reverted (LIFO stack, not FIFO queue).
pub struct CentralMomentsKleinKbn {
    n: usize,
    m1: KleinKbnAccumulator,
    m2: KleinKbnAccumulator,
    m3: KleinKbnAccumulator,
    m4: KleinKbnAccumulator,
    ddof: usize,
    bias: bool,
    fisher: bool,
}

impl CentralMomentsKleinKbn {
    pub fn new(ddof: usize, bias: bool, fisher: bool) -> Self {
        Self {
            n: 0,
            m1: KleinKbnAccumulator::new(),
            m2: KleinKbnAccumulator::new(),
            m3: KleinKbnAccumulator::new(),
            m4: KleinKbnAccumulator::new(),
            ddof,
            bias,
            fisher,
        }
    }

    /// Clears all accumulated state.
    pub fn reset(&mut self) {
        self.n = 0;
        self.m1.reset();
        self.m2.reset();
        self.m3.reset();
        self.m4.reset();
    }

    /// Adds a new sample `x` using Pébay's central moment update formulas.
    pub fn update(&mut self, x: f64) {
        let n_old = self.n;
        let n_new = n_old + 1;
        self.n = n_new;
        let delta = x - self.m1.value();
        let delta_n = delta / (n_new as f64);
        let delta_n2 = delta_n * delta_n;
        let n_old_f = n_old as f64;
        let n_new_f = n_new as f64;
        let term = delta * delta_n * n_old_f;

        self.m1.update(delta_n);
        self.m4.update(
            term * delta_n2 * (n_new_f * n_new_f - 3.0 * n_new_f + 3.0)
                + 6.0 * delta_n2 * self.m2.value()
                - 4.0 * delta_n * self.m3.value(),
        );
        self.m3.update(term * delta_n * (n_new_f - 2.0) - 3.0 * delta_n * self.m2.value());
        self.m2.update(term);
    }

    /// Removes the most recently added sample `x` (LIFO).
    ///
    /// Uses inverse Pébay formulas to restore prior state. The `KleinKbnAccumulator::set()`
    /// method is used for m1–m4, which resets the compensation terms to zero.
    pub fn revert(&mut self, x: f64) {
        let n_new = self.n;
        if n_new == 0 {
            panic!("cannot revert below 0");
        }
        let n_old = n_new - 1;
        if n_old == 0 {
            self.n = 0;
            self.m1.reset();
            self.m2.reset();
            self.m3.reset();
            self.m4.reset();
            return;
        }

        let m1_new = self.m1.value();
        let m2_new = self.m2.value();
        let m3_new = self.m3.value();
        let m4_new = self.m4.value();

        let n_new_f = n_new as f64;
        let n_old_f = n_old as f64;

        let m1_old = (n_new_f * m1_new - x) / n_old_f;
        let delta = x - m1_old;
        let delta_n = delta / n_new_f;
        let delta_n2 = delta_n * delta_n;
        let term = delta * delta_n * n_old_f;

        let m2_old = m2_new - term;
        let m3_old = m3_new - (term * delta_n * (n_new_f - 2.0) - 3.0 * delta_n * m2_old);
        let m4_old = m4_new
            - (term * delta_n2 * (n_new_f * n_new_f - 3.0 * n_new_f + 3.0)
                + 6.0 * delta_n2 * m2_old
                - 4.0 * delta_n * m3_old);

        self.n = n_old;
        self.m1.set(m1_old);
        self.m2.set(m2_old);
        self.m3.set(m3_old);
        self.m4.set(m4_old);
    }

    /// Returns the current arithmetic mean.
    pub fn mean(&self) -> f64 {
        self.m1.value()
    }

    /// Returns the current variance.
    /// Returns NaN if `n <= ddof`.
    pub fn variance(&self) -> f64 {
        let n = (self.n as f64) - (self.ddof as f64);
        if n <= 0.0 {
            return f64::NAN;
        }
        self.m2.value() / n
    }

    /// Returns the current standard deviation.
    /// Returns NaN if `n <= ddof`.
    pub fn standard_deviation(&self) -> f64 {
        let n = (self.n as f64) - (self.ddof as f64);
        if n <= 0.0 {
            return f64::NAN;
        }
        (self.m2.value() / n).sqrt()
    }

    /// Returns the current skewness.
    /// Returns NaN if `n < 3` or `m2 <= 0`.
    pub fn skewness(&self) -> f64 {
        let nu = self.n;
        if nu < 3 || self.m2.value() <= 0.0 {
            return f64::NAN;
        }
        let n = nu as f64;
        let g1 = n.sqrt() * self.m3.value() / self.m2.value().powf(1.5);
        if self.bias {
            g1
        } else {
            g1 * (n * (n - 1.0)).sqrt() / (n - 2.0)
        }
    }

    /// Returns the current kurtosis.
    /// Returns NaN if `n < 4` or `m2 <= 0`.
    pub fn kurtosis(&self) -> f64 {
        let nu = self.n;
        if nu < 4 || self.m2.value() <= 0.0 {
            return f64::NAN;
        }
        let n = nu as f64;
        let raw = n * self.m4.value() / (self.m2.value() * self.m2.value());
        if !self.bias {
            let adj =
                ((n * n - 1.0) * raw - 3.0 * (n - 1.0) * (n - 1.0)) / ((n - 2.0) * (n - 3.0));
            if self.fisher { adj } else { adj + 3.0 }
        } else if self.fisher {
            raw - 3.0
        } else {
            raw
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn almost_equal(a: f64, b: f64, eps: f64) -> bool {
        (a - b).abs() < eps
    }

    const BACON_DATA: [f64; 24] = [
        0.003, 0.026, 0.011, -0.010, 0.015, 0.025, 0.016, 0.067,
        -0.014, 0.040, -0.005, 0.081, 0.040, -0.037, -0.061, 0.017,
        -0.049, -0.022, 0.070, 0.058, -0.065, 0.024, -0.005, -0.009,
    ];

    #[test]
    fn test_simple_update() {
        let mut m = CentralMomentsKleinKbn::new(0, true, true);
        for &x in &[1.0, 2.0, 3.0, 4.0] {
            m.update(x);
        }
        assert!(almost_equal(m.mean(), 2.5, 1e-15),
            "mean = {}, want 2.5", m.mean());
        assert!(almost_equal(m.variance(), 1.25, 1e-15),
            "var = {}, want 1.25", m.variance());
        assert!(almost_equal(m.skewness(), 0.0, 1e-14),
            "skew = {}, want 0.0", m.skewness());
        assert!(almost_equal(m.kurtosis(), -1.36, 1e-13),
            "kurt = {}, want -1.36", m.kurtosis());
    }

    #[test]
    fn test_compare_scipy() {
        let mut m = CentralMomentsKleinKbn::new(0, true, true);
        for &x in &BACON_DATA {
            m.update(x);
        }
        assert!(almost_equal(m.mean(), 0.009000000000000001, 1e-15),
            "mean = {}, want 0.009000000000000001", m.mean());
        assert!(almost_equal(m.variance(), 0.0014989166666666668, 1e-15),
            "var = {}, want 0.0014989166666666668", m.variance());
        assert!(almost_equal(m.skewness(), -0.08256245520856803, 1e-14),
            "skew = {}, want -0.08256245520856803", m.skewness());
        assert!(almost_equal(m.kurtosis(), -0.5675462058921261, 1e-13),
            "kurt = {}, want -0.5675462058921261", m.kurtosis());
    }

    #[test]
    fn test_compare_scipy_bias_false() {
        let mut m = CentralMomentsKleinKbn::new(0, false, true);
        for &x in &BACON_DATA {
            m.update(x);
        }
        assert!(almost_equal(m.skewness(), -0.08817174934967532, 1e-14),
            "skew = {}, want -0.08817174934967532", m.skewness());
        assert!(almost_equal(m.kurtosis(), -0.4076603211860876, 1e-13),
            "kurt = {}, want -0.4076603211860876", m.kurtosis());
    }

    #[test]
    fn test_ddof() {
        let mut m = CentralMomentsKleinKbn::new(1, true, true);
        for &x in &[1.0, 2.0, 3.0] {
            m.update(x);
        }
        assert!(almost_equal(m.variance(), 1.0, 1e-15),
            "var = {}, want 1.0", m.variance());
    }

    #[test]
    fn test_revert_lifo_simple() {
        let data = [10.0, 18.0, 5.0];
        let mut full = CentralMomentsKleinKbn::new(0, true, true);
        let mut part = CentralMomentsKleinKbn::new(0, true, true);
        for &x in &data {
            full.update(x);
        }
        for &x in &data[..2] {
            part.update(x);
        }
        full.revert(data[2]);

        assert!(almost_equal(full.mean(), part.mean(), 1e-15),
            "mean full={} part={}", full.mean(), part.mean());
        assert!(almost_equal(full.variance(), part.variance(), 1e-15),
            "var full={} part={}", full.variance(), part.variance());
        assert!(full.skewness().is_nan() && part.skewness().is_nan(),
            "skew full={} part={}", full.skewness(), part.skewness());
        assert!(full.kurtosis().is_nan() && part.kurtosis().is_nan(),
            "kurt full={} part={}", full.kurtosis(), part.kurtosis());
    }

    #[test]
    fn test_revert_lifo_bacon() {
        let mut full = CentralMomentsKleinKbn::new(0, true, true);
        let mut part = CentralMomentsKleinKbn::new(0, true, true);
        for &x in &BACON_DATA {
            full.update(x);
        }
        for &x in &BACON_DATA[..BACON_DATA.len() - 1] {
            part.update(x);
        }
        full.revert(BACON_DATA[BACON_DATA.len() - 1]);

        assert!(almost_equal(full.mean(), part.mean(), 1e-15),
            "mean full={} part={}", full.mean(), part.mean());
        assert!(almost_equal(full.variance(), part.variance(), 1e-15),
            "var full={} part={}", full.variance(), part.variance());
        assert!(almost_equal(full.skewness(), part.skewness(), 1e-14),
            "skew full={} part={}", full.skewness(), part.skewness());
        assert!(almost_equal(full.kurtosis(), part.kurtosis(), 1e-13),
            "kurt full={} part={}", full.kurtosis(), part.kurtosis());
    }

    #[test]
    fn test_revert_lifo_roundtrip() {
        let mut m = CentralMomentsKleinKbn::new(0, true, true);
        for &x in &BACON_DATA {
            m.update(x);
        }
        for i in (0..BACON_DATA.len()).rev() {
            m.revert(BACON_DATA[i]);
        }
        assert_eq!(m.n, 0, "n = {}, want 0", m.n);
        assert_eq!(m.mean(), 0.0, "mean = {}, want 0", m.mean());
        assert!(m.variance().is_nan(), "var = {}, want NaN", m.variance());
    }

    #[test]
    fn test_reset() {
        let mut m = CentralMomentsKleinKbn::new(0, true, true);
        m.update(10.0);
        m.reset();
        assert_eq!(m.n, 0, "n = {}, want 0", m.n);
        assert_eq!(m.mean(), 0.0, "mean = {}, want 0", m.mean());
        assert!(m.variance().is_nan(), "var = {}, want NaN", m.variance());
    }
}
