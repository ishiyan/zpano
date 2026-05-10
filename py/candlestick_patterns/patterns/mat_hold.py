"""Mat Hold pattern (5-candle bullish)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body, is_real_body_gap_up

MAT_HOLD_PENETRATION_FACTOR: float = 0.5


def mat_hold(self) -> int:
    """Mat Hold: a five-candle bullish continuation pattern.

    Must have:
    - first candle: long white,
    - second candle: small, black, gaps up from first,
    - third and fourth candles: small,
    - reaction candles (2-4) are falling, hold within first body
      (penetration check),
    - fifth candle: white, opens above prior close, closes above
      highest high of reaction candles.

    Returns:
        +100 for bullish, 0 for no pattern.
    """
    if not self._enough(5, self._long_body, self._short_body):
        return 0

    o1, h1, l1, c1 = self._bar(5)
    o2, h2, l2, c2 = self._bar(4)
    o3, h3, l3, c3 = self._bar(3)
    o4, h4, l4, c4 = self._bar(2)
    o5, h5, l5, c5 = self._bar(1)

    penetration = MAT_HOLD_PENETRATION_FACTOR

    # 1st long, then 3 small
    if not (real_body(o1, c1) > self._avg(self._long_body, 5) and
            real_body(o2, c2) < self._avg(self._short_body, 4) and
            real_body(o3, c3) < self._avg(self._short_body, 3) and
            real_body(o4, c4) < self._avg(self._short_body, 2)):
        return 0

    # White, black, ?, ?, white
    if not (is_white(o1, c1) and is_black(o2, c2) and is_white(o5, c5)):
        return 0

    # Upside gap 1st to 2nd
    if not is_real_body_gap_up(o1, c1, o2, c2):
        return 0

    # 3rd to 4th hold within 1st: a part of real body within 1st range
    if not (min(o3, c3) < c1 and min(o4, c4) < c1):
        return 0

    # Reaction days don't penetrate first body too much
    if not (min(o3, c3) > c1 - real_body(o1, c1) * penetration and
            min(o4, c4) > c1 - real_body(o1, c1) * penetration):
        return 0

    # 2nd to 4th are falling
    if not (max(o3, c3) < o2 and
            max(o4, c4) < max(o3, c3)):
        return 0

    # 5th opens above prior close
    if not (o5 > c4):
        return 0

    # 5th closes above highest high of reaction candles
    if c5 > max(h2, h3, h4):
        return 100

    return 0
