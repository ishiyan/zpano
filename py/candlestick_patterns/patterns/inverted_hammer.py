"""Inverted Hammer pattern (2-candle bullish)."""
from __future__ import annotations

from ..core.primitives import real_body, upper_shadow, lower_shadow, is_real_body_gap_down


def inverted_hammer(self) -> int:
    """Inverted Hammer: a two-candle bullish pattern.

    Must have:
    - small real body,
    - long upper shadow,
    - very short lower shadow,
    - gap down from the previous candle's real body.

    Returns:
        +100 for bullish, 0 for no pattern.
    """
    if not self._enough(2, self._short_body, self._long_shadow,
                        self._very_short_shadow):
        return 0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    if (real_body(o2, c2) < self._avg(self._short_body, 1) and
            upper_shadow(o2, h2, c2) > self._avg(self._long_shadow, 1) and
            lower_shadow(o2, l2, c2) < self._avg(self._very_short_shadow, 1) and
            is_real_body_gap_down(o1, c1, o2, c2)):
        return 100

    return 0
