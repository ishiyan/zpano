"""In Neck pattern (2-candle bearish continuation)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body
from ...fuzzy import t_product_all


def in_neck(self) -> float:
    """In Neck: a two-candle bearish continuation pattern.

    Must have:
    - first candle: long black,
    - second candle: white, opens below the prior low, closes slightly
      into the prior real body (close near the prior close).

    The meaning of "long" is specified with self._long_body.
    The meaning of "near" is specified with self._near.

    Category A: always bearish (continuous).

    Returns:
        Continuous float in [-100, 0].  Always bearish.
    """
    if not self._enough(2, self._long_body, self._near):
        return 0.0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    # Crisp gates: color checks and open below prior low.
    if not (is_black(o1, c1) and is_white(o2, c2) and o2 < l1):
        return 0.0

    # Fuzzy conditions.
    mu_long1 = self._mu_greater(real_body(o1, c1), self._long_body, 2)

    # Close near prior close: crisp was abs(c2-c1) < near_avg.
    # Model as mu_less(abs_diff, near_avg) — crossover at near boundary.
    mu_near_close = self._mu_less(abs(c2 - c1), self._near, 1)

    confidence = t_product_all(mu_long1, mu_near_close)

    return -1.0 * confidence * 100.0
