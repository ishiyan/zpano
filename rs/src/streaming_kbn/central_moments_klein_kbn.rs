// Central moments with Klein KBN (Kahan-Babuška-Neumaier)
// compensated summation for improved numerical stability.
//
// References:
//   P. Pébay, "Formulas for Robust, One-Pass Parallel Computation
//     of Covariances and Arbitrary-Order Statistical Moments",
//     Sandia Report SAND2008-6212 (2008).
//   https://www.johndcook.com/skewness_kurtosis.html
//   https://github.com/kuiperzone/Compensated-Accumulators

use super::klein_kbn_accumulator::KleinKbnAccumulator;

/// Streaming mean, variance, skewness, kurtosis via Pébay's central moment
/// update with Klein KBN (Kahan-Babuška-Neumaier) compensated accumulation.
///
/// Maintains the mean M₁ and the sums of central powers
///
/// ```text
/// M₂ = Σ(x - x̄)²,   M₃ = Σ(x - x̄)³,   M₄ = Σ(x - x̄)⁴
/// ```
///
/// (each as a `KleinKbnAccumulator`), updated in O(1) per sample.
/// The population central moments are μₖ = Mₖ / n.
///
/// Avoids the catastrophic cancellation inherent in converting raw power
/// sums Σxᵏ to central moments.  This matters for data with a large mean
/// relative to its spread.  Inverse updates can remove any previously
/// added sample, including the oldest sample in a FIFO rolling window.
/// Reversion clears the compensation terms, so repeated removals can
/// accumulate rounding error.
///
/// # Parameters
///
/// - `ddof` (default 1): delta degrees of freedom for variance.
///   variance = M₂ / (n - ddof).  ddof=0 gives population, ddof=1 gives sample.
/// - `bias` (default true): if true, return the biased (population) skewness
///   and kurtosis.  If false, apply the bias corrections (see Notes).
/// - `fisher` (default true): if true, return excess kurtosis (subtract 3 so
///   Gaussian→0).  If false, return raw (Pearson) kurtosis (Gaussian→3).
///   Applied after the bias correction when bias=false.
///
/// # Notes
///
/// The results match scipy.stats.skew(bias=...) and
/// scipy.stats.kurtosis(bias=..., fisher=...).
///
/// ```text
/// Skewness (bias=true), requires n ≥ 2:
///     g₁ = μ₃ / μ₂^1.5 = √n · M₃ / M₂^1.5
///
/// Skewness (bias=false), requires n ≥ 3:
///     G₁ = g₁ · √(n·(n-1)) / (n-2)
///
/// Kurtosis (bias=true), requires n ≥ 2:
///     β₂ = μ₄ / μ₂² = n · M₄ / M₂²
///     fisher=true:  g₂ = β₂ - 3
///     fisher=false: β₂
///
/// Kurtosis (bias=false), requires n ≥ 4:
///     G₂ = ((n²-1) · β₂  -  3·(n-1)²) / ((n-2)·(n-3))
///     fisher=true:  G₂
///     fisher=false: G₂ + 3
/// ```
///
/// Skewness and kurtosis are NaN when M₂ = 0 (constant data).
#[derive(Debug, Clone)]
pub struct CentralMomentsKleinKbn {
    ddof: usize,
    bias: bool,
    fisher: bool,
    n: usize,
    m1: KleinKbnAccumulator,
    m2: KleinKbnAccumulator,
    m3: KleinKbnAccumulator,
    m4: KleinKbnAccumulator,
}

impl CentralMomentsKleinKbn {
    /// Creates an empty instance with the given ddof, bias and fisher.
    pub fn new(ddof: usize, bias: bool, fisher: bool) -> Self {
        Self {
            ddof,
            bias,
            fisher,
            n: 0,
            m1: KleinKbnAccumulator::new(),
            m2: KleinKbnAccumulator::new(),
            m3: KleinKbnAccumulator::new(),
            m4: KleinKbnAccumulator::new(),
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

    /// Adds a sample x using Pébay's update (n = count after adding x):
    ///
    /// ```text
    /// δ    = x − M₁
    /// δₙ   = δ / n
    /// term = δ · δₙ · (n − 1)
    ///
    /// M₁ += δₙ
    /// M₄ += term·δₙ²·(n²−3n+3) + 6·δₙ²·M₂ − 4·δₙ·M₃
    /// M₃ += term·δₙ·(n−2) − 3·δₙ·M₂
    /// M₂ += term
    /// ```
    ///
    /// M₄ and M₃ are updated before M₂ and M₃ respectively, because
    /// they use the values from before x was added.
    pub fn update(&mut self, x: f64) {
        let n_old = self.n;
        let n_new = n_old + 1;
        self.n = n_new;
        let n_old = n_old as f64;
        let n_new = n_new as f64;
        let delta = x - self.m1.value();
        let delta_n = delta / n_new;
        let delta_n2 = delta_n * delta_n;
        let term = delta * delta_n * n_old;
        let m2 = self.m2.value();
        let m3 = self.m3.value();
        self.m1.update(delta_n);
        self.m4.update(
            term * delta_n2 * (n_new * n_new - 3.0 * n_new + 3.0) + 6.0 * delta_n2 * m2
                - 4.0 * delta_n * m3,
        );
        self.m3.update(term * delta_n * (n_new - 2.0) - 3.0 * delta_n * m2);
        self.m2.update(term);
    }

    /// Removes a previously added sample x, regardless of insertion order.
    /// Reverting a value that was never added corrupts the state.
    ///
    /// The restored M₁–M₄ are written with `KleinKbnAccumulator::set()`, which
    /// clears their compensation terms.  Subsequent updates rebuild the
    /// compensation from the restored values.  Repeated reverts can
    /// accumulate rounding error, especially for large-offset data.
    ///
    /// Inverse formulas (where nₙ = count before revert, nₒ = nₙ − 1):
    ///
    /// ```text
    /// M₁_old = (nₙ · M₁_new − x) / nₒ            [mean undo]
    /// δ      = x − M₁_old
    /// δₙ     = δ / nₙ
    /// term   = δ · δₙ · nₒ
    ///
    /// M₂_old = M₂_new − term
    /// M₃_old = M₃_new − (term·δₙ·(nₙ−2) − 3·δₙ·M₂_old)
    /// M₄_old = M₄_new − (term·δₙ²·(nₙ²−3nₙ+3)
    ///                     + 6·δₙ²·M₂_old − 4·δₙ·M₃_old)
    /// ```
    ///
    /// # Panics
    ///
    /// Panics if there are no samples.
    pub fn revert(&mut self, x: f64) {
        let n_new = self.n;
        if n_new == 0 {
            panic!("Cannot revert from an empty accumulator");
        }
        let n_old = n_new - 1;
        if n_old == 0 {
            self.reset();
            return;
        }

        let m1_new = self.m1.value();
        let m2_new = self.m2.value();
        let m3_new = self.m3.value();
        let m4_new = self.m4.value();

        let nn = n_new as f64;
        let no = n_old as f64;
        let m1_old = (nn * m1_new - x) / no;
        let delta = x - m1_old;
        let delta_n = delta / nn;
        let delta_n2 = delta_n * delta_n;
        let term = delta * delta_n * no;

        let m2_old = m2_new - term;
        let m3_old = m3_new - (term * delta_n * (nn - 2.0) - 3.0 * delta_n * m2_old);
        let m4_old = m4_new
            - (term * delta_n2 * (nn * nn - 3.0 * nn + 3.0) + 6.0 * delta_n2 * m2_old
                - 4.0 * delta_n * m3_old);

        self.n = n_old;
        self.m1.set(m1_old);
        self.m2.set(m2_old);
        self.m3.set(m3_old);
        self.m4.set(m4_old);
    }

    /// Delta degrees of freedom used by variance.
    pub fn ddof(&self) -> usize {
        self.ddof
    }

    /// Sets the delta degrees of freedom.
    pub fn set_ddof(&mut self, ddof: usize) {
        self.ddof = ddof;
    }

    /// If true, skewness and kurtosis are biased (population).
    pub fn bias(&self) -> bool {
        self.bias
    }

    /// Sets the bias parameter.
    pub fn set_bias(&mut self, bias: bool) {
        self.bias = bias;
    }

    /// If true, kurtosis is excess kurtosis.
    pub fn fisher(&self) -> bool {
        self.fisher
    }

    /// Sets the fisher parameter.
    pub fn set_fisher(&mut self, fisher: bool) {
        self.fisher = fisher;
    }

    /// The number of samples.
    pub fn n(&self) -> usize {
        self.n
    }

    /// The arithmetic mean (0.0 when empty).
    pub fn mean(&self) -> f64 {
        self.m1.value()
    }

    /// The variance M₂ / (n - ddof), NaN when n ≤ ddof.
    ///
    /// A slightly negative M₂ caused by rounding after revert()
    /// is clamped to zero.
    pub fn variance(&self) -> f64 {
        if self.n <= self.ddof {
            return f64::NAN;
        }
        let d = self.n - self.ddof;
        let m2 = self.m2.value();
        // Python's max(m2, 0.0).
        let m2 = if 0.0 > m2 { 0.0 } else { m2 };
        m2 / d as f64
    }

    /// The square root of the variance, NaN when n ≤ ddof.
    pub fn standard_deviation(&self) -> f64 {
        let v = self.variance();
        if v.is_nan() {
            v
        } else {
            v.sqrt()
        }
    }

    /// The skewness g₁ (bias=true) or G₁ (bias=false); see the type Notes.
    pub fn skewness(&self) -> f64 {
        let n = self.n;
        let m2 = self.m2.value();
        if n < 2 || m2 <= 0.0 {
            return f64::NAN;
        }
        let nf = n as f64;
        let g1 = nf.sqrt() * self.m3.value() / (m2 * m2.sqrt());
        if self.bias {
            return g1;
        }
        if n < 3 {
            return f64::NAN;
        }
        g1 * (nf * (nf - 1.0)).sqrt() / (nf - 2.0)
    }

    /// The kurtosis selected by bias and fisher; see the type Notes.
    pub fn kurtosis(&self) -> f64 {
        let n = self.n;
        let m2 = self.m2.value();
        if n < 2 || m2 <= 0.0 {
            return f64::NAN;
        }
        let nf = n as f64;
        let b2 = nf * self.m4.value() / (m2 * m2);
        if self.bias {
            return if self.fisher { b2 - 3.0 } else { b2 };
        }
        if n < 4 {
            return f64::NAN;
        }
        let g2 = ((nf * nf - 1.0) * b2 - 3.0 * (nf - 1.0) * (nf - 1.0)) / ((nf - 2.0) * (nf - 3.0));
        if self.fisher { g2 } else { g2 + 3.0 }
    }
}

impl Default for CentralMomentsKleinKbn {
    /// ddof=1, bias=true, fisher=true.
    fn default() -> Self {
        Self::new(1, true, true)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::streaming_kbn::test_support::{almost_equal, fsum};

    // Bacon, Carl R., Practical Portfolio Performance Measurement and
    // Attribution, 2nd ed., Wiley, 2008, p. 65 (portfolio returns).
    const BACON: [f64; 24] = [
        0.003, 0.026, 0.011, -0.010,
        0.015, 0.025, 0.016, 0.067,
        -0.014, 0.040, -0.005, 0.081,
        0.040, -0.037, -0.061, 0.017,
        -0.049, -0.022, 0.070, 0.058,
        -0.065, 0.024, -0.005, -0.009,
    ];

    // Reference values for BACON, computed with exact rational arithmetic
    // (fractions.Fraction on the binary float inputs, square roots with
    // 50-digit decimal.Decimal), rounded to the nearest float.  The names
    // follow scipy.stats: skew(bias=...), kurtosis(bias=..., fisher=...).
    const MEAN: f64 = 0.009000000000000001;
    const VARIANCE_DDOF_0: f64 = 0.0014989166666666668;
    const VARIANCE_DDOF_1: f64 = 0.0015640869565217393;
    const STD_DDOF_0: f64 = 0.03871584516275819;
    const STD_DDOF_1: f64 = 0.039548539246370897;
    const SKEW_BIASED: f64 = -0.08256245520856804; // skew(bias=True)
    const SKEW_UNBIASED: f64 = -0.08817174934967535; // skew(bias=False)
    const KURT_BIASED_FISHER: f64 = -0.5675462058921257; // kurtosis(bias=True, fisher=True)
    const KURT_BIASED_PEARSON: f64 = 2.4324537941078743; // kurtosis(bias=True, fisher=False)
    const KURT_UNBIASED_FISHER: f64 = -0.40766032118608714; // kurtosis(bias=False, fisher=True)
    const KURT_UNBIASED_PEARSON: f64 = 2.592339678813913; // kurtosis(bias=False, fisher=False)

    // Same statistics for [1e4 + x for x in BACON], exact for the shifted floats.
    const OFFSET_SKEW_BIASED: f64 = -0.08256245521966786;
    const OFFSET_KURT_BIASED_FISHER: f64 = -0.5675462058934164;

    fn feed(mut m: CentralMomentsKleinKbn, data: &[f64]) -> CentralMomentsKleinKbn {
        for &x in data {
            m.update(x);
        }
        m
    }

    fn with_ddof(ddof: usize) -> CentralMomentsKleinKbn {
        CentralMomentsKleinKbn::new(ddof, true, true)
    }

    /// Python `statistics.fmean`: fsum(data) / n.
    fn fmean(data: &[f64]) -> f64 {
        fsum(data) / data.len() as f64
    }

    /// Python `statistics.pvariance` (exact in Python); approximated here
    /// with compensated sums, accurate far beyond the tested tolerance.
    fn pvariance(data: &[f64]) -> f64 {
        let mean = fmean(data);
        let d: Vec<f64> = data.iter().map(|&x| (x - mean) * (x - mean)).collect();
        fsum(&d) / data.len() as f64
    }

    #[test]
    fn test_simple_update() {
        let m = feed(with_ddof(0), &[1.0, 2.0, 3.0, 4.0]);
        assert_eq!(m.n(), 4);
        assert!(almost_equal(m.mean(), 2.5, 15));
        assert!(almost_equal(m.variance(), 1.25, 15));
        assert!(almost_equal(m.skewness(), 0.0, 14));
        assert!(almost_equal(m.kurtosis(), -1.36, 13));
    }

    #[test]
    fn test_bacon_mean_variance() {
        let m0 = feed(with_ddof(0), &BACON);
        let m1 = feed(with_ddof(1), &BACON);
        assert!(almost_equal(m0.mean(), MEAN, 16));
        assert!(almost_equal(m0.variance(), VARIANCE_DDOF_0, 16));
        assert!(almost_equal(m1.variance(), VARIANCE_DDOF_1, 16));
        assert!(almost_equal(m0.standard_deviation(), STD_DDOF_0, 15));
        assert!(almost_equal(m1.standard_deviation(), STD_DDOF_1, 15));
    }

    #[test]
    fn test_bacon_skewness_kurtosis() {
        let cases = [
            (true, true, SKEW_BIASED, KURT_BIASED_FISHER),
            (true, false, SKEW_BIASED, KURT_BIASED_PEARSON),
            (false, true, SKEW_UNBIASED, KURT_UNBIASED_FISHER),
            (false, false, SKEW_UNBIASED, KURT_UNBIASED_PEARSON),
        ];
        for (bias, fisher, skew, kurt) in cases {
            let m = feed(CentralMomentsKleinKbn::new(0, bias, fisher), &BACON);
            assert!(almost_equal(m.skewness(), skew, 14), "bias={} fisher={}", bias, fisher);
            assert!(almost_equal(m.kurtosis(), kurt, 13), "bias={} fisher={}", bias, fisher);
        }
    }

    #[test]
    fn test_large_offset() {
        // Central moments don't suffer from the cancellation of raw power sums.
        let data: Vec<f64> = BACON.iter().map(|&x| 1e4 + x).collect();
        let m = feed(with_ddof(0), &data);
        assert!(almost_equal(m.mean(), 1e4 + MEAN, 11));
        assert!(almost_equal(m.variance(), VARIANCE_DDOF_0, 13));
        assert!(almost_equal(m.skewness(), OFFSET_SKEW_BIASED, 10));
        assert!(almost_equal(m.kurtosis(), OFFSET_KURT_BIASED_FISHER, 10));
    }

    #[test]
    fn test_scale_invariance() {
        let data: Vec<f64> = BACON.iter().map(|&x| x * 1e-6).collect();
        let m = feed(with_ddof(0), &data);
        assert!(almost_equal(m.skewness(), SKEW_BIASED, 14));
        assert!(almost_equal(m.kurtosis(), KURT_BIASED_FISHER, 13));
    }

    #[test]
    fn test_empty() {
        let m = CentralMomentsKleinKbn::default();
        assert_eq!(m.n(), 0);
        assert_eq!(m.mean(), 0.0);
        assert!(m.variance().is_nan());
        assert!(m.standard_deviation().is_nan());
        assert!(m.skewness().is_nan());
        assert!(m.kurtosis().is_nan());
        assert_eq!(m.ddof(), 1);
        assert!(m.bias());
        assert!(m.fisher());
    }

    #[test]
    fn test_ddof() {
        let m = feed(with_ddof(1), &[1.0, 2.0, 3.0]);
        assert!(almost_equal(m.variance(), 1.0, 15));
        assert!(almost_equal(m.standard_deviation(), 1.0, 15));
        let m = feed(with_ddof(1), &[1.0]);
        assert!(m.variance().is_nan());
        assert!(m.standard_deviation().is_nan());
    }

    // test_invalid_ddof is not ported: ddof is a usize, so negative and
    // non-integer values are rejected at compile time.

    #[test]
    fn test_setters() {
        let mut m = feed(with_ddof(0), &BACON);
        m.set_ddof(1);
        assert_eq!(m.ddof(), 1);
        assert!(almost_equal(m.variance(), VARIANCE_DDOF_1, 16));
        m.set_bias(false);
        m.set_fisher(false);
        assert!(!m.bias());
        assert!(!m.fisher());
        assert!(almost_equal(m.skewness(), SKEW_UNBIASED, 14));
        assert!(almost_equal(m.kurtosis(), KURT_UNBIASED_PEARSON, 13));
    }

    #[test]
    fn test_large_ddof() {
        for ddof in [usize::MAX, 1usize << (usize::BITS - 1)] {
            let mut m = CentralMomentsKleinKbn::new(ddof, true, true);
            assert!(m.variance().is_nan());
            assert!(m.standard_deviation().is_nan());
            m.update(1.0);
            m.update(3.0);
            assert!(m.variance().is_nan());
            assert!(m.standard_deviation().is_nan());
            m.set_ddof(0);
            assert_eq!(m.variance(), 1.0);
            assert_eq!(m.standard_deviation(), 1.0);
            m.set_ddof(ddof);
            assert!(m.variance().is_nan());
            assert!(m.standard_deviation().is_nan());
            m.set_ddof(1);
            assert_eq!(m.variance(), 2.0);
        }
    }

    #[test]
    fn test_minimum_sample_sizes() {
        let data = [1.0, 2.0, 4.0, 8.0];
        // (bias, fisher) -> minimum n for (skewness, kurtosis)
        let cases = [
            ((true, true), (2usize, 2usize)),
            ((true, false), (2, 2)),
            ((false, true), (3, 4)),
            ((false, false), (3, 4)),
        ];
        for ((bias, fisher), (skew_n, kurt_n)) in cases {
            let mut m = CentralMomentsKleinKbn::new(0, bias, fisher);
            for (i, &x) in data.iter().enumerate() {
                m.update(x);
                let n = i + 1;
                assert_eq!(m.skewness().is_nan(), n < skew_n, "bias={} fisher={} n={}", bias, fisher, n);
                assert_eq!(m.kurtosis().is_nan(), n < kurt_n, "bias={} fisher={} n={}", bias, fisher, n);
            }
        }
    }

    #[test]
    fn test_constant_data() {
        let m = feed(with_ddof(0), &[3.0; 5]);
        assert_eq!(m.mean(), 3.0);
        assert_eq!(m.variance(), 0.0);
        assert_eq!(m.standard_deviation(), 0.0);
        assert!(m.skewness().is_nan());
        assert!(m.kurtosis().is_nan());
    }

    #[test]
    fn test_revert_lifo_simple() {
        let data = [10.0, 18.0, 5.0];
        let mut m_full = feed(with_ddof(0), &data);
        let m_part = feed(with_ddof(0), &data[..2]);
        m_full.revert(data[2]);

        assert_eq!(m_full.n(), 2);
        assert!(almost_equal(m_full.mean(), m_part.mean(), 15));
        assert!(almost_equal(m_full.variance(), m_part.variance(), 15));
        assert!(almost_equal(m_full.skewness(), m_part.skewness(), 14));
        assert!(almost_equal(m_full.kurtosis(), m_part.kurtosis(), 13));
    }

    #[test]
    fn test_revert_lifo_bacon() {
        for (bias, fisher) in [(true, true), (false, false)] {
            let mut m_full = feed(CentralMomentsKleinKbn::new(0, bias, fisher), &BACON);
            let m_part = feed(CentralMomentsKleinKbn::new(0, bias, fisher), &BACON[..BACON.len() - 1]);
            m_full.revert(BACON[BACON.len() - 1]);

            assert!(almost_equal(m_full.mean(), m_part.mean(), 15));
            assert!(almost_equal(m_full.variance(), m_part.variance(), 15));
            assert!(almost_equal(m_full.skewness(), m_part.skewness(), 13));
            assert!(almost_equal(m_full.kurtosis(), m_part.kurtosis(), 12));
        }
    }

    #[test]
    fn test_revert_then_update() {
        let mut m = feed(with_ddof(0), &BACON);
        for &x in BACON[12..].iter().rev() {
            m.revert(x);
        }
        let m = feed(m, &BACON[12..]);
        assert!(almost_equal(m.mean(), MEAN, 15));
        assert!(almost_equal(m.variance(), VARIANCE_DDOF_0, 15));
        assert!(almost_equal(m.skewness(), SKEW_BIASED, 12));
        assert!(almost_equal(m.kurtosis(), KURT_BIASED_FISHER, 12));
    }

    #[test]
    fn test_revert_lifo_roundtrip() {
        let mut m = feed(with_ddof(0), &BACON);
        for &x in BACON.iter().rev() {
            m.revert(x);
        }
        assert_eq!(m.n(), 0);
        assert_eq!(m.mean(), 0.0);
        assert!(m.variance().is_nan());
    }

    #[test]
    fn test_revert_oldest_and_middle() {
        let mut data = vec![0.0, 1.0, 2.0, 4.0, 8.0];
        let mut m = feed(with_ddof(0), &data);
        for removed in [0.0, 2.0] {
            m.revert(removed);
            let pos = data.iter().position(|&v| v == removed).unwrap();
            data.remove(pos);
            let len = data.len() as f64;
            let mean = fmean(&data);
            let p2: Vec<f64> = data.iter().map(|&x| (x - mean).powf(2.0)).collect();
            let p3: Vec<f64> = data.iter().map(|&x| (x - mean).powf(3.0)).collect();
            let p4: Vec<f64> = data.iter().map(|&x| (x - mean).powf(4.0)).collect();
            let mu2 = fsum(&p2) / len;
            let mu3 = fsum(&p3) / len;
            let mu4 = fsum(&p4) / len;
            assert_eq!(m.n(), data.len());
            assert!(almost_equal(m.mean(), mean, 14));
            assert!(almost_equal(m.variance(), mu2, 14));
            assert!(almost_equal(m.skewness(), mu3 / mu2.powf(1.5), 13));
            assert!(almost_equal(m.kurtosis(), mu4 / mu2.powf(2.0) - 3.0, 13));
        }
    }

    #[test]
    fn test_fifo_rolling_window() {
        let mut m = with_ddof(0);
        let width = 6usize;
        for (i, &x) in BACON.iter().enumerate() {
            m.update(x);
            if i >= width {
                m.revert(BACON[i - width]);
            }
            let window = &BACON[(i + 1).saturating_sub(width)..i + 1];
            assert_eq!(m.n(), window.len());
            assert!(almost_equal(m.mean(), fmean(window), 14), "step {}", i);
            assert!(almost_equal(m.variance(), pvariance(window), 14), "step {}", i);
        }
    }

    #[test]
    #[should_panic(expected = "Cannot revert from an empty accumulator")]
    fn test_revert_empty_raises() {
        let mut m = CentralMomentsKleinKbn::default();
        m.revert(1.0);
    }

    #[test]
    fn test_standard_deviation_is_real_after_revert() {
        // Reverting to two equal samples can leave a tiny negative M2.
        let mut m = with_ddof(0);
        for x in [0.1, 0.1, 0.7] {
            m.update(x);
        }
        m.revert(0.7);
        assert!(!m.standard_deviation().is_nan());
        assert!(m.variance() >= 0.0);
        assert!(almost_equal(m.standard_deviation(), 0.0, 15));
    }

    #[test]
    fn test_reset() {
        let mut m = feed(CentralMomentsKleinKbn::default(), &BACON);
        m.reset();
        assert_eq!(m.n(), 0);
        assert_eq!(m.mean(), 0.0);
        assert!(m.variance().is_nan());
        let m = feed(m, &[1.0, 2.0, 3.0]);
        assert!(almost_equal(m.variance(), 1.0, 15));
    }
}
