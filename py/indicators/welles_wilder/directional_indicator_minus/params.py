from dataclasses import dataclass


@dataclass
class DirectionalIndicatorMinusParams:
    """Parameters for the Directional Indicator Minus indicator."""
    length: int = 14
    """The smoothing length (the number of time periods). Must be >= 1. The default value is 14."""


def default_params() -> DirectionalIndicatorMinusParams:
    """Returns default parameters."""
    return DirectionalIndicatorMinusParams()
