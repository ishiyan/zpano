import math

from .klein_kbn_accumulator import KleinKBNAccumulator
from .raw_moments_klein_kbn import RawMomentsKleinKBN


class LinearRegressionKleinKBN:
    r"""
    Streaming ordinary least squares (OLS) regression y = a + b·x with
    Klein KBN (Kahan-Babuška-Neumaier) compensated accumulation.

    Tracks the means and variances of x and y (as RawMomentsKleinKBN with
    ddof=0) and the co-moment S_xy = Σ(x − x̄)(y − ȳ), updated in O(1) per
    sample (n = count before adding the sample):

        S_xy += (x − x̄)·(y − ȳ)·n / (n + 1)

    where x̄ and ȳ are the means before adding the sample.

    Like RawMomentsKleinKBN, revert() may remove any previously added
    (x, y) pair, not only the most recent one, so the class works for FIFO
    rolling windows.

    Derived quantities, with S_xx = Σ(x − x̄)² and S_yy = Σ(y − ȳ)²:

        slope       b = S_xy / S_xx
        intercept   a = ȳ − b·x̄
        correlation r = S_xy / √(S_xx·S_yy)
        covariance      S_xy / n   (population)
    """
    def __init__(self) -> None:
        self._n = 0
        self._x_moments: RawMomentsKleinKBN = RawMomentsKleinKBN(ddof=0)
        self._y_moments: RawMomentsKleinKBN = RawMomentsKleinKBN(ddof=0)
        self._s_xy: KleinKBNAccumulator = KleinKBNAccumulator()

    def reset(self) -> None:
        """Clears all accumulated state."""
        self._n = 0
        self._x_moments.reset()
        self._y_moments.reset()
        self._s_xy.reset()

    def update(self, x: float, y: float) -> None:
        """Adds a sample (x, y)."""
        n_old = self._n
        self._n += 1
        term = (self._x_moments.mean - x) * (self._y_moments.mean - y) * n_old / (n_old + 1)
        self._s_xy.update(term)
        self._x_moments.update(x)
        self._y_moments.update(y)

    def revert(self, x: float, y: float) -> None:
        """
        Removes a previously added sample (x, y), not necessarily the most
        recent one.

        Raises ValueError if there are no samples.
        """
        if self._n == 0:
            raise ValueError("Cannot revert from an empty regression")
        if self._n == 1:
            self.reset()
            return
        self._x_moments.revert(x)
        self._y_moments.revert(y)
        # The means are now those without (x, y), as in update().
        n = self._n - 1
        term = (self._x_moments.mean - x) * (self._y_moments.mean - y) * n / (n + 1)
        self._s_xy.revert(term)
        self._n = n

    @property
    def n(self) -> int:
        """The number of samples."""
        return self._n

    @property
    def mean_x(self) -> float:
        """The mean of x (0.0 when empty)."""
        return self._x_moments.mean

    @property
    def mean_y(self) -> float:
        """The mean of y (0.0 when empty)."""
        return self._y_moments.mean

    @property
    def variance_x(self) -> float:
        """The population variance of x, S_xx / n (NaN when empty)."""
        return self._x_moments.variance

    @property
    def variance_y(self) -> float:
        """The population variance of y, S_yy / n (NaN when empty)."""
        return self._y_moments.variance

    @property
    def co_moment(self) -> float:
        """The co-moment S_xy = Σ(x − x̄)(y − ȳ) (0.0 when empty)."""
        return self._s_xy.value

    @property
    def covariance(self) -> float:
        """The population covariance S_xy / n (NaN when empty)."""
        n = self._n
        if n < 1:
            return math.nan
        return self._s_xy.value / n

    @property
    def slope(self) -> float:
        """
        The OLS slope b = S_xy / S_xx.

        NaN when n < 2 or all x are equal (S_xx = 0).
        """
        n = self._n
        if n < 2:
            return math.nan
        s_xx = self._x_moments.variance * n
        return self._s_xy.value / s_xx if s_xx != 0 else math.nan

    @property
    def intercept(self) -> float:
        """The OLS intercept a = ȳ − b·x̄ (NaN when the slope is NaN)."""
        return self._y_moments.mean - self.slope * self._x_moments.mean

    @property
    def correlation(self) -> float:
        """
        The Pearson correlation coefficient r = S_xy / √(S_xx·S_yy),
        clamped to [−1, 1] to absorb rounding.

        NaN when n < 2 or either x or y is constant.
        """
        n = self._n
        if n < 2:
            return math.nan
        t = self._x_moments.standard_deviation * self._y_moments.standard_deviation
        if t == 0:
            return math.nan
        r = self._s_xy.value / (t * n)
        return max(-1.0, min(1.0, r))
