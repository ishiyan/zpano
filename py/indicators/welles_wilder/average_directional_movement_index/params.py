"""Parameters for the Average Directional Movement Index indicator."""

from dataclasses import dataclass


@dataclass
class AverageDirectionalMovementIndexParams:
    """Parameters for the Average Directional Movement Index indicator."""

    length: int = 14


def default_params() -> AverageDirectionalMovementIndexParams:
    """Returns default parameters."""
    return AverageDirectionalMovementIndexParams()
