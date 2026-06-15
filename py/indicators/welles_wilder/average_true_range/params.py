from dataclasses import dataclass


@dataclass
class AverageTrueRangeParams:
    """Parameters for the Average True Range indicator."""
    length: int = 14
    """The number of time periods. Must be >= 1. The default value is 14."""


def default_params() -> AverageTrueRangeParams:
    """Returns default parameters."""
    return AverageTrueRangeParams()
