"""Evening Doji Star pattern (3-candle bearish reversal)."""
from __future__ import annotations

from ..core.primitives import (
    is_white, is_black, real_body, is_real_body_gap_up,
)
from ...fuzzy import t_product_all

EVENING_DOJI_STAR_PENETRATION_FACTOR: float = 0.3


def evening_doji_star(self) -> float:
    """Evening Doji Star: a three-candle bearish reversal pattern.

    Must have:
    - first candle: long white real body,
    - second candle: doji that gaps up (real body gap up from the first),
    - third candle: black real body that moves well within the first candle's
      real body.

    The meaning of "long" is specified with self._long_body.
    The meaning of "doji" is specified with self._doji_body.
    The meaning of "short" is specified with self._short_body.

    Category A: always bearish (continuous).

    Returns:
        Continuous float in [-100, 0].  Always bearish.
    """
    if not self._enough(3, self._long_body, self._doji_body, self._short_body):
        return 0.0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    penetration = EVENING_DOJI_STAR_PENETRATION_FACTOR

    # Crisp gates: color checks and gap.
    if not (is_white(o1, c1) and
            is_real_body_gap_up(o1, c1, o2, c2) and
            is_black(o3, c3)):
        return 0.0

    # Fuzzy conditions.
    mu_long1 = self._mu_greater(real_body(o1, c1), self._long_body, 3)
    mu_doji2 = self._mu_less(real_body(o2, c2), self._doji_body, 2)

    # c3 < c1 - rb1 * penetration
    rb1 = real_body(o1, c1)
    threshold = c1 - rb1 * penetration
    width = self._fuzz_ratio * rb1 * penetration
    mu_penetration = self._mu_lt_raw(c3, threshold, width)

    confidence = t_product_all(mu_long1, mu_doji2, mu_penetration)

    return -1.0 * confidence * 100.0
