"""Long Legged Doji pattern (1-candle)."""
from __future__ import annotations

from ..core.primitives import real_body, upper_shadow, lower_shadow


def long_legged_doji(self) -> int:
    """Long Legged Doji: a one-candle pattern.

    Must have:
    - doji body (very small real body),
    - one or both shadows are long.

    The meaning of "doji" is specified with self._doji_body.
    The meaning of "long" for shadows is specified with self._long_shadow.

    Returns:
        +100 for pattern detected, 0 for no pattern.
    """
    if not self._enough(1, self._doji_body, self._long_shadow):
        return 0

    o1, h1, l1, c1 = self._bar(1)

    if (real_body(o1, c1) <= self._avg(self._doji_body, 1) and
            (upper_shadow(o1, h1, c1) > self._avg(self._long_shadow, 1) or
             lower_shadow(o1, l1, c1) > self._avg(self._long_shadow, 1))):
        return 100

    return 0
