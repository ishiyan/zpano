use super::klein_kbn_accumulator::KleinKbnAccumulator;

/// Counted compensated sum: a `KleinKbnAccumulator` plus a sample count,
/// which also gives the arithmetic mean.
///
/// See `KleinKbnAccumulator` for the summation algorithm.
///
/// Every call to update() increments the count, including calls with
/// x == 0; every call to revert() decrements it.  Because only sums are
/// stored, revert() may remove any previously added value (not only
/// the most recent one), so the summator works for FIFO rolling windows.
#[derive(Debug, Clone, Copy, Default)]
pub struct KleinKbnSummator {
    n: usize,
    sum: KleinKbnAccumulator,
}

impl KleinKbnSummator {
    /// Creates an empty summator.
    pub const fn new() -> Self {
        Self { n: 0, sum: KleinKbnAccumulator::new() }
    }

    /// Clears the count and the sum.
    pub fn reset(&mut self) {
        self.n = 0;
        self.sum.reset();
    }

    /// Removes a previously added value x.
    /// Removing the final sample clears the sum and its compensation terms.
    ///
    /// # Panics
    ///
    /// Panics if the summator is empty.
    pub fn revert(&mut self, x: f64) {
        if self.n == 0 {
            panic!("Cannot revert from an empty summator");
        }
        if self.n == 1 {
            self.reset();
            return;
        }
        self.n -= 1;
        // Adding zero leaves the accumulator unchanged, so skip it.
        if x != 0.0 {
            self.sum.revert(x);
        }
    }

    /// Adds a value x.
    pub fn update(&mut self, x: f64) {
        self.n += 1;
        // Adding zero leaves the accumulator unchanged, so skip it.
        if x != 0.0 {
            self.sum.update(x);
        }
    }

    /// The compensated sum of all added values (0.0 when empty).
    pub fn value(&self) -> f64 {
        self.sum.value()
    }

    /// The arithmetic mean, sum / n (NaN when empty).
    pub fn mean(&self) -> f64 {
        let n = self.n;
        if n == 0 {
            return f64::NAN;
        }
        self.sum.value() / n as f64
    }

    /// The number of added values.
    pub fn n(&self) -> usize {
        self.n
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::streaming_kbn::test_support::{almost_equal, fsum};

    #[test]
    fn test_empty() {
        let s = KleinKbnSummator::new();
        assert_eq!(s.n(), 0);
        assert_eq!(s.value(), 0.0);
        assert!(s.mean().is_nan());
        let d = KleinKbnSummator::default();
        assert_eq!(d.n(), 0);
        assert_eq!(d.value(), 0.0);
    }

    #[test]
    fn test_update() {
        let mut s = KleinKbnSummator::new();
        for x in [1.0, 2.0, 3.0, 4.0] {
            s.update(x);
        }
        assert_eq!(s.n(), 4);
        assert_eq!(s.value(), 10.0);
        assert_eq!(s.mean(), 2.5);
    }

    #[test]
    fn test_zero_is_counted() {
        let mut s = KleinKbnSummator::new();
        for x in [0.0, 3.0, 0.0] {
            s.update(x);
        }
        assert_eq!(s.n(), 3);
        assert_eq!(s.value(), 3.0);
        assert_eq!(s.mean(), 1.0);
    }

    #[test]
    fn test_compensated() {
        // Peters' example: naive summation yields 0.0.
        let mut s = KleinKbnSummator::new();
        for x in [1.0, 1e100, 1.0, -1e100] {
            s.update(x);
        }
        assert_eq!(s.value(), 2.0);
        assert_eq!(s.mean(), 0.5);
    }

    #[test]
    fn test_revert() {
        let mut s = KleinKbnSummator::new();
        for x in [1.0, 2.0, 0.0, 4.0] {
            s.update(x);
        }
        s.revert(0.0);
        assert_eq!(s.n(), 3);
        assert_eq!(s.value(), 7.0);
        s.revert(1.0); // not the most recent value
        assert_eq!(s.n(), 2);
        assert_eq!(s.value(), 6.0);
        assert_eq!(s.mean(), 3.0);
    }

    #[test]
    fn test_rolling_window() {
        let data = [0.003, 0.026, 0.011, -0.010, 0.015, 0.025, 0.016, 0.067];
        let w = 3usize;
        let mut s = KleinKbnSummator::new();
        for (i, &x) in data.iter().enumerate() {
            s.update(x);
            if i >= w {
                s.revert(data[i - w]);
            }
            let window = &data[(i + 1).saturating_sub(w)..i + 1];
            assert_eq!(s.n(), window.len());
            assert!(
                almost_equal(s.value(), fsum(window), 17),
                "step {}: {} vs {}",
                i,
                s.value(),
                fsum(window)
            );
        }
    }

    #[test]
    fn test_revert_to_empty() {
        let mut s = KleinKbnSummator::new();
        s.update(5.0);
        s.revert(5.0);
        assert_eq!(s.n(), 0);
        assert_eq!(s.value(), 0.0);
        assert!(s.mean().is_nan());
    }

    #[test]
    #[should_panic(expected = "Cannot revert from an empty summator")]
    fn test_revert_empty_raises() {
        let mut s = KleinKbnSummator::new();
        s.revert(1.0);
    }

    #[test]
    fn test_reset() {
        let mut s = KleinKbnSummator::new();
        for x in [1.0, 2.0] {
            s.update(x);
        }
        s.reset();
        assert_eq!(s.n(), 0);
        assert_eq!(s.value(), 0.0);
        assert!(s.mean().is_nan());
        s.update(3.0);
        assert_eq!(s.mean(), 3.0);
    }

    #[test]
    fn test_revert_to_empty_clears_compensation() {
        for final_zero in [false, true] {
            let mut s = KleinKbnSummator::new();
            for x in [0.1, 1e16, 1e32, 1e48] {
                s.update(x);
            }
            if final_zero {
                s.update(0.0);
            }
            for x in [1e16, 0.1, 1e32, 1e48] {
                s.revert(x);
            }
            if final_zero {
                assert_eq!(s.n(), 1);
                s.revert(0.0);
            }
            assert_eq!(s.n(), 0);
            assert_eq!(s.value(), 0.0);
            assert!(s.mean().is_nan());
            s.update(3.0);
            assert_eq!(s.n(), 1);
            assert_eq!(s.value(), 3.0);
            assert_eq!(s.mean(), 3.0);
        }
    }
}
