"""Doji pattern."""
from __future__ import annotations

from ..core.primitives import real_body


def doji(self) -> float:
    """Doji: open quite equal to close.

    Output is positive but this does not mean it is bullish:
    doji shows uncertainty and is neither bullish nor bearish when
    considered alone.

    The meaning of "doji" is specified with self._doji_body.

    Returns:
        Continuous float in [0, 100].  Higher = stronger doji signal.
    """
    if not self._enough(1, self._doji_body):
        return 0.0

    o1, h1, l1, c1 = self._bar(1)

    # Fuzzy: degree to which real_body <= doji_avg.
    confidence = self._mu_less(real_body(o1, c1), self._doji_body, 1)
    return confidence * 100.0
