"""Threshold crossing signals.

Fuzzy membership for indicator values relative to fixed thresholds
(e.g., RSI > 70, Stochastic < 20).
"""
from __future__ import annotations

from ..fuzzy import MembershipShape, mu_greater, mu_less


def mu_above(value: float, threshold: float,
             width: float = 5.0, shape: MembershipShape = MembershipShape.SIGMOID) -> float:
    """Degree to which *value* is above *threshold*.

    At ``value == threshold`` the membership is 0.5.  The transition
    zone spans ``threshold ± width/2``.

    Args:
        value: Current indicator value.
        threshold: Level to test against (e.g. 70 for RSI overbought).
        width: Fuzzy transition width.  Larger = softer boundary.
        shape: ``MembershipShape.SIGMOID`` (default) or ``MembershipShape.LINEAR``.

    Returns:
        Membership degree ∈ [0, 1].
    """
    return mu_greater(value, threshold, width, shape)


def mu_below(value: float, threshold: float,
             width: float = 5.0, shape: MembershipShape = MembershipShape.SIGMOID) -> float:
    """Degree to which *value* is below *threshold*.

    Complement of :func:`mu_above`.

    Args:
        value: Current indicator value.
        threshold: Level to test against (e.g. 30 for RSI oversold).
        width: Fuzzy transition width.
        shape: ``MembershipShape.SIGMOID`` or ``MembershipShape.LINEAR``.

    Returns:
        Membership degree ∈ [0, 1].
    """
    return mu_less(value, threshold, width, shape)


def mu_overbought(value: float, level: float = 70.0,
                  width: float = 5.0, shape: MembershipShape = MembershipShape.SIGMOID) -> float:
    """Degree of overbought condition.

    Convenience alias for ``mu_above(value, level, width, shape)``.
    Default level 70 matches common RSI interpretation.
    """
    return mu_greater(value, level, width, shape)


def mu_oversold(value: float, level: float = 30.0,
                width: float = 5.0, shape: MembershipShape = MembershipShape.SIGMOID) -> float:
    """Degree of oversold condition.

    Convenience alias for ``mu_below(value, level, width, shape)``.
    Default level 30 matches common RSI interpretation.
    """
    return mu_less(value, level, width, shape)
