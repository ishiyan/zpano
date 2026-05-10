"""Three Black Crows pattern (4-candle bearish)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, lower_shadow


def three_black_crows(self) -> int:
    """Three Black Crows: a four-candle bearish reversal pattern.

    Must have:
    - preceding candle (oldest) is white,
    - three consecutive black candles with declining closes,
    - each opens within the prior black candle's real body,
    - each has a very short lower shadow,
    - 1st black closes under the prior white candle's high.

    Returns:
        -100 for bearish, 0 for no pattern.
    """
    if not self._enough(4, self._very_short_shadow):
        return 0

    # bar(4)=white prior, bar(3)=1st black, bar(2)=2nd black, bar(1)=3rd black
    o0, h0, l0, c0 = self._bar(4)  # prior white
    o1, h1, l1, c1 = self._bar(3)  # 1st black
    o2, h2, l2, c2 = self._bar(2)  # 2nd black
    o3, h3, l3, c3 = self._bar(1)  # 3rd black

    if not is_white(o0, c0):
        return 0

    if not (is_black(o1, c1) and is_black(o2, c2) and is_black(o3, c3)):
        return 0

    # Very short lower shadows.
    if not (lower_shadow(o1, l1, c1) < self._avg(self._very_short_shadow, 3) and
            lower_shadow(o2, l2, c2) < self._avg(self._very_short_shadow, 2) and
            lower_shadow(o3, l3, c3) < self._avg(self._very_short_shadow, 1)):
        return 0

    # 2nd opens within 1st's real body (for black: open > close)
    # 3rd opens within 2nd's real body
    if not (o2 < o1 and o2 > c1 and
            o3 < o2 and o3 > c2):
        return 0

    # Prior white's high > 1st black's close
    if not (h0 > c1):
        return 0

    # Three declining closes
    if not (c1 > c2 and c2 > c3):
        return 0

    return -100
