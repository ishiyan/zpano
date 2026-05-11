"""Histogram sign-change signals.

Fuzzy membership for oscillator histograms turning positive or negative
(e.g., MACD histogram crossing the zero line).
"""
from __future__ import annotations

from ..fuzzy import MembershipShape, mu_greater, mu_less, t_product


def mu_turns_positive(prev_value: float, curr_value: float,
                      width: float = 0.0,
                      shape: MembershipShape = MembershipShape.SIGMOID) -> float:
    """Degree to which a histogram turned from non-positive to positive.

    Equivalent to ``mu_crosses_above(prev, curr, threshold=0, width)``.

    Args:
        prev_value: Histogram value at the previous bar.
        curr_value: Histogram value at the current bar.
        width: Fuzzy transition width around zero.
        shape: ``MembershipShape.SIGMOID`` or ``MembershipShape.LINEAR``.

    Returns:
        Membership degree ∈ [0, 1].
    """
    was_nonpositive = mu_less(prev_value, 0.0, width, shape)
    is_positive = mu_greater(curr_value, 0.0, width, shape)
    return t_product(was_nonpositive, is_positive)


def mu_turns_negative(prev_value: float, curr_value: float,
                      width: float = 0.0,
                      shape: MembershipShape = MembershipShape.SIGMOID) -> float:
    """Degree to which a histogram turned from non-negative to negative.

    Equivalent to ``mu_crosses_below(prev, curr, threshold=0, width)``.

    Args:
        prev_value: Histogram value at the previous bar.
        curr_value: Histogram value at the current bar.
        width: Fuzzy transition width around zero.
        shape: ``MembershipShape.SIGMOID`` or ``MembershipShape.LINEAR``.

    Returns:
        Membership degree ∈ [0, 1].
    """
    was_nonnegative = mu_greater(prev_value, 0.0, width, shape)
    is_negative = mu_less(curr_value, 0.0, width, shape)
    return t_product(was_nonnegative, is_negative)
