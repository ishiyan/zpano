"""Three Outside Up/Down pattern (3-candle reversal)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body
from ...fuzzy import t_product_all


def three_outside(self) -> float:
    """Three Outside Up/Down: a three-candle reversal pattern.

    Must have:
    - first and second candles form an engulfing pattern,
    - third candle confirms the direction by closing higher (up) or
      lower (down).

    Three Outside Up: first candle is black, second is white engulfing
    the first, third closes higher than the second.

    Three Outside Down: first candle is white, second is black engulfing
    the first, third closes lower than the second.

    Category C: both branches evaluated, return stronger signal.

    Returns:
        Continuous float in [-100, +100].
    """
    if not self._enough(3):
        return 0.0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    # Fuzzy engulfment width.
    eq_avg = self._avg(self._equal, 1) if hasattr(self, '_equal') else 0.0
    eq_width = self._fuzz_ratio * eq_avg if eq_avg > 0.0 else 0.0

    # Three Outside Up: black + white engulfing + 3rd closes higher.
    bull_signal = 0.0
    if is_black(o1, c1) and is_white(o2, c2):
        mu_enc_upper = self._mu_ge_raw(max(o2, c2), max(o1, c1), eq_width)
        mu_enc_lower = self._mu_lt_raw(min(o2, c2), min(o1, c1), eq_width)
        rb2 = real_body(o2, c2)
        width = self._fuzz_ratio * rb2 if rb2 > 0.0 else 0.0
        mu_close_higher = self._mu_gt_raw(c3, c2, width)
        conf = t_product_all(mu_enc_upper, mu_enc_lower, mu_close_higher)
        bull_signal = conf * 100.0

    # Three Outside Down: white + black engulfing + 3rd closes lower.
    bear_signal = 0.0
    if is_white(o1, c1) and is_black(o2, c2):
        mu_enc_upper = self._mu_ge_raw(max(o2, c2), max(o1, c1), eq_width)
        mu_enc_lower = self._mu_lt_raw(min(o2, c2), min(o1, c1), eq_width)
        rb2 = real_body(o2, c2)
        width = self._fuzz_ratio * rb2 if rb2 > 0.0 else 0.0
        mu_close_lower = self._mu_lt_raw(c3, c2, width)
        conf = t_product_all(mu_enc_upper, mu_enc_lower, mu_close_lower)
        bear_signal = -conf * 100.0

    if abs(bull_signal) >= abs(bear_signal):
        return bull_signal
    return bear_signal
