"""Short Line pattern (1-candle)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body, upper_shadow, lower_shadow
from ...fuzzy import t_product_all


def short_line(self) -> float:
    """Short Line: a one-candle pattern.

    A candle with a short body, short upper shadow, and short lower shadow.

    The meaning of "short" for body is specified with self._short_body.
    The meaning of "short" for shadows is specified with self._short_shadow.

    Category C: color determines sign.

    Returns:
        Continuous float in [-100, +100].
    """
    if not self._enough(1, self._short_body, self._short_shadow):
        return 0.0

    o, h, l, c = self._bar(1)

    mu_short_body = self._mu_less(real_body(o, c), self._short_body, 1)
    mu_short_us = self._mu_less(upper_shadow(o, h, c), self._short_shadow, 1)
    mu_short_ls = self._mu_less(lower_shadow(o, l, c), self._short_shadow, 1)

    confidence = t_product_all(mu_short_body, mu_short_us, mu_short_ls)

    if is_white(o, c):
        return confidence * 100.0
    if is_black(o, c):
        return -confidence * 100.0
    return 0.0
