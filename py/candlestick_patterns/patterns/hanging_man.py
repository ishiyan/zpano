"""Hanging Man pattern (2-candle bearish)."""
from __future__ import annotations

from ..core.primitives import real_body, lower_shadow, upper_shadow
from ...fuzzy import t_product_all


def hanging_man(self) -> float:
    """Hanging Man: a two-candle bearish pattern.

    Must have:
    - small real body,
    - long lower shadow,
    - no or very short upper shadow,
    - body is above or near the highs of the previous candle.

    Returns:
        Continuous float in [-100, 0].  More negative = stronger signal.
    """
    if not self._enough(2, self._short_body, self._long_shadow,
                        self._very_short_shadow, self._near):
        return 0.0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    near_avg = self._avg(self._near, 2)
    near_width = self._fuzz_ratio * near_avg if near_avg > 0.0 else 0.0

    mu_short = self._mu_less(real_body(o2, c2), self._short_body, 1)
    mu_long_ls = self._mu_greater(lower_shadow(o2, l2, c2), self._long_shadow, 1)
    mu_short_us = self._mu_less(upper_shadow(o2, h2, c2), self._very_short_shadow, 1)
    mu_near_high = self._mu_ge_raw(min(o2, c2), h1 - near_avg, near_width)

    confidence = t_product_all(mu_short, mu_long_ls, mu_short_us, mu_near_high)
    return -confidence * 100.0
