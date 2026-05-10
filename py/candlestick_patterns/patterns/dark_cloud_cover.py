"""Dark Cloud Cover pattern (2-candle bearish)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body

DARK_CLOUD_COVER_PENETRATION_FACTOR: float = 0.5


def dark_cloud_cover(self) -> int:
    """Dark Cloud Cover: a two-candle bearish reversal pattern.

    Must have:
    - first candle: long white candle,
    - second candle: black candle that opens above the prior high and
      closes well within the first candle's real body (below the midpoint).

    The penetration into the first candle's body is controlled by
    DARK_CLOUD_COVER_PENETRATION_FACTOR (default 0.5 = midpoint).

    The meaning of "long" is specified with self._long_body.

    Returns:
        -100 for bearish, 0 for no pattern.
    """
    if not self._enough(2, self._long_body):
        return 0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    penetration = DARK_CLOUD_COVER_PENETRATION_FACTOR

    if (is_white(o1, c1) and is_black(o2, c2) and
            real_body(o1, c1) > self._avg(self._long_body, 2) and
            o2 > h1 and
            c2 < c1 - real_body(o1, c1) * penetration and
            c2 > o1):
        return -100

    return 0
