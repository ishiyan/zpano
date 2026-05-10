"""Shooting Star pattern (2-candle bearish)."""
from __future__ import annotations

from ..core.primitives import real_body, upper_shadow, lower_shadow, is_real_body_gap_up


def shooting_star(self) -> int:
    """Shooting Star: a two-candle bearish reversal pattern.

    Must have:
    - gap up from the previous candle (real body gap up),
    - small real body,
    - long upper shadow,
    - very short lower shadow.

    The meaning of "short" is specified with self._short_body.
    The meaning of "long" (shadow) is specified with self._long_shadow.
    The meaning of "very short" (shadow) is specified with
    self._very_short_shadow.

    Returns:
        -100 for bearish, 0 for no pattern.
    """
    if not self._enough(2, self._short_body, self._long_shadow,
                        self._very_short_shadow):
        return 0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    if (real_body(o2, c2) < self._avg(self._short_body, 1) and
            upper_shadow(o2, h2, c2) > self._avg(self._long_shadow, 1) and
            lower_shadow(o2, l2, c2) < self._avg(self._very_short_shadow, 1) and
            is_real_body_gap_up(o1, c1, o2, c2)):
        return -100

    return 0
