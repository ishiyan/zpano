"""Parameters for the Directional Movement Index indicator."""

from dataclasses import dataclass


@dataclass
class DirectionalMovementIndexParams:
    """Parameters for the Directional Movement Index indicator."""

    length: int = 14


def default_params() -> DirectionalMovementIndexParams:
    """Returns default parameters."""
    return DirectionalMovementIndexParams()
