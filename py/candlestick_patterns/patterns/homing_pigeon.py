"""Homing Pigeon pattern (2-candle bullish)."""
from __future__ import annotations

from ..core.primitives import is_black, real_body


def homing_pigeon(self) -> int:
    """Homing Pigeon: a two-candle bullish pattern.

    Must have:
    - first candle: long black,
    - second candle: short black, real body engulfed by first candle's
      real body.

    The meaning of "long" is specified with self._long_body.
    The meaning of "short" is specified with self._short_body.

    Returns:
        +100 for bullish, 0 for no pattern.
    """
    if not self._enough(2, self._long_body, self._short_body):
        return 0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    if (is_black(o1, c1) and is_black(o2, c2) and
            real_body(o1, c1) > self._avg(self._long_body, 2) and
            real_body(o2, c2) <= self._avg(self._short_body, 1) and
            o2 < o1 and c2 > c1):
        return 100

    return 0
