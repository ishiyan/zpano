# https://www.wikiwand.com/en/Algorithms_for_calculating_variance#/Welford's_online_algorithm
# https://github.com/online-ml/river/blob/main/river/stats/mean.py
# https://www.johndcook.com/blog/distribution_chart/
# https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance

from __future__ import division
import collections

import math

##########################################################
# Mean and Variance with update/revert methods
##########################################################

class Mean:
    def __init__(self) -> None:
        self.n = 0
        self._mean = 0.0

    def update(self, x) -> None:
        self.n += 1
        self._mean += (1 / self.n) * (x - self._mean)

    def revert(self, x) -> None:
        self.n -= 1
        if self.n < 0:
            raise ValueError("Cannot go below 0")
        elif self.n == 0:
            self._mean = 0.0
        else:
            self._mean -= (1 / self.n) * (x - self._mean)

    @property
    def mean(self) -> float:
        return self._mean

class Variance:
    """Running variance using Welford's algorithm."""
    def __init__(self, ddof=1) -> None:
        """
        Args:
            ddof (int): Delta degrees of freedom. The divisor used in the variance calculation is (n - ddof).
        """

        self.ddof = ddof
        self._mean = Mean()
        self._S = 0

    def update(self, x) -> None:
        mean_old = self._mean.mean
        self._mean.update(x)
        mean_new = self._mean.mean
        self._S += (x - mean_old) * (x - mean_new)

    def revert(self, x) -> None:
        mean_old = self._mean.mean
        self._mean.revert(x)
        mean_new = self._mean.mean
        self._S -= (x - mean_old) * (x - mean_new)

    @property
    def mean(self) -> float:
        return self._mean.mean

    @property
    def variance(self) -> float:
        n = self._mean.n
        return self._S / (n - self.ddof) if n > self.ddof else 0.0

class RunningVariance:
    def __init__(self, ddof=1, window_size:int = None) -> None:
        self.window_size = None if window_size is not None and window_size <= 0 else window_size
        self.window = collections.deque(maxlen=window_size)
        self._variance: Variance = Variance(ddof=ddof)

    def update(self, x) -> None:
        win = self.window
        if self.window_size is not None and len(win) == self.window_size:
            x_old = win.popleft()
            self._variance.revert(x_old)
        self._variance.update(x)
        win.append(x)

    @property
    def mean(self) -> float:
        return self._variance.mean

    @property
    def variance(self) -> float:
        return self._variance.variance

##########################################################
# Central moments with update method.
# How to implement revert?
##########################################################

# https://www.johndcook.com/skewness_kurtosis.html
class CentralMoments:
    """
    Streaming mean, variance, skewness, kurtosis via Pébay's online central
    moment update (O(1) per sample, no compensation).

    Maintains running sums of central moments m₁, m₂, m₃, m₄ as plain floats.

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
        self.m1 = 0.0
        self.m2 = 0.0
        self.m3 = 0.0
        self.m4 = 0.0

    def update(self, x) -> None:
        n_old = self.n
        n_new = n_old + 1
        self.n = n_new
        delta = x - self.m1
        delta_n = delta / n_new
        delta_n2 = delta_n * delta_n
        term = delta * delta_n * n_old
        self.m1 += delta_n
        self.m4 += term * delta_n2 * (n_new * n_new - 3 * n_new + 3) + 6 * delta_n2 * self.m2 - 4 * delta_n * self.m3
        self.m3 += term * delta_n * (n_new - 2) - 3 * delta_n * self.m2
        self.m2 += term

    def revert(self, x) -> None:
        """
        LIFO revert: removes the most recently added sample x, restoring
        the state to exactly what it would be had x never been added.

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
            self.m1 = 0.0
            self.m2 = 0.0
            self.m3 = 0.0
            self.m4 = 0.0
            return

        m1_new = self.m1
        m2_new = self.m2
        m3_new = self.m3
        m4_new = self.m4

        m1_old = (n_new * m1_new - x) / n_old
        delta = x - m1_old
        delta_n = delta / n_new
        delta_n2 = delta_n * delta_n
        term = delta * delta_n * n_old

        m2_old = m2_new - term
        m3_old = m3_new - (term * delta_n * (n_new - 2) - 3 * delta_n * m2_old)
        m4_old = m4_new - (term * delta_n2 * (n_new * n_new - 3 * n_new + 3) + 6 * delta_n2 * m2_old - 4 * delta_n * m3_old)

        self.n = n_old
        self.m1 = m1_old
        self.m2 = m2_old
        self.m3 = m3_old
        self.m4 = m4_old

    @property
    def mean(self) -> float:
        return self.m1

    @property
    def variance(self) -> float:
        N = self.n - self.ddof
        return self.m2 / N if N > 0 else 0.0

    @property
    def standard_deviation(self) -> float:
        N = self.n - self.ddof
        return (self.m2 / N)**0.5 if N > 0 else 0.0

    @property
    def skewness(self) -> float:
        N = self.n
        if N < 3 or self.m2 <= 0:
            return 0.0
        g1 = math.sqrt(N) * self.m3 / (self.m2 ** 1.5)
        if self.bias:
            return g1
        return g1 * math.sqrt(N * (N - 1)) / (N - 2)

    @property
    def kurtosis(self) -> float:
        N = self.n
        if N <= 3 or self.m2 <= 0:
            return 0.0
        raw = N * self.m4 / (self.m2 * self.m2)
        if not self.bias:
            adj = ((N * N - 1) * raw - 3 * (N - 1) ** 2) / ((N - 2) * (N - 3))
            return adj if self.fisher else adj + 3.0
        return raw - 3.0 if self.fisher else raw

##########################################################
# Linear regression with update method.
# How to implement revert?
##########################################################

# https://www.johndcook.com/running_regression.html
class Regression:
    def __init__(self) -> None:
        self.n = 0
        self.S_xy = 0.0
        self.x_moments: CentralMoments = CentralMoments()
        self.y_moments: CentralMoments = CentralMoments()

    def update(self, x, y) -> None:
        n_old = self.n
        self.n += 1
        self.S_xy += (self.x_moments.mean - x) * (self.y_moments.mean - y) * n_old / (n_old + 1)
        self.x_moments.update(x)
        self.y_moments.update(y)

    @property
    def slope(self) -> float:
        S_xx = self.x_moments.variance * (self.n - 1)
        return self.S_xy / S_xx if S_xx != 0 else 0.0

    @property
    def intercept(self) -> float:
        return self.y_moments.mean - self.slope * self.x_moments.mean

    @property
    def correlation(self) -> float:
        t = self.x_moments.standard_deviation * self.y_moments.standard_deviation
        return self.S_xy / (t * (self.n - 1)) if self.n > 1 else 0.0


# https://stackoverflow.com/questions/5147378/rolling-variance-algorithm

class RunningStats1:
    def __init__(self, window_size=20) -> None:
        self.n = 0
        self._mean = 0.0
        self._var = 0.0
        self.window_size = window_size

        self.window = collections.deque(maxlen=window_size)

    def reset(self) -> None:
        self.n = 0
        self._mean = 0.0
        self._var = 0.0
        self.window.clear()

    def update(self, x) -> None:
        self.window.append(x)

        if self.n <= self.window_size:
            # Calculating first variance
            self.n += 1
            delta = x - self._mean
            self._mean += delta / self.n
            self._var += delta * (x - self._mean)
        else:
            # Adjusting variance
            x_removed = self.window.popleft() # or self.window[0] if you want to keep the window intact
            mean_old = self._mean
            self._mean += (x - x_removed) / self.window_size
            self._var += (x + x_removed - mean_old - self._mean) * (x - x_removed)

    @property
    def mean(self) -> float:
        return self._mean if self.n else 0.0

    @property
    def variance(self) -> float:
        return self._var / (self.window_size - 1) if self.n > 1 else 0.0

    @property
    def standard_deviation(self) -> float:
        return (self.variance)**0.5

# https://github.com/ajcr/rolling/blob/master/rolling/stats/mean.py
# https://github.com/ajcr/rolling/blob/master/rolling/stats/variance.py
# https://github.com/ajcr/rolling/blob/master/rolling/stats/skew.py
# https://github.com/ajcr/rolling/blob/master/rolling/stats/kurtosis.py

class RawMoments:
    """
    Streaming mean, variance, skewness, kurtosis via raw power sums (x¹..x⁴)
    with plain (uncompensated) accumulation.

    Accumulates Σx, Σx², Σx³, Σx⁴ as plain floats, plus a separate
    Welford-style variance tracker.  Converts raw sums to central moments
    at query time.  May suffer from catastrophic cancellation for large
    values — use RawMomentsKleinKBN or CentralMomentsKleinKBN for better numerical stability.

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
        self._x1 = 0.0
        self._x2 = 0.0
        self._x3 = 0.0
        self._x4 = 0.0
        self.ddof = ddof
        self.bias = bias
        self.fisher = fisher
        # variance is calculated separately
        # HOW TO COMBINE IT WITH _x1 ... _x4?
        self._mean = 0.0 # mean of values
        self._s = 0.0 # sum of squared values minus the mean

    def reset(self) -> None:
        self.n = 0
        self._x1 = 0.0
        self._x2 = 0.0
        self._x3 = 0.0
        self._x4 = 0.0
        self._mean = 0.0
        self._s = 0.0

    def update(self, x) -> None:
        self.n += 1
        self._x1 += x
        x2 = x * x
        self._x2 += x2
        x3 = x2 * x
        self._x3 += x3
        x4 = x3 * x
        self._x4 += x4
        # variance
        N = self.n
        delta = x - self._mean
        self._mean += delta / N if N > 0 else 0.0
        self._s += delta * (x - self._mean)

    def revert(self, x) -> None:
        self.n -= 1
        self._x1 -= x
        x2 = x * x
        self._x2 -= x2
        x3 = x2 * x
        self._x3 -= x3
        x4 = x3 * x
        self._x4 -= x4
        # variance
        N = self.n
        delta = x - self._mean
        self._mean -= delta / N if N > 0 else 0.0
        self._s -= delta * (x - self._mean)

    @property
    def mean(self) -> float:
        return self._mean

    @property
    def variance(self) -> float:
        N = self.n - self.ddof
        if N <= 0:
            return float('nan')
        elif self._s < 0:
            self._s = 0.0
            return float('nan')
        else:
            return self._s / N

    @property
    def skewness(self) -> float:
        N = self.n
        if N < 3:
            return float('nan')
        A = self._x1 / N
        B = self._x2 / N - A * A
        if B <= 1e-14:
            return float('nan')
        R = math.sqrt(B)
        C = self._x3 / N - A * A * A - 3 * A * B
        g1 = C / (R * R * R)
        if self.bias:
            return g1
        return g1 * math.sqrt(N * (N - 1)) / (N - 2)
        
    @property
    def kurtosis(self) -> float:
        N = self.n
        if N <= 3:
            return float('nan')
        A = self._x1 / N
        R = A * A
        B = self._x2 / N - R
        if B <= 1e-14:
            return float("nan")
        R *= A
        C = self._x3 / N - R - 3 * A * B
        R *= A
        D = self._x4 / N - R - 6 * B * A * A - 4 * C * A
        raw = D / (B * B)
        if not self.bias:
            adj = ((N * N - 1) * raw - 3 * (N - 1) ** 2) / ((N - 2) * (N - 3))
            return adj if self.fisher else adj + 3.0
        return raw - 3.0 if self.fisher else raw

# REFERENCE IS BROKEN
# stackoverflow.com/questions/6446729

class Stats3:
    def __init__(self, ddof=1) -> None:
        self.ddof = ddof
        self.n = 0
        self._m1 = 0.0
        self._m2 = 0.0
        self._m3 = 0.0
        self._m4 = 0.0
        self._s = 0.0
        self._sum = 0.0

    def reset(self) -> None:
        self.n = 0
        self._m1 = 0.0
        self._m2 = 0.0
        self._m3 = 0.0
        self._m4 = 0.0
        self._s = 0.0
        self._sum = 0.0

    def update(self, x) -> None:
        n_old = self.n
        self.n += 1
        n_new = self.n
        self._sum += x
        delta = x - self._m1
        delta_n = delta / n_new
        delta_n2 = delta_n * delta_n
        term1 = delta * delta_n * n_old
        self._m1 += delta_n
        self._m3 += term1 * delta_n * (n_new - 2) - 3 * delta_n * self._m2
        self._m4 += term1 * delta_n2 * (n_new * n_new - 3 * n_new + 3) + 6 * delta_n2 * self._m2 - 4 * delta_n * self._m3
        self._m2 += term1 # SHOULD THIS BE IN FRONT OF M3, M4 ???

    # THIS LOOKS VERY SUSPICIOUS, NEEDS TO BE CHECKED
    def revert(self, x) -> None:
        self._sum -= x;
        o = ((self._m1 * self.n) - x) / (self.n - 1) if self.n > 1 else 0.0
        v = self._m2
        y2 = (-(self.n - 1)*o*o + (2*(self.n - 1)*o*x)) + (self.n * (v-x*x)) + x*x/self.n
        self.n -= 1
        self._m1 = 0
        self._m2 = y2


class KahanWelfordVariance:
    def __init__(self, ddof=1) -> None:
        self.ddof = ddof
        self.n = 0
        self._mean = 0.0
        self._m2 = 0.0
        # Kahan compensation terms
        self._c_mean = 0.0
        self._c_m2 = 0.0

    def reset(self) -> None:
        self.n = 0
        self._mean = 0.0
        self._m2 = 0.0
        self._c_mean = 0.0
        self._c_m2 = 0.0

    def update(self, x) -> None:
        n_old = self.n
        self.n += 1
        n_new = self.n

        delta = x - self._mean
        # Update mean with Kahan compensation
        delta_kahan = delta/n_new - self._c_mean
        mean_new = self._mean + delta_kahan
        self._c_mean = (mean_new - self._mean) - delta_kahan
        self._mean = mean_new

        delta2 = x - self._mean
        term = delta * delta2
        # Update m2 with Kahan compensation
        term_kahan = term - self._c_m2
        m2_new = self._m2 + term_kahan
        self._c_m2 = (m2_new - self._m2) - term_kahan
        self._m2 = m2_new
    
    @property
    def mean(self) -> float:
        return self._mean

    @property
    def variance(self) -> float:
        N = self.n
        return self._m2 / (N - self.ddof) if N > self.ddof else 0.0

from .raw_moments_klein_kbn import RawMomentsKleinKBN
from .central_moments_klein_kbn import CentralMomentsKleinKBN
