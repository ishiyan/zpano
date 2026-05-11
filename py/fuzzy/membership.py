"""Membership functions for fuzzy logic.

Each function maps a crisp value to a membership degree μ ∈ [0, 1].
Two shapes are supported: SIGMOID (default, smooth) and LINEAR (trapezoidal ramp).
All functions degrade to crisp step functions when width = 0.
"""
from __future__ import annotations

import math
from enum import IntEnum


class MembershipShape(IntEnum):
    """Shape of the fuzzy membership transition curve.

    Members:
        SIGMOID: Smooth logistic curve.  Default for most applications.
        LINEAR:  Piecewise-linear ramp (trapezoidal/triangular).
    """
    SIGMOID = 0
    LINEAR = 1

# Steepness constant for sigmoid shape.
# k = _SIGMOID_K / width gives ≈0.997 at threshold ± width/2.
_SIGMOID_K = 12.0


def _sigmoid(x: float, threshold: float, k: float) -> float:
    """Logistic sigmoid: 1 / (1 + exp(k * (x - threshold))).

    Returns the "less-than" membership: high when x << threshold,
    low when x >> threshold, exactly 0.5 at x == threshold.
    """
    exponent = k * (x - threshold)
    # Clamp to avoid overflow in exp().
    if exponent > 500.0:
        return 0.0
    if exponent < -500.0:
        return 1.0
    return 1.0 / (1.0 + math.exp(exponent))


# -----------------------------------------------------------------------
# Core membership functions
# -----------------------------------------------------------------------

def mu_less(x: float, threshold: float, width: float,
            shape: MembershipShape = MembershipShape.SIGMOID) -> float:
    """Degree to which *x* is less than *threshold*.

    At ``threshold``: μ = 0.5.
    At ``threshold - width/2``: μ ≈ 0.997 (sigmoid) or 1.0 (linear).
    At ``threshold + width/2``: μ ≈ 0.003 (sigmoid) or 0.0 (linear).

    When *width* = 0 (crisp): 1.0 if x < threshold, 0.5 if x == threshold,
    0.0 if x > threshold.
    """
    if width <= 0.0:
        if x < threshold:
            return 1.0
        if x > threshold:
            return 0.0
        return 0.5

    if shape == MembershipShape.LINEAR:
        half = width * 0.5
        if x <= threshold - half:
            return 1.0
        if x >= threshold + half:
            return 0.0
        return (threshold + half - x) / width
    else:  # sigmoid
        return _sigmoid(x, threshold, _SIGMOID_K / width)


def mu_less_equal(x: float, threshold: float, width: float,
                  shape: MembershipShape = MembershipShape.SIGMOID) -> float:
    """Degree to which *x* ≤ *threshold*.

    Identical to :func:`mu_less` for continuous values — the distinction
    is conceptual (documenting intent).
    """
    return mu_less(x, threshold, width, shape)


def mu_greater(x: float, threshold: float, width: float,
               shape: MembershipShape = MembershipShape.SIGMOID) -> float:
    """Degree to which *x* > *threshold*.  Complement of :func:`mu_less`."""
    return 1.0 - mu_less(x, threshold, width, shape)


def mu_greater_equal(x: float, threshold: float, width: float,
                     shape: MembershipShape = MembershipShape.SIGMOID) -> float:
    """Degree to which *x* ≥ *threshold*.  Complement of :func:`mu_less_equal`."""
    return 1.0 - mu_less_equal(x, threshold, width, shape)


def mu_near(x: float, target: float, width: float,
            shape: MembershipShape = MembershipShape.SIGMOID) -> float:
    """Bell-shaped membership: degree to which *x* ≈ *target*.

    μ = 1.0 at ``x == target``.
    μ ≈ 0 at ``|x - target| ≥ width``.

    For *sigmoid* shape: Gaussian bell ``exp(-k * (x - target)²)``.
    For *linear* shape: triangular peak at target with base ``2 * width``.
    """
    if width <= 0.0:
        return 1.0 if x == target else 0.0

    if shape == MembershipShape.LINEAR:
        dist = abs(x - target)
        if dist >= width:
            return 0.0
        return 1.0 - dist / width
    else:  # sigmoid → Gaussian bell
        # Gaussian with σ chosen so that μ ≈ 0.003 at |x - target| = width.
        # exp(-(width/σ)²) ≈ 0.003 → (width/σ)² ≈ 5.8 → σ ≈ width / 2.41.
        sigma = width / 2.41
        return math.exp(-((x - target) / sigma) ** 2)


def mu_direction(o: float, c: float, body_avg: float,
                 steepness: float = 2.0) -> float:
    """Fuzzy candle direction ∈ [-1, +1].

    +1 = fully bullish (large white body).
     0 = neutral (doji-like).
    -1 = fully bearish (large black body).

    Uses ``tanh(steepness * (c - o) / body_avg)``.

    When *body_avg* ≤ 0: returns +1.0 if ``c ≥ o``, else -1.0 (crisp).
    """
    if body_avg <= 0.0:
        return 1.0 if c >= o else -1.0
    return math.tanh(steepness * (c - o) / body_avg)
