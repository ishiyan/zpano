"""Engulfing pattern (2-candle reversal)."""
from __future__ import annotations


def engulfing(self) -> int:
    """Engulfing: a two-candle reversal pattern.

    Must have:
    - first candle and second candle have opposite colors,
    - second candle's real body engulfs the first (at least one end strictly
      exceeds, the other can match).

    Returns 100 when both ends differ, 80 when one end matches.
    Direction from 2nd candle color.

    Returns:
        +/-100 or +/-80 for pattern detected, 0 for no pattern.
    """
    if not self._enough(2):
        return 0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    color1 = 1 if c1 >= o1 else -1
    color2 = 1 if c2 >= o2 else -1

    # Must be opposite colors.
    if color1 == color2:
        return 0

    matched = False
    if color2 == 1:
        # Bullish: white engulfs black.
        # 2nd close >= 1st open AND 2nd open < 1st close, OR
        # 2nd close > 1st open AND 2nd open <= 1st close.
        matched = ((c2 >= o1 and o2 < c1) or (c2 > o1 and o2 <= c1))
    else:
        # Bearish: black engulfs white.
        # 2nd open >= 1st close AND 2nd close < 1st open, OR
        # 2nd open > 1st close AND 2nd close <= 1st open.
        matched = ((o2 >= c1 and c2 < o1) or (o2 > c1 and c2 <= o1))

    if not matched:
        return 0

    # 100 when both endpoints differ, 80 when one matches.
    if o2 != c1 and c2 != o1:
        return color2 * 100
    return color2 * 80
