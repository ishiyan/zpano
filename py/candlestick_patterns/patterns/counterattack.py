"""Counterattack pattern (2-candle reversal)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body


def counterattack(self) -> int:
    """Counterattack: a two-candle reversal pattern.

    Two long candles of opposite color with closes that are equal
    (or very near equal).

    - bullish: first candle is long black, second is long white,
      closes are equal,
    - bearish: first candle is long white, second is long black,
      closes are equal.

    The meaning of "long" is specified with self._long_body.
    The meaning of "equal" is specified with self._equal.

    Returns:
        +100 for bullish, -100 for bearish, 0 for no pattern.
    """
    if not self._enough(2, self._long_body, self._equal):
        return 0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    color1 = 1 if c1 >= o1 else -1
    color2 = 1 if c2 >= o2 else -1

    eq = self._avg(self._equal, 2)

    if (color1 == -color2 and
            real_body(o1, c1) > self._avg(self._long_body, 2) and
            real_body(o2, c2) > self._avg(self._long_body, 1) and
            c2 <= c1 + eq and
            c2 >= c1 - eq):
        return color2 * 100

    return 0
