"""Kicking pattern (2-candle)."""
from __future__ import annotations

from ..core.primitives import (
    is_white, is_black, real_body, upper_shadow, lower_shadow,
    is_high_low_gap_up, is_high_low_gap_down,
)


def kicking(self) -> int:
    """Kicking: a two-candle pattern with opposite-color marubozus and gap.

    Must have:
    - first candle: marubozu (long body, very short shadows),
    - second candle: opposite-color marubozu with a high-low gap,
    - bullish: black marubozu followed by white marubozu gapping up,
    - bearish: white marubozu followed by black marubozu gapping down.

    Returns:
        +100 for bullish, -100 for bearish, 0 for no pattern.
    """
    if not self._enough(2, self._very_short_shadow, self._long_body):
        return 0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    color1 = 1 if c1 >= o1 else -1
    color2 = 1 if c2 >= o2 else -1

    if color1 == color2:
        return 0

    vs2 = self._avg(self._very_short_shadow, 2)
    vs1 = self._avg(self._very_short_shadow, 1)
    bl2 = self._avg(self._long_body, 2)
    bl1 = self._avg(self._long_body, 1)

    is_marubozu1 = (real_body(o1, c1) > bl2 and
                    upper_shadow(o1, h1, c1) < vs2 and
                    lower_shadow(o1, l1, c1) < vs2)
    is_marubozu2 = (real_body(o2, c2) > bl1 and
                    upper_shadow(o2, h2, c2) < vs1 and
                    lower_shadow(o2, l2, c2) < vs1)

    if not (is_marubozu1 and is_marubozu2):
        return 0

    # Gap check uses high-low gap (TA_CANDLEGAPUP / TA_CANDLEGAPDOWN).
    if color1 == -1 and is_high_low_gap_up(h1, l2):
        return color2 * 100
    if color1 == 1 and is_high_low_gap_down(l1, h2):
        return color2 * 100

    return 0
