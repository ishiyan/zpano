"""Up/Down-Gap Side-By-Side White Lines pattern (3-candle)."""
from __future__ import annotations

from ..core.primitives import (
    is_white, real_body,
    is_real_body_gap_up, is_real_body_gap_down,
)
from ...fuzzy import t_product_all


def up_down_gap_side_by_side_white_lines(self) -> float:
    """Up/Down-Gap Side-By-Side White Lines: a three-candle pattern.

    Must have:
    - first candle: white (for up gap) or black (for down gap),
    - gap (up or down) between the first and second candle — both 2nd AND
      3rd must gap from the 1st,
    - second and third candles are both white with similar size and
      approximately the same open.

    Up gap = bullish continuation, down gap = bearish continuation.

    Category C: both branches evaluated, return stronger signal.

    Returns:
        Continuous float in [-100, +100].
    """
    if not self._enough(3, self._near, self._equal):
        return 0.0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    # Crisp: both 2nd and 3rd must be white.
    if not (is_white(o2, c2) and is_white(o3, c3)):
        return 0.0

    # Both 2nd and 3rd must gap from 1st in the same direction — crisp.
    gap_up = is_real_body_gap_up(o1, c1, o2, c2) and is_real_body_gap_up(o1, c1, o3, c3)
    gap_down = is_real_body_gap_down(o1, c1, o2, c2) and is_real_body_gap_down(o1, c1, o3, c3)

    if not (gap_up or gap_down):
        return 0.0

    rb2 = real_body(o2, c2)
    rb3 = real_body(o3, c3)

    # Fuzzy: similar size and same open.
    mu_near_size = self._mu_less(abs(rb2 - rb3), self._near, 2)
    mu_equal_open = self._mu_less(abs(o3 - o2), self._equal, 2)

    conf = t_product_all(mu_near_size, mu_equal_open)

    if gap_up:
        return conf * 100.0
    return -conf * 100.0
