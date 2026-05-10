"""Marubozu pattern (1-candle)."""
from __future__ import annotations

from ..core.primitives import is_white, real_body, upper_shadow, lower_shadow


def marubozu(self) -> int:
    """Marubozu: a one-candle pattern.

    Must have:
    - long real body,
    - very short upper shadow,
    - very short lower shadow.

    The meaning of "long" is specified with self._long_body.
    The meaning of "very short" for shadows is specified with
    self._very_short_shadow.

    Returns:
        +100 for bullish (white), -100 for bearish (black), 0 for no pattern.
    """
    if not self._enough(1, self._long_body, self._very_short_shadow):
        return 0

    o, h, l, c = self._bar(1)

    vs = self._avg(self._very_short_shadow, 1)

    if (real_body(o, c) > self._avg(self._long_body, 1) and
            upper_shadow(o, h, c) < vs and
            lower_shadow(o, l, c) < vs):
        return 100 if is_white(o, c) else -100

    return 0
