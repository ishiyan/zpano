"""Spinning Top pattern (1-candle)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body, upper_shadow, lower_shadow


def spinning_top(self) -> int:
    """Spinning Top: a one-candle pattern.

    A candle with a small body and shadows longer than the body on both sides.

    The meaning of "short" is specified with self._short_body.

    Returns:
        +100 for white, -100 for black, 0 for no pattern.
    """
    if not self._enough(1, self._short_body):
        return 0

    o, h, l, c = self._bar(1)

    rb = real_body(o, c)

    if (rb < self._avg(self._short_body, 1) and
            upper_shadow(o, h, c) > rb and
            lower_shadow(o, l, c) > rb):
        if is_white(o, c):
            return 100
        if is_black(o, c):
            return -100

    return 0
