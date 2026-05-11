"""Matching Low pattern (2-candle bullish)."""
from __future__ import annotations

from ..core.primitives import is_black
from ...fuzzy import t_product_all


def matching_low(self) -> float:
    """Matching Low: a two-candle bullish pattern.

    Must have:
    - first candle: black,
    - second candle: black with close equal to the first candle's close.

    The meaning of "equal" is specified with self._equal.

    Category A: always bullish (continuous).

    Returns:
        Continuous float in [0, 100].  Always bullish.
    """
    if not self._enough(2, self._equal):
        return 0.0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    # Crisp gates: both black.
    if not (is_black(o1, c1) and is_black(o2, c2)):
        return 0.0

    # Fuzzy: close equal to prior close (two-sided band).
    mu_eq = self._mu_less(abs(c2 - c1), self._equal, 2)

    confidence = mu_eq

    return confidence * 100.0
