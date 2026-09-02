import math
import warnings

from ...streaming_kbn import KleinKBNAccumulator
from .min_max import MinMax

class CumulativeReturn:
    """
    Streaming cumulative returns.
    """
    def __init__(self, window_size: int) -> None:
        """
        The valid sliding `window_size` value is >= 1.
        We use zero value to indicate unbounded running window with infinite window size.
        """
        if window_size < 0:
            raise ValueError("window_size must be non-negative")

        self._cumlogret_sum: KleinKBNAccumulator = KleinKBNAccumulator()
        self._cumlogret_extrema: MinMax = MinMax(window_size=window_size)
        self._count: int = 0

    def reset(self) -> None:
        self._cumlogret_sum.reset()
        self._cumlogret_extrema.reset()
        self._count = 0

    def revert(self, ret: float) -> None:
        self._count -= 1
        self._cumlogret_sum.revert(math.log1p(ret) if ret != 0 else 0)
        # We don't revert the minmax, because it has monotonic queues
        # that are updated with every sliding-window step.

    def update(self, ret: float) -> None:
        self._count += 1
        self._cumlogret_sum.update(math.log1p(ret) if ret != 0 else 0)
        self._cumlogret_extrema.update(self._cumlogret_sum.value)

    @property
    def count(self) -> int:
        return self._count

    @property
    def cumulative_geometric_return(self) -> float:
        """
        Cumulative geometric return
        """
        return math.expm1(self._cumlogret_sum.value)

    @property
    def geometric_return_plus_1(self) -> float:
        """
        Cumulative geometric return + 1
        """
        return math.exp(self._cumlogret_sum.value)

    @property
    def geometric_mean_return(self) -> float:
        """
        The geometric mean of the returns.
        """
        return math.expm1(self._cumlogret_sum.value / self._count) if self._count > 0 else math.nan

    def annualized_geometric_mean_return(self, periods_per_year: float) -> float:
        if self._count == 0:
            return math.nan
        return math.expm1(self._cumlogret_sum.value * periods_per_year / self._count)

    @property
    def geometric_return_plus_1_min(self) -> float:
        return math.exp(self._cumlogret_extrema.min)

    @property
    def geometric_return_plus_1_max(self) -> float:
        return math.exp(self._cumlogret_extrema.max)
