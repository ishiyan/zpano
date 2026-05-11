"""Crossover signals.

Fuzzy membership for line crossings (e.g., fast MA crossing slow MA)
and threshold crossings (e.g., RSI crossing above 30).

A crossover is the conjunction of "was on one side" and "is now on the
other side".  The fuzzy version replaces the crisp boolean with a
product of two membership degrees.
"""
from __future__ import annotations

from ..fuzzy import MembershipShape, mu_greater, mu_less, t_product


def mu_crosses_above(prev_value: float, curr_value: float,
                     threshold: float, width: float = 0.0,
                     shape: MembershipShape = MembershipShape.SIGMOID) -> float:
    """Degree to which a value crossed above *threshold* from below.

    Computed as ``mu_below(prev, threshold) * mu_above(curr, threshold)``.
    Returns a high value only when the previous value was below the
    threshold AND the current value is above it.

    Args:
        prev_value: Indicator value at the previous bar.
        curr_value: Indicator value at the current bar.
        threshold: Level being crossed.
        width: Fuzzy transition width.  0 = crisp crossover.
        shape: ``MembershipShape.SIGMOID`` or ``MembershipShape.LINEAR``.

    Returns:
        Membership degree ∈ [0, 1].
    """
    was_below = mu_less(prev_value, threshold, width, shape)
    is_above = mu_greater(curr_value, threshold, width, shape)
    return t_product(was_below, is_above)


def mu_crosses_below(prev_value: float, curr_value: float,
                     threshold: float, width: float = 0.0,
                     shape: MembershipShape = MembershipShape.SIGMOID) -> float:
    """Degree to which a value crossed below *threshold* from above.

    Computed as ``mu_above(prev, threshold) * mu_below(curr, threshold)``.

    Args:
        prev_value: Indicator value at the previous bar.
        curr_value: Indicator value at the current bar.
        threshold: Level being crossed.
        width: Fuzzy transition width.
        shape: ``MembershipShape.SIGMOID`` or ``MembershipShape.LINEAR``.

    Returns:
        Membership degree ∈ [0, 1].
    """
    was_above = mu_greater(prev_value, threshold, width, shape)
    is_below = mu_less(curr_value, threshold, width, shape)
    return t_product(was_above, is_below)


def mu_line_crosses_above(prev_fast: float, curr_fast: float,
                          prev_slow: float, curr_slow: float,
                          width: float = 0.0,
                          shape: MembershipShape = MembershipShape.SIGMOID) -> float:
    """Degree to which a fast line crossed above a slow line.

    Reduces to a threshold crossing of the difference
    ``(fast - slow)`` crossing above zero.

    Args:
        prev_fast: Fast line value at the previous bar.
        curr_fast: Fast line value at the current bar.
        prev_slow: Slow line value at the previous bar.
        curr_slow: Slow line value at the current bar.
        width: Fuzzy transition width applied to the difference.
        shape: ``MembershipShape.SIGMOID`` or ``MembershipShape.LINEAR``.

    Returns:
        Membership degree ∈ [0, 1].
    """
    prev_diff = prev_fast - prev_slow
    curr_diff = curr_fast - curr_slow
    return mu_crosses_above(prev_diff, curr_diff, 0.0, width, shape)


def mu_line_crosses_below(prev_fast: float, curr_fast: float,
                          prev_slow: float, curr_slow: float,
                          width: float = 0.0,
                          shape: MembershipShape = MembershipShape.SIGMOID) -> float:
    """Degree to which a fast line crossed below a slow line.

    Args:
        prev_fast: Fast line value at the previous bar.
        curr_fast: Fast line value at the current bar.
        prev_slow: Slow line value at the previous bar.
        curr_slow: Slow line value at the current bar.
        width: Fuzzy transition width applied to the difference.
        shape: ``MembershipShape.SIGMOID`` or ``MembershipShape.LINEAR``.

    Returns:
        Membership degree ∈ [0, 1].
    """
    prev_diff = prev_fast - prev_slow
    curr_diff = curr_fast - curr_slow
    return mu_crosses_below(prev_diff, curr_diff, 0.0, width, shape)
