"""Abandoned Baby pattern (3-candle reversal)."""
from __future__ import annotations

from ..core.primitives import (
    is_white, is_black, real_body, is_high_low_gap_up, is_high_low_gap_down,
)
from ...fuzzy import t_product_all

ABANDONED_BABY_PENETRATION_FACTOR: float = 0.3


def abandoned_baby(self) -> float:
    """Abandoned Baby: a three-candle reversal pattern.

    Must have:
    - first candle: long real body,
    - second candle: doji,
    - third candle: real body longer than short, opposite color to 1st,
      closes well within 1st body,
    - upside/downside gap between 1st and doji (shadows don't touch),
    - downside/upside gap between doji and 3rd (shadows don't touch).

    Category C: both branches evaluated, return stronger signal.

    Returns:
        Continuous float in [-100, +100].
    """
    if not self._enough(3, self._long_body, self._doji_body, self._short_body):
        return 0.0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    # Shared fuzzy conditions: 1st long, 2nd doji, 3rd > short.
    mu_long1 = self._mu_greater(real_body(o1, c1), self._long_body, 3)
    mu_doji2 = self._mu_less(real_body(o2, c2), self._doji_body, 2)
    mu_short3 = self._mu_greater(real_body(o3, c3), self._short_body, 1)

    penetration = ABANDONED_BABY_PENETRATION_FACTOR

    # Bearish: white-doji-black, gap up then gap down.
    bear_signal = 0.0
    if is_white(o1, c1) and is_black(o3, c3):
        if is_high_low_gap_up(h1, l2) and is_high_low_gap_down(l2, h3):
            rb1 = real_body(o1, c1)
            pen_threshold = c1 - rb1 * penetration
            pen_width = self._fuzz_ratio * rb1 if rb1 > 0.0 else 0.0
            mu_pen = self._mu_lt_raw(c3, pen_threshold, pen_width)
            conf_bear = t_product_all(mu_long1, mu_doji2, mu_short3, mu_pen)
            bear_signal = -conf_bear * 100.0

    # Bullish: black-doji-white, gap down then gap up.
    bull_signal = 0.0
    if is_black(o1, c1) and is_white(o3, c3):
        if is_high_low_gap_down(l1, h2) and is_high_low_gap_up(h2, l3):
            rb1 = real_body(o1, c1)
            pen_threshold = c1 + rb1 * penetration
            pen_width = self._fuzz_ratio * rb1 if rb1 > 0.0 else 0.0
            mu_pen = self._mu_gt_raw(c3, pen_threshold, pen_width)
            conf_bull = t_product_all(mu_long1, mu_doji2, mu_short3, mu_pen)
            bull_signal = conf_bull * 100.0

    if abs(bull_signal) >= abs(bear_signal):
        return bull_signal
    return bear_signal
