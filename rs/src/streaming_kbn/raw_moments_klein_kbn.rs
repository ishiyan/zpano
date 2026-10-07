// Raw moments with Klein KBN (Kahan-Babuška-Neumaier)
// compensated summation for improved numerical stability.
//
// References:
//   https://github.com/kuiperzone/Compensated-Accumulators
//   https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance

use super::klein_kbn_accumulator::KleinKbnAccumulator;

/// Relative threshold below which the population variance μ₂, computed
/// from raw power sums as Σx²/n − (Σx/n)², is considered to be lost
/// in rounding error (i.e. indistinguishable from zero).
const CANCELLATION_EPSILON: f64 = 1e-14;

/// Python's `max(v, 0.0)`: returns `v` unless `0.0 > v`.
#[inline]
fn max_zero(v: f64) -> f64 {
    if 0.0 > v {
        0.0
    } else {
        v
    }
}

/// Streaming mean, variance, skewness, kurtosis via raw power sums (x¹..x⁴)
/// with Klein KBN (Kahan-Babuška-Neumaier) compensated accumulation.
///
/// Accumulates Σx, Σx², Σx³, Σx⁴ using a `KleinKbnAccumulator` for each,
/// plus a separate Welford mean/variance tracker (also KBN-compensated).
/// Mean and variance come from the Welford tracker; skewness and kurtosis
/// are converted from the raw power sums at query time.
///
/// Supports removal of any previously added sample with revert(), not
/// only the most recent one, because both the power sums and Welford's
/// mean/M₂ are symmetric functions of the samples.  This makes the type
/// suitable for FIFO rolling windows (update the new sample, revert the
/// oldest one).
///
/// Accuracy caveat: converting raw power sums to central moments suffers
/// catastrophic cancellation when the mean is large compared to the spread
/// (e.g. prices rather than returns).  For such data skewness and kurtosis
/// lose precision.  Use `CentralMomentsKleinKbn` for such data; it supports
/// FIFO removal, but repeated removals clear compensation and can accumulate
/// rounding error.
///
/// # Notation
///
/// μₖ' = Σxᵏ / n are the raw moments (getters x1..x4) and μₖ are the
/// population central moments derived from them:
///
/// ```text
/// μ₂ = μ₂' − μ₁'²
/// μ₃ = μ₃' − μ₁'³ − 3·μ₁'·μ₂
/// μ₄ = μ₄' − μ₁'⁴ − 6·μ₁'²·μ₂ − 4·μ₁'·μ₃
/// ```
///
/// g₁ = μ₃ / μ₂^1.5 is the population skewness and β₂ = μ₄ / μ₂² is the
/// population (Pearson) kurtosis.  All skewness and kurtosis getters
/// are NaN when n < 2 or μ₂ is zero relative to μ₂' (constant data).
///
/// # Parameters
///
/// - `ddof` (default 1): delta degrees of freedom for variance and
///   standard_deviation.  variance = Σ(x - x̄)² / (n - ddof).  ddof=0 gives
///   population, ddof=1 gives sample.
/// - `bias` (default true): selects the skewness and kurtosis getters, see below.
/// - `fisher` (default true): selects the kurtosis getter, see below.
///
/// The `skewness` and `kurtosis` getters dispatch as follows
/// (matching scipy.stats.skew and scipy.stats.kurtosis):
///
/// | bias  | fisher | skewness            | kurtosis                    |
/// |-------|--------|---------------------|-----------------------------|
/// | true  | true   | skewness_moment, g₁ | kurtosis_excess, β₂ − 3     |
/// | true  | false  | skewness_moment, g₁ | kurtosis_moment, β₂         |
/// | false | true   | skewness_fisher, G₁ | kurtosis_sample_excess, G₂  |
/// | false | false  | skewness_fisher, G₁ | kurtosis_sample, G₂ + 3     |
///
/// The remaining variants (skewness_sample, kurtosis_sample_corrected),
/// which match the R PerformanceAnalytics package, are available as
/// separate getters.
#[derive(Debug, Clone)]
pub struct RawMomentsKleinKbn {
    n: usize,
    x1: KleinKbnAccumulator,
    x2: KleinKbnAccumulator,
    x3: KleinKbnAccumulator,
    x4: KleinKbnAccumulator,
    ddof: usize,
    bias: bool,
    fisher: bool,
    // Welford's mean and sum of squared deviations Σ(x - x̄)².
    mean: KleinKbnAccumulator,
    s: KleinKbnAccumulator,
}

impl RawMomentsKleinKbn {
    /// Creates an empty instance with the given ddof, bias and fisher.
    pub fn new(ddof: usize, bias: bool, fisher: bool) -> Self {
        Self {
            n: 0,
            x1: KleinKbnAccumulator::new(),
            x2: KleinKbnAccumulator::new(),
            x3: KleinKbnAccumulator::new(),
            x4: KleinKbnAccumulator::new(),
            ddof,
            bias,
            fisher,
            mean: KleinKbnAccumulator::new(),
            s: KleinKbnAccumulator::new(),
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

    /// Adds a sample x.
    pub fn update(&mut self, x: f64) {
        self.n += 1;
        self.x1.update(x);
        let x2 = x * x;
        self.x2.update(x2);
        let x3 = x2 * x;
        self.x3.update(x3);
        let x4 = x3 * x;
        self.x4.update(x4);
        // Welford: mean += (x - mean_old) / n;  S += (x - mean_old)·(x - mean_new)
        let delta = x - self.mean.value();
        self.mean.update(delta / self.n as f64);
        self.s.update(delta * (x - self.mean.value()));
    }

    /// Removes a previously added sample x (any sample, not only the most
    /// recent one).  Reverting a value that was never added corrupts the
    /// state.
    ///
    /// # Panics
    ///
    /// Panics if there are no samples.
    pub fn revert(&mut self, x: f64) {
        if self.n == 0 {
            panic!("Cannot revert from an empty accumulator");
        }
        if self.n == 1 {
            self.reset();
            return;
        }
        self.n -= 1;
        self.x1.revert(x);
        let x2 = x * x;
        self.x2.revert(x2);
        let x3 = x2 * x;
        self.x3.revert(x3);
        let x4 = x3 * x;
        self.x4.revert(x4);
        // Inverse Welford: mean_old = mean_new - (x - mean_new) / (n - 1);
        // S -= (x - mean_new)·(x - mean_old)
        let delta = x - self.mean.value();
        self.mean.revert(delta / self.n as f64);
        self.s.revert(delta * (x - self.mean.value()));
    }

    /// Delta degrees of freedom used by variance and standard_deviation.
    pub fn ddof(&self) -> usize {
        self.ddof
    }

    /// Sets the delta degrees of freedom.
    pub fn set_ddof(&mut self, ddof: usize) {
        self.ddof = ddof;
    }

    /// Selects the skewness and kurtosis getters.
    pub fn bias(&self) -> bool {
        self.bias
    }

    /// Sets the bias parameter.
    pub fn set_bias(&mut self, bias: bool) {
        self.bias = bias;
    }

    /// Selects the kurtosis getter.
    pub fn fisher(&self) -> bool {
        self.fisher
    }

    /// Sets the fisher parameter.
    pub fn set_fisher(&mut self, fisher: bool) {
        self.fisher = fisher;
    }

    fn variance_with(&self, ddof: usize) -> f64 {
        // A slightly negative S caused by rounding after revert() is clamped to zero.
        if self.n <= ddof {
            return f64::NAN;
        }
        let d = self.n - ddof;
        max_zero(self.s.value()) / d as f64
    }

    fn standard_deviation_with(&self, ddof: usize) -> f64 {
        let v = self.variance_with(ddof);
        if v.is_nan() {
            v
        } else {
            v.sqrt()
        }
    }

    /// The arithmetic mean (0.0 when empty).
    pub fn mean(&self) -> f64 {
        self.mean.value()
    }

    /// The variance Σ(x - x̄)² / (n - ddof), NaN when n ≤ ddof.
    pub fn variance(&self) -> f64 {
        self.variance_with(self.ddof)
    }

    /// The population variance Σ(x - x̄)² / n, regardless of ddof.
    pub fn variance_ddof_0(&self) -> f64 {
        self.variance_with(0)
    }

    /// The sample variance Σ(x - x̄)² / (n - 1), regardless of ddof.
    pub fn variance_ddof_1(&self) -> f64 {
        self.variance_with(1)
    }

    /// The square root of variance.
    pub fn standard_deviation(&self) -> f64 {
        self.standard_deviation_with(self.ddof)
    }

    /// The population standard deviation, regardless of ddof.
    pub fn standard_deviation_ddof_0(&self) -> f64 {
        self.standard_deviation_with(0)
    }

    /// The sample standard deviation, regardless of ddof.
    pub fn standard_deviation_ddof_1(&self) -> f64 {
        self.standard_deviation_with(1)
    }

    /// Converts the raw power sums to population central moments
    /// (μ₂, μ₃, μ₄), see the type Notation.
    ///
    /// Returns None when n < 2 or μ₂ is lost in rounding error.
    fn central_moments(&self) -> Option<(f64, f64, f64)> {
        let n = self.n;
        if n < 2 {
            return None;
        }
        let nf = n as f64;
        let mu1 = self.x1.value() / nf;
        let mut r = mu1 * mu1;
        let mean_x2 = self.x2.value() / nf;
        let mu2 = mean_x2 - r;
        if mu2 <= CANCELLATION_EPSILON * mean_x2 {
            return None;
        }
        r *= mu1;
        let mu3 = self.x3.value() / nf - r - 3.0 * mu1 * mu2;
        r *= mu1;
        let mu4 = self.x4.value() / nf - r - 6.0 * mu2 * mu1 * mu1 - 4.0 * mu3 * mu1;
        Some((mu2, mu3, mu4))
    }

    /// The population skewness g₁ = μ₃ / μ₂^1.5.
    fn g1(&self) -> f64 {
        match self.central_moments() {
            None => f64::NAN,
            Some((mu2, mu3, _)) => mu3 / (mu2 * mu2.sqrt()),
        }
    }

    /// The population (Pearson) kurtosis β₂ = μ₄ / μ₂².
    fn b2(&self) -> f64 {
        match self.central_moments() {
            None => f64::NAN,
            Some((mu2, _, mu4)) => mu4 / (mu2 * mu2),
        }
    }

    /// The 'moment' (biased, population) skewness, requires n ≥ 2:
    ///
    /// ```text
    /// g₁ = μ₃ / μ₂^1.5
    /// ```
    ///
    /// Matches scipy.stats.skew(bias=True) and PerformanceAnalytics
    /// skewness(method="moment").
    pub fn skewness_moment(&self) -> f64 {
        self.g1()
    }

    /// The 'fisher' (bias-adjusted Fisher-Pearson) skewness, requires n ≥ 3:
    ///
    /// ```text
    /// G₁ = g₁ · √(n(n−1)) / (n−2)
    /// ```
    ///
    /// Matches scipy.stats.skew(bias=False) and PerformanceAnalytics
    /// skewness(method="fisher").
    pub fn skewness_fisher(&self) -> f64 {
        let g1 = self.g1();
        if g1.is_nan() {
            return f64::NAN;
        }
        let n = self.n;
        if n < 3 {
            return f64::NAN;
        }
        let nf = n as f64;
        g1 * (nf * (nf - 1.0)).sqrt() / (nf - 2.0)
    }

    /// The 'sample' skewness, requires n ≥ 3:
    ///
    /// ```text
    /// g₁ · n² / ((n−1)(n−2))
    /// ```
    ///
    /// Matches PerformanceAnalytics skewness(method="sample").
    /// Doesn't depend on the bias parameter.
    pub fn skewness_sample(&self) -> f64 {
        let g1 = self.g1();
        if g1.is_nan() {
            return f64::NAN;
        }
        let n = self.n;
        if n < 3 {
            return f64::NAN;
        }
        let nf = n as f64;
        g1 * (nf * nf) / ((nf - 1.0) * (nf - 2.0))
    }

    /// The skewness selected by the bias parameter:
    ///
    /// - bias=true:  skewness_moment, g₁
    /// - bias=false: skewness_fisher, G₁ = g₁ · √(n(n−1)) / (n−2)
    ///
    /// The third variant, skewness_sample, doesn't depend on bias.
    pub fn skewness(&self) -> f64 {
        if self.bias {
            self.skewness_moment()
        } else {
            self.skewness_fisher()
        }
    }

    /// The 'moment' (biased, population) Pearson kurtosis, requires n ≥ 2:
    ///
    /// ```text
    /// β₂ = μ₄ / μ₂²
    /// ```
    ///
    /// Matches scipy.stats.kurtosis(bias=True, fisher=False) and
    /// PerformanceAnalytics kurtosis(method="moment").
    pub fn kurtosis_moment(&self) -> f64 {
        self.b2()
    }

    /// The 'excess' (biased, population) excess kurtosis, requires n ≥ 2:
    ///
    /// ```text
    /// β₂ − 3
    /// ```
    ///
    /// Matches scipy.stats.kurtosis(bias=True, fisher=True) and
    /// PerformanceAnalytics kurtosis(method="excess").
    pub fn kurtosis_excess(&self) -> f64 {
        let b2 = self.b2();
        if b2.is_nan() {
            return f64::NAN;
        }
        b2 - 3.0
    }

    /// ((n²−1)·β₂ − 3(n−1)²) / ((n−2)(n−3)), NaN when β₂ is NaN or n ≤ 3.
    fn g2(&self) -> f64 {
        let b2 = self.b2();
        if b2.is_nan() {
            return f64::NAN;
        }
        let n = self.n;
        if n <= 3 {
            return f64::NAN;
        }
        let nf = n as f64;
        ((nf * nf - 1.0) * b2 - 3.0 * (nf - 1.0) * (nf - 1.0)) / ((nf - 2.0) * (nf - 3.0))
    }

    /// The 'sample excess' (unbiased) excess kurtosis, requires n ≥ 4:
    ///
    /// ```text
    /// G₂ = ((n²−1)·β₂ − 3(n−1)²) / ((n−2)(n−3))
    /// ```
    ///
    /// Matches scipy.stats.kurtosis(bias=False, fisher=True) and
    /// PerformanceAnalytics kurtosis(method="sample_excess").
    pub fn kurtosis_sample_excess(&self) -> f64 {
        self.g2()
    }

    /// The 'sample' (unbiased) Pearson kurtosis, requires n ≥ 4, calculated
    /// as the 'sample excess' kurtosis plus 3:
    ///
    /// ```text
    /// G₂ + 3 = ((n²−1)·β₂ − 3(n−1)²) / ((n−2)(n−3)) + 3
    /// ```
    ///
    /// Matches scipy.stats.kurtosis(bias=False, fisher=False).
    ///
    /// The PerformanceAnalytics kurtosis(method="sample") variant is
    /// available as kurtosis_sample_corrected; it is larger than this one
    /// by (9n−15) / ((n−2)(n−3)), approximately 0.44 for n = 24.
    pub fn kurtosis_sample(&self) -> f64 {
        self.g2() + 3.0
    }

    /// The PerformanceAnalytics 'sample' (unbiased) Pearson kurtosis,
    /// requires n ≥ 4:
    ///
    /// ```text
    /// (n²−1)·β₂ / ((n−2)(n−3))
    /// ```
    ///
    /// Matches PerformanceAnalytics kurtosis(method="sample").
    /// Doesn't depend on the bias and fisher parameters.
    ///
    /// It differs from kurtosis_sample (G₂ + 3) by (9n−15) / ((n−2)(n−3)),
    /// approximately 0.44 for n = 24.
    pub fn kurtosis_sample_corrected(&self) -> f64 {
        let b2 = self.b2();
        if b2.is_nan() {
            return f64::NAN;
        }
        let n = self.n;
        if n <= 3 {
            return f64::NAN;
        }
        let nf = n as f64;
        b2 * (nf * nf - 1.0) / ((nf - 2.0) * (nf - 3.0))
    }

    /// The kurtosis selected by the bias and fisher parameters:
    ///
    /// - bias=true,  fisher=true:  kurtosis_excess, β₂ − 3
    /// - bias=true,  fisher=false: kurtosis_moment, β₂
    /// - bias=false, fisher=true:  kurtosis_sample_excess, G₂
    /// - bias=false, fisher=false: kurtosis_sample, G₂ + 3
    pub fn kurtosis(&self) -> f64 {
        match (self.bias, self.fisher) {
            (true, true) => self.kurtosis_excess(),
            (true, false) => self.kurtosis_moment(),
            (false, true) => self.kurtosis_sample_excess(),
            (false, false) => self.kurtosis_sample(),
        }
    }

    /// The sum Σx.
    pub fn x1_sum(&self) -> f64 {
        self.x1.value()
    }

    /// The sum Σx².
    pub fn x2_sum(&self) -> f64 {
        self.x2.value()
    }

    /// The sum Σx³.
    pub fn x3_sum(&self) -> f64 {
        self.x3.value()
    }

    /// The sum Σx⁴.
    pub fn x4_sum(&self) -> f64 {
        self.x4.value()
    }

    fn raw_moment(&self, sum: &KleinKbnAccumulator) -> f64 {
        if self.n > 0 {
            sum.value() / self.n as f64
        } else {
            f64::NAN
        }
    }

    /// The first raw moment Σx / n (NaN when empty).
    pub fn x1(&self) -> f64 {
        self.raw_moment(&self.x1)
    }

    /// The second raw moment Σx² / n (NaN when empty).
    pub fn x2(&self) -> f64 {
        self.raw_moment(&self.x2)
    }

    /// The third raw moment Σx³ / n (NaN when empty).
    pub fn x3(&self) -> f64 {
        self.raw_moment(&self.x3)
    }

    /// The fourth raw moment Σx⁴ / n (NaN when empty).
    pub fn x4(&self) -> f64 {
        self.raw_moment(&self.x4)
    }

    /// The number of samples.
    pub fn n(&self) -> usize {
        self.n
    }
}

impl Default for RawMomentsKleinKbn {
    /// ddof=1, bias=true, fisher=true.
    fn default() -> Self {
        Self::new(1, true, true)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::streaming_kbn::test_support::almost_equal;

    // https://github.com/medo64/Medo/blob/main/tests/Tests.Medo/Math/WelfordVariance.cs
    // https://github.com/andrewuhl/RollingWindow/blob/master/src/RollingWindow.cpp
    // https://github.com/ajcr/rolling/blob/master/rolling/similarity.py

    type Getter = fn(&RawMomentsKleinKbn) -> f64;

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
    // 50-digit decimal.Decimal), rounded to the nearest float.
    const MEAN: f64 = 0.009000000000000001;
    const VARIANCE_DDOF_0: f64 = 0.0014989166666666668;
    const VARIANCE_DDOF_1: f64 = 0.0015640869565217393;
    const STANDARD_DEVIATION_DDOF_0: f64 = 0.03871584516275819;
    const STANDARD_DEVIATION_DDOF_1: f64 = 0.039548539246370897;
    const SKEWNESS_MOMENT: f64 = -0.08256245520856804; // scipy skew(bias=True)
    const SKEWNESS_FISHER: f64 = -0.08817174934967535; // scipy skew(bias=False)
    const SKEWNESS_SAMPLE: f64 = -0.09398413873544505; // R PerformanceAnalytics "sample"
    const KURTOSIS_MOMENT: f64 = 2.4324537941078743; // scipy kurtosis(bias=True, fisher=False)
    const KURTOSIS_EXCESS: f64 = -0.5675462058921257; // scipy kurtosis(bias=True, fisher=True)
    const KURTOSIS_SAMPLE_EXCESS: f64 = -0.40766032118608714; // scipy kurtosis(bias=False, fisher=True)
    const KURTOSIS_SAMPLE: f64 = 2.592339678813913; // scipy kurtosis(bias=False, fisher=False)
    const KURTOSIS_SAMPLE_CORRECTED: f64 = 3.027404613878848; // R PerformanceAnalytics "sample"
    const X1_SUM: f64 = 0.21600000000000003;
    const X2_SUM: f64 = 0.037918;
    const X3_SUM: f64 = 0.0008738040000000003;
    const X4_SUM: f64 = 0.00014466403;
    const X1: f64 = 0.009000000000000001;
    const X2: f64 = 0.0015799166666666668;
    const X3: f64 = 3.640850000000001e-05;
    const X4: f64 = 6.0276679166666674e-06;

    fn expected() -> [(&'static str, Getter, f64); 21] {
        [
            ("mean", RawMomentsKleinKbn::mean, MEAN),
            ("variance_ddof_0", RawMomentsKleinKbn::variance_ddof_0, VARIANCE_DDOF_0),
            ("variance_ddof_1", RawMomentsKleinKbn::variance_ddof_1, VARIANCE_DDOF_1),
            ("standard_deviation_ddof_0", RawMomentsKleinKbn::standard_deviation_ddof_0, STANDARD_DEVIATION_DDOF_0),
            ("standard_deviation_ddof_1", RawMomentsKleinKbn::standard_deviation_ddof_1, STANDARD_DEVIATION_DDOF_1),
            ("skewness_moment", RawMomentsKleinKbn::skewness_moment, SKEWNESS_MOMENT),
            ("skewness_fisher", RawMomentsKleinKbn::skewness_fisher, SKEWNESS_FISHER),
            ("skewness_sample", RawMomentsKleinKbn::skewness_sample, SKEWNESS_SAMPLE),
            ("kurtosis_moment", RawMomentsKleinKbn::kurtosis_moment, KURTOSIS_MOMENT),
            ("kurtosis_excess", RawMomentsKleinKbn::kurtosis_excess, KURTOSIS_EXCESS),
            ("kurtosis_sample_excess", RawMomentsKleinKbn::kurtosis_sample_excess, KURTOSIS_SAMPLE_EXCESS),
            ("kurtosis_sample", RawMomentsKleinKbn::kurtosis_sample, KURTOSIS_SAMPLE),
            ("kurtosis_sample_corrected", RawMomentsKleinKbn::kurtosis_sample_corrected, KURTOSIS_SAMPLE_CORRECTED),
            ("x1_sum", RawMomentsKleinKbn::x1_sum, X1_SUM),
            ("x2_sum", RawMomentsKleinKbn::x2_sum, X2_SUM),
            ("x3_sum", RawMomentsKleinKbn::x3_sum, X3_SUM),
            ("x4_sum", RawMomentsKleinKbn::x4_sum, X4_SUM),
            ("x1", RawMomentsKleinKbn::x1, X1),
            ("x2", RawMomentsKleinKbn::x2, X2),
            ("x3", RawMomentsKleinKbn::x3, X3),
            ("x4", RawMomentsKleinKbn::x4, X4),
        ]
    }

    // Defaults are ddof=1, bias=true, fisher=true.
    fn expected_by_dispatch() -> [(&'static str, Getter, f64); 4] {
        [
            ("mean", RawMomentsKleinKbn::mean, MEAN),
            ("variance", RawMomentsKleinKbn::variance, VARIANCE_DDOF_1),
            ("skewness", RawMomentsKleinKbn::skewness, SKEWNESS_MOMENT),
            ("kurtosis", RawMomentsKleinKbn::kurtosis, KURTOSIS_EXCESS),
        ]
    }

    fn feed(mut m: RawMomentsKleinKbn, data: &[f64]) -> RawMomentsKleinKbn {
        for &x in data {
            m.update(x);
        }
        m
    }

    fn with_ddof(ddof: usize) -> RawMomentsKleinKbn {
        RawMomentsKleinKbn::new(ddof, true, true)
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
    fn test_bacon_all_properties() {
        let m = feed(RawMomentsKleinKbn::default(), &BACON);
        for (name, getter, expected) in expected() {
            let actual = getter(&m);
            assert!(almost_equal(actual, expected, 14), "{}: {} vs {}", name, actual, expected);
        }
    }

    #[test]
    fn test_dispatch() {
        let cases: [(bool, bool, Getter, f64, Getter, f64); 4] = [
            (true, true, RawMomentsKleinKbn::skewness_moment, SKEWNESS_MOMENT,
             RawMomentsKleinKbn::kurtosis_excess, KURTOSIS_EXCESS),
            (true, false, RawMomentsKleinKbn::skewness_moment, SKEWNESS_MOMENT,
             RawMomentsKleinKbn::kurtosis_moment, KURTOSIS_MOMENT),
            (false, true, RawMomentsKleinKbn::skewness_fisher, SKEWNESS_FISHER,
             RawMomentsKleinKbn::kurtosis_sample_excess, KURTOSIS_SAMPLE_EXCESS),
            (false, false, RawMomentsKleinKbn::skewness_fisher, SKEWNESS_FISHER,
             RawMomentsKleinKbn::kurtosis_sample, KURTOSIS_SAMPLE),
        ];
        for (bias, fisher, skew, skew_expected, kurt, kurt_expected) in cases {
            let m = feed(RawMomentsKleinKbn::new(1, bias, fisher), &BACON);
            assert_eq!(m.bias(), bias);
            assert_eq!(m.fisher(), fisher);
            assert_eq!(m.skewness(), skew(&m), "bias={} fisher={}", bias, fisher);
            assert_eq!(m.kurtosis(), kurt(&m), "bias={} fisher={}", bias, fisher);
            assert!(almost_equal(m.skewness(), skew_expected, 14), "bias={} fisher={}", bias, fisher);
            assert!(almost_equal(m.kurtosis(), kurt_expected, 13), "bias={} fisher={}", bias, fisher);
        }
    }

    #[test]
    fn test_ddof() {
        let cases: [(usize, Getter, Getter); 2] = [
            (0, RawMomentsKleinKbn::variance_ddof_0, RawMomentsKleinKbn::standard_deviation_ddof_0),
            (1, RawMomentsKleinKbn::variance_ddof_1, RawMomentsKleinKbn::standard_deviation_ddof_1),
        ];
        for (ddof, variance, standard_deviation) in cases {
            let m = feed(with_ddof(ddof), &BACON);
            assert_eq!(m.ddof(), ddof);
            assert_eq!(m.variance(), variance(&m));
            assert_eq!(m.standard_deviation(), standard_deviation(&m));
        }
        let m = feed(with_ddof(1), &[1.0, 2.0, 3.0]);
        assert!(almost_equal(m.variance(), 1.0, 15));
        assert!(almost_equal(m.standard_deviation(), 1.0, 15));
    }

    // test_invalid_ddof is not ported: ddof is a usize, so negative and
    // non-integer values are rejected at compile time.

    #[test]
    fn test_large_ddof() {
        for ddof in [usize::MAX, 1usize << (usize::BITS - 1)] {
            let mut m = RawMomentsKleinKbn::new(ddof, true, true);
            assert!(m.variance().is_nan());
            assert!(m.standard_deviation().is_nan());
            m.update(1.0);
            m.update(3.0);
            assert!(m.variance().is_nan());
            assert!(m.standard_deviation().is_nan());
            assert_eq!(m.variance_ddof_0(), 1.0);
            assert_eq!(m.variance_ddof_1(), 2.0);
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
    fn test_setters() {
        let mut m = feed(RawMomentsKleinKbn::default(), &BACON);
        m.set_ddof(0);
        assert_eq!(m.variance(), m.variance_ddof_0());
        m.set_bias(false);
        assert_eq!(m.skewness(), m.skewness_fisher());
        m.set_fisher(false);
        assert_eq!(m.kurtosis(), m.kurtosis_sample());
    }

    #[test]
    fn test_kurtosis_sample_corrected_difference() {
        // kurtosis_sample_corrected - kurtosis_sample = (9n-15) / ((n-2)(n-3))
        let m = feed(RawMomentsKleinKbn::default(), &BACON);
        let n = BACON.len() as f64;
        assert!(almost_equal(
            m.kurtosis_sample_corrected() - m.kurtosis_sample(),
            (9.0 * n - 15.0) / ((n - 2.0) * (n - 3.0)),
            14
        ));
    }

    #[test]
    fn test_empty() {
        let m = RawMomentsKleinKbn::default();
        assert_eq!(m.n(), 0);
        assert_eq!(m.mean(), 0.0);
        let getters: [(&str, Getter); 8] = [
            ("variance", RawMomentsKleinKbn::variance),
            ("standard_deviation", RawMomentsKleinKbn::standard_deviation),
            ("skewness", RawMomentsKleinKbn::skewness),
            ("kurtosis", RawMomentsKleinKbn::kurtosis),
            ("x1", RawMomentsKleinKbn::x1),
            ("x2", RawMomentsKleinKbn::x2),
            ("x3", RawMomentsKleinKbn::x3),
            ("x4", RawMomentsKleinKbn::x4),
        ];
        for (name, getter) in getters {
            assert!(getter(&m).is_nan(), "{}", name);
        }
        assert_eq!(m.x1_sum(), 0.0);
    }

    #[test]
    fn test_minimum_sample_sizes() {
        let data = [1.0, 2.0, 4.0, 8.0];
        let minimum_n: [(&str, Getter, usize); 8] = [
            ("skewness_moment", RawMomentsKleinKbn::skewness_moment, 2),
            ("skewness_fisher", RawMomentsKleinKbn::skewness_fisher, 3),
            ("skewness_sample", RawMomentsKleinKbn::skewness_sample, 3),
            ("kurtosis_moment", RawMomentsKleinKbn::kurtosis_moment, 2),
            ("kurtosis_excess", RawMomentsKleinKbn::kurtosis_excess, 2),
            ("kurtosis_sample_excess", RawMomentsKleinKbn::kurtosis_sample_excess, 4),
            ("kurtosis_sample", RawMomentsKleinKbn::kurtosis_sample, 4),
            ("kurtosis_sample_corrected", RawMomentsKleinKbn::kurtosis_sample_corrected, 4),
        ];
        let mut m = RawMomentsKleinKbn::default();
        for (i, &x) in data.iter().enumerate() {
            m.update(x);
            let n = i + 1;
            for (name, getter, min_n) in minimum_n {
                assert_eq!(getter(&m).is_nan(), n < min_n, "{} n={}", name, n);
            }
        }
    }

    #[test]
    fn test_constant_data() {
        let m = feed(with_ddof(0), &[0.1; 5]);
        assert!(almost_equal(m.mean(), 0.1, 16));
        assert!(almost_equal(m.variance(), 0.0, 16));
        assert!(m.skewness().is_nan());
        assert!(m.kurtosis().is_nan());
    }

    #[test]
    fn test_scale_invariance() {
        // The cancellation threshold is relative, so tiny values work.
        let data: Vec<f64> = BACON.iter().map(|&x| x * 1e-6).collect();
        let m = feed(RawMomentsKleinKbn::default(), &data);
        assert!(almost_equal(m.skewness_moment(), SKEWNESS_MOMENT, 13));
        assert!(almost_equal(m.kurtosis_excess(), KURTOSIS_EXCESS, 13));
    }

    #[test]
    fn test_large_offset_preserves_variance_but_not_higher_moments() {
        // Welford's variance remains usable when raw-power cancellation
        // makes skewness and kurtosis unreliable.
        let m = feed(with_ddof(0), &[1e8, 1e8 + 1.0, 1e8 + 2.0]);
        assert!(almost_equal(m.mean(), 1e8 + 1.0, 10));
        assert!(almost_equal(m.variance(), 2.0 / 3.0, 14));
        assert!(m.skewness().is_nan());
        assert!(m.kurtosis().is_nan());
    }

    #[test]
    fn test_revert_partial() {
        let data = [10.0, 18.0, 5.0, 12.0, 7.0];
        let mut m_full = feed(with_ddof(0), &data);
        let m_part = feed(with_ddof(0), &data[..4]);
        m_full.revert(data[4]);
        assert_eq!(m_full.n(), 4);
        assert!(almost_equal(m_full.mean(), m_part.mean(), 15));
        assert!(almost_equal(m_full.variance(), m_part.variance(), 15));
        assert!(almost_equal(m_full.skewness(), m_part.skewness(), 14));
        assert!(almost_equal(m_full.kurtosis(), m_part.kurtosis(), 13));
    }

    #[test]
    fn test_revert_not_most_recent() {
        let mut data = BACON.to_vec();
        data.push(0.5);
        let mut m = feed(RawMomentsKleinKbn::default(), &data);
        m.revert(0.5);
        let mut data2 = vec![0.5];
        data2.extend_from_slice(&BACON);
        let mut m2 = feed(RawMomentsKleinKbn::default(), &data2);
        m2.revert(0.5); // the oldest sample
        for (name, getter, expected) in expected_by_dispatch() {
            assert!(almost_equal(getter(&m), expected, 13), "{}: {} vs {}", name, getter(&m), expected);
            assert!(almost_equal(getter(&m2), expected, 13), "{}: {} vs {}", name, getter(&m2), expected);
        }
    }

    #[test]
    fn test_revert_to_empty() {
        let mut m = feed(with_ddof(0), &BACON);
        for &x in &BACON {
            m.revert(x);
        }
        assert_eq!(m.n(), 0);
        assert_eq!(m.mean(), 0.0);
        assert_eq!(m.x1_sum(), 0.0);
        assert!(m.variance().is_nan());
        let m = feed(m, &[1.0, 2.0, 3.0, 4.0]);
        assert!(almost_equal(m.variance(), 1.25, 15));
    }

    #[test]
    #[should_panic(expected = "Cannot revert from an empty accumulator")]
    fn test_revert_empty_raises() {
        let mut m = RawMomentsKleinKbn::default();
        m.revert(1.0);
    }

    #[test]
    fn test_rolling_window() {
        let names: [(&str, Getter); 11] = [
            ("mean", RawMomentsKleinKbn::mean),
            ("variance", RawMomentsKleinKbn::variance),
            ("standard_deviation", RawMomentsKleinKbn::standard_deviation),
            ("skewness", RawMomentsKleinKbn::skewness),
            ("kurtosis", RawMomentsKleinKbn::kurtosis),
            ("skewness_sample", RawMomentsKleinKbn::skewness_sample),
            ("kurtosis_sample_corrected", RawMomentsKleinKbn::kurtosis_sample_corrected),
            ("x1", RawMomentsKleinKbn::x1),
            ("x2", RawMomentsKleinKbn::x2),
            ("x3", RawMomentsKleinKbn::x3),
            ("x4", RawMomentsKleinKbn::x4),
        ];
        let w = 5usize;
        let mut m = RawMomentsKleinKbn::new(1, false, true);
        for (i, &x) in BACON.iter().enumerate() {
            m.update(x);
            if i >= w {
                m.revert(BACON[i - w]);
            }
            let r = feed(RawMomentsKleinKbn::new(1, false, true), &BACON[(i + 1).saturating_sub(w)..i + 1]);
            assert_eq!(m.n(), r.n());
            for (name, getter) in names {
                let (actual, expected) = (getter(&m), getter(&r));
                if expected.is_nan() {
                    assert!(actual.is_nan(), "step {} {}: {}", i, name, actual);
                } else {
                    assert!(almost_equal(actual, expected, 13), "step {} {}: {} vs {}", i, name, actual, expected);
                }
            }
        }
    }

    #[test]
    fn test_standard_deviation_is_real_after_revert() {
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
    fn test_variance_getter_has_no_side_effects() {
        let m = feed(with_ddof(0), &BACON);
        let v = m.variance();
        let _ = m.standard_deviation();
        assert_eq!(m.variance(), v);
        assert_eq!(m.n(), BACON.len());
    }

    #[test]
    fn test_reset() {
        let mut m = feed(RawMomentsKleinKbn::default(), &BACON);
        m.reset();
        assert_eq!(m.n(), 0);
        assert_eq!(m.mean(), 0.0);
        assert_eq!(m.x4_sum(), 0.0);
        assert!(m.variance().is_nan());
        let m = feed(m, &[1.0, 2.0, 3.0]);
        assert!(almost_equal(m.variance(), 1.0, 15));
    }
}
