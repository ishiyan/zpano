"""Gravestone Doji pattern (1-candle)."""
from __future__ import annotations

from ..core.primitives import real_body, upper_shadow, lower_shadow
from ...fuzzy import t_product_all


def gravestone_doji(self) -> float:
    """Gravestone Doji: a one-candle pattern.

    Must have:
    - doji body (very small real body relative to high-low range),
    - no or very short lower shadow,
    - upper shadow is not very short.

    Returns:
        Continuous float in [0, 100].  Higher = stronger signal.
    """
    if not self._enough(1, self._doji_body, self._very_short_shadow):
        return 0.0

    o1, h1, l1, c1 = self._bar(1)

    mu_doji = self._mu_less(real_body(o1, c1), self._doji_body, 1)
    mu_short_ls = self._mu_less(lower_shadow(o1, l1, c1), self._very_short_shadow, 1)
    mu_long_us = self._mu_greater(upper_shadow(o1, h1, c1), self._very_short_shadow, 1)

    confidence = t_product_all(mu_doji, mu_short_ls, mu_long_us)
    return confidence * 100.0
