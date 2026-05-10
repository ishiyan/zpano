"""Stick Sandwich pattern (3-candle bullish)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black


def stick_sandwich(self) -> int:
    """Stick Sandwich: a three-candle bullish pattern.

    Must have:
    - first candle: black,
    - second candle: white, trades above the first candle's close
      (low > first close),
    - third candle: black, close equals the first candle's close.

    The meaning of "equal" is specified with self._equal.

    Returns:
        +100 for bullish, 0 for no pattern.
    """
    if not self._enough(3, self._equal):
        return 0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    eq = self._avg(self._equal, 3)

    if (is_black(o1, c1) and is_white(o2, c2) and is_black(o3, c3) and
            l2 > c1 and
            c3 <= c1 + eq and
            c3 >= c1 - eq):
        return 100

    return 0
