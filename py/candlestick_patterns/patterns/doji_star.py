"""Doji Star pattern (2-candle reversal)."""
from __future__ import annotations

from ..core.primitives import (
    is_white, is_black, real_body, is_real_body_gap_up, is_real_body_gap_down,
)
from ...fuzzy import t_product_all


def doji_star(self) -> float:
    """Doji Star: a two-candle reversal pattern.

    Must have:
    - first candle: long real body,
    - second candle: doji that gaps away from the first candle.

    - bearish: first candle is long white, doji gaps up,
    - bullish: first candle is long black, doji gaps down.

    The meaning of "long" is specified with self._long_body.
    The meaning of "doji" is specified with self._doji_body.

    Category B: direction from 1st candle color (continuous).

    Returns:
        Continuous float in [-100, +100].
    """
    if not self._enough(2, self._long_body, self._doji_body):
        return 0.0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    color1 = 1 if c1 >= o1 else -1

    # Crisp gates: gap direction must match color.
    if color1 == 1 and not is_real_body_gap_up(o1, c1, o2, c2):
        return 0.0
    if color1 == -1 and not is_real_body_gap_down(o1, c1, o2, c2):
        return 0.0

    # Fuzzy conditions.
    mu_long1 = self._mu_greater(real_body(o1, c1), self._long_body, 2)
    mu_doji2 = self._mu_less(real_body(o2, c2), self._doji_body, 1)

    confidence = t_product_all(mu_long1, mu_doji2)

    # Direction: opposite of 1st candle color.
    direction = -1.0 if color1 == 1 else 1.0

    return direction * confidence * 100.0
