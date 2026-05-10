"""Identical Three Crows pattern (3-candle bearish)."""
from __future__ import annotations

from ..core.primitives import is_black, real_body, lower_shadow


def identical_three_crows(self) -> int:
    """Identical Three Crows: a three-candle bearish pattern.

    Must have:
    - three consecutive declining black candles,
    - each opens very close to the prior candle's close (equal criterion),
    - very short lower shadows.

    The meaning of "equal" is specified with self._equal.
    The meaning of "very short" for shadows is specified with
    self._very_short_shadow.

    Returns:
        -100 for bearish, 0 for no pattern.
    """
    if not self._enough(3, self._equal, self._very_short_shadow):
        return 0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    if not (is_black(o1, c1) and is_black(o2, c2) and is_black(o3, c3)):
        return 0

    # TA-Lib uses Equal (not Near) with <= two-sided band.
    # Declining closes, opens very close to prior close, very short lower shadows.
    eq2 = self._avg(self._equal, 3)
    eq1 = self._avg(self._equal, 2)

    if (lower_shadow(o1, l1, c1) < self._avg(self._very_short_shadow, 3) and
            lower_shadow(o2, l2, c2) < self._avg(self._very_short_shadow, 2) and
            lower_shadow(o3, l3, c3) < self._avg(self._very_short_shadow, 1) and
            c1 > c2 and c2 > c3 and
            o2 <= c1 + eq2 and o2 >= c1 - eq2 and
            o3 <= c2 + eq1 and o3 >= c2 - eq1):
        return -100

    return 0
