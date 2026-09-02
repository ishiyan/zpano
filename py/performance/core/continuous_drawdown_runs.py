import math
import collections

from ...streaming_kbn import KleinKBNAccumulator

def dd_percent(logsum: float) -> float:
    """Convert a compounded log return into a percentage drawdown."""
    return math.expm1(logsum) * 100.0

class ContinuousDrawdownRuns:
    """
    Streaming 'continuous' drawdown runs for Burke-type measures..

    A continuous drawdown is the compounded loss over a maximal run of
    consecutive negative returns:

        DD = (prod(1 + r_i * 0.01) - 1) * 100

    The Burke denominator is:

        sqrt(sum(DD_j^2))

    where the sum is taken over all continuous losing runs in the current
    window.

    This class is a pure accumulator: the caller owns the rolling window
    and feeds evicted values to `revert()` and new values to `update()`.
    Within one step, call `revert(old)` BEFORE `update(new)` so run adjacency stays correct.

    The complexity is O(1) per call.
    """
    def __init__(self) -> None:
        # Eeach run: [logsum, count].
        self._runs: collections.deque[list] = collections.deque()

        # Sum of squared continuous drawdowns.
        self._sum_sq: KleinKBNAccumulator = KleinKBNAccumulator()

        # Whether the most recent return is negative.
        self._last_was_negative = False
        self._count: int = 0

    def reset(self) -> None:
        self._count = 0
        self._runs.clear()
        self._sum_sq.reset()
        self._last_was_negative = False

    def revert(self, old_ret: float) -> None:
        """Remove the oldest return from the left edge of the window."""
        if old_ret < 0:
            # oldest negative is the front of the left-most run, shrink it
            run = self._runs[0]
            self._sum_sq.revert(dd_percent(run[0]) ** 2)
            run[0] -= math.log1p(old_ret * 0.01)
            run[1] -= 1
            if run[1] == 0:
                self._runs.popleft() # run fully evicted
            else:
                self._sum_sq.update(dd_percent(run[0]) ** 2)
        # old_ret >= 0 is a separator, nothing to update

    def update(self, ret: float) -> None:
        """Add a new (most-recent) return at the right edge of the window."""
        if ret < 0:
            logr = math.log1p(ret * 0.01)
            if self._last_was_negative and self._runs:
                # Extend the currently-open (right-most) run.
                run = self._runs[-1]
                self._sum_sq.revert(dd_percent(run[0]) ** 2)
                run[0] += logr
                run[1] += 1
                self._sum_sq.update(dd_percent(run[0]) ** 2)
            else:
                # Start a new run.
                self._runs.append([logr, 1])
                self._sum_sq.update(dd_percent(logr) ** 2)
            self._last_was_negative = True
        else:
            # Non-negative return closes any open run (already counted) — a separator.
            self._last_was_negative = False

    @property
    def drawdowns(self) -> list[float]:
        """Continuous drawdowns (negative percentages), one value for each losing run."""
        return [dd_percent(logsum) for logsum, _ in self._runs]

    @property
    def sum_drawdowns_squared(self) -> float:
        """Sum of squared continuous drawdowns."""
        return max(self._sum_sq.value, 0.0)

    @property
    def sqrt_sum_drawdowns_squared(self) -> float:
        """
        Square root of the sum of squared continuous drawdowns.

        This is the denominator used by the Burke ratio.
        """
        return math.sqrt(max(self._sum_sq.value, 0.0))

    @property
    def run_count(self) -> int:
        """Number of continuous losing runs in the current window."""
        return len(self._runs)