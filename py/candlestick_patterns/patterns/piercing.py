"""Piercing pattern (2-candle bullish)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body
from ...fuzzy import t_product_all


def piercing(self) -> float:
    """Piercing: a two-candle bullish reversal pattern.

    Must have:
    - first candle: long black,
    - second candle: long white that opens below the prior low and closes
      above the midpoint of the first candle's real body but within the body.

    Returns:
        Continuous float in [0, 100].  Higher = stronger signal.
    """
    if not self._enough(2, self._long_body):
        return 0.0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    # Color checks stay crisp
    if not is_black(o1, c1) or not is_white(o2, c2):
        return 0.0

    rb1 = real_body(o1, c1)
    eq_avg = self._avg(self._equal, 1)
    eq_width = self._fuzz_ratio * eq_avg if eq_avg > 0.0 else 0.0

    mu_long1 = self._mu_greater(rb1, self._long_body, 2)
    mu_long2 = self._mu_greater(real_body(o2, c2), self._long_body, 1)
    mu_open_below = self._mu_lt_raw(o2, l1, eq_width)
    pen_threshold = c1 + rb1 * 0.5
    pen_width = self._fuzz_ratio * rb1 * 0.5 if rb1 > 0.0 else 0.0
    mu_pen = self._mu_gt_raw(c2, pen_threshold, pen_width)
    mu_below_open1 = self._mu_lt_raw(c2, o1, eq_width)

    confidence = t_product_all(mu_long1, mu_long2, mu_open_below,
                               mu_pen, mu_below_open1)
    return confidence * 100.0
