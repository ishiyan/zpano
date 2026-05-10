"""Rising/Falling Three Methods pattern (5-candle continuation)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body


def rising_falling_three_methods(self) -> int:
    """Rising/Falling Three Methods: a five-candle continuation pattern.

    Uses TA-Lib logic: opposite-color check via color multiplication,
    real-body overlap (not full candle containment), sequential closes,
    5th opens beyond 4th close.

    Returns:
        +100 for bullish (rising), -100 for bearish (falling),
        0 for no pattern.
    """
    if not self._enough(5, self._long_body, self._short_body):
        return 0

    o1, h1, l1, c1 = self._bar(5)
    o2, h2, l2, c2 = self._bar(4)
    o3, h3, l3, c3 = self._bar(3)
    o4, h4, l4, c4 = self._bar(2)
    o5, h5, l5, c5 = self._bar(1)

    # 1st long, then 3 small, 5th long
    if not (real_body(o1, c1) > self._avg(self._long_body, 5) and
            real_body(o2, c2) < self._avg(self._short_body, 4) and
            real_body(o3, c3) < self._avg(self._short_body, 3) and
            real_body(o4, c4) < self._avg(self._short_body, 2) and
            real_body(o5, c5) > self._avg(self._long_body, 1)):
        return 0

    # Determine color of 1st candle: +1 white, -1 black
    color1 = 1 if is_white(o1, c1) else -1

    # Color check: white, 3 black, white  OR  black, 3 white, black
    color2 = 1 if is_white(o2, c2) else -1
    color3 = 1 if is_white(o3, c3) else -1
    color4 = 1 if is_white(o4, c4) else -1
    color5 = 1 if is_white(o5, c5) else -1

    if not (color2 == -color1 and color3 == color2 and
            color4 == color3 and color5 == -color4):
        return 0

    # 2nd to 4th hold within 1st: a part of the real body overlaps 1st range
    if not (min(o2, c2) < h1 and max(o2, c2) > l1 and
            min(o3, c3) < h1 and max(o3, c3) > l1 and
            min(o4, c4) < h1 and max(o4, c4) > l1):
        return 0

    # 2nd to 4th are falling (rising) — using color multiply trick
    if not (c3 * color1 < c2 * color1 and
            c4 * color1 < c3 * color1):
        return 0

    # 5th opens above (below) the prior close
    if not (o5 * color1 > c4 * color1):
        return 0

    # 5th closes above (below) the 1st close
    if c5 * color1 > c1 * color1:
        return 100 * color1

    return 0
