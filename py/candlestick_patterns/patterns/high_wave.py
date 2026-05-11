"""High Wave pattern (1-candle)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body, upper_shadow, lower_shadow
from ...fuzzy import t_product_all


def high_wave(self) -> float:
    """High Wave: a one-candle pattern.

    Must have:
    - short real body,
    - very long upper shadow,
    - very long lower shadow.

    The meaning of "short" is specified with self._short_body.
    The meaning of "very long" (shadow) is specified with self._very_long_shadow.

    Category C: color determines sign.

    Returns:
        Continuous float in [-100, +100].
    """
    if not self._enough(1, self._short_body, self._very_long_shadow):
        return 0.0

    o1, h1, l1, c1 = self._bar(1)

    mu_short = self._mu_less(real_body(o1, c1), self._short_body, 1)
    mu_long_us = self._mu_greater(upper_shadow(o1, h1, c1), self._very_long_shadow, 1)
    mu_long_ls = self._mu_greater(lower_shadow(o1, l1, c1), self._very_long_shadow, 1)

    confidence = t_product_all(mu_short, mu_long_us, mu_long_ls)

    if is_white(o1, c1):
        return confidence * 100.0
    return -confidence * 100.0
