"""Concealing Baby Swallow pattern (4-candle bullish)."""
from __future__ import annotations

from ..core.primitives import (
    is_black, real_body, upper_shadow, lower_shadow,
    is_real_body_gap_down,
)


def concealing_baby_swallow(self) -> int:
    """Concealing Baby Swallow: a four-candle bullish pattern.

    Must have:
    - first candle: black marubozu (very short shadows),
    - second candle: black marubozu (very short shadows),
    - third candle: black, opens gapping down, upper shadow extends into
      the prior real body (upper shadow > very-short avg),
    - fourth candle: black, completely engulfs the third candle including
      shadows (strict > / <).

    The meaning of "very short" for shadows is specified with
    self._very_short_shadow.

    Returns:
        +100 for bullish, 0 for no pattern.
    """
    if not self._enough(4, self._very_short_shadow):
        return 0

    o1, h1, l1, c1 = self._bar(4)
    o2, h2, l2, c2 = self._bar(3)
    o3, h3, l3, c3 = self._bar(2)
    o4, h4, l4, c4 = self._bar(1)

    if not (is_black(o1, c1) and is_black(o2, c2) and
            is_black(o3, c3) and is_black(o4, c4)):
        return 0

    vs4 = self._avg(self._very_short_shadow, 4)
    vs3 = self._avg(self._very_short_shadow, 3)
    vs2 = self._avg(self._very_short_shadow, 2)

    # First and second candles are marubozu (very short shadows).
    if not (lower_shadow(o1, l1, c1) < vs4 and
            upper_shadow(o1, h1, c1) < vs4 and
            lower_shadow(o2, l2, c2) < vs3 and
            upper_shadow(o2, h2, c2) < vs3):
        return 0

    # Third candle: opens gapping down, upper shadow > very-short avg,
    # upper shadow extends into prior body.
    if not (is_real_body_gap_down(o2, c2, o3, c3) and
            upper_shadow(o3, h3, c3) > vs2 and
            h3 > c2):
        return 0

    # Fourth candle: engulfs third including shadows (strict).
    if h4 > h3 and l4 < l3:
        return 100

    return 0
