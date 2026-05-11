"""Defuzzification utilities.

Provides alpha-cut conversion from continuous fuzzy output back to crisp
discrete values for backward compatibility with TA-Lib-style integer outputs.
"""
from __future__ import annotations

import math


def alpha_cut(value: float, alpha: float = 0.5,
              scale: float = 100.0) -> int:
    """Convert a continuous fuzzy output to a crisp discrete value.

    The confidence is ``abs(value) / scale``.  If confidence ≥ *alpha*,
    the output is rounded to the nearest multiple of *scale* with the
    original sign preserved.  Otherwise 0 is returned.

    Args:
        value: Continuous fuzzy output (e.g. -87.3, +156.8).
        alpha: Confidence threshold ∈ [0, 1].  Default 0.5.
        scale: Base scale for rounding.  Default 100.0.

    Returns:
        Crisp integer value (e.g. -100, 0, +100, +200).

    Examples:
        >>> alpha_cut(-87.3)
        -100
        >>> alpha_cut(-32.1)
        0
        >>> alpha_cut(156.8)
        200
        >>> alpha_cut(-87.3, alpha=0.9)
        0
    """
    if scale <= 0.0:
        return 0
    confidence = abs(value) / scale
    if confidence < alpha - 1e-10:
        return 0
    sign = 1 if value >= 0 else -1
    # Round to nearest multiple of scale.
    level = max(1, round(confidence))
    return sign * int(level * scale)
