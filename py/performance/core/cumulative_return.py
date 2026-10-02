import math

from ...streaming_kbn import KleinKBNAccumulator

class CumulativeReturn:
    """
    Streaming cumulative (geometric) returns.

    Accumulates the sum of log returns, log(1 + r), with compensated
    summation.  Because only a sum is stored, revert() may remove any
    previously added return, so the class works for FIFO rolling windows:
    the caller owns the window and feeds evicted returns to revert().
    """
    def __init__(self) -> None:
        self._cumlogret_sum: KleinKBNAccumulator = KleinKBNAccumulator()
        self._count: int = 0

    def reset(self) -> None:
        self._cumlogret_sum.reset()
        self._count = 0

    def revert(self, ret: float) -> None:
        """
        Removes a previously added return.

        Raises ValueError if there are no returns.
        """
        if self._count <= 0:
            raise ValueError("Cannot revert from an empty accumulator")
        self._count -= 1
        self._cumlogret_sum.revert(math.log1p(ret) if ret != 0 else 0)

    def update(self, ret: float) -> None:
        """Adds a return, expressed as a decimal (must be > -1)."""
        self._count += 1
        self._cumlogret_sum.update(math.log1p(ret) if ret != 0 else 0)

    @property
    def count(self) -> int:
        return self._count

    @property
    def cumulative_geometric_return(self) -> float:
        """
        Cumulative geometric return, prod(1 + r) - 1 (0.0 when empty).
        """
        return math.expm1(self._cumlogret_sum.value)

    @property
    def geometric_mean_return(self) -> float:
        """
        The geometric mean of the returns, prod(1 + r)^(1/n) - 1
        (NaN when empty).
        """
        return math.expm1(self._cumlogret_sum.value / self._count) if self._count > 0 else math.nan

    def annualized_geometric_mean_return(self, periods_per_year: float) -> float:
        """
        The annualized geometric mean, prod(1 + r)^(periods_per_year/n) - 1
        (NaN when empty).
        """
        if self._count == 0:
            return math.nan
        return math.expm1(self._cumlogret_sum.value * periods_per_year / self._count)
