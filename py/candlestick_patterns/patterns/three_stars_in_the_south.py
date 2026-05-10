"""Three Stars In The South pattern (3-candle bullish)."""
from __future__ import annotations

from ..core.primitives import (
    is_black, real_body, lower_shadow, upper_shadow,
)


def three_stars_in_the_south(self) -> int:
    """Three Stars In The South: a three-candle bullish pattern.

    Must have:
    - all three candles are black,
    - first candle: long body with long lower shadow,
    - second candle: smaller body, opens within or above prior range,
      trades lower but its low does not go below the first candle's low,
    - third candle: small marubozu (very short shadows) engulfed by the
      second candle's range.

    The meaning of "long" is specified with self._long_body.
    The meaning of "short" is specified with self._short_body.
    The meaning of "long" for shadows is specified with self._long_shadow.
    The meaning of "very short" for shadows is specified with
    self._very_short_shadow.

    Returns:
        +100 for bullish, 0 for no pattern.
    """
    if not self._enough(
        3, self._long_body, self._short_body,
        self._long_shadow, self._very_short_shadow,
    ):
        return 0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    if not (is_black(o1, c1) and is_black(o2, c2) and is_black(o3, c3)):
        return 0

    rb1 = real_body(o1, c1)
    rb2 = real_body(o2, c2)

    # First candle: long body with long lower shadow.
    if not (rb1 > self._avg(self._long_body, 3) and
            lower_shadow(o1, l1, c1) > self._avg(self._long_shadow, 3)):
        return 0

    # Second candle: smaller body, opens within or above prior range,
    # trades lower but not below prior low.
    if not (rb2 < rb1 and o2 <= h1 and o2 >= l1 and l2 >= l1):
        return 0

    # Third candle: small marubozu engulfed by second candle's range.
    vs1 = self._avg(self._very_short_shadow, 1)
    if not (real_body(o3, c3) < self._avg(self._short_body, 1) and
            upper_shadow(o3, h3, c3) < vs1 and
            lower_shadow(o3, l3, c3) < vs1 and
            h3 <= h2 and l3 >= l2):
        return 0

    return 100
