"""In Neck pattern (2-candle bearish continuation)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body


def in_neck(self) -> int:
    """In Neck: a two-candle bearish continuation pattern.

    Must have:
    - first candle: long black,
    - second candle: white, opens below the prior low, closes slightly
      into the prior real body (close near the prior close).

    The meaning of "long" is specified with self._long_body.
    The meaning of "near" is specified with self._near.

    Returns:
        -100 for bearish, 0 for no pattern.
    """
    if not self._enough(2, self._long_body, self._near):
        return 0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    if (is_black(o1, c1) and is_white(o2, c2) and
            real_body(o1, c1) > self._avg(self._long_body, 2) and
            o2 < l1 and
            abs(c2 - c1) < self._avg(self._near, 1)):
        return -100

    return 0
