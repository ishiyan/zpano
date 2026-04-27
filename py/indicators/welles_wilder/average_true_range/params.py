from dataclasses import dataclass


@dataclass
class AverageTrueRangeParams:
    """Parameters for the Average True Range indicator."""
    length: int = 14


def default_params() -> AverageTrueRangeParams:
    """Returns default parameters."""
    return AverageTrueRangeParams()
