from dataclasses import dataclass


@dataclass
class NormalizedAverageTrueRangeParams:
    """Parameters for the Normalized Average True Range indicator."""
    length: int = 14
    """The number of time periods. Must be >= 1. The default value is 14."""


def default_params() -> NormalizedAverageTrueRangeParams:
    """Returns default parameters."""
    return NormalizedAverageTrueRangeParams()
