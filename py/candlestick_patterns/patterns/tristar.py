"""Tristar pattern (3-candle reversal)."""
from __future__ import annotations

from ..core.primitives import (
    real_body, is_real_body_gap_up, is_real_body_gap_down,
)
from ...fuzzy import t_product_all


def tristar(self) -> float:
    """Tristar: a three-candle reversal pattern with three dojis.

    Must have:
    - three consecutive doji candles,
    - if the second doji gaps up from the first and the third does not
      close higher than the second: bearish,
    - if the second doji gaps down from the first and the third does not
      close lower than the second: bullish.

    Category A: fixed direction per branch (bullish or bearish).

    Returns:
        Continuous float in [-100, +100].
    """
    if not self._enough(3, self._doji_body):
        return 0.0

    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    # Fuzzy: all three must be dojis.
    mu_doji1 = self._mu_less(real_body(o1, c1), self._doji_body, 3)
    mu_doji2 = self._mu_less(real_body(o2, c2), self._doji_body, 2)
    mu_doji3 = self._mu_less(real_body(o3, c3), self._doji_body, 1)

    # Bearish: second gaps up, third is not higher than second — crisp direction checks.
    if (is_real_body_gap_up(o1, c1, o2, c2) and
            max(o3, c3) < max(o2, c2)):
        conf = t_product_all(mu_doji1, mu_doji2, mu_doji3)
        return -conf * 100.0

    # Bullish: second gaps down, third is not lower than second.
    if (is_real_body_gap_down(o1, c1, o2, c2) and
            min(o3, c3) > min(o2, c2)):
        conf = t_product_all(mu_doji1, mu_doji2, mu_doji3)
        return conf * 100.0

    return 0.0
