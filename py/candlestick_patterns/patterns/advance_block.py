"""Advance Block pattern (3-candle bearish)."""
from __future__ import annotations

from ..core.primitives import is_white, real_body, upper_shadow
from ...fuzzy import t_product_all


def advance_block(self) -> float:
    """Advance Block: a bearish three-candle pattern.

    Three white candles with consecutively higher closes and opens, but
    showing signs of weakening (diminishing bodies, growing upper shadows).

    Category A: always bearish (continuous).

    Returns:
        Continuous float in [-100, 0].  Always bearish.
    """
    if not self._enough(
        3, self._long_body, self._short_shadow, self._long_shadow,
        self._near, self._far,
    ):
        return 0.0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    # Crisp gates: all white with rising closes.
    if not (is_white(o1, c1) and is_white(o2, c2) and is_white(o3, c3) and
            c3 > c2 > c1):
        return 0.0

    # Crisp: 2nd opens above 1st open.
    if not (o2 > o1):
        return 0.0

    # Crisp: 3rd opens above 2nd open.
    if not (o3 > o2):
        return 0.0

    rb1 = real_body(o1, c1)
    rb2 = real_body(o2, c2)
    rb3 = real_body(o3, c3)

    # Fuzzy: 2nd opens within/near 1st body (upper bound).
    near3 = self._avg(self._near, 3)
    near3_width = self._fuzz_ratio * near3 if near3 > 0.0 else 0.0
    mu_o2_near = self._mu_lt_raw(o2, c1 + near3, near3_width)

    # Fuzzy: 3rd opens within/near 2nd body (upper bound).
    near2 = self._avg(self._near, 2)
    near2_width = self._fuzz_ratio * near2 if near2 > 0.0 else 0.0
    mu_o3_near = self._mu_lt_raw(o3, c2 + near2, near2_width)

    # Fuzzy: first candle long body.
    mu_long1 = self._mu_greater(rb1, self._long_body, 3)

    # Fuzzy: first candle short upper shadow.
    mu_us1 = self._mu_less(upper_shadow(o1, h1, c1), self._short_shadow, 3)

    # At least one weakness condition must hold (OR → max).
    far2 = self._avg(self._far, 3)
    far2_width = self._fuzz_ratio * far2 if far2 > 0.0 else 0.0
    far1 = self._avg(self._far, 2)
    far1_width = self._fuzz_ratio * far1 if far1 > 0.0 else 0.0
    near1 = self._avg(self._near, 2)
    near1_width = self._fuzz_ratio * near1 if near1 > 0.0 else 0.0

    # Branch 1: 2 far smaller than 1 AND 3 not longer than 2
    mu_b1_a = self._mu_lt_raw(rb2, rb1 - far2, far2_width)
    mu_b1_b = self._mu_lt_raw(rb3, rb2 + near1, near1_width)
    branch1 = t_product_all(mu_b1_a, mu_b1_b)

    # Branch 2: 3 far smaller than 2
    branch2 = self._mu_lt_raw(rb3, rb2 - far1, far1_width)

    # Branch 3: 3 < 2 AND 2 < 1 AND (3 or 2 has non-short upper shadow)
    rb3_width = self._fuzz_ratio * rb2 if rb2 > 0.0 else 0.0
    rb2_width = self._fuzz_ratio * rb1 if rb1 > 0.0 else 0.0
    mu_b3_a = self._mu_lt_raw(rb3, rb2, rb3_width)
    mu_b3_b = self._mu_lt_raw(rb2, rb1, rb2_width)
    mu_b3_us3 = self._mu_greater(upper_shadow(o3, h3, c3), self._short_shadow, 1)
    mu_b3_us2 = self._mu_greater(upper_shadow(o2, h2, c2), self._short_shadow, 2)
    branch3 = t_product_all(mu_b3_a, mu_b3_b, max(mu_b3_us3, mu_b3_us2))

    # Branch 4: 3 < 2 AND 3 has long upper shadow
    mu_b4_a = self._mu_lt_raw(rb3, rb2, rb3_width)
    mu_b4_b = self._mu_greater(upper_shadow(o3, h3, c3), self._long_shadow, 1)
    branch4 = t_product_all(mu_b4_a, mu_b4_b)

    weakness = max(branch1, branch2, branch3, branch4)

    confidence = t_product_all(mu_o2_near, mu_o3_near, mu_long1, mu_us1,
                               weakness)

    return -confidence * 100.0
