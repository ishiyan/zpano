"""Upside Gap Two Crows pattern (3-candle bearish)."""
from __future__ import annotations

from ..core.primitives import (
    is_white, is_black, real_body, white_real_body, black_real_body,
    is_real_body_gap_up,
)
from ...fuzzy import t_product_all


def upside_gap_two_crows(self) -> float:
    """Upside Gap Two Crows: a three-candle bearish pattern.

    Must have:
    - first candle: long white,
    - second candle: small black that gaps up from the first,
    - third candle: black that engulfs the second candle's body and
      closes above the first candle's close (gap not filled).

    The meaning of "long" is specified with self._long_body.
    The meaning of "short" is specified with self._short_body.

    Category A: always bearish (continuous).

    Returns:
        Continuous float in [-100, 0].  Always bearish.
    """
    if not self._enough(3, self._long_body, self._short_body):
        return 0.0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    # Crisp gates: colors.
    if not (is_white(o1, c1) and is_black(o2, c2) and is_black(o3, c3)):
        return 0.0

    # Crisp: gap up from first to second.
    if not is_real_body_gap_up(o1, c1, o2, c2):
        return 0.0

    # Crisp: third engulfs second (o3 > o2 and c3 < c2) and closes above c1.
    if not (o3 > o2 and c3 < c2 and c3 > c1):
        return 0.0

    # Fuzzy: first candle is long.
    mu_long1 = self._mu_greater(real_body(o1, c1), self._long_body, 3)

    # Fuzzy: second candle is short.
    mu_short2 = self._mu_less(real_body(o2, c2), self._short_body, 2)

    confidence = t_product_all(mu_long1, mu_short2)

    return -confidence * 100.0
