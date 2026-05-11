"""Two Crows pattern (3-candle bearish)."""
from __future__ import annotations

from ..core.primitives import (
    is_white, is_black, real_body, white_real_body,
    is_real_body_gap_up,
)
from ...fuzzy import t_product_all


def two_crows(self) -> float:
    """Two Crows: a three-candle bearish pattern.

    Must have:
    - first candle: long white,
    - second candle: black, gaps up (real body gap up from the first),
    - third candle: black, opens within the second candle's real body,
      closes within the first candle's real body.

    The meaning of "long" is specified with self._long_body.

    Category A: always bearish (continuous).

    Returns:
        Continuous float in [-100, 0].  Always bearish.
    """
    if not self._enough(3, self._long_body):
        return 0.0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    # Crisp gates: color checks.
    if not (is_white(o1, c1) and is_black(o2, c2) and is_black(o3, c3)):
        return 0.0

    # Crisp: gap up.
    if not is_real_body_gap_up(o1, c1, o2, c2):
        return 0.0

    # Crisp: third opens within second body (o3 < o2 and o3 > c2).
    if not (o3 < o2 and o3 > c2):
        return 0.0

    # Crisp: third closes within first body (c3 > o1 and c3 < c1).
    if not (c3 > o1 and c3 < c1):
        return 0.0

    # Fuzzy: first candle is long.
    mu_long1 = self._mu_greater(real_body(o1, c1), self._long_body, 3)

    confidence = mu_long1

    return -confidence * 100.0
