"""Matching Low pattern (2-candle bullish)."""
from __future__ import annotations

from ..core.primitives import is_black


def matching_low(self) -> int:
    """Matching Low: a two-candle bullish pattern.

    Must have:
    - first candle: black,
    - second candle: black with close equal to the first candle's close.

    The meaning of "equal" is specified with self._equal.

    Returns:
        +100 for bullish, 0 for no pattern.
    """
    if not self._enough(2, self._equal):
        return 0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    eq = self._avg(self._equal, 2)

    if (is_black(o1, c1) and is_black(o2, c2) and
            c2 <= c1 + eq and
            c2 >= c1 - eq):
        return 100

    return 0
