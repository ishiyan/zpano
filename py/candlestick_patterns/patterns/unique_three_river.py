"""Unique Three River pattern (3-candle bullish)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body
from ...fuzzy import t_product_all


def unique_three_river(self) -> float:
    """Unique Three River: a three-candle bullish pattern.

    Must have:
    - first candle: long black,
    - second candle: black harami (body within first body) with a lower
      low than the first candle,
    - third candle: small white, opens not lower than the second candle's
      low.

    The meaning of "long" is specified with self._long_body.
    The meaning of "short" is specified with self._short_body.

    Category A: always bullish (continuous).

    Returns:
        Continuous float in [0, 100].  Always bullish.
    """
    if not self._enough(3, self._long_body, self._short_body):
        return 0.0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    # Crisp gates: colors.
    if not (is_black(o1, c1) and is_black(o2, c2) and is_white(o3, c3)):
        return 0.0

    # Crisp: harami body containment and lower low.
    if not (c2 > c1 and o2 <= o1 and l2 < l1):
        return 0.0

    # Crisp: third opens not lower than second's low.
    if not (o3 >= l2):
        return 0.0

    # Fuzzy: first candle is long.
    mu_long1 = self._mu_greater(real_body(o1, c1), self._long_body, 3)

    # Fuzzy: third candle is short.
    mu_short3 = self._mu_less(real_body(o3, c3), self._short_body, 1)

    confidence = t_product_all(mu_long1, mu_short3)

    return confidence * 100.0
