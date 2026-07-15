import math

from .klein_kbn_accumulator import KleinKBNAccumulator


##########################################################
# Central moments with KBN (Kahan-Babuška-Neumaier) compensated
# summation for improved numerical stability.
# https://github.com/kuiperzone/Compensated-Accumulators/tree/master/CompensatedAccumulators
# https://www.johndcook.com/skewness_kurtosis.html
# How to implement revert?
##########################################################

class CentralMomentsKleinKBN:
    """
    Streaming mean, variance, skewness, kurtosis via Pébay's central moment
    update with KBN (Kahan-Babuška-Neumaier) double-compensated accumulation.

    Maintains running sums of central moments m₂, m₃, m₄ (as KleinKBNAccumulators)
    updated in O(1) per sample.  Preferred over RawMomentsKleinKBN for forward-only
    computation (no revert) because it avoids the numerical cancellation
    inherent in converting raw power sums to central moments.

    Parameters
    ----------
    ddof : int, default=1
        Delta degrees of freedom for variance.
        variance = m₂ / (n - ddof).  ddof=0 gives population, ddof=1 gives sample.
    bias : bool, default=True
        If True, compute population standardized moments (m₃/m₂^1.5).
        If False, apply the Fisher-Pearson adjusted (bias-corrected) factor:
        skewness_bcf = skewness_pop · √(n·(n-1)) / (n-2)
    fisher : bool, default=True
        If True, return excess kurtosis (subtract 3 so Gaussian→0).
        If False, return raw kurtosis (Gaussian→3).
        Applied after the bias correction when bias=False.

    Notes
    -----
    Skewness (bias=True):
        g₁ = √n · m₃ / m₂^1.5

    Skewness (bias=False):
        G₁ = g₁ · √(n·(n-1)) / (n-2)

    Kurtosis (bias=True, fisher=True):
        g₂ = n · m₄ / m₂²  -  3

    Kurtosis (bias=True, fisher=False):
        g₂ = n · m₄ / m₂²

    Kurtosis (bias=False, fisher=True):
        G₂ = ((n²-1) · n·m₄/m₂²  -  3·(n-1)²) / ((n-2)·(n-3))

    Kurtosis (bias=False, fisher=False):
        G₂ = ((n²-1) · n·m₄/m₂²  -  3·(n-1)²) / ((n-2)·(n-3))  +  3
    """
    def __init__(self, ddof=1, bias=True, fisher=True) -> None:
        self.ddof = ddof
        self.bias = bias
        self.fisher = fisher
        self.n = 0
        self.m1: KleinKBNAccumulator = KleinKBNAccumulator()
        self.m2: KleinKBNAccumulator = KleinKBNAccumulator()
        self.m3: KleinKBNAccumulator = KleinKBNAccumulator()
        self.m4: KleinKBNAccumulator = KleinKBNAccumulator()

    def reset(self) -> None:
        self.n = 0
        self.m1.reset()
        self.m2.reset()
        self.m3.reset()
        self.m4.reset()

    def update(self, x) -> None:
        n_old = self.n
        n_new = n_old + 1
        self.n = n_new
        delta = x - self.m1.value
        delta_n = delta / n_new
        delta_n2 = delta_n * delta_n
        term = delta * delta_n * n_old
        self.m1.update(delta_n)
        self.m4.update(term * delta_n2 * (n_new * n_new - 3 * n_new + 3) + 6 * delta_n2 * self.m2.value - 4 * delta_n * self.m3.value)
        self.m3.update(term * delta_n * (n_new - 2) - 3 * delta_n * self.m2.value)
        self.m2.update(term)

    def revert(self, x) -> None:
        """
        LIFO revert: removes the most recently added sample x, restoring
        the state to exactly what it would be had x never been added.

        Uses the same inverse Pébay formulas as CentralMoments.revert().
        KleinKBNAccumulator.set() is used for m₁–m₄, which resets the
        compensation terms (_cs, _ccs to zero).  This means subsequent
        updates rebuild compensation from the restored value — a minor
        loss of error correction for each revert operation.

        Only the most recent sample can be reverted (LIFO stack, not FIFO
        queue).  For rolling-window FIFO removal use a wrapper that keeps
        a sample deque and calls reset() + forward replay of the reduced
        window, or use RawMoments/RawMomentsKleinKBN which support FIFO revert natively.

        Inverse formulas (where nₙ = count before revert, nₒ = nₙ - 1):

            m₁_old = (nₙ · m₁_new − x) / nₒ            [mean undo]
            δ      = x − m₁_old
            δₙ     = δ / nₙ
            δₙ²    = δₙ · δₙ
            term   = δ · δₙ · nₒ

            m₂_old = m₂_new − term                      [2nd moment undo]
            m₃_old = m₃_new − (term·δₙ·(nₙ−2) − 3·δₙ·m₂_old)
                                                        [3rd moment undo]
            m₄_old = m₄_new − (term·δₙ²·(nₙ²−3nₙ+3)
                                + 6·δₙ²·m₂_old − 4·δₙ·m₃_old)
                                                        [4th moment undo]
        """
        n_new = self.n
        if n_new == 0:
            raise ValueError("Cannot go below 0")
        n_old = n_new - 1
        if n_old == 0:
            self.n = 0
            self.m1.reset()
            self.m2.reset()
            self.m3.reset()
            self.m4.reset()
            return

        m1_new = self.m1.value
        m2_new = self.m2.value
        m3_new = self.m3.value
        m4_new = self.m4.value

        m1_old = (n_new * m1_new - x) / n_old
        delta = x - m1_old
        delta_n = delta / n_new
        delta_n2 = delta_n * delta_n
        term = delta * delta_n * n_old

        m2_old = m2_new - term
        m3_old = m3_new - (term * delta_n * (n_new - 2) - 3 * delta_n * m2_old)
        m4_old = m4_new - (term * delta_n2 * (n_new * n_new - 3 * n_new + 3) + 6 * delta_n2 * m2_old - 4 * delta_n * m3_old)

        self.n = n_old
        self.m1.set(m1_old)
        self.m2.set(m2_old)
        self.m3.set(m3_old)
        self.m4.set(m4_old)

    @property
    def mean(self) -> float:
        return self.m1.value

    @property
    def variance(self) -> float:
        N = self.n - self.ddof
        return self.m2.value / N if N > 0 else float('nan')

    @property
    def standard_deviation(self) -> float:
        N = self.n - self.ddof
        return (self.m2.value / N)**0.5 if N > 0 else float('nan')

    @property
    def skewness(self) -> float:
        N = self.n
        if N < 3 or self.m2.value <= 0:
            return float('nan')
        g1 = math.sqrt(N) * self.m3.value / (self.m2.value ** 1.5)
        if self.bias:
            return g1
        return g1 * math.sqrt(N * (N - 1)) / (N - 2)

    @property
    def kurtosis(self) -> float:
        N = self.n
        if N <= 3 or self.m2.value <= 0:
            return float('nan')
        raw = N * self.m4.value / (self.m2.value * self.m2.value)
        if not self.bias:
            adj = ((N * N - 1) * raw - 3 * (N - 1) ** 2) / ((N - 2) * (N - 3))
            return adj if self.fisher else adj + 3.0
        return raw - 3.0 if self.fisher else raw
