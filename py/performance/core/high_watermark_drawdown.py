import collections
import math

from ...streaming_kbn import KleinKBNAccumulator

class HighWaterMarkDrawdown:
    """
    Rolling high-water-mark drawdown.

    Drawdown at each observation is measured from the high-water mark,
    the highest equity value reached up to that observation within the
    current rolling window, including the equity at the start of the window:

        drawdown_t = equity_t / max(equity_start, equity_1, ..., equity_t) - 1

    This matches R PerformanceAnalytics ``Drawdowns()``, which uses
    ``cummax(c(1, cumprod(1 + R)))``: a first negative return already
    produces a drawdown.  For a rolling window, ``equity_start`` is the
    equity just before the first observation in the window, so the result
    equals a fresh calculation over the window's returns.

    Drawdowns are expressed as decimals and are non-positive.

    Cumulative log-equity is maintained internally so that returns can be
    accumulated accurately. Running sums of drawdowns and squared drawdowns
    are maintained using compensated floating-point accumulation.

    When an observation leaves the window, the window's starting equity
    changes.  The drawdowns in the window are recomputed only when this
    changes their high-water marks; otherwise, the update is O(1).

    A window size of zero means an expanding (unbounded) window.
    """
    def __init__(self, window_size: int) -> None:
        self._window_size = window_size if window_size and window_size > 0 else 0
        maxlen = self._window_size or None

        # Cumulative log-equity at each observation.
        self._cumlog: collections.deque[float] = collections.deque(maxlen=maxlen)

        # Drawdown at each observation, as a decimal (<= 0).
        self._dd: collections.deque[float] = collections.deque(maxlen=maxlen)

        # Cumulative log return.
        self._c: KleinKBNAccumulator = KleinKBNAccumulator()

        # Log-equity just before the first observation in the window.
        self._base: float = 0.0

        # Current high-water mark in log-equity space.
        self._peak: float = 0.0

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
        self._base = 0.0
        self._peak = 0.0

    def _recompute(self) -> None:
        """
        Recompute all drawdowns from the cumulative log-equity values.

        This is required when an observation leaving the rolling window
        changes the high-water marks of the remaining observations.
        """
        self._dd.clear()
        self._sum_dd.reset()
        self._sum_dd2.reset()
        peak = self._base
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
                Period return expressed as a decimal.
                For example, ``0.02`` represents a 2% return
                and ``-0.015`` represents a -1.5% return.

        Returns:
            True if the rolling window required a drawdown recomputation,
            otherwise False.
        """
        old_base = None
        if self._window_size and len(self._cumlog) == self._window_size:
            old_c = self._cumlog.popleft()
            old_dd = self._dd.popleft()

            self._sum_dd.revert(old_dd)
            self._sum_dd2.revert(old_dd * old_dd)

            # The evicted observation's equity is the new starting equity.
            # High-water marks of the remaining observations can only
            # change if the old starting equity was above the evicted one.
            if old_c < self._base:
                old_base = self._base
            self._base = old_c

        # Global cumulative log-equity.
        self._c.update(math.log1p(ret))
        c = self._c.value
        self._cumlog.append(c)

        # Peaks of all remaining observations were max(old_base, c0, ..., cj);
        # without old_base they are max(c0, c1, ..., cj).  They differ only
        # if the new first observation is also below old_base.
        if old_base is not None and self._cumlog[0] < old_base:
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
    def drawdowns(self) -> collections.deque[float]:
        """
        Drawdowns for observations currently in the window.

        Returns the internal deque without a defensive copy, because
        this class is private to the package; callers must not modify it.
        """
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
