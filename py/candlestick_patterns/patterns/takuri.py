"""Takuri pattern (1-candle)."""
from __future__ import annotations

from ..core.primitives import real_body, upper_shadow, lower_shadow
from ...fuzzy import t_product_all


def takuri(self) -> float:
    """Takuri (Dragonfly Doji with very long lower shadow): a one-candle pattern.

    A doji body with a very short upper shadow and a very long lower shadow.

    Returns:
        Continuous float in [0, 100].  Higher = stronger signal.
    """
    if not self._enough(1, self._doji_body, self._very_short_shadow,
                        self._very_long_shadow):
        return 0.0

    o, h, l, c = self._bar(1)

    mu_doji = self._mu_less(real_body(o, c), self._doji_body, 1)
    mu_short_us = self._mu_less(upper_shadow(o, h, c), self._very_short_shadow, 1)
    mu_long_ls = self._mu_greater(lower_shadow(o, l, c), self._very_long_shadow, 1)

    confidence = t_product_all(mu_doji, mu_short_us, mu_long_ls)
    return confidence * 100.0
