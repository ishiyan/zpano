"""Takuri pattern (1-candle)."""
from __future__ import annotations

from ..core.primitives import real_body, upper_shadow, lower_shadow


def takuri(self) -> int:
    """Takuri (Dragonfly Doji with very long lower shadow): a one-candle pattern.

    A doji body with a very short upper shadow and a very long lower shadow.

    The meaning of "doji" is specified with self._doji_body.
    The meaning of "very short" for shadows is specified with
    self._very_short_shadow.
    The meaning of "very long" for shadows is specified with
    self._very_long_shadow.

    Returns:
        100 if detected, 0 for no pattern.
    """
    if not self._enough(1, self._doji_body, self._very_short_shadow,
                        self._very_long_shadow):
        return 0

    o, h, l, c = self._bar(1)

    if (real_body(o, c) <= self._avg(self._doji_body, 1) and
            upper_shadow(o, h, c) < self._avg(self._very_short_shadow, 1) and
            lower_shadow(o, l, c) > self._avg(self._very_long_shadow, 1)):
        return 100

    return 0
