"""Mat Hold pattern (5-candle bullish)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body, is_real_body_gap_up
from ...fuzzy import t_product_all

MAT_HOLD_PENETRATION_FACTOR: float = 0.5


def mat_hold(self) -> float:
    """Mat Hold: a five-candle bullish continuation pattern.

    Must have:
    - first candle: long white,
    - second candle: small, black, gaps up from first,
    - third and fourth candles: small,
    - reaction candles (2-4) are falling, hold within first body
      (penetration check),
    - fifth candle: white, opens above prior close, closes above
      highest high of reaction candles.

    Category A: always bullish (continuous).

    Returns:
        Continuous float in [0, 100].  Always bullish.
    """
    if not self._enough(5, self._long_body, self._short_body):
        return 0.0

    o1, h1, l1, c1 = self._bar(5)
    o2, h2, l2, c2 = self._bar(4)
    o3, h3, l3, c3 = self._bar(3)
    o4, h4, l4, c4 = self._bar(2)
    o5, h5, l5, c5 = self._bar(1)

    penetration = MAT_HOLD_PENETRATION_FACTOR

    # Crisp gates: colors.
    if not (is_white(o1, c1) and is_black(o2, c2) and is_white(o5, c5)):
        return 0.0

    # Crisp: gap up from 1st to 2nd.
    if not is_real_body_gap_up(o1, c1, o2, c2):
        return 0.0

    # Crisp: 3rd to 4th hold within 1st range.
    if not (min(o3, c3) < c1 and min(o4, c4) < c1):
        return 0.0

    # Crisp: reaction days don't penetrate first body too much.
    rb1 = real_body(o1, c1)
    if not (min(o3, c3) > c1 - rb1 * penetration and
            min(o4, c4) > c1 - rb1 * penetration):
        return 0.0

    # Crisp: 2nd to 4th are falling.
    if not (max(o3, c3) < o2 and
            max(o4, c4) < max(o3, c3)):
        return 0.0

    # Crisp: 5th opens above prior close.
    if not (o5 > c4):
        return 0.0

    # Crisp: 5th closes above highest high of reaction candles.
    if not (c5 > max(h2, h3, h4)):
        return 0.0

    # Fuzzy: first candle long.
    mu_long1 = self._mu_greater(rb1, self._long_body, 5)

    # Fuzzy: 2nd, 3rd, 4th short.
    mu_short2 = self._mu_less(real_body(o2, c2), self._short_body, 4)
    mu_short3 = self._mu_less(real_body(o3, c3), self._short_body, 3)
    mu_short4 = self._mu_less(real_body(o4, c4), self._short_body, 2)

    confidence = t_product_all(mu_long1, mu_short2, mu_short3, mu_short4)

    return confidence * 100.0
