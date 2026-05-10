"""Three-Line Strike pattern (4-candle)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black


def three_line_strike(self) -> int:
    """Three-Line Strike: a four-candle pattern.

    Bullish: three white candles with rising closes, each opening within/near
    the prior body, 4th black opens above 3rd close and closes below 1st open.

    Bearish: three black candles with falling closes, each opening within/near
    the prior body, 4th white opens below 3rd close and closes above 1st open.

    Returns:
        +100 for bullish, -100 for bearish, 0 for no pattern.
    """
    if not self._enough(4, self._near):
        return 0

    o1, h1, l1, c1 = self._bar(4)
    o2, h2, l2, c2 = self._bar(3)
    o3, h3, l3, c3 = self._bar(2)
    o4, h4, l4, c4 = self._bar(1)

    # Three same color
    color1 = 1 if is_white(o1, c1) else -1
    color2 = 1 if is_white(o2, c2) else -1
    color3 = 1 if is_white(o3, c3) else -1
    color4 = 1 if is_white(o4, c4) else -1

    if not (color1 == color2 and color2 == color3 and color4 == -color3):
        return 0

    # 2nd opens within/near 1st real body
    near4 = self._avg(self._near, 4)
    near3 = self._avg(self._near, 3)
    if not (o2 >= min(o1, c1) - near4 and
            o2 <= max(o1, c1) + near4):
        return 0

    # 3rd opens within/near 2nd real body
    if not (o3 >= min(o2, c2) - near3 and
            o3 <= max(o2, c2) + near3):
        return 0

    # Bullish: three white, rising closes, 4th opens above 3rd close, closes below 1st open
    if (color3 == 1 and
            c3 > c2 and c2 > c1 and
            o4 > c3 and
            c4 < o1):
        return 100

    # Bearish: three black, falling closes, 4th opens below 3rd close, closes above 1st open
    if (color3 == -1 and
            c3 < c2 and c2 < c1 and
            o4 < c3 and
            c4 > o1):
        return -100

    return 0
