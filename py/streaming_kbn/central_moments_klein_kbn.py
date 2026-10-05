import math

from .klein_kbn_accumulator import KleinKBNAccumulator


##########################################################
# Central moments with Klein KBN (Kahan-Babuška-Neumaier)
# compensated summation for improved numerical stability.
#
# References:
#   P. Pébay, "Formulas for Robust, One-Pass Parallel Computation
#     of Covariances and Arbitrary-Order Statistical Moments",
#     Sandia Report SAND2008-6212 (2008).
#   https://www.johndcook.com/skewness_kurtosis.html
#   https://github.com/kuiperzone/Compensated-Accumulators
##########################################################

class CentralMomentsKleinKBN:
    r"""
    Streaming mean, variance, skewness, kurtosis via Pébay's central moment
    update with Klein KBN (Kahan-Babuška-Neumaier) compensated accumulation.

    Maintains the mean M₁ and the sums of central powers

        M₂ = Σ(x - x̄)²,   M₃ = Σ(x - x̄)³,   M₄ = Σ(x - x̄)⁴

    (each as a KleinKBNAccumulator), updated in O(1) per sample.
    The population central moments are μₖ = Mₖ / n.

    Avoids the catastrophic cancellation inherent in converting raw power
    sums Σxᵏ to central moments.  This matters for data with a large mean
    relative to its spread.  Inverse updates can remove any previously
    added sample, including the oldest sample in a FIFO rolling window.
    Reversion clears the compensation terms, so repeated removals can
    accumulate rounding error.

    Parameters
    ----------
    ddof : nonnegative int, default=1
        Delta degrees of freedom for variance.
        variance = M₂ / (n - ddof).  ddof=0 gives population, ddof=1 gives sample.
    bias : bool, default=True
        If True, return the biased (population) skewness and kurtosis.
        If False, apply the bias corrections (see Notes).
    fisher : bool, default=True
        If True, return excess kurtosis (subtract 3 so Gaussian→0).
        If False, return raw (Pearson) kurtosis (Gaussian→3).
        Applied after the bias correction when bias=False.

    Notes
    -----
    The results match scipy.stats.skew(bias=...) and
    scipy.stats.kurtosis(bias=..., fisher=...).

    Skewness (bias=True), requires n ≥ 2:
        g₁ = μ₃ / μ₂^1.5 = √n · M₃ / M₂^1.5

    Skewness (bias=False), requires n ≥ 3:
        G₁ = g₁ · √(n·(n-1)) / (n-2)

    Kurtosis (bias=True), requires n ≥ 2:
        β₂ = μ₄ / μ₂² = n · M₄ / M₂²
        fisher=True:  g₂ = β₂ - 3
        fisher=False: β₂

    Kurtosis (bias=False), requires n ≥ 4:
        G₂ = ((n²-1) · β₂  -  3·(n-1)²) / ((n-2)·(n-3))
        fisher=True:  G₂
        fisher=False: G₂ + 3

    Skewness and kurtosis are NaN when M₂ = 0 (constant data).
    """
    def __init__(self, ddof=1, bias=True, fisher=True) -> None:
        if type(ddof) is not int or ddof < 0:
            raise ValueError("ddof must be a nonnegative integer")
        self.ddof = ddof
        self.bias = bias
        self.fisher = fisher
        self._n = 0
        self._m1: KleinKBNAccumulator = KleinKBNAccumulator()
        self._m2: KleinKBNAccumulator = KleinKBNAccumulator()
        self._m3: KleinKBNAccumulator = KleinKBNAccumulator()
        self._m4: KleinKBNAccumulator = KleinKBNAccumulator()

    def reset(self) -> None:
        """Clears all accumulated state."""
        self._n = 0
        self._m1.reset()
        self._m2.reset()
        self._m3.reset()
        self._m4.reset()

    def update(self, x: float) -> None:
        r"""
        Adds a sample x using Pébay's update (n = count after adding x):

            δ    = x − M₁
            δₙ   = δ / n
            term = δ · δₙ · (n − 1)

            M₁ += δₙ
            M₄ += term·δₙ²·(n²−3n+3) + 6·δₙ²·M₂ − 4·δₙ·M₃
            M₃ += term·δₙ·(n−2) − 3·δₙ·M₂
            M₂ += term

        M₄ and M₃ are updated before M₂ and M₃ respectively, because
        they use the values from before x was added.
        """
        n_old = self._n
        n_new = n_old + 1
        self._n = n_new
        delta = x - self._m1.value
        delta_n = delta / n_new
        delta_n2 = delta_n * delta_n
        term = delta * delta_n * n_old
        m2 = self._m2.value
        m3 = self._m3.value
        self._m1.update(delta_n)
        self._m4.update(term * delta_n2 * (n_new * n_new - 3 * n_new + 3) + 6 * delta_n2 * m2 - 4 * delta_n * m3)
        self._m3.update(term * delta_n * (n_new - 2) - 3 * delta_n * m2)
        self._m2.update(term)

    def revert(self, x: float) -> None:
        r"""
        Removes a previously added sample x, regardless of insertion order.
        Reverting a value that was never added corrupts the state.

        The restored M₁–M₄ are written with KleinKBNAccumulator.set(), which
        clears their compensation terms.  Subsequent updates rebuild the
        compensation from the restored values.  Repeated reverts can
        accumulate rounding error, especially for large-offset data.

        Inverse formulas (where nₙ = count before revert, nₒ = nₙ − 1):

            M₁_old = (nₙ · M₁_new − x) / nₒ            [mean undo]
            δ      = x − M₁_old
            δₙ     = δ / nₙ
            term   = δ · δₙ · nₒ

            M₂_old = M₂_new − term
            M₃_old = M₃_new − (term·δₙ·(nₙ−2) − 3·δₙ·M₂_old)
            M₄_old = M₄_new − (term·δₙ²·(nₙ²−3nₙ+3)
                                + 6·δₙ²·M₂_old − 4·δₙ·M₃_old)

        Raises ValueError if there are no samples.
        """
        n_new = self._n
        if n_new == 0:
            raise ValueError("Cannot revert from an empty accumulator")
        n_old = n_new - 1
        if n_old == 0:
            self.reset()
            return

        m1_new = self._m1.value
        m2_new = self._m2.value
        m3_new = self._m3.value
        m4_new = self._m4.value

        m1_old = (n_new * m1_new - x) / n_old
        delta = x - m1_old
        delta_n = delta / n_new
        delta_n2 = delta_n * delta_n
        term = delta * delta_n * n_old

        m2_old = m2_new - term
        m3_old = m3_new - (term * delta_n * (n_new - 2) - 3 * delta_n * m2_old)
        m4_old = m4_new - (term * delta_n2 * (n_new * n_new - 3 * n_new + 3) + 6 * delta_n2 * m2_old - 4 * delta_n * m3_old)

        self._n = n_old
        self._m1.set(m1_old)
        self._m2.set(m2_old)
        self._m3.set(m3_old)
        self._m4.set(m4_old)

    @property
    def n(self) -> int:
        """The number of samples."""
        return self._n

    @property
    def mean(self) -> float:
        """The arithmetic mean (0.0 when empty)."""
        return self._m1.value

    @property
    def variance(self) -> float:
        """
        The variance M₂ / (n - ddof), NaN when n ≤ ddof.

        A slightly negative M₂ caused by rounding after revert()
        is clamped to zero.
        """
        d = self._n - self.ddof
        if d <= 0:
            return math.nan
        return max(self._m2.value, 0.0) / d

    @property
    def standard_deviation(self) -> float:
        """The square root of the variance, NaN when n ≤ ddof."""
        v = self.variance
        return v if math.isnan(v) else math.sqrt(v)

    @property
    def skewness(self) -> float:
        """The skewness g₁ (bias=True) or G₁ (bias=False); see the class Notes."""
        n = self._n
        m2 = self._m2.value
        if n < 2 or m2 <= 0:
            return math.nan
        g1 = math.sqrt(n) * self._m3.value / (m2 * math.sqrt(m2))
        if self.bias:
            return g1
        if n < 3:
            return math.nan
        return g1 * math.sqrt(n * (n - 1)) / (n - 2)

    @property
    def kurtosis(self) -> float:
        """The kurtosis selected by bias and fisher; see the class Notes."""
        n = self._n
        m2 = self._m2.value
        if n < 2 or m2 <= 0:
            return math.nan
        b2 = n * self._m4.value / (m2 * m2)
        if self.bias:
            return b2 - 3.0 if self.fisher else b2
        if n < 4:
            return math.nan
        g2 = ((n * n - 1) * b2 - 3 * (n - 1) ** 2) / ((n - 2) * (n - 3))
        return g2 if self.fisher else g2 + 3.0
