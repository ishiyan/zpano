"""Shooting Star pattern (2-candle bearish)."""
from __future__ import annotations

from ..core.primitives import real_body, upper_shadow, lower_shadow, is_real_body_gap_up
from ...fuzzy import t_product_all


def shooting_star(self) -> float:
    """Shooting Star: a two-candle bearish reversal pattern.

    Must have:
    - gap up from the previous candle (real body gap up),
    - small real body,
    - long upper shadow,
    - very short lower shadow.

    Returns:
        Continuous float in [-100, 0].  More negative = stronger signal.
    """
    if not self._enough(2, self._short_body, self._long_shadow,
                        self._very_short_shadow):
        return 0.0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    if not is_real_body_gap_up(o1, c1, o2, c2):
        return 0.0

    mu_short = self._mu_less(real_body(o2, c2), self._short_body, 1)
    mu_long_us = self._mu_greater(upper_shadow(o2, h2, c2), self._long_shadow, 1)
    mu_short_ls = self._mu_less(lower_shadow(o2, l2, c2), self._very_short_shadow, 1)

    confidence = t_product_all(mu_short, mu_long_us, mu_short_ls)
    return -confidence * 100.0
