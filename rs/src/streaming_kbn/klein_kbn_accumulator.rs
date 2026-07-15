/// Klein second-order Kahan-Babuška-Neumaier compensated summation accumulator.
///
/// Maintains `sum + cs + ccs` where `sum` is the primary sum, `cs` is the
/// first-level KBN correction, and `ccs` is a second-level KBN correction
/// applied to the first correction term (Klein's generalisation).
#[derive(Debug, Clone, Copy)]
pub struct KleinKbnAccumulator {
    sum: f64,
    cs: f64,
    ccs: f64,
}

impl KleinKbnAccumulator {
    pub fn new() -> Self {
        Self { sum: 0.0, cs: 0.0, ccs: 0.0 }
    }

    /// Overwrites the accumulator value and resets both compensation terms to zero.
    pub fn set(&mut self, x: f64) {
        self.sum = x;
        self.cs = 0.0;
        self.ccs = 0.0;
    }

    /// Resets the accumulator to zero.
    pub fn reset(&mut self) {
        self.set(0.0);
    }

    /// Returns the current compensated sum: `sum + cs + ccs`.
    pub fn value(&self) -> f64 {
        self.sum + self.cs + self.ccs
    }

    /// Adds `x` to the accumulator using Klein second-order KBN compensated summation.
    pub fn update(&mut self, x: f64) {
        let s = self.sum;
        let t = s + x;

        let c = if s.abs() >= x.abs() {
            (s - t) + x
        } else {
            (x - t) + s
        };
        self.sum = t;

        let cs = self.cs;
        let t = cs + c;
        let cc = if cs.abs() >= c.abs() {
            (cs - t) + c
        } else {
            (c - t) + cs
        };
        self.cs = t;
        self.ccs = cc;
    }

    /// Removes `x` from the accumulator by adding `-x`.
    pub fn revert(&mut self, x: f64) {
        self.update(-x);
    }
}

impl Default for KleinKbnAccumulator {
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

    struct NaiveSum(f64);

    impl NaiveSum {
        fn new() -> Self {
            Self(0.0)
        }
        fn reset(&mut self) {
            self.0 = 0.0;
        }
        fn set(&mut self, x: f64) {
            self.0 = x;
        }
        fn update(&mut self, x: f64) {
            self.0 += x;
        }
        fn value(&self) -> f64 {
            self.0
        }
    }

    /// SplitMix64 PRNG — deterministic, no external deps.
    struct SimpleRng(u64);

    impl SimpleRng {
        fn new(seed: u64) -> Self {
            Self(seed)
        }
        fn next_f64(&mut self) -> f64 {
            self.0 = self.0.wrapping_add(0x9e3779b97f4a7c15);
            let mut z = self.0;
            z = (z ^ (z >> 30)).wrapping_mul(0xbf58476d1ce4e5b9);
            z = (z ^ (z >> 27)).wrapping_mul(0x94d049bb133111eb);
            z ^= z >> 31;
            (z >> 11) as f64 * (1.0 / (1u64 << 53) as f64)
        }
    }

    #[test]
    fn test_peters() {
        let data = [1.0, 1e100, 1.0, -1e100];
        let mut naive = NaiveSum::new();
        let mut kbn = KleinKbnAccumulator::new();
        for &x in &data {
            naive.update(x);
            kbn.update(x);
        }
        assert!(almost_equal(kbn.value(), 2.0, 1e-15),
            "KBN sum = {}, want 2.0", kbn.value());
        assert!(kbn.value().abs() > naive.value().abs(),
            "KBN sum {} not more accurate than naive sum {}", kbn.value(), naive.value());
    }

    #[test]
    fn test_numpy() {
        let data = [
            -0.41253261766461263,
            41287272281118.43,
            -1.4727977348624173e-14,
            5670.3302557520055,
            2.119245229045646e-11,
            -0.003679264134906428,
            -6.892634568678797e-14,
            -0.0006984744181630712,
            -4054136.048352595,
            -1003.101760720037,
            -1.4436349910427172e-17,
            -41287268231649.57,
        ];
        let expected = -0.377392919181026;
        let mut kbn = KleinKbnAccumulator::new();
        for &x in &data {
            kbn.update(x);
        }
        assert!(almost_equal(kbn.value(), expected, 1e-16),
            "KBN sum = {}, want {}", kbn.value(), expected);
    }

    #[test]
    fn test_better_accuracy_than_naive() {
        let spread = 1e7;
        let mut naive = NaiveSum::new();
        let mut kbn = KleinKbnAccumulator::new();

        let mut rng = SimpleRng::new(42);
        for _ in 0..1_000_000 {
            let x = rng.next_f64() * spread;
            naive.update(x);
            kbn.update(x);
        }

        let mut rng = SimpleRng::new(42);
        for _ in 0..1_000_000 {
            let x = rng.next_f64() * spread;
            naive.update(-x);
            kbn.update(-x);
        }

        assert!(kbn.value().abs() <= naive.value().abs(),
            "KBN sum {} is not more accurate than naive sum {}", kbn.value(), naive.value());
    }

    #[test]
    fn test_revert() {
        let mut kbn = KleinKbnAccumulator::new();
        assert!(almost_equal(kbn.value(), 0.0, 1e-15),
            "initial value = {}, want 0.0", kbn.value());

        kbn.update(1.5);
        kbn.update(2.5);
        kbn.revert(2.5);
        assert!(almost_equal(kbn.value(), 1.5, 1e-15),
            "after revert 2.5: {}, want 1.5", kbn.value());

        kbn.revert(1.5);
        assert!(almost_equal(kbn.value(), 0.0, 1e-15),
            "after revert 1.5: {}, want 0.0", kbn.value());
    }

    #[test]
    fn test_reset() {
        let mut kbn = KleinKbnAccumulator::new();
        kbn.update(1.5);
        kbn.reset();
        assert!(almost_equal(kbn.value(), 0.0, 1e-15),
            "after reset: {}, want 0.0", kbn.value());

        kbn.update(1.5);
        assert!(almost_equal(kbn.value(), 1.5, 1e-15),
            "after update 1.5: {}, want 1.5", kbn.value());
    }
}
