"""Gravestone Doji pattern (1-candle)."""
from __future__ import annotations

from ..core.primitives import real_body, upper_shadow, lower_shadow


def gravestone_doji(self) -> int:
    """Gravestone Doji: a one-candle pattern.

    Must have:
    - doji body (very small real body relative to high-low range),
    - no or very short lower shadow,
    - upper shadow is not very short.

    The meaning of "doji" is specified with self._doji_body.
    The meaning of "very short" is specified with self._very_short_shadow.

    Returns:
        +100 for pattern detected, 0 for no pattern.
    """
    if not self._enough(1, self._doji_body, self._very_short_shadow):
        return 0

    o1, h1, l1, c1 = self._bar(1)

    if (real_body(o1, c1) <= self._avg(self._doji_body, 1) and
            lower_shadow(o1, l1, c1) < self._avg(self._very_short_shadow, 1) and
            upper_shadow(o1, h1, c1) > self._avg(self._very_short_shadow, 1)):
        return 100

    return 0
