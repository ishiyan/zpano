"""Ladder Bottom pattern (5-candle bullish)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body, upper_shadow
from ...fuzzy import t_product_all


def ladder_bottom(self) -> float:
    """Ladder Bottom: a five-candle bullish pattern.

    Must have:
    - first three candles: descending black candles (each closes lower),
    - fourth candle: black with a long upper shadow,
    - fifth candle: white, opens above the fourth candle's real body,
      closes above the fourth candle's high.

    The meaning of "long" for shadows is specified with self._long_shadow.

    Category A: always bullish (continuous).

    Returns:
        Continuous float in [0, 100].  Always bullish.
    """
    if not self._enough(5, self._very_short_shadow):
        return 0.0

    o1, h1, l1, c1 = self._bar(5)
    o2, h2, l2, c2 = self._bar(4)
    o3, h3, l3, c3 = self._bar(3)
    o4, h4, l4, c4 = self._bar(2)
    o5, h5, l5, c5 = self._bar(1)

    # Crisp gates: colors.
    if not (is_black(o1, c1) and is_black(o2, c2) and
            is_black(o3, c3) and is_black(o4, c4) and
            is_white(o5, c5)):
        return 0.0

    # Crisp: three descending opens and closes.
    if not (o1 > o2 and o2 > o3 and c1 > c2 and c2 > c3):
        return 0.0

    # Crisp: fifth opens above fourth's open, closes above fourth's high.
    if not (o5 > o4 and c5 > h4):
        return 0.0

    # Fuzzy: fourth candle has upper shadow > very short avg.
    mu_us4 = self._mu_greater(upper_shadow(o4, h4, c4), self._very_short_shadow, 2)

    confidence = mu_us4

    return confidence * 100.0
