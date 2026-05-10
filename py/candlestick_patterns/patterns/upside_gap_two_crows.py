"""Upside Gap Two Crows pattern (3-candle bearish)."""
from __future__ import annotations

from ..core.primitives import (
    is_white, is_black, real_body, white_real_body, black_real_body,
    is_real_body_gap_up,
)


def upside_gap_two_crows(self) -> int:
    """Upside Gap Two Crows: a three-candle bearish pattern.

    Must have:
    - first candle: long white,
    - second candle: small black that gaps up from the first,
    - third candle: black that engulfs the second candle's body and
      closes above the first candle's close (gap not filled).

    The meaning of "long" is specified with self._long_body.
    The meaning of "short" is specified with self._short_body.

    Returns:
        -100 for bearish, 0 for no pattern.
    """
    if not self._enough(3, self._long_body, self._short_body):
        return 0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    if not (is_white(o1, c1) and is_black(o2, c2) and is_black(o3, c3)):
        return 0

    # First candle is long white.
    if not (real_body(o1, c1) > self._avg(self._long_body, 3)):
        return 0

    # Second candle is small and gaps up.
    if not (real_body(o2, c2) <= self._avg(self._short_body, 2) and
            is_real_body_gap_up(o1, c1, o2, c2)):
        return 0

    # Third candle: engulfs second body, closes above first close.
    if (o3 > o2 and c3 < c2 and c3 > c1):
        return -100

    return 0
