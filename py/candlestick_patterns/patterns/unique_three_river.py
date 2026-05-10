"""Unique Three River pattern (3-candle bullish)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body


def unique_three_river(self) -> int:
    """Unique Three River: a three-candle bullish pattern.

    Must have:
    - first candle: long black,
    - second candle: black harami (body within first body) with a lower
      low than the first candle,
    - third candle: small white, opens not lower than the second candle's
      low.

    The meaning of "long" is specified with self._long_body.
    The meaning of "short" is specified with self._short_body.

    Returns:
        +100 for bullish, 0 for no pattern.
    """
    if not self._enough(3, self._long_body, self._short_body):
        return 0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    if not (is_black(o1, c1) and is_black(o2, c2) and is_white(o3, c3)):
        return 0

    # First candle is long.
    if not (real_body(o1, c1) > self._avg(self._long_body, 3)):
        return 0

    # Second candle: harami (body within first body) with lower low.
    if not (c2 > c1 and o2 <= o1 and l2 < l1):
        return 0

    # Third candle: small, opens not lower than second candle's low.
    if not (real_body(o3, c3) < self._avg(self._short_body, 1) and
            o3 >= l2):
        return 0

    return 100
