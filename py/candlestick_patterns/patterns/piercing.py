"""Piercing pattern (2-candle bullish)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body


def piercing(self) -> int:
    """Piercing: a two-candle bullish reversal pattern.

    Must have:
    - first candle: long black,
    - second candle: long white that opens below the prior low and closes
      above the midpoint of the first candle's real body but within the body.

    The meaning of "long" is specified with self._long_body.

    Returns:
        +100 for bullish, 0 for no pattern.
    """
    if not self._enough(2, self._long_body):
        return 0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    if (is_black(o1, c1) and is_white(o2, c2) and
            real_body(o1, c1) > self._avg(self._long_body, 2) and
            real_body(o2, c2) > self._avg(self._long_body, 1) and
            o2 < l1 and
            c2 > c1 + real_body(o1, c1) * 0.5 and
            c2 < o1):
        return 100

    return 0
