from dataclasses import dataclass


@dataclass
class NormalizedAverageTrueRangeParams:
    """Parameters for the Normalized Average True Range indicator."""
    length: int = 14


def default_params() -> NormalizedAverageTrueRangeParams:
    """Returns default parameters."""
    return NormalizedAverageTrueRangeParams()
