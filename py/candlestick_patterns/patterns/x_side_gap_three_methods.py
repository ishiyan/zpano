"""Up/Down-side Gap Three Methods pattern (3-candle continuation)."""
from __future__ import annotations

from ..core.primitives import (
    is_white, is_black, real_body,
    is_real_body_gap_up, is_real_body_gap_down,
)
from ...fuzzy import t_product_all


def x_side_gap_three_methods(self) -> float:
    """Up/Down-side Gap Three Methods: a three-candle continuation pattern.

    Must have:
    - first and second candles are the same color with a gap between them,
    - third candle is opposite color, opens within the second candle's
      real body and closes within the first candle's real body (fills the
      gap).

    Upside gap: two white candles with gap up, third is black = bullish.
    Downside gap: two black candles with gap down, third is white = bearish.

    Category C: both branches evaluated, return stronger signal.

    Returns:
        Continuous float in [-100, +100].
    """
    if not self._enough(3):
        return 0.0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    # Upside gap: two whites gap up, third black fills.
    bull_signal = 0.0
    if (is_white(o1, c1) and is_white(o2, c2) and is_black(o3, c3) and
            is_real_body_gap_up(o1, c1, o2, c2)):
        rb2 = real_body(o2, c2)
        width = self._fuzz_ratio * rb2 if rb2 > 0.0 else 0.0
        # o3 within 2nd body: o3 < c2 and o3 > o2
        mu_o3_lt_c2 = self._mu_lt_raw(o3, c2, width)
        mu_o3_gt_o2 = self._mu_gt_raw(o3, o2, width)
        # c3 within 1st body: c3 > o1 and c3 < c1
        rb1 = real_body(o1, c1)
        width1 = self._fuzz_ratio * rb1 if rb1 > 0.0 else 0.0
        mu_c3_gt_o1 = self._mu_gt_raw(c3, o1, width1)
        mu_c3_lt_c1 = self._mu_lt_raw(c3, c1, width1)
        conf = t_product_all(mu_o3_lt_c2, mu_o3_gt_o2, mu_c3_gt_o1, mu_c3_lt_c1)
        bull_signal = conf * 100.0

    # Downside gap: two blacks gap down, third white fills.
    bear_signal = 0.0
    if (is_black(o1, c1) and is_black(o2, c2) and is_white(o3, c3) and
            is_real_body_gap_down(o1, c1, o2, c2)):
        rb2 = real_body(o2, c2)
        width = self._fuzz_ratio * rb2 if rb2 > 0.0 else 0.0
        # o3 within 2nd body: o3 > c2 and o3 < o2
        mu_o3_gt_c2 = self._mu_gt_raw(o3, c2, width)
        mu_o3_lt_o2 = self._mu_lt_raw(o3, o2, width)
        # c3 within 1st body: c3 < o1 and c3 > c1
        rb1 = real_body(o1, c1)
        width1 = self._fuzz_ratio * rb1 if rb1 > 0.0 else 0.0
        mu_c3_lt_o1 = self._mu_lt_raw(c3, o1, width1)
        mu_c3_gt_c1 = self._mu_gt_raw(c3, c1, width1)
        conf = t_product_all(mu_o3_gt_c2, mu_o3_lt_o2, mu_c3_lt_o1, mu_c3_gt_c1)
        bear_signal = -conf * 100.0

    if abs(bull_signal) >= abs(bear_signal):
        return bull_signal
    return bear_signal
