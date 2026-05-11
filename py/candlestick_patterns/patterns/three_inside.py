"""Three Inside Up/Down pattern (3-candle reversal)."""
from __future__ import annotations

from ..core.primitives import (
    is_white, is_black, real_body, white_real_body, black_real_body,
    is_real_body_encloses_real_body,
)
from ...fuzzy import t_product_all


def three_inside(self) -> float:
    """Three Inside Up/Down: a three-candle reversal pattern.

    Three Inside Up (bullish):
    - first candle: long black,
    - second candle: short, engulfed by the first candle's real body,
    - third candle: white, closes above the first candle's open.

    Three Inside Down (bearish):
    - first candle: long white,
    - second candle: short, engulfed by the first candle's real body,
    - third candle: black, closes below the first candle's open.

    The meaning of "long" is specified with self._long_body.
    The meaning of "short" is specified with self._short_body.

    Category C: both branches evaluated, return stronger signal.

    Returns:
        Continuous float in [-100, +100].
    """
    if not self._enough(3, self._long_body, self._short_body):
        return 0.0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    # Shared fuzzy conditions.
    mu_long1 = self._mu_greater(real_body(o1, c1), self._long_body, 3)
    mu_short2 = self._mu_less(real_body(o2, c2), self._short_body, 2)

    # Fuzzy containment: 1st body encloses 2nd body.
    eq_avg = self._avg(self._equal, 2) if hasattr(self, '_equal') else 0.0
    eq_width = self._fuzz_ratio * eq_avg if eq_avg > 0.0 else 0.0
    mu_enc_upper = self._mu_ge_raw(max(o1, c1), max(o2, c2), eq_width)
    mu_enc_lower = self._mu_lt_raw(min(o1, c1), min(o2, c2), eq_width)

    # Three Inside Up: long black, short engulfed, white closes above 1st open.
    bull_signal = 0.0
    if is_black(o1, c1) and is_white(o3, c3):
        width = self._fuzz_ratio * real_body(o1, c1) if real_body(o1, c1) > 0.0 else 0.0
        mu_close_above = self._mu_gt_raw(c3, o1, width)
        conf = t_product_all(mu_long1, mu_short2, mu_enc_upper, mu_enc_lower, mu_close_above)
        bull_signal = conf * 100.0

    # Three Inside Down: long white, short engulfed, black closes below 1st open.
    bear_signal = 0.0
    if is_white(o1, c1) and is_black(o3, c3):
        width = self._fuzz_ratio * real_body(o1, c1) if real_body(o1, c1) > 0.0 else 0.0
        mu_close_below = self._mu_lt_raw(c3, o1, width)
        conf = t_product_all(mu_long1, mu_short2, mu_enc_upper, mu_enc_lower, mu_close_below)
        bear_signal = -conf * 100.0

    if abs(bull_signal) >= abs(bear_signal):
        return bull_signal
    return bear_signal
