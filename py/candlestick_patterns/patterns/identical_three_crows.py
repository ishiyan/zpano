"""Identical Three Crows pattern (3-candle bearish)."""
from __future__ import annotations

from ..core.primitives import is_black, real_body, lower_shadow
from ...fuzzy import t_product_all


def identical_three_crows(self) -> float:
    """Identical Three Crows: a three-candle bearish pattern.

    Must have:
    - three consecutive declining black candles,
    - each opens very close to the prior candle's close (equal criterion),
    - very short lower shadows.

    The meaning of "equal" is specified with self._equal.
    The meaning of "very short" for shadows is specified with
    self._very_short_shadow.

    Category A: always bearish (continuous).

    Returns:
        Continuous float in [-100, 0].  Always bearish.
    """
    if not self._enough(3, self._equal, self._very_short_shadow):
        return 0.0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    # Crisp gates: all black, declining closes.
    if not (is_black(o1, c1) and is_black(o2, c2) and is_black(o3, c3)):
        return 0.0
    if not (c1 > c2 and c2 > c3):
        return 0.0

    # Fuzzy conditions.
    mu_ls1 = self._mu_less(lower_shadow(o1, l1, c1), self._very_short_shadow, 3)
    mu_ls2 = self._mu_less(lower_shadow(o2, l2, c2), self._very_short_shadow, 2)
    mu_ls3 = self._mu_less(lower_shadow(o3, l3, c3), self._very_short_shadow, 1)

    # Opens near prior close (equal criterion, two-sided band).
    mu_eq2 = self._mu_less(abs(o2 - c1), self._equal, 3)
    mu_eq3 = self._mu_less(abs(o3 - c2), self._equal, 2)

    confidence = t_product_all(mu_ls1, mu_ls2, mu_ls3, mu_eq2, mu_eq3)

    return -1.0 * confidence * 100.0
