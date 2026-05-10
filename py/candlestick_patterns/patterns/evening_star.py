"""Evening Star pattern (3-candle bearish reversal)."""
from __future__ import annotations

from ..core.primitives import (
    is_white, is_black, real_body, is_real_body_gap_up,
)

EVENING_STAR_PENETRATION_FACTOR: float = 0.3


def evening_star(self) -> int:
    """Evening Star: a three-candle bearish reversal pattern.

    Must have:
    - first candle: long white real body,
    - second candle: short real body that gaps up (real body gap up from the
      first),
    - third candle: black real body that moves well within the first candle's
      real body.

    The meaning of "long" is specified with self._long_body.
    The meaning of "short" is specified with self._short_body.

    Returns:
        -100 for bearish, 0 for no pattern.
    """
    if not self._enough(3, self._long_body, self._short_body):
        return 0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    penetration = EVENING_STAR_PENETRATION_FACTOR

    if (is_white(o1, c1) and
            real_body(o1, c1) > self._avg(self._long_body, 3) and
            real_body(o2, c2) <= self._avg(self._short_body, 2) and
            is_real_body_gap_up(o1, c1, o2, c2) and
            is_black(o3, c3) and
            c3 < c1 - real_body(o1, c1) * penetration):
        return -100

    return 0
