"""Parameters for the Average Directional Movement Index indicator."""

from dataclasses import dataclass


@dataclass
class AverageDirectionalMovementIndexParams:
    """Parameters for the Average Directional Movement Index indicator."""

    length: int = 14
    """The smoothing length (the number of time periods). Must be >= 1. The default value is 14."""


def default_params() -> AverageDirectionalMovementIndexParams:
    """Returns default parameters."""
    return AverageDirectionalMovementIndexParams()
