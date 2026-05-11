"""Kicking By Length pattern (2-candle)."""
from __future__ import annotations

from ..core.primitives import (
    is_white, is_black, real_body, upper_shadow, lower_shadow,
    is_high_low_gap_up, is_high_low_gap_down,
)
from ...fuzzy import t_product_all


def kicking_by_length(self) -> float:
    """Kicking By Length: like Kicking but direction determined by the longer marubozu.

    Must have:
    - first candle: marubozu (long body, very short shadows),
    - second candle: opposite-color marubozu with a high-low gap,
    - bull/bear determined by which marubozu has the longer real body.

    Category B: direction from longer marubozu's color.

    Returns:
        Continuous float in [-100, +100].
    """
    if not self._enough(2, self._very_short_shadow, self._long_body):
        return 0.0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    color1 = 1 if c1 >= o1 else -1
    color2 = 1 if c2 >= o2 else -1

    # Crisp: opposite colors.
    if color1 == color2:
        return 0.0

    # Crisp: gap check.
    has_gap = False
    if color1 == -1 and is_high_low_gap_up(h1, l2):
        has_gap = True
    elif color1 == 1 and is_high_low_gap_down(l1, h2):
        has_gap = True

    if not has_gap:
        return 0.0

    rb1 = real_body(o1, c1)
    rb2 = real_body(o2, c2)

    # Fuzzy: both are marubozu (long body, very short shadows).
    mu_long1 = self._mu_greater(rb1, self._long_body, 2)
    mu_vs_us1 = self._mu_less(upper_shadow(o1, h1, c1), self._very_short_shadow, 2)
    mu_vs_ls1 = self._mu_less(lower_shadow(o1, l1, c1), self._very_short_shadow, 2)

    mu_long2 = self._mu_greater(rb2, self._long_body, 1)
    mu_vs_us2 = self._mu_less(upper_shadow(o2, h2, c2), self._very_short_shadow, 1)
    mu_vs_ls2 = self._mu_less(lower_shadow(o2, l2, c2), self._very_short_shadow, 1)

    confidence = t_product_all(mu_long1, mu_vs_us1, mu_vs_ls1,
                               mu_long2, mu_vs_us2, mu_vs_ls2)

    # Direction determined by the longer marubozu's color.
    direction = color2 if rb2 > rb1 else color1

    return direction * confidence * 100.0
