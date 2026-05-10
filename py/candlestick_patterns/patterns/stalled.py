"""Stalled pattern (3-candle bearish)."""
from __future__ import annotations

from ..core.primitives import is_white, real_body, upper_shadow


def stalled(self) -> int:
    """Stalled (Deliberation): a three-candle bearish pattern.

    Three white candles with progressively higher closes:
    - first candle: long white body,
    - second candle: long white body, opens within or near the first
      candle's body, very short upper shadow,
    - third candle: small body that rides on the shoulder of the second
      (opens near the second's close, accounting for its own body size).

    Returns:
        -100 for bearish, 0 for no pattern.
    """
    if not self._enough(3, self._long_body, self._short_body,
                        self._very_short_shadow, self._near):
        return 0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    if not (is_white(o1, c1) and is_white(o2, c2) and is_white(o3, c3)):
        return 0

    if not (c3 > c2 and c2 > c1):
        return 0

    rb3 = real_body(o3, c3)

    # TA-Lib: open[3rd] >= close[2nd] - body[3rd] - near_avg(2nd)
    # No upper bound check.
    if (real_body(o1, c1) > self._avg(self._long_body, 3) and
            real_body(o2, c2) > self._avg(self._long_body, 2) and
            upper_shadow(o2, h2, c2) < self._avg(self._very_short_shadow, 2) and
            o2 > o1 and
            o2 <= c1 + self._avg(self._near, 3) and
            rb3 < self._avg(self._short_body, 1) and
            o3 >= c2 - rb3 - self._avg(self._near, 2)):
        return -100

    return 0
