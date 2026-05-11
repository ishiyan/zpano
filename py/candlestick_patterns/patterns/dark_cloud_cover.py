"""Dark Cloud Cover pattern (2-candle bearish)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body
from ...fuzzy import t_product_all

DARK_CLOUD_COVER_PENETRATION_FACTOR: float = 0.5


def dark_cloud_cover(self) -> float:
    """Dark Cloud Cover: a two-candle bearish reversal pattern.

    Must have:
    - first candle: long white candle,
    - second candle: black candle that opens above the prior high and
      closes well within the first candle's real body (below the midpoint).

    Returns:
        Continuous float in [-100, 0].  More negative = stronger signal.
    """
    if not self._enough(2, self._long_body):
        return 0.0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    # Color checks stay crisp
    if not is_white(o1, c1) or not is_black(o2, c2):
        return 0.0

    penetration = DARK_CLOUD_COVER_PENETRATION_FACTOR
    rb1 = real_body(o1, c1)
    eq_avg = self._avg(self._equal, 1)
    eq_width = self._fuzz_ratio * eq_avg if eq_avg > 0.0 else 0.0

    mu_long = self._mu_greater(rb1, self._long_body, 2)
    mu_open_above = self._mu_gt_raw(o2, h1, eq_width)
    pen_threshold = c1 - rb1 * penetration
    pen_width = self._fuzz_ratio * rb1 * penetration if rb1 * penetration > 0.0 else 0.0
    mu_pen = self._mu_lt_raw(c2, pen_threshold, pen_width)
    mu_above_open1 = self._mu_gt_raw(c2, o1, eq_width)

    confidence = t_product_all(mu_long, mu_open_above, mu_pen, mu_above_open1)
    return -confidence * 100.0
