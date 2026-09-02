import math

from .klein_kbn_accumulator import KleinKBNAccumulator
from .raw_moments_klein_kbn import RawMomentsKleinKBN


class LinearRegressionKleinKBN:
    def __init__(self) -> None:
        self.n = 0
        self._x_moments: RawMomentsKleinKBN = RawMomentsKleinKBN(ddof=0)
        self._y_moments: RawMomentsKleinKBN = RawMomentsKleinKBN(ddof=0)
        self._s_xy: KleinKBNAccumulator = KleinKBNAccumulator()

    def reset(self) -> None:
        self.n = 0
        self._x_moments.reset()
        self._y_moments.reset()
        self._s_xy.reset()

    def update(self, x: float, y: float) -> None:
        n_old = self.n
        self.n += 1
        term = (self._x_moments.mean - x) * (self._y_moments.mean - y) * n_old / (n_old + 1)
        self._s_xy.update(term)
        self._x_moments.update(x)
        self._y_moments.update(y)

    def revert(self, x: float, y: float) -> None:
        if self.n == 0:
            return
        if self.n == 1:
            self.reset()
            return
        self._x_moments.revert(x)
        self._y_moments.revert(y)
        n = self.n - 1
        term = (self._x_moments.mean - x) * (self._y_moments.mean - y) * n / (n + 1)
        self._s_xy.revert(term)
        self.n = n

    @property
    def variance_x(self) -> float:
        return self._x_moments.variance

    @property
    def covariance(self) -> float:
        return self._s_xy.value

    @property
    def slope(self) -> float:
        n = self.n
        if n < 2:
            return math.nan
        s_xx = self._x_moments.variance * n
        return self._s_xy.value / s_xx if s_xx != 0 else math.nan

    @property
    def intercept(self) -> float:
        return self._y_moments.mean - self.slope * self._x_moments.mean

    @property
    def correlation(self) -> float:
        n = self.n
        if n < 2:
            return math.nan
        t = self._x_moments.standard_deviation * self._y_moments.standard_deviation
        return self._s_xy.value / (t * n) if t != 0 else math.nan
