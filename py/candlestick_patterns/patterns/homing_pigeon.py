"""Homing Pigeon pattern (2-candle bullish)."""
from __future__ import annotations

from ..core.primitives import is_black, real_body
from ...fuzzy import t_product_all


def homing_pigeon(self) -> float:
    """Homing Pigeon: a two-candle bullish pattern.

    Must have:
    - first candle: long black,
    - second candle: short black, real body engulfed by first candle's
      real body.

    The meaning of "long" is specified with self._long_body.
    The meaning of "short" is specified with self._short_body.

    Category A: always bullish (continuous).

    Returns:
        Continuous float in [0, 100].  Always bullish.
    """
    if not self._enough(2, self._long_body, self._short_body):
        return 0.0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    # Crisp gates: both black.
    if not (is_black(o1, c1) and is_black(o2, c2)):
        return 0.0

    # Fuzzy conditions.
    mu_long1 = self._mu_greater(real_body(o1, c1), self._long_body, 2)
    mu_short2 = self._mu_less(real_body(o2, c2), self._short_body, 1)

    # Containment: second body engulfed by first body.
    # For black candles: open > close, so upper = open, lower = close.
    eq_width = self._fuzz_ratio * self._avg(self._equal, 2)
    mu_enc_upper = self._mu_lt_raw(o2, o1, eq_width)
    mu_enc_lower = self._mu_gt_raw(c2, c1, eq_width)

    confidence = t_product_all(mu_long1, mu_short2, mu_enc_upper, mu_enc_lower)

    return confidence * 100.0
