"""Three Black Crows pattern (4-candle bearish)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, lower_shadow
from ...fuzzy import t_product_all


def three_black_crows(self) -> float:
    """Three Black Crows: a four-candle bearish reversal pattern.

    Must have:
    - preceding candle (oldest) is white,
    - three consecutive black candles with declining closes,
    - each opens within the prior black candle's real body,
    - each has a very short lower shadow,
    - 1st black closes under the prior white candle's high.

    Category A: always bearish (continuous).

    Returns:
        Continuous float in [-100, 0].  Always bearish.
    """
    if not self._enough(4, self._very_short_shadow):
        return 0.0

    o0, h0, l0, c0 = self._bar(4)  # prior white
    o1, h1, l1, c1 = self._bar(3)  # 1st black
    o2, h2, l2, c2 = self._bar(2)  # 2nd black
    o3, h3, l3, c3 = self._bar(1)  # 3rd black

    # Crisp gates: colors, declining closes, opens within prior body.
    if not is_white(o0, c0):
        return 0.0
    if not (is_black(o1, c1) and is_black(o2, c2) and is_black(o3, c3)):
        return 0.0
    if not (c1 > c2 and c2 > c3):
        return 0.0
    # Opens within prior black body (crisp containment for strict ordering).
    if not (o2 < o1 and o2 > c1 and o3 < o2 and o3 > c2):
        return 0.0
    # Prior white's high > 1st black's close (crisp).
    if not (h0 > c1):
        return 0.0

    # Fuzzy: very short lower shadows.
    mu_ls1 = self._mu_less(lower_shadow(o1, l1, c1), self._very_short_shadow, 3)
    mu_ls2 = self._mu_less(lower_shadow(o2, l2, c2), self._very_short_shadow, 2)
    mu_ls3 = self._mu_less(lower_shadow(o3, l3, c3), self._very_short_shadow, 1)

    confidence = t_product_all(mu_ls1, mu_ls2, mu_ls3)

    return -1.0 * confidence * 100.0
