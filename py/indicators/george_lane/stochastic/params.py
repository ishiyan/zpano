from dataclasses import dataclass
from enum import IntEnum


class MovingAverageType(IntEnum):
    """Type of moving average for smoothing."""
    SMA = 0
    EMA = 1


@dataclass
class StochasticParams:
    """Parameters for the Stochastic Oscillator indicator."""
    fast_k_length: int = 5
    slow_k_length: int = 3
    slow_d_length: int = 3
    slow_k_ma_type: MovingAverageType = MovingAverageType.SMA
    slow_d_ma_type: MovingAverageType = MovingAverageType.SMA
    first_is_average: bool = False


def default_params() -> StochasticParams:
    """Returns default parameters."""
    return StochasticParams()
