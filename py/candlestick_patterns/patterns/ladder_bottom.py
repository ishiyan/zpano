"""Ladder Bottom pattern (5-candle bullish)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body, upper_shadow


def ladder_bottom(self) -> int:
    """Ladder Bottom: a five-candle bullish pattern.

    Must have:
    - first three candles: descending black candles (each closes lower),
    - fourth candle: black with a long upper shadow,
    - fifth candle: white, opens above the fourth candle's real body,
      closes above the fourth candle's high.

    The meaning of "long" for shadows is specified with self._long_shadow.

    Returns:
        +100 for bullish, 0 for no pattern.
    """
    if not self._enough(5, self._very_short_shadow):
        return 0

    o1, h1, l1, c1 = self._bar(5)
    o2, h2, l2, c2 = self._bar(4)
    o3, h3, l3, c3 = self._bar(3)
    o4, h4, l4, c4 = self._bar(2)
    o5, h5, l5, c5 = self._bar(1)

    if not (is_black(o1, c1) and is_black(o2, c2) and
            is_black(o3, c3) and is_black(o4, c4) and
            is_white(o5, c5)):
        return 0

    # Three descending black candles with consecutively lower opens and closes.
    if not (o1 > o2 and o2 > o3 and c1 > c2 and c2 > c3):
        return 0

    # Fourth candle: black with an upper shadow (> very short average).
    if upper_shadow(o4, h4, c4) <= self._avg(self._very_short_shadow, 2):
        return 0

    # Fifth candle: white, opens above prior candle's open, closes above prior candle's high.
    if o5 > o4 and c5 > h4:
        return 100

    return 0
