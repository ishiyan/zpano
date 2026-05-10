"""Separating Lines pattern (2-candle continuation)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body, upper_shadow, lower_shadow


def separating_lines(self) -> int:
    """Separating Lines: a two-candle continuation pattern.

    Opposite colors with the same open. The second candle is a belt hold
    (long body with no shadow on the opening side).

    - bullish: first candle is black, second is white with same open,
      long body, very short lower shadow,
    - bearish: first candle is white, second is black with same open,
      long body, very short upper shadow.

    The meaning of "long" is specified with self._long_body.
    The meaning of "very short" for shadows is specified with
    self._very_short_shadow.
    The meaning of "equal" is specified with self._equal.

    Returns:
        +100 for bullish, -100 for bearish, 0 for no pattern.
    """
    if not self._enough(2, self._long_body, self._very_short_shadow,
                        self._equal):
        return 0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    color1 = 1 if c1 >= o1 else -1
    color2 = 1 if c2 >= o2 else -1

    if color1 == color2:
        return 0

    eq = self._avg(self._equal, 2)

    if o2 > o1 + eq or o2 < o1 - eq:
        return 0

    if real_body(o2, c2) <= self._avg(self._long_body, 1):
        return 0

    vs = self._avg(self._very_short_shadow, 1)

    # Bullish: white belt hold (very short lower shadow).
    if color2 == 1 and lower_shadow(o2, l2, c2) < vs:
        return 100

    # Bearish: black belt hold (very short upper shadow).
    if color2 == -1 and upper_shadow(o2, h2, c2) < vs:
        return -100

    return 0
