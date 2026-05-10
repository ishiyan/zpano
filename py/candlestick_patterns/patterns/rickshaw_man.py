"""Rickshaw Man pattern (1-candle)."""
from __future__ import annotations

from ..core.primitives import real_body, upper_shadow, lower_shadow


def rickshaw_man(self) -> int:
    """Rickshaw Man: a one-candle doji pattern.

    Must have:
    - doji body (very small real body),
    - two long shadows,
    - body near the midpoint of the high-low range.

    The meaning of "doji" is specified with self._doji_body.
    The meaning of "long" for shadows is specified with self._long_shadow.
    The meaning of "near" is specified with self._near.

    Returns:
        +100 for pattern detected, 0 for no pattern.
    """
    if not self._enough(1, self._doji_body, self._long_shadow, self._near):
        return 0

    o, h, l, c = self._bar(1)

    hl_range = h - l
    near_avg = self._avg(self._near, 1)

    if (real_body(o, c) <= self._avg(self._doji_body, 1) and
            upper_shadow(o, h, c) > self._avg(self._long_shadow, 1) and
            lower_shadow(o, l, c) > self._avg(self._long_shadow, 1) and
            min(o, c) <= l + hl_range / 2.0 + near_avg and
            max(o, c) >= l + hl_range / 2.0 - near_avg):
        return 100

    return 0
