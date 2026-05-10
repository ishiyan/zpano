"""Two Crows pattern (3-candle bearish)."""
from __future__ import annotations

from ..core.primitives import (
    is_white, is_black, real_body, white_real_body,
    is_real_body_gap_up,
)


def two_crows(self) -> int:
    """Two Crows: a three-candle bearish pattern.

    Must have:
    - first candle: long white,
    - second candle: black, gaps up (real body gap up from the first),
    - third candle: black, opens within the second candle's real body,
      closes within the first candle's real body.

    The meaning of "long" is specified with self._long_body.

    Returns:
        -100 for bearish, 0 for no pattern.
    """
    if not self._enough(3, self._long_body):
        return 0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    if not (is_white(o1, c1) and is_black(o2, c2) and is_black(o3, c3)):
        return 0

    if not (real_body(o1, c1) > self._avg(self._long_body, 3)):
        return 0

    # Second candle gaps up from first.
    if not is_real_body_gap_up(o1, c1, o2, c2):
        return 0

    # Third candle opens within second body and closes within first body.
    wb1 = white_real_body(o1, c1)  # (o1, c1) -> open, close of white
    if (o3 < o2 and o3 > c2 and
            c3 > o1 and c3 < c1):
        return -100

    return 0
