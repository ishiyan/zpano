"""Advance Block pattern (3-candle bearish)."""
from __future__ import annotations

from ..core.primitives import is_white, real_body, upper_shadow


def advance_block(self) -> int:
    """Advance Block: a bearish three-candle pattern.

    Three white candles with consecutively higher closes and opens, but
    showing signs of weakening (diminishing bodies, growing upper shadows).

    Returns:
        -100 for bearish, 0 for no pattern.
    """
    if not self._enough(
        3, self._long_body, self._short_shadow, self._long_shadow,
        self._near, self._far,
    ):
        return 0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    # All three candles must be white with rising closes.
    if not (is_white(o1, c1) and is_white(o2, c2) and is_white(o3, c3) and
            c3 > c2 > c1):
        return 0

    # 2nd opens within/near 1st real body
    if not (o2 > o1 and
            o2 <= c1 + self._avg(self._near, 3)):
        return 0

    # 3rd opens within/near 2nd real body
    if not (o3 > o2 and
            o3 <= c2 + self._avg(self._near, 2)):
        return 0

    rb1 = real_body(o1, c1)
    rb2 = real_body(o2, c2)
    rb3 = real_body(o3, c3)

    # First candle must be long with a short upper shadow.
    if not (rb1 > self._avg(self._long_body, 3) and
            upper_shadow(o1, h1, c1) < self._avg(self._short_shadow, 3)):
        return 0

    # At least one weakness condition must hold:
    far2 = self._avg(self._far, 3)     # Far avg at 1st candle position
    far1 = self._avg(self._far, 2)     # Far avg at 2nd candle position
    near1 = self._avg(self._near, 2)   # Near avg at 2nd candle position

    if (
        # (2 far smaller than 1 && 3 not longer than 2)
        (rb2 < rb1 - far2 and
         rb3 < rb2 + near1) or
        # 3 far smaller than 2
        (rb3 < rb2 - far1) or
        # (3 < 2 && 2 < 1 && (3 or 2 has non-short upper shadow))
        (rb3 < rb2 and rb2 < rb1 and
         (upper_shadow(o3, h3, c3) > self._avg(self._short_shadow, 1) or
          upper_shadow(o2, h2, c2) > self._avg(self._short_shadow, 2))) or
        # (3 < 2 && 3 has long upper shadow)
        (rb3 < rb2 and
         upper_shadow(o3, h3, c3) > self._avg(self._long_shadow, 1))
    ):
        return -100

    return 0
