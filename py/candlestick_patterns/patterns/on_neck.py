"""On Neck pattern (2-candle bearish)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body


def on_neck(self) -> int:
    """On Neck: a two-candle bearish continuation pattern.

    Must have:
    - first candle: long black,
    - second candle: white that opens below the prior low and closes
      equal to the prior candle's low.

    The meaning of "long" is specified with self._long_body.
    The meaning of "equal" is specified with self._equal.

    Returns:
        -100 for bearish, 0 for no pattern.
    """
    if not self._enough(2, self._long_body, self._equal):
        return 0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    eq = self._avg(self._equal, 1)

    if (is_black(o1, c1) and
            real_body(o1, c1) > self._avg(self._long_body, 2) and
            is_white(o2, c2) and
            o2 < l1 and
            abs(c2 - l1) <= eq):
        return -100

    return 0
