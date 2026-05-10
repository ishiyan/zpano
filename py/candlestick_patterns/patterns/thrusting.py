"""Thrusting pattern (2-candle bearish)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body


def thrusting(self) -> int:
    """Thrusting: a two-candle bearish continuation pattern.

    Must have:
    - first candle: long black,
    - second candle: white, opens below the prior candle's low, closes
      into the prior candle's real body but below the midpoint, and the
      close is not equal to the prior candle's close (to distinguish
      from in-neck).

    The meaning of "long" is specified with self._long_body.
    The meaning of "equal" is specified with self._equal.

    Returns:
        -100 for bearish, 0 for no pattern.
    """
    if not self._enough(2, self._long_body, self._equal):
        return 0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    rb1 = real_body(o1, c1)
    eq = self._avg(self._equal, 2)

    if (is_black(o1, c1) and is_white(o2, c2) and
            rb1 > self._avg(self._long_body, 2) and
            o2 < l1 and
            c2 > c1 + eq and
            c2 <= c1 + rb1 * 0.5):
        return -100

    return 0
