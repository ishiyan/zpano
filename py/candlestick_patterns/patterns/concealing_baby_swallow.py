"""Concealing Baby Swallow pattern (4-candle bullish)."""
from __future__ import annotations

from ..core.primitives import (
    is_black, real_body, upper_shadow, lower_shadow,
    is_real_body_gap_down,
)
from ...fuzzy import t_product_all


def concealing_baby_swallow(self) -> float:
    """Concealing Baby Swallow: a four-candle bullish pattern.

    Must have:
    - first candle: black marubozu (very short shadows),
    - second candle: black marubozu (very short shadows),
    - third candle: black, opens gapping down, upper shadow extends into
      the prior real body (upper shadow > very-short avg),
    - fourth candle: black, completely engulfs the third candle including
      shadows (strict > / <).

    The meaning of "very short" for shadows is specified with
    self._very_short_shadow.

    Category A: always bullish (continuous).

    Returns:
        Continuous float in [0, 100].  Always bullish.
    """
    if not self._enough(4, self._very_short_shadow):
        return 0.0

    o1, h1, l1, c1 = self._bar(4)
    o2, h2, l2, c2 = self._bar(3)
    o3, h3, l3, c3 = self._bar(2)
    o4, h4, l4, c4 = self._bar(1)

    # Crisp gates: all black.
    if not (is_black(o1, c1) and is_black(o2, c2) and
            is_black(o3, c3) and is_black(o4, c4)):
        return 0.0

    # Crisp: gap down and upper shadow extends into prior body.
    if not (is_real_body_gap_down(o2, c2, o3, c3) and h3 > c2):
        return 0.0

    # Crisp: fourth engulfs third including shadows (strict).
    if not (h4 > h3 and l4 < l3):
        return 0.0

    # Fuzzy: first and second are marubozu (very short shadows).
    mu_ls1 = self._mu_less(lower_shadow(o1, l1, c1), self._very_short_shadow, 4)
    mu_us1 = self._mu_less(upper_shadow(o1, h1, c1), self._very_short_shadow, 4)
    mu_ls2 = self._mu_less(lower_shadow(o2, l2, c2), self._very_short_shadow, 3)
    mu_us2 = self._mu_less(upper_shadow(o2, h2, c2), self._very_short_shadow, 3)

    # Fuzzy: third candle upper shadow > very-short avg.
    mu_us3_long = self._mu_greater(upper_shadow(o3, h3, c3), self._very_short_shadow, 2)

    confidence = t_product_all(mu_ls1, mu_us1, mu_ls2, mu_us2, mu_us3_long)

    return confidence * 100.0
