"""Breakaway pattern (5-candle reversal)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body, is_real_body_gap_down, is_real_body_gap_up


def breakaway(self) -> int:
    """Breakaway: a five-candle reversal pattern.

    Bullish: first candle is long black, second candle is black gapping down,
    third and fourth candles have consecutively lower highs and lows, fifth
    candle is white closing into the gap (between first and second candle's
    real bodies).

    Bearish: mirror image with colors reversed and gaps reversed.

    The meaning of "long" is specified with self._long_body.

    Returns:
        +100 for bullish, -100 for bearish, 0 for no pattern.
    """
    if not self._enough(5, self._long_body):
        return 0

    o1, h1, l1, c1 = self._bar(5)
    o2, h2, l2, c2 = self._bar(4)
    o3, h3, l3, c3 = self._bar(3)
    o4, h4, l4, c4 = self._bar(2)
    o5, h5, l5, c5 = self._bar(1)

    # Bullish breakaway.
    if (is_black(o1, c1) and is_black(o2, c2) and
            is_black(o4, c4) and is_white(o5, c5) and
            h3 < h2 and l3 < l2 and
            h4 < h3 and l4 < l3 and
            c5 > o2 and c5 < c1 and
            is_real_body_gap_down(o1, c1, o2, c2) and
            real_body(o1, c1) > self._avg(self._long_body, 5)):
        return 100

    # Bearish breakaway.
    if (is_white(o1, c1) and is_white(o2, c2) and
            is_white(o4, c4) and is_black(o5, c5) and
            h3 > h2 and l3 > l2 and
            h4 > h3 and l4 > l3 and
            c5 < o2 and c5 > c1 and
            is_real_body_gap_up(o1, c1, o2, c2) and
            real_body(o1, c1) > self._avg(self._long_body, 5)):
        return -100

    return 0


def breakaway_bullish(self) -> int:
    """Convenience: returns only the bullish breakaway signal."""
    result = breakaway(self)
    return result if result > 0 else 0


def breakaway_bearish(self) -> int:
    """Convenience: returns only the bearish breakaway signal."""
    result = breakaway(self)
    return result if result < 0 else 0
