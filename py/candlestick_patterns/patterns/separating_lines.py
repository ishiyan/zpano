"""Separating Lines pattern (2-candle continuation)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body, upper_shadow, lower_shadow
from ...fuzzy import t_product_all


def separating_lines(self) -> float:
    """Separating Lines: a two-candle continuation pattern.

    Opposite colors with the same open. The second candle is a belt hold
    (long body with no shadow on the opening side).

    - bullish: first candle is black, second is white with same open,
      long body, very short lower shadow,
    - bearish: first candle is white, second is black with same open,
      long body, very short upper shadow.

    The meaning of "long" is specified with self._long_body.
    The meaning of "very short" for shadows is specified with
    self._very_short_shadow.
    The meaning of "equal" is specified with self._equal.

    Category C: both branches evaluated, return stronger signal.

    Returns:
        Continuous float in [-100, +100].
    """
    if not self._enough(2, self._long_body, self._very_short_shadow,
                        self._equal):
        return 0.0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    # Opposite colors — crisp gate.
    color1 = 1 if c1 >= o1 else -1
    color2 = 1 if c2 >= o2 else -1
    if color1 == color2:
        return 0.0

    # Opens near equal — fuzzy (crisp was abs(o2-o1) <= eq).
    mu_eq = self._mu_less(abs(o2 - o1), self._equal, 2)

    # Long body on 2nd candle — fuzzy.
    mu_long = self._mu_greater(real_body(o2, c2), self._long_body, 1)

    # Bullish: white belt hold (very short lower shadow).
    bull_signal = 0.0
    if color2 == 1:
        mu_vs = self._mu_less(lower_shadow(o2, l2, c2), self._very_short_shadow, 1)
        conf = t_product_all(mu_eq, mu_long, mu_vs)
        bull_signal = conf * 100.0

    # Bearish: black belt hold (very short upper shadow).
    bear_signal = 0.0
    if color2 == -1:
        mu_vs = self._mu_less(upper_shadow(o2, h2, c2), self._very_short_shadow, 1)
        conf = t_product_all(mu_eq, mu_long, mu_vs)
        bear_signal = -conf * 100.0

    if abs(bull_signal) >= abs(bear_signal):
        return bull_signal
    return bear_signal
