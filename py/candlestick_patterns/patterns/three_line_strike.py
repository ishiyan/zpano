"""Three-Line Strike pattern (4-candle)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black
from ...fuzzy import t_product_all


def three_line_strike(self) -> float:
    """Three-Line Strike: a four-candle pattern.

    Bullish: three white candles with rising closes, each opening within/near
    the prior body, 4th black opens above 3rd close and closes below 1st open.

    Bearish: three black candles with falling closes, each opening within/near
    the prior body, 4th white opens below 3rd close and closes above 1st open.

    Category C: both branches evaluated, return stronger signal.

    Returns:
        Continuous float in [-100, +100].
    """
    if not self._enough(4, self._near):
        return 0.0

    o1, h1, l1, c1 = self._bar(4)
    o2, h2, l2, c2 = self._bar(3)
    o3, h3, l3, c3 = self._bar(2)
    o4, h4, l4, c4 = self._bar(1)

    # Three same color — crisp gate.
    color1 = 1 if is_white(o1, c1) else -1
    color2 = 1 if is_white(o2, c2) else -1
    color3 = 1 if is_white(o3, c3) else -1
    color4 = 1 if is_white(o4, c4) else -1

    if not (color1 == color2 and color2 == color3 and color4 == -color3):
        return 0.0

    # 2nd opens within/near 1st real body — fuzzy.
    near4 = self._avg(self._near, 4)
    near3 = self._avg(self._near, 3)
    near_width4 = self._fuzz_ratio * near4 if near4 > 0.0 else 0.0
    near_width3 = self._fuzz_ratio * near3 if near3 > 0.0 else 0.0

    mu_o2_ge = self._mu_ge_raw(o2, min(o1, c1) - near4, near_width4)
    mu_o2_le = self._mu_lt_raw(o2, max(o1, c1) + near4, near_width4)

    # 3rd opens within/near 2nd real body — fuzzy.
    mu_o3_ge = self._mu_ge_raw(o3, min(o2, c2) - near3, near_width3)
    mu_o3_le = self._mu_lt_raw(o3, max(o2, c2) + near3, near_width3)

    # Bullish: three white, rising closes, 4th opens above 3rd close, closes below 1st open.
    bull_signal = 0.0
    if color3 == 1 and c3 > c2 and c2 > c1:
        rb1 = abs(c1 - o1)
        width = self._fuzz_ratio * rb1 if rb1 > 0.0 else 0.0
        mu_o4_above = self._mu_gt_raw(o4, c3, width)
        mu_c4_below = self._mu_lt_raw(c4, o1, width)
        conf = t_product_all(mu_o2_ge, mu_o2_le, mu_o3_ge, mu_o3_le,
                             mu_o4_above, mu_c4_below)
        bull_signal = conf * 100.0

    # Bearish: three black, falling closes, 4th opens below 3rd close, closes above 1st open.
    bear_signal = 0.0
    if color3 == -1 and c3 < c2 and c2 < c1:
        rb1 = abs(c1 - o1)
        width = self._fuzz_ratio * rb1 if rb1 > 0.0 else 0.0
        mu_o4_below = self._mu_lt_raw(o4, c3, width)
        mu_c4_above = self._mu_gt_raw(c4, o1, width)
        conf = t_product_all(mu_o2_ge, mu_o2_le, mu_o3_ge, mu_o3_le,
                             mu_o4_below, mu_c4_above)
        bear_signal = -conf * 100.0

    if abs(bull_signal) >= abs(bear_signal):
        return bull_signal
    return bear_signal
