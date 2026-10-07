import math

from .klein_kbn_accumulator import KleinKBNAccumulator


class KleinKBNSummator:
    """
    Counted compensated sum: a KleinKBNAccumulator plus a sample count,
    which also gives the arithmetic mean.

    See KleinKBNAccumulator for the summation algorithm.

    Every call to update() increments the count, including calls with
    x == 0; every call to revert() decrements it.  Because only sums are
    stored, revert() may remove any previously added value (not only
    the most recent one), so the summator works for FIFO rolling windows.
    """

    def __init__(self) -> None:
        self._n = 0
        self._sum: KleinKBNAccumulator = KleinKBNAccumulator()

    def reset(self) -> None:
        """Clears the count and the sum."""
        self._n = 0
        self._sum.reset()

    def revert(self, x: float) -> None:
        """
        Removes a previously added value x.

        Removing the final sample clears the sum and its compensation terms.

        Raises ValueError if the summator is empty.
        """
        if self._n <= 0:
            raise ValueError("Cannot revert from an empty summator")
        if self._n == 1:
            self.reset()
            return
        self._n -= 1
        # Adding zero leaves the accumulator unchanged, so skip it.
        if x != 0:
            self._sum.revert(x)

    def update(self, x: float) -> None:
        """Adds a value x."""
        self._n += 1
        # Adding zero leaves the accumulator unchanged, so skip it.
        if x != 0:
            self._sum.update(x)

    @property
    def value(self) -> float:
        """The compensated sum of all added values (0.0 when empty)."""
        return self._sum.value

    @property
    def mean(self) -> float:
        """The arithmetic mean, sum / n (NaN when empty)."""
        n = self._n
        if n <= 0:
            return math.nan
        return self._sum.value / n

    @property
    def n(self) -> int:
        """The number of added values."""
        return self._n
