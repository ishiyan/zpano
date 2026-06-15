"""Moving Mini-Max output enum."""

from enum import IntEnum


class MovingMiniMaxOutput(IntEnum):
    """Describes the outputs of the indicator."""

    UP = 0
    """The up mini-max value at the most recent bar (emphasizes local maxima)."""

    DOWN = 1
    """The down mini-max value at the most recent bar (emphasizes local minima)."""

    RESISTANCES = 2
    """The detected resistance levels, sorted by strength (strongest first)."""

    SUPPORTS = 3
    """The detected support levels, sorted by strength (strongest first)."""

    UP_DISTRIBUTION = 4
    """The full up mini-max probability distribution over the window (sums to 1.0)."""

    DOWN_DISTRIBUTION = 5
    """The full down mini-max probability distribution over the window (sums to 1.0)."""
