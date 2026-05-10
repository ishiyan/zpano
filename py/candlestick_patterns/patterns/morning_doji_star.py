"""Morning Doji Star pattern (3-candle bullish reversal)."""
from __future__ import annotations

from ..core.primitives import (
    is_white, is_black, real_body, is_real_body_gap_down,
)

MORNING_DOJI_STAR_PENETRATION_FACTOR: float = 0.3


def morning_doji_star(self) -> int:
    """Morning Doji Star: a three-candle bullish reversal pattern.

    Must have:
    - first candle: long black real body,
    - second candle: doji that gaps down (real body gap down from the first),
    - third candle: white real body that closes well within the first candle's
      real body.

    The meaning of "long" is specified with self._long_body.
    The meaning of "doji" is specified with self._doji_body.
    The meaning of "short" is specified with self._short_body.

    Returns:
        +100 for bullish, 0 for no pattern.
    """
    if not self._enough(3, self._long_body, self._doji_body, self._short_body):
        return 0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    penetration = MORNING_DOJI_STAR_PENETRATION_FACTOR

    if (is_black(o1, c1) and
            real_body(o1, c1) > self._avg(self._long_body, 3) and
            real_body(o2, c2) <= self._avg(self._doji_body, 2) and
            is_real_body_gap_down(o1, c1, o2, c2) and
            is_white(o3, c3) and
            c3 > c1 + real_body(o1, c1) * penetration):
        return 100

    return 0
