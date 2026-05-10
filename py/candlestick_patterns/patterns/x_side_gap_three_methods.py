"""Up/Down-side Gap Three Methods pattern (3-candle continuation)."""
from __future__ import annotations

from ..core.primitives import (
    is_white, is_black,
    is_real_body_gap_up, is_real_body_gap_down,
)


def x_side_gap_three_methods(self) -> int:
    """Up/Down-side Gap Three Methods: a three-candle continuation pattern.

    Must have:
    - first and second candles are the same color with a gap between them,
    - third candle is opposite color, opens within the second candle's
      real body and closes within the first candle's real body (fills the
      gap).

    Upside gap: two white candles with gap up, third is black = +100.
    Downside gap: two black candles with gap down, third is white = -100.

    No criteria needed (just requires at least 3 candles).

    Returns:
        +100 for bullish, -100 for bearish, 0 for no pattern.
    """
    if not self._enough(3):
        return 0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    # Upside gap: two whites gap up, third black fills.
    if (is_white(o1, c1) and is_white(o2, c2) and is_black(o3, c3) and
            is_real_body_gap_up(o1, c1, o2, c2) and
            o3 < c2 and o3 > o2 and
            c3 > o1 and c3 < c1):
        return 100

    # Downside gap: two blacks gap down, third white fills.
    if (is_black(o1, c1) and is_black(o2, c2) and is_white(o3, c3) and
            is_real_body_gap_down(o1, c1, o2, c2) and
            o3 > c2 and o3 < o2 and
            c3 < o1 and c3 > c1):
        return -100

    return 0
