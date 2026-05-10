"""Harami pattern (2-candle reversal)."""
from __future__ import annotations

from ..core.primitives import real_body


def harami(self) -> int:
    """Harami: a two-candle reversal pattern.

    Must have:
    - first candle: long real body,
    - second candle: short real body contained within the first candle's
      real body.

    Returns strictly inside = 100, edge-touching = 80, direction from 1st
    candle color: positive if 1st is black (bullish), negative if 1st is
    white (bearish).

    The meaning of "long" is specified with self._long_body.
    The meaning of "short" is specified with self._short_body.

    Returns:
        +/-100 or +/-80 for pattern detected, 0 for no pattern.
    """
    if not self._enough(2, self._long_body, self._short_body):
        return 0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    # First candle must be long, second must be short.
    if not (real_body(o1, c1) > self._avg(self._long_body, 2) and
            real_body(o2, c2) <= self._avg(self._short_body, 1)):
        return 0

    # Direction from 1st candle color: black (falling) -> bullish reversal.
    color1 = 1 if c1 >= o1 else -1

    # Strictly inside: both ends of 2nd body are strictly within 1st body.
    if (max(o2, c2) < max(o1, c1) and min(o2, c2) > min(o1, c1)):
        return -color1 * 100

    # Edge-touching: 2nd body is within or touching 1st body boundaries.
    if (max(o2, c2) <= max(o1, c1) and min(o2, c2) >= min(o1, c1)):
        return -color1 * 80

    return 0
