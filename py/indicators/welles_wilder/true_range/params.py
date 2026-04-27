from dataclasses import dataclass


@dataclass
class TrueRangeParams:
    """Parameters for the True Range indicator."""
    pass


def default_params() -> TrueRangeParams:
    """Returns default parameters."""
    return TrueRangeParams()
