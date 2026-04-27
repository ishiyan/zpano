"""Classifies whether an indicator adapts its parameters to market conditions."""

from enum import IntEnum


class Adaptivity(IntEnum):
    """Classifies whether an indicator adapts its parameters to market conditions."""

    # An indicator with fixed parameters.
    STATIC = 0

    # An indicator that adapts parameters to market conditions.
    ADAPTIVE = 1
