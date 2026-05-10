"""Abandoned Baby pattern (3-candle reversal)."""
from __future__ import annotations

from ..core.primitives import (
    is_white, is_black, real_body, is_high_low_gap_up, is_high_low_gap_down,
)

ABANDONED_BABY_PENETRATION_FACTOR: float = 0.3


def abandoned_baby(self) -> int:
    """Abandoned Baby: a three-candle reversal pattern.

    Must have:
    - first candle: long real body,
    - second candle: doji,
    - third candle: real body longer than short, opposite color to 1st,
      closes well within 1st body,
    - upside/downside gap between 1st and doji (shadows don't touch),
    - downside/upside gap between doji and 3rd (shadows don't touch).

    Returns:
        +100 for bullish, -100 for bearish, 0 for no pattern.
    """
    if not self._enough(3, self._long_body, self._doji_body, self._short_body):
        return 0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    # 1st: long body, 2nd: doji, 3rd: longer than short
    if not (real_body(o1, c1) > self._avg(self._long_body, 3) and
            real_body(o2, c2) <= self._avg(self._doji_body, 2) and
            real_body(o3, c3) > self._avg(self._short_body, 1)):
        return 0

    penetration = ABANDONED_BABY_PENETRATION_FACTOR

    # Bearish: white-doji-black, gap up then gap down
    if (is_white(o1, c1) and is_black(o3, c3) and
            c3 < c1 - real_body(o1, c1) * penetration and
            is_high_low_gap_up(h1, l2) and
            is_high_low_gap_down(l2, h3)):
        return -100

    # Bullish: black-doji-white, gap down then gap up
    if (is_black(o1, c1) and is_white(o3, c3) and
            c3 > c1 + real_body(o1, c1) * penetration and
            is_high_low_gap_down(l1, h2) and
            is_high_low_gap_up(h2, l3)):
        return 100

    return 0
