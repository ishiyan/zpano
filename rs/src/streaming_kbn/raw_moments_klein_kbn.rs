use super::klein_kbn_accumulator::KleinKbnAccumulator;

/// Streaming mean, variance, skewness, and kurtosis via raw power sums (x¹..x⁴)
/// with KBN double-compensated accumulation.
///
/// Accumulates Σx, Σx², Σx³, Σx⁴ using `KleinKbnAccumulator` for each, plus a
/// separate Welford-style variance tracker (also KBN-compensated). Raw sums are
/// converted to central moments at query time.
///
/// Supports both LIFO revert and FIFO rolling window (via revert/update cycle)
/// because subtracting from a linear sum preserves the KBN compensation state.
pub struct RawMomentsKleinKbn {
    n: usize,
    x1: KleinKbnAccumulator,
    x2: KleinKbnAccumulator,
    x3: KleinKbnAccumulator,
    x4: KleinKbnAccumulator,
    mean: KleinKbnAccumulator,
    s: KleinKbnAccumulator,
    ddof: usize,
    bias: bool,
    fisher: bool,
}

impl RawMomentsKleinKbn {
    pub fn new(ddof: usize, bias: bool, fisher: bool) -> Self {
        Self {
            n: 0,
            x1: KleinKbnAccumulator::new(),
            x2: KleinKbnAccumulator::new(),
            x3: KleinKbnAccumulator::new(),
            x4: KleinKbnAccumulator::new(),
            mean: KleinKbnAccumulator::new(),
            s: KleinKbnAccumulator::new(),
            ddof,
            bias,
            fisher,
        }
    }

    /// Clears all accumulated state.
    pub fn reset(&mut self) {
        self.n = 0;
        self.x1.reset();
        self.x2.reset();
        self.x3.reset();
        self.x4.reset();
        self.mean.reset();
        self.s.reset();
    }

    /// Adds a new sample `x` to the accumulator.
    pub fn update(&mut self, x: f64) {
        self.n += 1;
        self.x1.update(x);
        let x2 = x * x;
        self.x2.update(x2);
        let x3 = x2 * x;
        self.x3.update(x3);
        let x4 = x3 * x;
        self.x4.update(x4);

        let n = self.n as f64;
        let delta = x - self.mean.value();
        self.mean.update(delta / n);
        self.s.update(delta * (x - self.mean.value()));
    }

    /// Removes a previously added sample `x` from the accumulator.
    pub fn revert(&mut self, x: f64) {
        self.n -= 1;
        self.x1.revert(x);
        let x2 = x * x;
        self.x2.revert(x2);
        let x3 = x2 * x;
        self.x3.revert(x3);
        let x4 = x3 * x;
        self.x4.revert(x4);

        let delta = x - self.mean.value();
        let n = self.n as f64;
        self.mean.revert(delta / n);
        self.s.revert(delta * (x - self.mean.value()));
    }

    /// Returns the current arithmetic mean.
    pub fn mean(&self) -> f64 {
        self.mean.value()
    }

    /// Returns the current variance.
    /// Returns NaN if `n <= ddof`.
    pub fn variance(&mut self) -> f64 {
        let n = (self.n as f64) - (self.ddof as f64);
        if n <= 0.0 {
            return f64::NAN;
        }
        let sv = self.s.value();
        if sv < 0.0 {
            self.s.reset();
            return f64::NAN;
        }
        sv / n
    }

    /// Returns the current standard deviation.
    /// Returns NaN if `n <= ddof`.
    pub fn standard_deviation(&self) -> f64 {
        let n = (self.n as f64) - (self.ddof as f64);
        if n <= 0.0 {
            return f64::NAN;
        }
        (self.s.value() / n).sqrt()
    }

    /// Returns the current skewness.
    /// Returns NaN if `n < 3`.
    pub fn skewness(&self) -> f64 {
        let nu = self.n;
        if nu < 3 {
            return f64::NAN;
        }
        let n = nu as f64;
        let a = self.x1.value() / n;
        let b = self.x2.value() / n - a * a;
        if b <= 1e-14 {
            return f64::NAN;
        }
        let r = b.sqrt();
        let c = self.x3.value() / n - a * a * a - 3.0 * a * b;
        let g1 = c / (r * r * r);
        if self.bias {
            g1
        } else {
            g1 * (n * (n - 1.0)).sqrt() / (n - 2.0)
        }
    }

    /// Returns the current kurtosis.
    /// Returns NaN if `n < 4`.
    pub fn kurtosis(&self) -> f64 {
        let nu = self.n;
        if nu < 4 {
            return f64::NAN;
        }
        let n = nu as f64;
        let a = self.x1.value() / n;
        let r = a * a;
        let b = self.x2.value() / n - r;
        if b <= 1e-14 {
            return f64::NAN;
        }
        let r = r * a;
        let c = self.x3.value() / n - r - 3.0 * a * b;
        let r = r * a;
        let d = self.x4.value() / n - r - 6.0 * b * a * a - 4.0 * c * a;
        let raw = d / (b * b);
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
        let mut m = RawMomentsKleinKbn::new(0, true, true);
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
    fn test_standard_deviation() {
        let mut m = RawMomentsKleinKbn::new(0, true, true);
        for &x in &[1.0, 2.0, 3.0, 4.0] {
            m.update(x);
        }
        let expected = m.variance().sqrt();
        assert!(almost_equal(m.standard_deviation(), expected, 1e-15),
            "std = {}, want {}", m.standard_deviation(), expected);
    }

    #[test]
    fn test_ddof() {
        let mut m = RawMomentsKleinKbn::new(1, true, true);
        for &x in &[1.0, 2.0, 3.0] {
            m.update(x);
        }
        assert!(almost_equal(m.variance(), 1.0, 1e-15),
            "var = {}, want 1.0", m.variance());
    }

    #[test]
    fn test_compare_scipy() {
        let mut m = RawMomentsKleinKbn::new(0, true, true);
        for &x in &BACON_DATA {
            m.update(x);
        }
        assert!(almost_equal(m.mean(), 0.009000000000000001, 1e-15),
            "mean = {}, want 0.009000000000000001", m.mean());
        assert!(almost_equal(m.variance(), 0.0014989166666666666, 1e-14),
            "var = {}, want 0.0014989166666666666", m.variance());
        assert!(almost_equal(m.skewness(), -0.08256245520856798, 1e-14),
            "skew = {}, want -0.08256245520856798", m.skewness());
        assert!(almost_equal(m.kurtosis(), -0.5675462058921257, 1e-13),
            "kurt = {}, want -0.5675462058921257", m.kurtosis());
    }

    #[test]
    fn test_compare_scipy_bias_false() {
        let mut m = RawMomentsKleinKbn::new(0, false, true);
        for &x in &BACON_DATA {
            m.update(x);
        }
        assert!(almost_equal(m.skewness(), -0.08817174934967527, 1e-14),
            "skew = {}, want -0.08817174934967527", m.skewness());
        assert!(almost_equal(m.kurtosis(), -0.40766032118608714, 1e-13),
            "kurt = {}, want -0.40766032118608714", m.kurtosis());
    }

    #[test]
    fn test_revert_roundtrip() {
        let data = [1.0, 2.0, 3.0, 4.0, 5.0];
        let mut m = RawMomentsKleinKbn::new(0, true, true);
        for &x in &data {
            m.update(x);
        }
        for i in (0..data.len() - 1).rev() {
            m.revert(data[i + 1]);
        }
        assert_eq!(m.n, 1, "n = {}, want 1", m.n);
        assert!(almost_equal(m.mean(), 1.0, 1e-15),
            "mean = {}, want 1.0", m.mean());
    }

    #[test]
    fn test_revert_partial() {
        let data = [10.0, 18.0, 5.0, 12.0, 7.0];

        let mut full = RawMomentsKleinKbn::new(0, true, true);
        let mut part = RawMomentsKleinKbn::new(0, true, true);
        for &x in &data {
            full.update(x);
        }
        for &x in &data[..4] {
            part.update(x);
        }
        full.revert(data[4]);

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
    fn test_reset() {
        let mut m = RawMomentsKleinKbn::new(0, true, true);
        m.update(10.0);
        m.reset();
        assert_eq!(m.n, 0, "n = {}, want 0", m.n);
        assert!(m.variance().is_nan(), "var = {}, want NaN", m.variance());
    }
}
