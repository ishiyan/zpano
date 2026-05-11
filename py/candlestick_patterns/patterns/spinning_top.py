"""Spinning Top pattern (1-candle)."""
from __future__ import annotations

from ..core.primitives import is_white, is_black, real_body, upper_shadow, lower_shadow
from ...fuzzy import t_product_all


def spinning_top(self) -> float:
    """Spinning Top: a one-candle pattern.

    A candle with a small body and shadows longer than the body on both sides.

    The meaning of "short" is specified with self._short_body.

    Category C: color determines sign.

    Returns:
        Continuous float in [-100, +100].
    """
    if not self._enough(1, self._short_body):
        return 0.0

    o, h, l, c = self._bar(1)

    rb = real_body(o, c)

    mu_short = self._mu_less(rb, self._short_body, 1)

    # Shadows > body: positional comparisons.
    us = upper_shadow(o, h, c)
    ls = lower_shadow(o, l, c)
    width_us = self._fuzz_ratio * rb if rb > 0.0 else 0.0
    width_ls = self._fuzz_ratio * rb if rb > 0.0 else 0.0
    mu_us_gt_rb = self._mu_gt_raw(us, rb, width_us)
    mu_ls_gt_rb = self._mu_gt_raw(ls, rb, width_ls)

    confidence = t_product_all(mu_short, mu_us_gt_rb, mu_ls_gt_rb)

    if is_white(o, c):
        return confidence * 100.0
    if is_black(o, c):
        return -confidence * 100.0
    return 0.0
