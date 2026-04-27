from dataclasses import dataclass


@dataclass
class DirectionalMovementPlusParams:
    """Parameters for the Directional Movement Plus indicator."""
    length: int = 14


def default_params() -> DirectionalMovementPlusParams:
    """Returns default parameters."""
    return DirectionalMovementPlusParams()
