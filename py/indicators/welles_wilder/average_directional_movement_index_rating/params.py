"""Parameters for the Average Directional Movement Index Rating indicator."""

from dataclasses import dataclass


@dataclass
class AverageDirectionalMovementIndexRatingParams:
    """Parameters for the Average Directional Movement Index Rating indicator."""

    length: int = 14
    """The smoothing length (the number of time periods). Must be >= 1. The default value is 14."""


def default_params() -> AverageDirectionalMovementIndexRatingParams:
    """Returns default parameters."""
    return AverageDirectionalMovementIndexRatingParams()
