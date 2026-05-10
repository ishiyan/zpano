"""Hanging Man pattern (2-candle bearish)."""
from __future__ import annotations

from ..core.primitives import real_body, lower_shadow, upper_shadow


def hanging_man(self) -> int:
    """Hanging Man: a two-candle bearish pattern.

    Must have:
    - small real body,
    - long lower shadow,
    - no or very short upper shadow,
    - body is above or near the highs of the previous candle.

    The meaning of "short" is specified with self._short_body.
    The meaning of "long" (shadow) is specified with self._long_shadow.
    The meaning of "very short" (shadow) is specified with self._very_short_shadow.
    The meaning of "near" is specified with self._near.

    Returns:
        -100 for bearish, 0 for no pattern.
    """
    if not self._enough(2, self._short_body, self._long_shadow,
                        self._very_short_shadow, self._near):
        return 0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    near_avg = self._avg(self._near, 1)

    if (real_body(o2, c2) < self._avg(self._short_body, 1) and
            lower_shadow(o2, l2, c2) > self._avg(self._long_shadow, 1) and
            upper_shadow(o2, h2, c2) < self._avg(self._very_short_shadow, 1) and
            max(o2, c2) >= h1 - near_avg):
        return -100

    return 0
