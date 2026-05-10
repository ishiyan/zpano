"""Three Outside Up/Down pattern (3-candle reversal)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, is_real_body_encloses_real_body


def three_outside(self) -> int:
    """Three Outside Up/Down: a three-candle reversal pattern.

    Must have:
    - first and second candles form an engulfing pattern,
    - third candle confirms the direction by closing higher (up) or
      lower (down).

    Three Outside Up: first candle is black, second is white engulfing
    the first, third closes higher than the second.

    Three Outside Down: first candle is white, second is black engulfing
    the first, third closes lower than the second.

    No criteria needed (just requires at least 3 candles).

    Returns:
        +100 for bullish, -100 for bearish, 0 for no pattern.
    """
    if not self._enough(3):
        return 0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    # Three Outside Up: black + white engulfing + closes higher.
    if (is_black(o1, c1) and is_white(o2, c2) and
            is_real_body_encloses_real_body(o2, c2, o1, c1) and
            c3 > c2):
        return 100

    # Three Outside Down: white + black engulfing + closes lower.
    if (is_white(o1, c1) and is_black(o2, c2) and
            is_real_body_encloses_real_body(o2, c2, o1, c1) and
            c3 < c2):
        return -100

    return 0
