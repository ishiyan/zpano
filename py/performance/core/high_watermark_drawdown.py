import collections
import math

from ...streaming_kbn import KleinKBNAccumulator

class HighWaterMarkDrawdown:
    """
    Rolling high-water-mark drawdown.

    Drawdown at each observation is measured from the highest cumulative
    equity value reached up to that observation within the current rolling
    window.

    Drawdowns are expressed as percentage decimals and are non-positive:

        drawdown = (equity / high_water_mark - 1)

    Cumulative log-equity is maintained internally so that returns can be
    accumulated accurately. Running sums of drawdowns and squared drawdowns
    are maintained using compensated floating-point accumulation.

    When the observation leaving the window was a high-water mark, all
    drawdowns in the window are recomputed because the applicable peak may
    change. Otherwise, the update is O(1).

    A window size of zero means an expanding (unbounded) window.
    """
    def __init__(self, window_size: int) -> None:
        self._window_size = window_size if window_size and window_size > 0 else 0
        maxlen = self._window_size or None

        # Cumulative log-equity at each observation.
        self._cumlog: collections.deque[float] = collections.deque(maxlen=maxlen)

        # Drawdown at each observation, in percent (<= 0).
        self._dd: collections.deque[float] = collections.deque(maxlen=maxlen)

        # Cumulative log return.
        self._c: KleinKBNAccumulator = KleinKBNAccumulator()

        # Current high-water mark in log-equity space.
        self._peak: float = -math.inf

        # Running drawdown aggregates.
        self._sum_dd: KleinKBNAccumulator = KleinKBNAccumulator()
        self._sum_dd2: KleinKBNAccumulator = KleinKBNAccumulator()

    def reset(self) -> None:
        """Reset the accumulator to its initial empty state."""
        self._cumlog.clear()
        self._dd.clear()
        self._sum_dd.reset()
        self._sum_dd2.reset()
        self._c.reset()
        self._peak = -math.inf

    def _recompute(self) -> None:
        """
         Recompute all drawdowns from the cumulative log-equity values.

        This is required when the previous high-water mark leaves the
        rolling window.
        """
        self._dd.clear()
        self._sum_dd.reset()
        self._sum_dd2.reset()
        peak = -math.inf
        for c in self._cumlog:
            if c >= peak:
                peak = c
                dd = 0.0
            else:
                dd = math.expm1(c - peak)
            self._dd.append(dd)
            self._sum_dd.update(dd)
            self._sum_dd2.update(dd * dd)
        self._peak = peak

    def update(self, ret: float) -> bool:
        """
        Add a return observation.

        If the rolling window is full, the oldest observation is removed
        before the new observation is added.

        Args:
            ret:
                Period return expressed as a percentag decimal.
                For example, ``0.02`` represents a 2% return
                and ``-0.015`` represents a -1.5% return.

        Returns:
            True if the rolling window required a drawdown recomputation,
            otherwise False.
        """
        evicted_peak = False
        if self._window_size and len(self._cumlog) == self._window_size:
            old_c = self._cumlog.popleft()
            old_dd = self._dd.popleft()

            self._sum_dd.revert(old_dd)
            self._sum_dd2.revert(old_dd * old_dd)

            # A zero drawdown identifies an observation at the
            # high-water mark. If it leaves, the peak may change.
            evicted_peak = old_dd == 0.0

        # Global cumulative log-equity.
        self._c.update(math.log1p(ret))
        c = self._c.value
        self._cumlog.append(c)

        if evicted_peak:
            self._recompute()
            return True
        if c >= self._peak:
            self._peak = c
            dd = 0.0
        else:
            dd = math.expm1(c - self._peak)

        self._dd.append(dd)
        self._sum_dd.update(dd)
        self._sum_dd2.update(dd * dd)
        return False

    @property
    def drawdowns(self) -> list[float]:
        """Drawdowns for observations currently in the window."""
        # This was a defencive copy, disabled because we don't want
        # to expose this class outside.
        # return list(self._dd)
        return self._dd

    @property
    def drawdown(self) -> float:
        """Return the most recent drawdown in the current window."""
        return self._dd[-1] if self._dd else math.nan

    @property
    def maximum_drawdown(self) -> float:
        """
        Maximum drawdown in the current window.

        Drawdowns are non-positive, the largest loss
        is the minimum drawdown value.
        """
        return min(self._dd) if self._dd else math.nan

    @property
    def drawdowns_mean(self) -> float:
        """Arithmetic mean of drawdowns in the current window."""
        n = len(self._dd)
        return self._sum_dd.value / n if n else math.nan

    @property
    def drawdowns_squared_mean(self) -> float:
        """Mean squared drawdown in the current window."""
        n = len(self._dd)
        return self._sum_dd2.value / n if n else math.nan

    @property
    def drawdowns_count(self) -> int:
        """Number of observations in the current window."""
        return len(self._dd)
