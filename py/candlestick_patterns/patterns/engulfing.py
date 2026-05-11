"""Engulfing pattern (2-candle reversal)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black
from ...fuzzy import t_product_all


def engulfing(self) -> float:
    """Engulfing: a two-candle reversal pattern.

    Must have:
    - first candle and second candle have opposite colors,
    - second candle's real body engulfs the first (at least one end strictly
      exceeds, the other can match).

    Category B: direction from 2nd candle color (continuous).
    Opposite-color check stays crisp (doji edge case).

    Returns:
        Continuous float in [-100, +100].  Sign from 2nd candle direction.
    """
    if not self._enough(2):
        return 0.0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    # Opposite colors — crisp gate (TA-Lib convention: c >= o is white).
    color1 = 1 if c1 >= o1 else -1
    color2 = 1 if c2 >= o2 else -1
    if color1 == color2:
        return 0.0

    # Fuzzy engulfment: 2nd body upper >= 1st body upper AND
    #                    2nd body lower <= 1st body lower.
    upper1 = max(o1, c1)
    lower1 = min(o1, c1)
    upper2 = max(o2, c2)
    lower2 = min(o2, c2)

    # Width based on the equal criterion for tight comparisons.
    eq_avg = self._avg(self._equal, 1)
    eq_width = self._fuzz_ratio * eq_avg if eq_avg > 0.0 else 0.0

    mu_upper = self._mu_ge_raw(upper2, upper1, eq_width)
    mu_lower = self._mu_lt_raw(lower2, lower1, eq_width)

    confidence = t_product_all(mu_upper, mu_lower)

    # Direction sign from 2nd candle (TA-Lib: c >= o is bullish).
    direction = 1.0 if c2 >= o2 else -1.0

    return direction * confidence * 100.0
