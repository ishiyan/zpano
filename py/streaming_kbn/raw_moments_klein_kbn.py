import math

from .klein_kbn_accumulator import KleinKBNAccumulator


##########################################################
# Raw moments with Klein KBN (Kahan-Babuška-Neumaier)
# compensated summation for improved numerical stability.
#
# References:
#   https://github.com/kuiperzone/Compensated-Accumulators
#   https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance
##########################################################

# Relative threshold below which the population variance μ₂, computed
# from raw power sums as Σx²/n − (Σx/n)², is considered to be lost
# in rounding error (i.e. indistinguishable from zero).
_CANCELLATION_EPSILON = 1e-14


class RawMomentsKleinKBN:
    r"""
    Streaming mean, variance, skewness, kurtosis via raw power sums (x¹..x⁴)
    with Klein KBN (Kahan-Babuška-Neumaier) compensated accumulation.

    Accumulates Σx, Σx², Σx³, Σx⁴ using a KleinKBNAccumulator for each,
    plus a separate Welford mean/variance tracker (also KBN-compensated).
    Mean and variance come from the Welford tracker; skewness and kurtosis
    are converted from the raw power sums at query time.

    Supports removal of any previously added sample with revert(), not
    only the most recent one, because both the power sums and Welford's
    mean/M₂ are symmetric functions of the samples.  This makes the class
    suitable for FIFO rolling windows (update the new sample, revert the
    oldest one).

    Accuracy caveat: converting raw power sums to central moments suffers
    catastrophic cancellation when the mean is large compared to the spread
    (e.g. prices rather than returns).  For such data skewness and kurtosis
    lose precision.  Use CentralMomentsKleinKBN if FIFO revert is not needed.

    Notation
    --------
    μₖ' = Σxᵏ / n are the raw moments (properties x1..x4) and μₖ are the
    population central moments derived from them:

        μ₂ = μ₂' − μ₁'²
        μ₃ = μ₃' − μ₁'³ − 3·μ₁'·μ₂
        μ₄ = μ₄' − μ₁'⁴ − 6·μ₁'²·μ₂ − 4·μ₁'·μ₃

    g₁ = μ₃ / μ₂^1.5 is the population skewness and β₂ = μ₄ / μ₂² is the
    population (Pearson) kurtosis.  All skewness and kurtosis properties
    are NaN when n < 2 or μ₂ is zero relative to μ₂' (constant data).

    Parameters
    ----------
    ddof : int, default=1
        Delta degrees of freedom for variance and standard_deviation.
        variance = Σ(x - x̄)² / (n - ddof).  ddof=0 gives population, ddof=1 gives sample.
    bias : bool, default=True
        Selects the skewness and kurtosis properties, see below.
    fisher : bool, default=True
        Selects the kurtosis property, see below.

    The ``skewness`` and ``kurtosis`` properties dispatch as follows
    (matching scipy.stats.skew and scipy.stats.kurtosis):

    ==========  ==========  ======================  ==============================
    bias        fisher      skewness                kurtosis
    ==========  ==========  ======================  ==============================
    True        True        skewness_moment, g₁     kurtosis_excess, β₂ − 3
    True        False       skewness_moment, g₁     kurtosis_moment, β₂
    False       True        skewness_fisher, G₁     kurtosis_sample_excess, G₂
    False       False       skewness_fisher, G₁     kurtosis_sample, G₂ + 3
    ==========  ==========  ======================  ==============================

    The remaining variants (skewness_sample, kurtosis_sample_corrected),
    which match the R PerformanceAnalytics package, are available as
    separate properties.
    """
    def __init__(self, ddof=1, bias=True, fisher=True) -> None:
        self._n = 0
        self._x1: KleinKBNAccumulator = KleinKBNAccumulator()
        self._x2: KleinKBNAccumulator = KleinKBNAccumulator()
        self._x3: KleinKBNAccumulator = KleinKBNAccumulator()
        self._x4: KleinKBNAccumulator = KleinKBNAccumulator()
        self.ddof = ddof
        self.bias = bias
        self.fisher = fisher
        # Welford's mean and sum of squared deviations Σ(x - x̄)².
        self._mean: KleinKBNAccumulator = KleinKBNAccumulator()
        self._s: KleinKBNAccumulator = KleinKBNAccumulator()

    def reset(self) -> None:
        """Clears all accumulated state."""
        self._n = 0
        self._x1.reset()
        self._x2.reset()
        self._x3.reset()
        self._x4.reset()
        self._mean.reset()
        self._s.reset()

    def update(self, x: float) -> None:
        """Adds a sample x."""
        self._n += 1
        self._x1.update(x)
        x2 = x * x
        self._x2.update(x2)
        x3 = x2 * x
        self._x3.update(x3)
        x4 = x3 * x
        self._x4.update(x4)
        # Welford: mean += (x - mean_old) / n;  S += (x - mean_old)·(x - mean_new)
        delta = x - self._mean.value
        self._mean.update(delta / self._n)
        self._s.update(delta * (x - self._mean.value))

    def revert(self, x: float) -> None:
        """
        Removes a previously added sample x (any sample, not only the most
        recent one).  Reverting a value that was never added corrupts the
        state.

        Raises ValueError if there are no samples.
        """
        if self._n <= 0:
            raise ValueError("Cannot revert from an empty accumulator")
        if self._n == 1:
            self.reset()
            return
        self._n -= 1
        self._x1.revert(x)
        x2 = x * x
        self._x2.revert(x2)
        x3 = x2 * x
        self._x3.revert(x3)
        x4 = x3 * x
        self._x4.revert(x4)
        # Inverse Welford: mean_old = mean_new - (x - mean_new) / (n - 1);
        # S -= (x - mean_new)·(x - mean_old)
        delta = x - self._mean.value
        self._mean.revert(delta / self._n)
        self._s.revert(delta * (x - self._mean.value))

    def _variance(self, ddof: int) -> float:
        # A slightly negative S caused by rounding after revert() is clamped to zero.
        d = self._n - ddof
        if d <= 0:
            return math.nan
        return max(self._s.value, 0.0) / d

    def _standard_deviation(self, ddof: int) -> float:
        v = self._variance(ddof)
        return v if math.isnan(v) else math.sqrt(v)

    @property
    def mean(self) -> float:
        """The arithmetic mean (0.0 when empty)."""
        return self._mean.value

    @property
    def variance(self) -> float:
        """The variance Σ(x - x̄)² / (n - ddof), NaN when n ≤ ddof."""
        return self._variance(self.ddof)

    @property
    def variance_ddof_0(self) -> float:
        """The population variance Σ(x - x̄)² / n, regardless of ddof."""
        return self._variance(0)

    @property
    def variance_ddof_1(self) -> float:
        """The sample variance Σ(x - x̄)² / (n - 1), regardless of ddof."""
        return self._variance(1)

    @property
    def standard_deviation(self) -> float:
        """The square root of variance."""
        return self._standard_deviation(self.ddof)

    @property
    def standard_deviation_ddof_0(self) -> float:
        """The population standard deviation, regardless of ddof."""
        return self._standard_deviation(0)

    @property
    def standard_deviation_ddof_1(self) -> float:
        """The sample standard deviation, regardless of ddof."""
        return self._standard_deviation(1)

    def _central_moments(self) -> tuple[float, float, float] | None:
        """
        Converts the raw power sums to population central moments
        (μ₂, μ₃, μ₄), see the class Notation.

        Returns None when n < 2 or μ₂ is lost in rounding error.
        """
        n = self._n
        if n < 2:
            return None
        mu1 = self._x1.value / n
        r = mu1 * mu1
        mean_x2 = self._x2.value / n
        mu2 = mean_x2 - r
        if mu2 <= _CANCELLATION_EPSILON * mean_x2:
            return None
        r *= mu1
        mu3 = self._x3.value / n - r - 3 * mu1 * mu2
        r *= mu1
        mu4 = self._x4.value / n - r - 6 * mu2 * mu1 * mu1 - 4 * mu3 * mu1
        return mu2, mu3, mu4

    @property
    def _g1(self) -> float:
        r"""The population skewness g₁ = μ₃ / μ₂^1.5."""
        cm = self._central_moments()
        if cm is None:
            return math.nan
        mu2, mu3, _ = cm
        return mu3 / (mu2 * math.sqrt(mu2))

    @property
    def _b2(self) -> float:
        r"""The population (Pearson) kurtosis β₂ = μ₄ / μ₂²."""
        cm = self._central_moments()
        if cm is None:
            return math.nan
        mu2, _, mu4 = cm
        return mu4 / (mu2 * mu2)

    @property
    def skewness_moment(self) -> float:
        r"""
        The 'moment' (biased, population) skewness, requires n ≥ 2:

            g₁ = μ₃ / μ₂^1.5

        Matches scipy.stats.skew(bias=True) and PerformanceAnalytics
        skewness(method="moment").
        """
        return self._g1

    @property
    def skewness_fisher(self) -> float:
        r"""
        The 'fisher' (bias-adjusted Fisher-Pearson) skewness, requires n ≥ 3:

            G₁ = g₁ · √(n(n−1)) / (n−2)

        Matches scipy.stats.skew(bias=False) and PerformanceAnalytics
        skewness(method="fisher").
        """
        g1 = self._g1
        if math.isnan(g1):
            return math.nan
        n = self._n
        return math.nan if n < 3 else g1 * math.sqrt(n * (n - 1)) / (n - 2)

    @property
    def skewness_sample(self) -> float:
        r"""
        The 'sample' skewness, requires n ≥ 3:

            g₁ · n² / ((n−1)(n−2))

        Matches PerformanceAnalytics skewness(method="sample").
        Doesn't depend on the bias parameter.
        """
        g1 = self._g1
        if math.isnan(g1):
            return math.nan
        n = self._n
        return math.nan if n < 3 else g1 * (n * n) / ((n - 1) * (n - 2))

    @property
    def skewness(self) -> float:
        r"""
        The skewness selected by the bias parameter:

        - bias=True:  skewness_moment, g₁
        - bias=False: skewness_fisher, G₁ = g₁ · √(n(n−1)) / (n−2)

        The third variant, skewness_sample, doesn't depend on bias.
        """
        return self.skewness_moment if self.bias else self.skewness_fisher

    @property
    def kurtosis_moment(self) -> float:
        r"""
        The 'moment' (biased, population) Pearson kurtosis, requires n ≥ 2:

            β₂ = μ₄ / μ₂²

        Matches scipy.stats.kurtosis(bias=True, fisher=False) and
        PerformanceAnalytics kurtosis(method="moment").
        """
        return self._b2

    @property
    def kurtosis_excess(self) -> float:
        r"""
        The 'excess' (biased, population) excess kurtosis, requires n ≥ 2:

            β₂ − 3

        Matches scipy.stats.kurtosis(bias=True, fisher=True) and
        PerformanceAnalytics kurtosis(method="excess").
        """
        b2 = self._b2
        if math.isnan(b2):
            return math.nan
        return b2 - 3

    @property
    def kurtosis_sample_excess(self) -> float:
        r"""
        The 'sample excess' (unbiased) excess kurtosis, requires n ≥ 4:

            G₂ = ((n²−1)·β₂ − 3(n−1)²) / ((n−2)(n−3))

        Matches scipy.stats.kurtosis(bias=False, fisher=True) and
        PerformanceAnalytics kurtosis(method="sample_excess").
        """
        b2 = self._b2
        if math.isnan(b2):
            return math.nan
        n = self._n
        if n <= 3:
            return math.nan
        return ((n * n - 1) * b2 - 3 * (n - 1) ** 2) / ((n - 2) * (n - 3))

    @property
    def kurtosis_sample(self) -> float:
        r"""
        The 'sample' (unbiased) Pearson kurtosis, requires n ≥ 4, calculated
        as the 'sample excess' kurtosis plus 3:

            G₂ + 3 = ((n²−1)·β₂ − 3(n−1)²) / ((n−2)(n−3)) + 3

        Matches scipy.stats.kurtosis(bias=False, fisher=False).

        The PerformanceAnalytics kurtosis(method="sample") variant is
        available as kurtosis_sample_corrected; it is larger than this one
        by (9n−15) / ((n−2)(n−3)), approximately 0.44 for n = 24.
        """
        b2 = self._b2
        if math.isnan(b2):
            return math.nan
        n = self._n
        if n <= 3:
            return math.nan
        return ((n * n - 1) * b2 - 3 * (n - 1) ** 2) / ((n - 2) * (n - 3)) + 3

    @property
    def kurtosis_sample_corrected(self) -> float:
        r"""
        The PerformanceAnalytics 'sample' (unbiased) Pearson kurtosis,
        requires n ≥ 4:

            (n²−1)·β₂ / ((n−2)(n−3))

        Matches PerformanceAnalytics kurtosis(method="sample").
        Doesn't depend on the bias and fisher parameters.

        It differs from kurtosis_sample (G₂ + 3) by (9n−15) / ((n−2)(n−3)),
        approximately 0.44 for n = 24.
        """
        b2 = self._b2
        if math.isnan(b2):
            return math.nan
        n = self._n
        if n <= 3:
            return math.nan
        return b2 * (n * n - 1) / ((n - 2) * (n - 3))

    @property
    def kurtosis(self) -> float:
        r"""
        The kurtosis selected by the bias and fisher parameters:

        - bias=True,  fisher=True:  kurtosis_excess, β₂ − 3
        - bias=True,  fisher=False: kurtosis_moment, β₂
        - bias=False, fisher=True:  kurtosis_sample_excess, G₂
        - bias=False, fisher=False: kurtosis_sample, G₂ + 3
        """
        if self.bias:
            return self.kurtosis_excess if self.fisher else self.kurtosis_moment
        return self.kurtosis_sample_excess if self.fisher else self.kurtosis_sample

    @property
    def x1_sum(self) -> float:
        """The sum Σx."""
        return self._x1.value

    @property
    def x2_sum(self) -> float:
        """The sum Σx²."""
        return self._x2.value

    @property
    def x3_sum(self) -> float:
        """The sum Σx³."""
        return self._x3.value

    @property
    def x4_sum(self) -> float:
        """The sum Σx⁴."""
        return self._x4.value

    @property
    def x1(self) -> float:
        """The first raw moment Σx / n (NaN when empty)."""
        return self._x1.value / self._n if self._n > 0 else math.nan

    @property
    def x2(self) -> float:
        """The second raw moment Σx² / n (NaN when empty)."""
        return self._x2.value / self._n if self._n > 0 else math.nan

    @property
    def x3(self) -> float:
        """The third raw moment Σx³ / n (NaN when empty)."""
        return self._x3.value / self._n if self._n > 0 else math.nan

    @property
    def x4(self) -> float:
        """The fourth raw moment Σx⁴ / n (NaN when empty)."""
        return self._x4.value / self._n if self._n > 0 else math.nan

    @property
    def n(self) -> int:
        """The number of samples."""
        return self._n
