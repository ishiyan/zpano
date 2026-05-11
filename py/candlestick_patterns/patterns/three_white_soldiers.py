"""Three White Soldiers pattern (3-candle bullish)."""
from __future__ import annotations

from ..core.primitives import is_white, real_body, upper_shadow
from ...fuzzy import t_product_all


def three_white_soldiers(self) -> float:
    """Three White Soldiers: a three-candle bullish pattern.

    Must have:
    - three consecutive white candles with consecutively higher closes,
    - all three have very short upper shadows,
    - each opens within or near the prior candle's real body,
    - none is far shorter than the prior candle,
    - third candle is not short.

    Category A: always bullish (continuous).

    Returns:
        Continuous float in [0, 100].  Always bullish.
    """
    if not self._enough(
        3, self._short_body, self._very_short_shadow,
        self._near, self._far,
    ):
        return 0.0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    # Crisp gates: all white with consecutively higher closes.
    if not (is_white(o1, c1) and is_white(o2, c2) and is_white(o3, c3) and
            c3 > c2 > c1):
        return 0.0

    rb1 = real_body(o1, c1)
    rb2 = real_body(o2, c2)
    rb3 = real_body(o3, c3)

    # Crisp: each opens above the prior open (ordering).
    if not (o2 > o1 and o3 > o2):
        return 0.0

    # Fuzzy: very short upper shadows (all three).
    mu_us1 = self._mu_less(upper_shadow(o1, h1, c1), self._very_short_shadow, 3)
    mu_us2 = self._mu_less(upper_shadow(o2, h2, c2), self._very_short_shadow, 2)
    mu_us3 = self._mu_less(upper_shadow(o3, h3, c3), self._very_short_shadow, 1)

    # Fuzzy: each opens within or near the prior body (upper bound).
    near3 = self._avg(self._near, 3)
    near3_width = self._fuzz_ratio * near3 if near3 > 0.0 else 0.0
    mu_o2_near = self._mu_lt_raw(o2, c1 + near3, near3_width)

    near2 = self._avg(self._near, 2)
    near2_width = self._fuzz_ratio * near2 if near2 > 0.0 else 0.0
    mu_o3_near = self._mu_lt_raw(o3, c2 + near2, near2_width)

    # Fuzzy: not far shorter than prior candle.
    far3 = self._avg(self._far, 3)
    far3_width = self._fuzz_ratio * far3 if far3 > 0.0 else 0.0
    mu_notfar2 = self._mu_gt_raw(rb2, rb1 - far3, far3_width)

    far2 = self._avg(self._far, 2)
    far2_width = self._fuzz_ratio * far2 if far2 > 0.0 else 0.0
    mu_notfar3 = self._mu_gt_raw(rb3, rb2 - far2, far2_width)

    # Fuzzy: third candle is not short.
    mu_not_short3 = self._mu_greater(rb3, self._short_body, 1)

    confidence = t_product_all(mu_us1, mu_us2, mu_us3,
                               mu_o2_near, mu_o3_near,
                               mu_notfar2, mu_notfar3,
                               mu_not_short3)

    return confidence * 100.0
