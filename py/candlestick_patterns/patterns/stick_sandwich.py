"""Stick Sandwich pattern (3-candle bullish)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black
from ...fuzzy import t_product_all


def stick_sandwich(self) -> float:
    """Stick Sandwich: a three-candle bullish pattern.

    Must have:
    - first candle: black,
    - second candle: white, trades above the first candle's close
      (low > first close),
    - third candle: black, close equals the first candle's close.

    The meaning of "equal" is specified with self._equal.

    Category A: always bullish (continuous).

    Returns:
        Continuous float in [0, 100].  Always bullish.
    """
    if not self._enough(3, self._equal):
        return 0.0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    # Crisp gates: colors and gap.
    if not (is_black(o1, c1) and is_white(o2, c2) and is_black(o3, c3)
            and l2 > c1):
        return 0.0

    # Fuzzy: third close equals first close (two-sided band).
    mu_eq = self._mu_less(abs(c3 - c1), self._equal, 3)

    confidence = mu_eq

    return confidence * 100.0
