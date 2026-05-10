"""Up/Down-Gap Side-By-Side White Lines pattern (3-candle)."""
from __future__ import annotations

from ..core.primitives import (
    is_white, real_body,
    is_real_body_gap_up, is_real_body_gap_down,
)


def up_down_gap_side_by_side_white_lines(self) -> int:
    """Up/Down-Gap Side-By-Side White Lines: a three-candle pattern.

    Must have:
    - first candle: white (for up gap) or black (for down gap),
    - gap (up or down) between the first and second candle — both 2nd AND
      3rd must gap from the 1st,
    - second and third candles are both white with similar size and
      approximately the same open.

    Up gap = +100 (bullish continuation), down gap = -100 (bearish
    continuation).

    The meaning of "near" is specified with self._near.
    The meaning of "equal" is specified with self._equal.

    Returns:
        +100 for bullish, -100 for bearish, 0 for no pattern.
    """
    if not self._enough(3, self._near, self._equal):
        return 0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    if not (is_white(o2, c2) and is_white(o3, c3)):
        return 0

    # Both 2nd and 3rd must gap from 1st in the same direction.
    gap_up = is_real_body_gap_up(o1, c1, o2, c2) and is_real_body_gap_up(o1, c1, o3, c3)
    gap_down = is_real_body_gap_down(o1, c1, o2, c2) and is_real_body_gap_down(o1, c1, o3, c3)

    if not (gap_up or gap_down):
        return 0

    rb2 = real_body(o2, c2)
    rb3 = real_body(o3, c3)
    near_avg = self._avg(self._near, 2)   # at the 2nd candle
    eq_avg = self._avg(self._equal, 2)     # at the 2nd candle

    # Similar size (two-sided band) and same open.
    if not (rb3 >= rb2 - near_avg and rb3 <= rb2 + near_avg and
            o3 >= o2 - eq_avg and o3 <= o2 + eq_avg):
        return 0

    if gap_up:
        return 100
    return -100
