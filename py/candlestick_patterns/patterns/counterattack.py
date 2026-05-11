"""Counterattack pattern (2-candle reversal)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body
from ...fuzzy import t_product_all


def counterattack(self) -> float:
    """Counterattack: a two-candle reversal pattern.

    Two long candles of opposite color with closes that are equal
    (or very near equal).

    - bullish: first candle is long black, second is long white,
      closes are equal,
    - bearish: first candle is long white, second is long black,
      closes are equal.

    The meaning of "long" is specified with self._long_body.
    The meaning of "equal" is specified with self._equal.

    Category B: direction from 2nd candle color (continuous).

    Returns:
        Continuous float in [-100, +100].
    """
    if not self._enough(2, self._long_body, self._equal):
        return 0.0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    # Opposite colors — crisp gate.
    color1 = 1 if c1 >= o1 else -1
    color2 = 1 if c2 >= o2 else -1
    if color1 == color2:
        return 0.0

    # Fuzzy conditions.
    mu_long1 = self._mu_greater(real_body(o1, c1), self._long_body, 2)
    mu_long2 = self._mu_greater(real_body(o2, c2), self._long_body, 1)

    # Closes near equal: crisp was abs(c2-c1) <= eq.
    # Model as mu_less(abs_diff, eq_avg) — crossover at eq boundary.
    mu_eq = self._mu_less(abs(c2 - c1), self._equal, 2)

    confidence = t_product_all(mu_long1, mu_long2, mu_eq)

    # Direction from 2nd candle color.
    direction = 1.0 if c2 >= o2 else -1.0

    return direction * confidence * 100.0
