from dataclasses import dataclass


@dataclass
class DirectionalIndicatorPlusParams:
    """Parameters for the Directional Indicator Plus indicator."""
    length: int = 14


def default_params() -> DirectionalIndicatorPlusParams:
    """Returns default parameters."""
    return DirectionalIndicatorPlusParams()
