"""Tasuki Gap pattern (3-candle continuation)."""
from __future__ import annotations

from ..core.primitives import (
    is_white, is_black, real_body,
    is_real_body_gap_up, is_real_body_gap_down,
)


def tasuki_gap(self) -> int:
    """Tasuki Gap: a three-candle continuation pattern.

    Upside Tasuki Gap (bullish):
    - real-body gap up between 1st and 2nd candles,
    - 2nd candle: white,
    - 3rd candle: black, opens within 2nd white body, closes below 2nd
      open but above 1st candle's real body top (inside the gap),
    - 2nd and 3rd have near-equal body sizes.

    Downside Tasuki Gap (bearish):
    - real-body gap down between 1st and 2nd candles,
    - 2nd candle: black,
    - 3rd candle: white, opens within 2nd black body, closes above 2nd
      open but below 1st candle's real body bottom (inside the gap),
    - 2nd and 3rd have near-equal body sizes.

    Returns:
        +100 for bullish, -100 for bearish, 0 for no pattern.
    """
    if not self._enough(3, self._near):
        return 0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    near2 = self._avg(self._near, 2)

    # Upside Tasuki Gap (bullish).
    if (is_real_body_gap_up(o1, c1, o2, c2) and
            is_white(o2, c2) and is_black(o3, c3) and
            o3 < c2 and o3 > o2 and          # opens within 2nd white rb
            c3 < o2 and                        # closes below 2nd open
            c3 > max(c1, o1) and              # closes inside the gap (above 1st body top)
            abs(real_body(o2, c2) - real_body(o3, c3)) < near2):
        return 100

    # Downside Tasuki Gap (bearish).
    if (is_real_body_gap_down(o1, c1, o2, c2) and
            is_black(o2, c2) and is_white(o3, c3) and
            o3 < o2 and o3 > c2 and          # opens within 2nd black rb
            c3 > o2 and                        # closes above 2nd open
            c3 < min(c1, o1) and              # closes inside the gap (below 1st body bottom)
            abs(real_body(o2, c2) - real_body(o3, c3)) < near2):
        return -100

    return 0
