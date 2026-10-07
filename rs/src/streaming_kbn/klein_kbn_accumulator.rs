// Klein second-order Kahan-Babuška-Neumaier (KBN) compensated summation.
//
// Kahan (1965) introduced single-level compensated summation.
// Neumaier (1974) improved it with a branch on |sum| >= |x|
// (the KBN algorithm proper).  Klein (2006) generalised KBN
// to arbitrary order; this is the second-order variant, which
// applies the same KBN trick to the correction term itself.
//
// Level 1 (KBN):      t = sum + x
//                     if |sum| >= |x|: c = (sum - t) + x
//                     else:            c = (x - t) + sum
//                     sum = t
// Level 2 (Klein):    t = cs + c
//                     if |cs| >= |c|:  cc = (cs - t) + c
//                     else:            cc = (c - t) + cs
//                     cs = t
//                     ccs += cc
//
// The corrected sum is: sum + cs + ccs.
//
// References:
//   A. Klein, "A Generalized Kahan-Babuška-Summation-Algorithm",
//     Computing 76, 279-293 (2006).
//   https://github.com/kuiperzone/Compensated-Accumulators
//   https://en.wikipedia.org/wiki/Kahan_summation_algorithm

/// Klein second-order Kahan-Babuška-Neumaier (KBN) floating-point accumulator.
///
/// Maintains three terms whose sum is the corrected total:
///
/// - `sum`: the primary (naive) running sum;
/// - `cs`:  the running sum of first-level KBN corrections;
/// - `ccs`: the running sum of second-level corrections, i.e. the
///   rounding errors made while accumulating `cs` (Klein's
///   generalisation).
///
/// Unlike naive summation, KBN correctly sums sequences with extreme
/// magnitude differences (e.g. Peters' example [1.0, 1e100, 1.0, -1e100]
/// → 2.0, while naive and standard Kahan summation return 0.0).
///
/// Level 1 (Kahan-Babuška-Neumaier):
///
/// ```text
/// t = sum + x
/// if |sum| >= |x|:  c = (sum - t) + x
/// else:             c = (x - t) + sum
/// sum = t
/// ```
///
/// The branch makes sure the larger operand comes first, so the
/// expression recovers exactly the low-order bits that were lost
/// when rounding `sum + x` to `t`.
///
/// Level 2 (Klein generalisation) applies the same technique to the
/// addition `cs + c` and accumulates its rounding error `cc`
/// into `ccs`.
///
/// The accumulator only stores sums, so `revert(x)` (adding `-x`)
/// removes any previously added value, not only the most recent one.
/// This makes it suitable for FIFO rolling windows.
#[derive(Debug, Clone, Copy, Default)]
pub struct KleinKbnAccumulator {
    sum: f64,
    cs: f64,
    ccs: f64,
}

impl KleinKbnAccumulator {
    /// Creates a new accumulator with the value zero.
    pub const fn new() -> Self {
        Self { sum: 0.0, cs: 0.0, ccs: 0.0 }
    }

    /// Sets the accumulator to zero.
    pub fn reset(&mut self) {
        self.set(0.0);
    }

    /// Overwrites the accumulated value with x and clears both
    /// compensation terms.
    ///
    /// Prefer set() over constructing a new instance when the
    /// accumulator is stored in a struct field.
    pub fn set(&mut self, x: f64) {
        self.sum = x;
        self.cs = 0.0;
        self.ccs = 0.0;
    }

    /// Removes a previously added value x (equivalent to update(-x)).
    pub fn revert(&mut self, x: f64) {
        self.update(-x);
    }

    /// Adds x to the accumulator.
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
        self.ccs += cc;
    }

    /// The compensated sum of all added values.
    pub fn value(&self) -> f64 {
        self.sum + self.cs + self.ccs
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::streaming_kbn::test_support::{almost_equal, fsum, Rng};

    fn naive_sum(data: &[f64]) -> f64 {
        let mut s = 0.0;
        for &x in data {
            s += x;
        }
        s
    }

    fn kbn_sum(data: &[f64]) -> f64 {
        let mut kbn = KleinKbnAccumulator::new();
        for &x in data {
            kbn.update(x);
        }
        kbn.value()
    }

    // https://en.wikipedia.org/wiki/Kahan_summation_algorithm
    // A simple example due to Peters: summing [1.0, +1e100, 1.0, -1e100]
    // in double precision, Kahan's algorithm yields 0.0, whereas
    // Neumaier's algorithm yields the correct value 2.0.
    const PETERS_DATA: [f64; 4] = [1.0, 1e100, 1.0, -1e100];

    // https://github.com/numpy/numpy/issues/8786
    // A badly conditioned sum, condition number ~2.188e+14.
    const NUMPY_DATA: [f64; 12] = [
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
    const NUMPY_EXPECTED: f64 = -0.377392919181026;

    // A sequence where the second-level correction is non-zero at more
    // than one step, so it must be accumulated (ccs += cc), not
    // overwritten (ccs = cc).  Overwriting yields -1.0000000000000001e-16.
    const KLEIN_DATA: [f64; 8] = [1e-16, -1e16, 1.0, 1e-16, -1.0, -1e-16, -1e-32, 1e16];
    const KLEIN_EXPECTED: f64 = 9.999999999999999e-17; // math.fsum(klein_data)

    #[test]
    fn test_initial_value_is_zero() {
        assert_eq!(KleinKbnAccumulator::new().value(), 0.0);
        assert_eq!(KleinKbnAccumulator::default().value(), 0.0);
    }

    #[test]
    fn test_peters() {
        assert_eq!(naive_sum(&PETERS_DATA), 0.0);
        assert_eq!(kbn_sum(&PETERS_DATA), 2.0);
    }

    #[test]
    fn test_numpy_issue() {
        assert!(almost_equal(kbn_sum(&NUMPY_DATA), NUMPY_EXPECTED, 16));
        assert!(!almost_equal(naive_sum(&NUMPY_DATA), NUMPY_EXPECTED, 3));
    }

    #[test]
    fn test_second_level_correction_is_accumulated() {
        assert_eq!(kbn_sum(&KLEIN_DATA), KLEIN_EXPECTED);
        assert_eq!(fsum(&KLEIN_DATA), KLEIN_EXPECTED);
    }

    #[test]
    fn test_matches_fsum_on_mixed_magnitudes() {
        let mut rng = Rng::new(42);
        for _ in 0..200 {
            let data: Vec<f64> = (0..100)
                .map(|_| {
                    let u = rng.uniform(-1.0, 1.0);
                    u * rng.choice(&[1e-8, 1.0, 1e8])
                })
                .collect();
            assert_eq!(kbn_sum(&data), fsum(&data));
        }
    }

    #[test]
    fn test_better_accuracy_than_naive() {
        // Add and then subtract the same values, so the exact sum is 0.
        let mut rng = Rng::new(42);
        let mut data: Vec<f64> = (0..100_000).map(|_| rng.uniform(0.0, 1e7)).collect();
        let negated: Vec<f64> = data.iter().map(|&x| -x).collect();
        data.extend(negated);
        let k = kbn_sum(&data);
        let v = naive_sum(&data);
        assert_eq!(k, 0.0);
        assert_ne!(v, 0.0);
    }

    #[test]
    fn test_update_zero_keeps_compensation() {
        let mut kbn = KleinKbnAccumulator::new();
        for &x in &KLEIN_DATA {
            kbn.update(x);
        }
        kbn.update(0.0);
        assert_eq!(kbn.value(), KLEIN_EXPECTED);
    }

    #[test]
    fn test_revert() {
        let mut kbn = KleinKbnAccumulator::new();
        kbn.update(1.5);
        kbn.update(2.5);
        kbn.revert(2.5);
        assert_eq!(kbn.value(), 1.5);
        kbn.revert(1.5);
        assert_eq!(kbn.value(), 0.0);
    }

    #[test]
    fn test_revert_not_most_recent() {
        let mut kbn = KleinKbnAccumulator::new();
        for &x in &PETERS_DATA {
            kbn.update(x);
        }
        kbn.revert(1e100); // not the most recent value
        assert_eq!(kbn.value(), 2.0 - 1e100);
        kbn.revert(-1e100);
        assert_eq!(kbn.value(), 2.0);
    }

    #[test]
    fn test_revert_restores_compensated_sum() {
        let mut kbn = KleinKbnAccumulator::new();
        for &x in &NUMPY_DATA {
            kbn.update(x);
        }
        kbn.update(1e20);
        kbn.revert(1e20);
        assert!(almost_equal(kbn.value(), NUMPY_EXPECTED, 16));
    }

    #[test]
    fn test_set() {
        let mut kbn = KleinKbnAccumulator::new();
        for &x in &PETERS_DATA {
            kbn.update(x);
        }
        kbn.set(5.0);
        assert_eq!(kbn.value(), 5.0);
        // Compensation terms are cleared, so only the new values count.
        kbn.update(1e100);
        kbn.update(-1e100);
        assert_eq!(kbn.value(), 5.0);
    }

    #[test]
    fn test_reset() {
        let mut kbn = KleinKbnAccumulator::new();
        for &x in &PETERS_DATA {
            kbn.update(x);
        }
        kbn.reset();
        assert_eq!(kbn.value(), 0.0);
        kbn.update(1.5);
        assert_eq!(kbn.value(), 1.5);
    }
}
