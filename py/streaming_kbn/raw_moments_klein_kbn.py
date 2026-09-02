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
        self._n = 0
        self._x1: KleinKBNAccumulator = KleinKBNAccumulator()
        self._x2: KleinKBNAccumulator = KleinKBNAccumulator()
        self._x3: KleinKBNAccumulator = KleinKBNAccumulator()
        self._x4: KleinKBNAccumulator = KleinKBNAccumulator()
        self.ddof = ddof
        self.bias = bias
        self.fisher = fisher
        # Mean and variance are calculated separately
        self._mean: KleinKBNAccumulator = KleinKBNAccumulator()
        self._s: KleinKBNAccumulator = KleinKBNAccumulator()

    def reset(self) -> None:
        self._n = 0
        self._x1.reset()
        self._x2.reset()
        self._x3.reset()
        self._x4.reset()
        self._mean.reset()
        self._s.reset()

    def update(self, x) -> None:
        self._n += 1
        self._x1.update(x)
        x2 = x * x
        self._x2.update(x2)
        x3 = x2 * x
        self._x3.update(x3)
        x4 = x3 * x
        self._x4.update(x4)
        # variance
        N = self._n
        delta = x - self._mean.value
        self._mean.update(delta / self._n)
        self._s.update(delta * (x - self._mean.value))


    def revert(self, x) -> None:
        self._n -= 1
        self._x1.revert(x)
        x2 = x * x
        self._x2.revert(x2)
        x3 = x2 * x
        self._x3.revert(x3)
        x4 = x3 * x
        self._x4.revert(x4)
        # mean and variance
        delta = x - self._mean.value
        self._mean.revert(delta / self._n)
        self._s.revert(delta * (x - self._mean.value))

    def _variance(self, ddof: int) -> float:
        n = self._n - ddof
        if n <= 0:
            return math.nan
        elif self._s.value < 0:
            self._s.reset()
            return math.nan
        else:
            return self._s.value / n

    def _standard_deviation(self, ddof: int) -> float:
        n = self._n - ddof
        return (self._s.value / n)**0.5 if n > 0 else math.nan

    @property
    def mean(self) -> float:
        return self._mean.value

    @property
    def variance(self) -> float:
        return self._variance(self.ddof)

    @property
    def variance_ddof_0(self) -> float:
        return self._variance(0)

    @property
    def variance_ddof_1(self) -> float:
        return self._variance(1)

    @property
    def standard_deviation(self) -> float:
        return self._standard_deviation(self.ddof)

    @property
    def standard_deviation_ddof_0(self) -> float:
        return self._standard_deviation(0)

    @property
    def standard_deviation_ddof_1(self) -> float:
        return self._standard_deviation(1)

    @property
    def _g1(self) -> float:
        """
        Calculates central moments $\\mu_k$ from the raw moments $\\mu_k^{'}=E[X^k]$

        When calculated, returns

        $$g_1=frac{\\mu_3}{\\mu_2^{3/2}}$$
        """
        n = self._n
        if n < 2:
            return math.nan
        # Conversion from raw moments to central moments
        # $\mu_1=\frac{\mu_1^'}{n}$, 1st central moment (mean)
        mu1 = self._x1.value / n
        # $\mu_1^2$
        r = mu1 * mu1
        # $\mu_2=\frac{\mu_2^'}{n}-\mu_1^2$, 2nd central moment (population variance)
        mu2 = self._x2.value / n - r
        if mu2 <= 1e-14:
            return math.nan
        # $\mu_3=\frac{\mu_3^'}{n}-\mu_1^3-3\mu_1 \cdot \mu_2$, 3rd central moment
        mu3 = self._x3.value / n - r * mu1 - 3 * mu1 * mu2
        # $g_1=]frac{\mu_3}{\mu_2^{3/2}}$
        return mu3 / (mu2 * math.sqrt(mu2))

    @property
    def skewness_moment(self) -> float:
        """
        The 'moment' skewness is calculated as

        $$g_1=\\frac{\\mu_3}{\\mu_2^{3/2}}$$
        """
        # bias=True, 'moment', $g_1=\frac{\mu_3}{\mu_2^{3/2}}$
        return self._g1

    @property
    def skewness_fisher(self) -> float:
        """
        The 'fisher' skewness is calculated as

        $$g_1\\cdot \\frac{\\sqrt{n(n-1)}}{n-2}$$

        where

        $$g_1=\\frac{\\mu_3}{\\mu_2^{3/2}}$$
        """
        # bias=False, 'fisher', $g_1\cdot \frac{\sqrt{n(n-1)}}{n-2}$
        g1 = self._g1
        if math.isnan(g1):
            return math.nan
        n = self._n
        return math.nan if n == 2 else g1 * math.sqrt(n * (n - 1)) / (n - 2)

    @property
    def skewness_sample(self) -> float:
        """
        The 'sample' skewness is calculated as

        $$g_1\\cdot \\frac{n^2}{(n-1)(n-2)}$$

        where

        $$g_1=\\frac{\\mu_3}{\\mu_2^{3/2}}$$
        """
        g1 = self._g1
        if math.isnan(g1):
            return math.nan
        n = self._n
        # $g_1\cdot \frac{n^2}{(n-1)(n-2)}$
        return math.nan if n < 3 else g1 * (n*n) / ((n-1)*(n-2))

    @property
    def skewness(self) -> float:
        """
        The scalculation method depends on `bias` parameter:

        - bias=True: 'moment' method is calculated as
          $$g_1$$
        - bias=False: 'fisher' method is calculated as
          $$g_1\\cdot \\frac{\\sqrt{n(n-1)}}{n-2}$$

        where

        $$g_1=\\frac{\\mu_3}{\\mu_2^{3/2}}$$

        There is a third method, 'sample', which is implemented as a separate
        `skewness_sample` property because it doesn't depend on the `bias` parameter.
        """
        return self.skewness_moment if self.bias else self.skewness_fisher

    @property
    def _b2(self) -> float:
        """
        Calculates central moments $\\mu_k$ from the raw moments $\\mu_k^{'}=E[X^k]$

        When calculated, returns

        $$\\beta_2=\\frac{\\mu_4}{\\mu_2^2}$$
        """
        n = self._n
        if n < 2:
            return math.nan
        # Conversion from raw moments to central moments
        # $\mu_1=\frac{\mu_1^'}{n}$, 1st central moment (mean)
        mu1 = self._x1.value / n
        # $\mu_1^2$
        r = mu1 * mu1
         # $\mu_2=\frac{\mu_2^'}{n}-\mu_1^2$, 2nd central moment (population variance)
        mu2 = self._x2.value / n - r
        if mu2 <= 1e-14:
            return math.nan
        # $\mu_1^3$
        r *= mu1
        # $\mu_3=\frac{\mu_3^'}{n}-\mu_1^3-3\mu_1 \cdot \mu_2$, 3rd central moment
        mu3 = self._x3.value / n - r - 3 * mu1 * mu2
        # $\mu_1^4$
        r *= mu1
        # $\mu_4=\frac{\mu_4^'}{n}-\mu_1^4-6\mu_1^2\mu_2-4\mu_3\mu_1$, 4th central moment
        mu4 = self._x4.value / n - r - 6 * mu2 * mu1 * mu1 - 4 * mu3 * mu1
        # $\beta_2=\frac{\mu_4}{\mu_2^2}$, population moment kurtosis (bias=True, fisher=False)
        return mu4 / (mu2 * mu2)

    @property
    def kurtosis_moment(self) -> float:
        """
        The 'moment' biased Pearson (population) kurtosis is calculated as

        $$\\beta_2=\\frac{\\mu_4}{\\mu_2^2}$$
        """
        # bias=True, fisher=False,'moment', $\beta_2=\frac{\mu_4}{\mu_2^2}$
        return self._b2

    @property
    def kurtosis_excess(self) -> float:
        """
        The 'excess' biased excess kurtosis is calculated as

        $$\\beta_2-3$$

        where

        $$\\beta_2=\\frac{\\mu_4}{\\mu_2^2}$$
        """
        # bias=True, fisher=True,'excess', $\beta_2-3=\frac{\mu_4}{\mu_2^2}-3$
        b2 = self._b2
        if math.isnan(b2):
            return math.nan
        return b2 - 3

    @property
    def kurtosis_sample_excess(self) -> float:
        """
        The 'sample excess' unbiased excess kurtosis is calculated as

        $$\\frac{(n^2-1)\\beta_2-3(n - 1)^2}{(n - 2)(n - 3)}$$

        where

        $$\\beta_2=\\frac{\\mu_4}{\\mu_2^2}$$
        """
        # bias=False, fisher=True,'sample excess', unbiased excess kurtosis, $\frac{(n^2-1)\beta_2-3(n - 1)^2}{(n - 2)(n - 3)}$
        b2 = self._b2
        if math.isnan(b2):
            return math.nan
        n = self._n
        if n <= 3:
            return math.nan
        return ((n * n - 1) * b2 - 3 * (n - 1) ** 2) / ((n - 2) * (n - 3))

    @property
    def kurtosis_sample_corrected(self) -> float:
        """
        The 'sample' unbiased Pearson kurtosis is calculated as

        $$\\frac{(n^2 - 1)\\beta_2}{(n-2)(n-3)}$$

        where

        $$\\beta_2=\\frac{\\mu_4}{\\mu_2^2}$$

        This variant is compatible with `PerformanceAnalytics` R package implementation.
        There is another version of the calculation which is calculated as the 'sample excess' kurtosis plus 3,

        $$\\frac{(n^2-1)\\beta_2-3(n - 1)^2}{(n - 2)(n - 3)}+3$$

        This version differs by

        $$\frac{9n-15}{(n-2)(n-3)}$$

        from the R package version mentioned above.
        The difference is approximately $0.44$ for $n=24$.
        """
        # bias=False, fisher=False -> 'sample', unbiased Pearson kurtosis, $\frac{(n^2 - 1)\beta_2}{(n-2)(n-3)}$
        b2 = self._b2
        if math.isnan(b2):
            return math.nan
        n = self._n
        if n <= 3:
            return math.nan
        return b2 * (n*n - 1) / ((n-2)*(n-3))

    @property
    def kurtosis_sample(self) -> float:
        """
        The 'sample' unbiased Pearson kurtosis is calculated as the 'sample excess' kurtosis plus 3,

        $$\\frac{(n^2-1)\\beta_2-3(n - 1)^2}{(n - 2)(n - 3)}+3$$

        where

        $$\\beta_2=\\frac{\\mu_4}{\\mu_2^2}$$

        There is another version of the calculation which is compatible with `PerformanceAnalytics` R package implementation,

        $$\\frac{(n^2 - 1)\\beta_2}{(n-2)(n-3)}$$

        This version differs by

        $$\frac{9n-15}{(n-2)(n-3)}$$

        from the R package version mentioned above.
        The difference is approximately $0.44$ for $n=24$.
        """
        # bias=False, fisher=False -> 'sample', unbiased Pearson kurtosis, $\frac{(n^2-1)\beta_2-3(n - 1)^2}{(n - 2)(n - 3)}+3$
        b2 = self._b2
        if math.isnan(b2):
            return math.nan
        n = self._n
        if n <= 3:
            return math.nan
        return ((n * n - 1) * b2 - 3 * (n - 1) ** 2) / ((n - 2) * (n - 3)) + 3

    @property
    def kurtosis(self) -> float:
        if self.bias: # Biased estimator
            return self.kurtosis_excess if self.fisher else self.kurtosis_moment
        else: # Unbiased estimator
            return self.kurtosis_sample_excess if self.fisher else self.kurtosis_sample_corrected

    @property
    def x1_sum(self) -> float:
        return self._x1.value

    @property
    def x2_sum(self) -> float:
        return self._x2.value

    @property
    def x3_sum(self) -> float:
        return self._x3.value

    @property
    def x4_sum(self) -> float:
        return self._x4.value

    @property
    def x1(self) -> float:
        return self._x1.value / self._n if self._n > 0 else math.nan

    @property
    def x2(self) -> float:
        return self._x2.value / self._n if self._n > 0 else math.nan

    @property
    def x3(self) -> float:
        return self._x3.value / self._n if self._n > 0 else math.nan

    @property
    def x4(self) -> float:
        return self._x4.value / self._n if self._n > 0 else math.nan

    @property
    def n(self) -> int:
        return self._n
