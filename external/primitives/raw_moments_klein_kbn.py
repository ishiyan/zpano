import math

from .klein_kbn_accumulator import KleinKBNAccumulator


##########################################################
# RawMoments with KBN (Kahan-Babuška-Neumaier) compensated summation
# for improved numerical stability.
# https://github.com/kuiperzone/Compensated-Accumulators/tree/master/CompensatedAccumulators
##########################################################

class RawMomentsKleinKBN:
    """
    Streaming mean, variance, skewness, kurtosis via raw power sums (x¹..x⁴)
    with KBN (Kahan-Babuška-Neumaier) double-compensated accumulation.

    Accumulates Σx, Σx², Σx³, Σx⁴ using KleinKBNAccumulator for each,
    plus a separate Welford-style variance tracker (also KBN-compensated).
    Converts raw sums to central moments at query time.

    Supports both LIFO revert (undo the most recent update) and FIFO
    rolling window (via the revert/update cycle) because subtracting
    from a linear sum preserves the KBN compensation state.

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
        self.n = 0
        self._x1: KleinKBNAccumulator = KleinKBNAccumulator()
        self._x2: KleinKBNAccumulator = KleinKBNAccumulator()
        self._x3: KleinKBNAccumulator = KleinKBNAccumulator()
        self._x4: KleinKBNAccumulator = KleinKBNAccumulator()
        self.ddof = ddof
        self.bias = bias
        self.fisher = fisher
        # variance is calculated separately
        # HOW TO COMBINE IT WITH _x1 ... _x4?
        self._mean: KleinKBNAccumulator = KleinKBNAccumulator()
        self._s: KleinKBNAccumulator = KleinKBNAccumulator()

    def reset(self) -> None:
        self.n = 0
        self._x1.reset()
        self._x2.reset()
        self._x3.reset()
        self._x4.reset()
        self._mean.reset()
        self._s.reset()

    def update(self, x) -> None:
        self.n += 1
        self._x1.update(x)
        x2 = x * x
        self._x2.update(x2)
        x3 = x2 * x
        self._x3.update(x3)
        x4 = x3 * x
        self._x4.update(x4)
        # variance
        N = self.n
        delta = x - self._mean.value
        self._mean.update(delta / self.n)
        self._s.update(delta * (x - self._mean.value))


    def revert(self, x) -> None:
        self.n -= 1
        self._x1.update(-x)
        x2 = x * x
        self._x2.update(-x2)
        x3 = x2 * x
        self._x3.update(-x3)
        x4 = x3 * x
        self._x4.update(-x4)
        # mean and variance
        delta = x - self._mean.value
        self._mean.update(-delta / self.n)
        self._s.update(-delta * (x - self._mean.value))

    @property
    def mean(self) -> float:
        return self._mean.value

    @property
    def variance(self) -> float:
        N = self.n - self.ddof
        if N <= 0:
            return float('nan')
        elif self._s.value < 0:
            self._s.reset()
            return float('nan')
        else:
            return self._s.value / N

    @property
    def standard_deviation(self) -> float:
        N = self.n - self.ddof
        return (self._s.value / N)**0.5 if N > 0 else float('nan')

    @property
    def skewness(self) -> float:
        N = self.n
        if N < 3:
            return float('nan')
        A = self._x1.value / N
        B = self._x2.value / N - A * A
        if B <= 1e-14:
            return float('nan')
        R = math.sqrt(B)
        C = self._x3.value / N - A * A * A - 3 * A * B
        g1 = C / (R * R * R)
        if self.bias:
            return g1
        return g1 * math.sqrt(N * (N - 1)) / (N - 2)

    @property
    def kurtosis(self) -> float:
        N = self.n
        if N <= 3:
            return float('nan')
        A = self._x1.value / N
        R = A * A
        B = self._x2.value / N - R
        if B <= 1e-14:
            return float("nan")
        R *= A
        C = self._x3.value / N - R - 3 * A * B
        R *= A
        D = self._x4.value / N - R - 6 * B * A * A - 4 * C * A
        raw = D / (B * B)
        if not self.bias:
            adj = ((N * N - 1) * raw - 3 * (N - 1) ** 2) / ((N - 2) * (N - 3))
            return adj if self.fisher else adj + 3.0
        return raw - 3.0 if self.fisher else raw
