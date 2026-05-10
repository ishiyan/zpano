"""Doji Star pattern (2-candle reversal)."""
from __future__ import annotations

from ..core.primitives import (
    is_white, is_black, real_body, is_real_body_gap_up, is_real_body_gap_down,
)


def doji_star(self) -> int:
    """Doji Star: a two-candle reversal pattern.

    Must have:
    - first candle: long real body,
    - second candle: doji that gaps away from the first candle.

    - bearish: first candle is long white, doji gaps up,
    - bullish: first candle is long black, doji gaps down.

    The meaning of "long" is specified with self._long_body.
    The meaning of "doji" is specified with self._doji_body.

    Returns:
        +100 for bullish, -100 for bearish, 0 for no pattern.
    """
    if not self._enough(2, self._long_body, self._doji_body):
        return 0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    color1 = 1 if c1 >= o1 else -1

    if (real_body(o1, c1) > self._avg(self._long_body, 2) and
            real_body(o2, c2) <= self._avg(self._doji_body, 1) and
            ((color1 == 1 and is_real_body_gap_up(o1, c1, o2, c2)) or
             (color1 == -1 and is_real_body_gap_down(o1, c1, o2, c2)))):
        return -color1 * 100

    return 0
