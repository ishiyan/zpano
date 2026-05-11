"""Tasuki Gap pattern (3-candle continuation)."""
from __future__ import annotations

from ..core.primitives import (
    is_white, is_black, real_body,
    is_real_body_gap_up, is_real_body_gap_down,
)
from ...fuzzy import t_product_all


def tasuki_gap(self) -> float:
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

    Category C: both branches evaluated, return stronger signal.

    Returns:
        Continuous float in [-100, +100].
    """
    if not self._enough(3, self._near):
        return 0.0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    near2 = self._avg(self._near, 2)

    # Upside Tasuki Gap (bullish).
    bull_signal = 0.0
    if (is_real_body_gap_up(o1, c1, o2, c2) and
            is_white(o2, c2) and is_black(o3, c3)):
        rb2 = real_body(o2, c2)
        rb3 = real_body(o3, c3)
        width = self._fuzz_ratio * rb2 if rb2 > 0.0 else 0.0
        # o3 within 2nd body: o3 < c2 and o3 > o2
        mu_o3_lt_c2 = self._mu_lt_raw(o3, c2, width)
        mu_o3_gt_o2 = self._mu_gt_raw(o3, o2, width)
        # c3 below o2
        mu_c3_lt_o2 = self._mu_lt_raw(c3, o2, width)
        # c3 above 1st body top (inside gap)
        body1_top = max(c1, o1)
        mu_c3_gt_top1 = self._mu_gt_raw(c3, body1_top, width)
        # near-equal bodies
        mu_near = self._mu_less(abs(rb2 - rb3), self._near, 2)
        conf = t_product_all(mu_o3_lt_c2, mu_o3_gt_o2, mu_c3_lt_o2,
                             mu_c3_gt_top1, mu_near)
        bull_signal = conf * 100.0

    # Downside Tasuki Gap (bearish).
    bear_signal = 0.0
    if (is_real_body_gap_down(o1, c1, o2, c2) and
            is_black(o2, c2) and is_white(o3, c3)):
        rb2 = real_body(o2, c2)
        rb3 = real_body(o3, c3)
        width = self._fuzz_ratio * rb2 if rb2 > 0.0 else 0.0
        # o3 within 2nd body: o3 < o2 and o3 > c2
        mu_o3_lt_o2 = self._mu_lt_raw(o3, o2, width)
        mu_o3_gt_c2 = self._mu_gt_raw(o3, c2, width)
        # c3 above o2
        mu_c3_gt_o2 = self._mu_gt_raw(c3, o2, width)
        # c3 below 1st body bottom (inside gap)
        body1_bot = min(c1, o1)
        mu_c3_lt_bot1 = self._mu_lt_raw(c3, body1_bot, width)
        # near-equal bodies
        mu_near = self._mu_less(abs(rb2 - rb3), self._near, 2)
        conf = t_product_all(mu_o3_lt_o2, mu_o3_gt_c2, mu_c3_gt_o2,
                             mu_c3_lt_bot1, mu_near)
        bear_signal = -conf * 100.0

    if abs(bull_signal) >= abs(bear_signal):
        return bull_signal
    return bear_signal
