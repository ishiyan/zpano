from dataclasses import dataclass


@dataclass
class DirectionalIndicatorMinusParams:
    """Parameters for the Directional Indicator Minus indicator."""
    length: int = 14


def default_params() -> DirectionalIndicatorMinusParams:
    """Returns default parameters."""
    return DirectionalIndicatorMinusParams()
