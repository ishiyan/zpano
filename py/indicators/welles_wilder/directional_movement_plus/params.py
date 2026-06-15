from dataclasses import dataclass


@dataclass
class DirectionalMovementPlusParams:
    """Parameters for the Directional Movement Plus indicator."""
    length: int = 14
    """The smoothing length (the number of time periods). Must be >= 1. The default value is 14. A length of 1 means no smoothing."""


def default_params() -> DirectionalMovementPlusParams:
    """Returns default parameters."""
    return DirectionalMovementPlusParams()
