"""Belt Hold pattern (1-candle)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body, upper_shadow, lower_shadow


def belt_hold(self) -> int:
    """Belt Hold: a one-candle pattern.

    A long candle with a very short shadow on the opening side:
    - bullish: long white candle with very short lower shadow,
    - bearish: long black candle with very short upper shadow.

    The meaning of "long" is specified with self._long_body.
    The meaning of "very short" for shadows is specified with
    self._very_short_shadow.

    Returns:
        +100 for bullish, -100 for bearish, 0 for no pattern.
    """
    if not self._enough(1, self._long_body, self._very_short_shadow):
        return 0

    o, h, l, c = self._bar(1)

    if real_body(o, c) > self._avg(self._long_body, 1):
        vs = self._avg(self._very_short_shadow, 1)
        if is_white(o, c) and lower_shadow(o, l, c) < vs:
            return 100
        if is_black(o, c) and upper_shadow(o, h, c) < vs:
            return -100

    return 0
