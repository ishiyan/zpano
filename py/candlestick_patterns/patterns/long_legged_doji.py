"""Long Legged Doji pattern (1-candle)."""
from __future__ import annotations

from ..core.primitives import real_body, upper_shadow, lower_shadow
from ...fuzzy import t_product_all, s_max


def long_legged_doji(self) -> float:
    """Long Legged Doji: a one-candle pattern.

    Must have:
    - doji body (very small real body),
    - one or both shadows are long.

    Returns:
        Continuous float in [0, 100].  Higher = stronger signal.
    """
    if not self._enough(1, self._doji_body, self._long_shadow):
        return 0.0

    o1, h1, l1, c1 = self._bar(1)

    mu_doji = self._mu_less(real_body(o1, c1), self._doji_body, 1)
    mu_long_us = self._mu_greater(upper_shadow(o1, h1, c1), self._long_shadow, 1)
    mu_long_ls = self._mu_greater(lower_shadow(o1, l1, c1), self._long_shadow, 1)
    mu_any_long = s_max(mu_long_us, mu_long_ls)

    confidence = t_product_all(mu_doji, mu_any_long)
    return confidence * 100.0
