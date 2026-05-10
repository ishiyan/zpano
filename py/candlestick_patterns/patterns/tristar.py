"""Tristar pattern (3-candle reversal)."""
from __future__ import annotations

from ..core.primitives import (
    real_body, is_real_body_gap_up, is_real_body_gap_down,
)


def tristar(self) -> int:
    """Tristar: a three-candle reversal pattern with three dojis.

    Must have:
    - three consecutive doji candles,
    - if the second doji gaps up from the first and the third does not
      close higher than the second: bearish (-100),
    - if the second doji gaps down from the first and the third does not
      close lower than the second: bullish (+100).

    The meaning of "doji" is specified with self._doji_body.

    Returns:
        +100 for bullish, -100 for bearish, 0 for no pattern.
    """
    if not self._enough(3, self._doji_body):
        return 0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    doji_avg = self._avg(self._doji_body, 3)

    # All three must be dojis (TA-Lib uses the same average for all three).
    if not (real_body(o1, c1) <= doji_avg and
            real_body(o2, c2) <= doji_avg and
            real_body(o3, c3) <= doji_avg):
        return 0

    # Bearish: second gaps up, third is not higher than second.
    if (is_real_body_gap_up(o1, c1, o2, c2) and
            max(o3, c3) < max(o2, c2)):
        return -100

    # Bullish: second gaps down, third is not lower than second.
    if (is_real_body_gap_down(o1, c1, o2, c2) and
            min(o3, c3) > min(o2, c2)):
        return 100

    return 0
