"""High Wave pattern (1-candle)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body, upper_shadow, lower_shadow


def high_wave(self) -> int:
    """High Wave: a one-candle pattern.

    Must have:
    - short real body,
    - very long upper shadow,
    - very long lower shadow.

    The meaning of "short" is specified with self._short_body.
    The meaning of "very long" (shadow) is specified with self._very_long_shadow.

    Returns:
        +100 for white candle, -100 for black candle, 0 for no pattern.
    """
    if not self._enough(1, self._short_body, self._very_long_shadow):
        return 0

    o1, h1, l1, c1 = self._bar(1)

    if (real_body(o1, c1) < self._avg(self._short_body, 1) and
            upper_shadow(o1, h1, c1) > self._avg(self._very_long_shadow, 1) and
            lower_shadow(o1, l1, c1) > self._avg(self._very_long_shadow, 1)):
        if is_white(o1, c1):
            return 100
        return -100

    return 0
