//! Streaming (one-pass, O(1) per sample) statistics with Klein second-order
//! Kahan-Babuška-Neumaier (KBN) compensated summation.
//!
//! Uses only the Rust standard library to keep the algorithms portable.
//!
//! Inputs are assumed to be finite floats. The module does not validate NaN or
//! infinity, or recover when an intermediate product overflows.
//!
//! - [`KleinKbnAccumulator`]: compensated sum; supports set() and revert().
//! - [`KleinKbnSummator`]: compensated sum plus sample count and mean.
//! - [`RawMomentsKleinKbn`]: mean, variance, skewness and kurtosis from raw power
//!   sums; revert() removes any previously added sample (FIFO rolling windows).
//! - [`CentralMomentsKleinKbn`]: mean, variance, skewness and kurtosis from Pébay's
//!   central moment updates; more accurate for data with a large mean and
//!   supports removal of any previously added sample.
//! - [`LinearRegressionKleinKbn`]: OLS slope, intercept, correlation and covariance;
//!   revert() removes any previously added sample.

pub mod klein_kbn_accumulator;
pub mod klein_kbn_summator;
pub mod raw_moments_klein_kbn;
pub mod central_moments_klein_kbn;
pub mod linear_regression_klein_kbn;

pub use klein_kbn_accumulator::KleinKbnAccumulator;
pub use klein_kbn_summator::KleinKbnSummator;
pub use raw_moments_klein_kbn::RawMomentsKleinKbn;
pub use central_moments_klein_kbn::CentralMomentsKleinKbn;
pub use linear_regression_klein_kbn::LinearRegressionKleinKbn;

/// Shared test helpers (Python `assertAlmostEqual`, `math.fsum`, and a
/// deterministic PRNG standing in for `random.Random`).
#[cfg(test)]
pub(crate) mod test_support {
    /// Python `assertAlmostEqual(a, b, places=p)`: `round(|a - b|, p) == 0`.
    pub fn almost_equal(a: f64, b: f64, places: i32) -> bool {
        a == b || (a - b).abs() <= 0.5 * 10f64.powi(-places)
    }

    /// Faithful port of CPython's `math.fsum` (Shewchuk partials with the
    /// final half-even rounding correction), for finite inputs.
    pub fn fsum(data: &[f64]) -> f64 {
        let mut p: Vec<f64> = Vec::new();
        for &xx in data {
            let mut x = xx;
            let mut i = 0usize;
            for j in 0..p.len() {
                let mut y = p[j];
                if x.abs() < y.abs() {
                    std::mem::swap(&mut x, &mut y);
                }
                let hi = x + y;
                let lo = y - (hi - x);
                if lo != 0.0 {
                    p[i] = lo;
                    i += 1;
                }
                x = hi;
            }
            p.truncate(i);
            if x != 0.0 {
                p.push(x);
            }
        }
        let mut n = p.len();
        let mut hi = 0.0;
        let mut lo = 0.0;
        if n > 0 {
            n -= 1;
            hi = p[n];
            while n > 0 {
                let x = hi;
                n -= 1;
                let y = p[n];
                hi = x + y;
                let yr = hi - x;
                lo = y - yr;
                if lo != 0.0 {
                    break;
                }
            }
            if n > 0 && ((lo < 0.0 && p[n - 1] < 0.0) || (lo > 0.0 && p[n - 1] > 0.0)) {
                let y = lo * 2.0;
                let x = hi + y;
                let yr = x - hi;
                if y == yr {
                    hi = x;
                }
            }
        }
        hi
    }

    /// SplitMix64 PRNG producing doubles in [0, 1); stands in for
    /// Python's `random.Random(42)` (the sequences differ, the
    /// distributions are the same).
    pub struct Rng(u64);

    impl Rng {
        pub fn new(seed: u64) -> Self {
            Self(seed)
        }

        pub fn next_u64(&mut self) -> u64 {
            self.0 = self.0.wrapping_add(0x9e3779b97f4a7c15);
            let mut z = self.0;
            z = (z ^ (z >> 30)).wrapping_mul(0xbf58476d1ce4e5b9);
            z = (z ^ (z >> 27)).wrapping_mul(0x94d049bb133111eb);
            z ^ (z >> 31)
        }

        /// A double uniformly distributed in [0, 1).
        pub fn random(&mut self) -> f64 {
            (self.next_u64() >> 11) as f64 * (1.0 / (1u64 << 53) as f64)
        }

        /// Python `random.uniform(a, b)` = a + (b - a) * random().
        pub fn uniform(&mut self, a: f64, b: f64) -> f64 {
            a + (b - a) * self.random()
        }

        /// Python `random.choice(items)`.
        pub fn choice<T: Copy>(&mut self, items: &[T]) -> T {
            items[(self.next_u64() % items.len() as u64) as usize]
        }
    }

    #[test]
    fn test_fsum() {
        assert_eq!(fsum(&[]), 0.0);
        assert_eq!(fsum(&[1.0, 1e100, 1.0, -1e100]), 2.0);
        assert_eq!(fsum(&[0.1; 10]), 1.0);
        // Half-even correction: 1 + 2^-53 + 2^-106 rounds up.
        assert_eq!(fsum(&[1.0, 2f64.powi(-53), 2f64.powi(-106)]), 1.0 + 2f64.powi(-52));
    }
}
