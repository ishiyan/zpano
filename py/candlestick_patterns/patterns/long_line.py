"""Long Line pattern (1-candle)."""
from __future__ import annotations

from ..core.primitives import is_white, real_body, upper_shadow, lower_shadow
from ...fuzzy import t_product_all


def long_line(self) -> float:
    """Long Line: a one-candle pattern.

    Must have:
    - long real body,
    - short upper shadow,
    - short lower shadow.

    The meaning of "long" is specified with self._long_body.
    The meaning of "short" for shadows is specified with self._short_shadow.

    Category B: direction from candle color.

    Returns:
        Continuous float in [-100, +100].
    """
    if not self._enough(1, self._long_body, self._short_shadow):
        return 0.0

    o, h, l, c = self._bar(1)

    # Fuzzy: long body, short shadows.
    mu_long = self._mu_greater(real_body(o, c), self._long_body, 1)
    mu_us = self._mu_less(upper_shadow(o, h, c), self._short_shadow, 1)
    mu_ls = self._mu_less(lower_shadow(o, l, c), self._short_shadow, 1)

    confidence = t_product_all(mu_long, mu_us, mu_ls)

    # Crisp direction from color.
    direction = 1 if is_white(o, c) else -1

    return direction * confidence * 100.0
