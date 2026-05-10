"""Three White Soldiers pattern (3-candle bullish)."""
from __future__ import annotations

from ..core.primitives import is_white, real_body, upper_shadow


def three_white_soldiers(self) -> int:
    """Three White Soldiers: a three-candle bullish pattern.

    Must have:
    - three consecutive white candles with consecutively higher closes,
    - all three have very short upper shadows,
    - each opens within or near the prior candle's real body,
    - none is far shorter than the prior candle,
    - third candle is not short.

    Returns:
        +100 for bullish, 0 for no pattern.
    """
    if not self._enough(
        3, self._short_body, self._very_short_shadow,
        self._near, self._far,
    ):
        return 0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    # All three must be white with consecutively higher closes.
    if not (is_white(o1, c1) and is_white(o2, c2) and is_white(o3, c3) and
            c3 > c2 > c1):
        return 0

    rb1 = real_body(o1, c1)
    rb2 = real_body(o2, c2)
    rb3 = real_body(o3, c3)

    # Very short upper shadows — ALL THREE candles (TA-Lib checks all 3).
    if not (upper_shadow(o1, h1, c1) < self._avg(self._very_short_shadow, 3) and
            upper_shadow(o2, h2, c2) < self._avg(self._very_short_shadow, 2) and
            upper_shadow(o3, h3, c3) < self._avg(self._very_short_shadow, 1)):
        return 0

    # Each opens within or near the prior body.
    if not (o2 > o1 and
            o2 <= c1 + self._avg(self._near, 3) and
            o3 > o2 and
            o3 <= c2 + self._avg(self._near, 2)):
        return 0

    # Not far shorter than the prior candle.
    # TA-Lib: rb2 > rb1 - far_avg(1st), rb3 > rb2 - far_avg(2nd)
    if not (rb2 > rb1 - self._avg(self._far, 3) and
            rb3 > rb2 - self._avg(self._far, 2)):
        return 0

    # Third candle is not short.
    if rb3 < self._avg(self._short_body, 1):
        return 0

    return 100
