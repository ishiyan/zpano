"""Rickshaw Man pattern (1-candle)."""
from __future__ import annotations

from ..core.primitives import real_body, upper_shadow, lower_shadow
from ...fuzzy import t_product_all


def rickshaw_man(self) -> float:
    """Rickshaw Man: a one-candle doji pattern.

    Must have:
    - doji body (very small real body),
    - two long shadows,
    - body near the midpoint of the high-low range.

    Returns:
        Continuous float in [0, 100].  Higher = stronger signal.
    """
    if not self._enough(1, self._doji_body, self._long_shadow, self._near):
        return 0.0

    o, h, l, c = self._bar(1)

    hl_range = h - l
    near_avg = self._avg(self._near, 1)
    near_width = self._fuzz_ratio * near_avg if near_avg > 0.0 else 0.0

    mu_doji = self._mu_less(real_body(o, c), self._doji_body, 1)
    mu_long_us = self._mu_greater(upper_shadow(o, h, c), self._long_shadow, 1)
    mu_long_ls = self._mu_greater(lower_shadow(o, l, c), self._long_shadow, 1)
    midpoint = l + hl_range / 2.0
    mu_near_mid_lo = self._mu_lt_raw(min(o, c), midpoint + near_avg, near_width)
    mu_near_mid_hi = self._mu_ge_raw(max(o, c), midpoint - near_avg, near_width)

    confidence = t_product_all(mu_doji, mu_long_us, mu_long_ls,
                               mu_near_mid_lo, mu_near_mid_hi)
    return confidence * 100.0
