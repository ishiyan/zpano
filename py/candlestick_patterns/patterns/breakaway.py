"""Breakaway pattern (5-candle reversal)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body, is_real_body_gap_down, is_real_body_gap_up
from ...fuzzy import t_product_all


def breakaway(self) -> float:
    """Breakaway: a five-candle reversal pattern.

    Bullish: first candle is long black, second candle is black gapping down,
    third and fourth candles have consecutively lower highs and lows, fifth
    candle is white closing into the gap (between first and second candle's
    real bodies).

    Bearish: mirror image with colors reversed and gaps reversed.

    Category C: both branches evaluated, return stronger signal.

    Returns:
        Continuous float in [-100, +100].
    """
    if not self._enough(5, self._long_body):
        return 0.0

    o1, h1, l1, c1 = self._bar(5)
    o2, h2, l2, c2 = self._bar(4)
    o3, h3, l3, c3 = self._bar(3)
    o4, h4, l4, c4 = self._bar(2)
    o5, h5, l5, c5 = self._bar(1)

    # Fuzzy: 1st candle is long.
    mu_long1 = self._mu_greater(real_body(o1, c1), self._long_body, 5)

    # Bullish breakaway.
    bull_signal = 0.0
    if (is_black(o1, c1) and is_black(o2, c2) and
            is_black(o4, c4) and is_white(o5, c5) and
            h3 < h2 and l3 < l2 and
            h4 < h3 and l4 < l3 and
            is_real_body_gap_down(o1, c1, o2, c2)):
        # Fuzzy: c5 > o2 and c5 < c1 (closing into the gap).
        rb1 = real_body(o1, c1)
        width = self._fuzz_ratio * rb1 if rb1 > 0.0 else 0.0
        mu_c5_above_o2 = self._mu_gt_raw(c5, o2, width)
        mu_c5_below_c1 = self._mu_lt_raw(c5, c1, width)
        conf = t_product_all(mu_long1, mu_c5_above_o2, mu_c5_below_c1)
        bull_signal = conf * 100.0

    # Bearish breakaway.
    bear_signal = 0.0
    if (is_white(o1, c1) and is_white(o2, c2) and
            is_white(o4, c4) and is_black(o5, c5) and
            h3 > h2 and l3 > l2 and
            h4 > h3 and l4 > l3 and
            is_real_body_gap_up(o1, c1, o2, c2)):
        rb1 = real_body(o1, c1)
        width = self._fuzz_ratio * rb1 if rb1 > 0.0 else 0.0
        mu_c5_below_o2 = self._mu_lt_raw(c5, o2, width)
        mu_c5_above_c1 = self._mu_gt_raw(c5, c1, width)
        conf = t_product_all(mu_long1, mu_c5_below_o2, mu_c5_above_c1)
        bear_signal = -conf * 100.0

    if abs(bull_signal) >= abs(bear_signal):
        return bull_signal
    return bear_signal


def breakaway_bullish(self) -> float:
    """Convenience: returns only the bullish breakaway signal."""
    result = breakaway(self)
    return result if result > 0 else 0.0


def breakaway_bearish(self) -> float:
    """Convenience: returns only the bearish breakaway signal."""
    result = breakaway(self)
    return result if result < 0 else 0.0
