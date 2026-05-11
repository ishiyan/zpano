"""Closing Marubozu pattern (1-candle)."""
from __future__ import annotations

from ..core.primitives import is_white, real_body, upper_shadow, lower_shadow
from ...fuzzy import t_product_all


def closing_marubozu(self) -> float:
    """Closing Marubozu: a one-candle pattern.

    A long candle with a very short shadow on the closing side:
    - bullish (white): very short upper shadow,
    - bearish (black): very short lower shadow.

    The meaning of "long" is specified with self._long_body.
    The meaning of "very short" for shadows is specified with
    self._very_short_shadow.

    Category C: both branches evaluated, return stronger signal.

    Returns:
        Continuous float in [-100, +100].
    """
    if not self._enough(1, self._long_body, self._very_short_shadow):
        return 0.0

    o, h, l, c = self._bar(1)

    mu_long = self._mu_greater(real_body(o, c), self._long_body, 1)

    # Bullish: white + very short upper shadow.
    bull_signal = 0.0
    if is_white(o, c):
        mu_vs = self._mu_less(upper_shadow(o, h, c), self._very_short_shadow, 1)
        conf = t_product_all(mu_long, mu_vs)
        bull_signal = conf * 100.0

    # Bearish: black (not white) + very short lower shadow.
    bear_signal = 0.0
    if not is_white(o, c):
        mu_vs = self._mu_less(lower_shadow(o, l, c), self._very_short_shadow, 1)
        conf = t_product_all(mu_long, mu_vs)
        bear_signal = -conf * 100.0

    if abs(bull_signal) >= abs(bear_signal):
        return bull_signal
    return bear_signal
