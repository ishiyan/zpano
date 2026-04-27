"""Parameters for the Average Directional Movement Index Rating indicator."""

from dataclasses import dataclass


@dataclass
class AverageDirectionalMovementIndexRatingParams:
    """Parameters for the Average Directional Movement Index Rating indicator."""

    length: int = 14


def default_params() -> AverageDirectionalMovementIndexRatingParams:
    """Returns default parameters."""
    return AverageDirectionalMovementIndexRatingParams()
