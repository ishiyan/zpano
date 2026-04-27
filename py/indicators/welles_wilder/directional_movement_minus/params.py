from dataclasses import dataclass


@dataclass
class DirectionalMovementMinusParams:
    """Parameters for the Directional Movement Minus indicator."""
    length: int = 14


def default_params() -> DirectionalMovementMinusParams:
    """Returns default parameters."""
    return DirectionalMovementMinusParams()
