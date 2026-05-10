"""Long Line pattern (1-candle)."""
from __future__ import annotations

from ..core.primitives import is_white, real_body, upper_shadow, lower_shadow


def long_line(self) -> int:
    """Long Line: a one-candle pattern.

    Must have:
    - long real body,
    - short upper shadow,
    - short lower shadow.

    The meaning of "long" is specified with self._long_body.
    The meaning of "short" for shadows is specified with self._short_shadow.

    Returns:
        +100 for bullish (white), -100 for bearish (black), 0 for no pattern.
    """
    if not self._enough(1, self._long_body, self._short_shadow):
        return 0

    o, h, l, c = self._bar(1)

    if (real_body(o, c) > self._avg(self._long_body, 1) and
            upper_shadow(o, h, c) < self._avg(self._short_shadow, 1) and
            lower_shadow(o, l, c) < self._avg(self._short_shadow, 1)):
        return 100 if is_white(o, c) else -100

    return 0
