"""Three Inside Up/Down pattern (3-candle reversal)."""
from __future__ import annotations

from ..core.primitives import (
    is_white, is_black, real_body, white_real_body, black_real_body,
    is_real_body_encloses_real_body,
)


def three_inside(self) -> int:
    """Three Inside Up/Down: a three-candle reversal pattern.

    Three Inside Up (bullish):
    - first candle: long black,
    - second candle: short, engulfed by the first candle's real body,
    - third candle: white, closes above the first candle's open.

    Three Inside Down (bearish):
    - first candle: long white,
    - second candle: short, engulfed by the first candle's real body,
    - third candle: black, closes below the first candle's open.

    The meaning of "long" is specified with self._long_body.
    The meaning of "short" is specified with self._short_body.

    Returns:
        +100 for bullish, -100 for bearish, 0 for no pattern.
    """
    if not self._enough(3, self._long_body, self._short_body):
        return 0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    rb1 = real_body(o1, c1)
    rb2 = real_body(o2, c2)

    # Three Inside Up: long black, short engulfed, white closes above 1st open.
    if (is_black(o1, c1) and
            rb1 > self._avg(self._long_body, 3) and
            rb2 < self._avg(self._short_body, 2) and
            is_real_body_encloses_real_body(o1, c1, o2, c2) and
            is_white(o3, c3) and c3 > o1):
        return 100

    # Three Inside Down: long white, short engulfed, black closes below 1st open.
    if (is_white(o1, c1) and
            rb1 > self._avg(self._long_body, 3) and
            rb2 < self._avg(self._short_body, 2) and
            is_real_body_encloses_real_body(o1, c1, o2, c2) and
            is_black(o3, c3) and c3 < o1):
        return -100

    return 0
