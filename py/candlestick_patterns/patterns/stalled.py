"""Stalled pattern (3-candle bearish)."""
from __future__ import annotations

from ..core.primitives import is_white, real_body, upper_shadow
from ...fuzzy import t_product_all


def stalled(self) -> float:
    """Stalled (Deliberation): a three-candle bearish pattern.

    Three white candles with progressively higher closes:
    - first candle: long white body,
    - second candle: long white body, opens within or near the first
      candle's body, very short upper shadow,
    - third candle: small body that rides on the shoulder of the second
      (opens near the second's close, accounting for its own body size).

    Category A: always bearish (continuous).

    Returns:
        Continuous float in [-100, 0].  Always bearish.
    """
    if not self._enough(3, self._long_body, self._short_body,
                        self._very_short_shadow, self._near):
        return 0.0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    # Crisp gates: all white, rising closes.
    if not (is_white(o1, c1) and is_white(o2, c2) and is_white(o3, c3)):
        return 0.0
    if not (c3 > c2 and c2 > c1):
        return 0.0
    # Crisp: o2 > o1 (opens above prior open).
    if not (o2 > o1):
        return 0.0

    rb3 = real_body(o3, c3)

    # Fuzzy conditions.
    mu_long1 = self._mu_greater(real_body(o1, c1), self._long_body, 3)
    mu_long2 = self._mu_greater(real_body(o2, c2), self._long_body, 2)
    mu_us2 = self._mu_less(upper_shadow(o2, h2, c2), self._very_short_shadow, 2)

    # o2 <= c1 + near_avg (opens within or near prior body).
    near3 = self._avg(self._near, 3)
    near3_width = self._fuzz_ratio * near3 if near3 > 0.0 else 0.0
    mu_o2_near = self._mu_lt_raw(o2, c1 + near3, near3_width)

    # Third candle: short body.
    mu_short3 = self._mu_less(rb3, self._short_body, 1)

    # o3 >= c2 - rb3 - near_avg (rides on shoulder).
    near2 = self._avg(self._near, 2)
    near2_width = self._fuzz_ratio * near2 if near2 > 0.0 else 0.0
    mu_o3_shoulder = self._mu_ge_raw(o3, c2 - rb3 - near2, near2_width)

    confidence = t_product_all(mu_long1, mu_long2, mu_us2,
                               mu_o2_near, mu_short3, mu_o3_shoulder)

    return -1.0 * confidence * 100.0
