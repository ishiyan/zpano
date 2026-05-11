"""Harami Cross pattern (2-candle reversal)."""
from __future__ import annotations

from ..core.primitives import real_body
from ...fuzzy import t_product_all


def harami_cross(self) -> float:
    """Harami Cross: a two-candle reversal pattern.

    Like Harami, but the second candle is a doji instead of just short.

    Must have:
    - first candle: long real body,
    - second candle: doji body contained within the first candle's real body.

    Category B: direction from 1st candle color (continuous).

    Returns:
        Continuous float in [-100, +100].
    """
    if not self._enough(2, self._long_body, self._doji_body):
        return 0.0

    o1, h1, l1, c1 = self._bar(2)
    o2, h2, l2, c2 = self._bar(1)

    # Fuzzy size conditions.
    mu_long1 = self._mu_greater(real_body(o1, c1), self._long_body, 2)
    mu_doji2 = self._mu_less(real_body(o2, c2), self._doji_body, 1)

    # Fuzzy containment: 1st body encloses 2nd body.
    eq_avg = self._avg(self._equal, 1)
    eq_width = self._fuzz_ratio * eq_avg if eq_avg > 0.0 else 0.0

    mu_enc_upper = self._mu_ge_raw(max(o1, c1), max(o2, c2), eq_width)
    mu_enc_lower = self._mu_lt_raw(min(o1, c1), min(o2, c2), eq_width)

    confidence = t_product_all(mu_long1, mu_doji2, mu_enc_upper, mu_enc_lower)

    # Direction: opposite of 1st candle color.
    direction = -1.0 if c1 >= o1 else 1.0

    return direction * confidence * 100.0
