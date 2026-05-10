"""Short Line pattern (1-candle)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body, upper_shadow, lower_shadow


def short_line(self) -> int:
    """Short Line: a one-candle pattern.

    A candle with a short body, short upper shadow, and short lower shadow.

    The meaning of "short" for body is specified with self._short_body.
    The meaning of "short" for shadows is specified with self._short_shadow.

    Returns:
        +100 for white, -100 for black, 0 for no pattern.
    """
    if not self._enough(1, self._short_body, self._short_shadow):
        return 0

    o, h, l, c = self._bar(1)

    if (real_body(o, c) < self._avg(self._short_body, 1) and
            upper_shadow(o, h, c) < self._avg(self._short_shadow, 1) and
            lower_shadow(o, l, c) < self._avg(self._short_shadow, 1)):
        if is_white(o, c):
            return 100
        if is_black(o, c):
            return -100

    return 0
