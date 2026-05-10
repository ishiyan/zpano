"""Doji pattern."""
from __future__ import annotations

from ..core.primitives import real_body


def doji(self) -> int:
    """Doji: open quite equal to close.

    Output is positive (100) but this does not mean it is bullish:
    doji shows uncertainty and is neither bullish nor bearish when
    considered alone.

    The meaning of "doji" is specified with self._doji_body.

    Returns:
        100 if doji detected, 0 otherwise.
    """
    if not self._enough(1, self._doji_body):
        return 0

    o1, h1, l1, c1 = self._bar(1)

    if real_body(o1, c1) <= self._avg(self._doji_body, 1):
        return 100

    return 0
