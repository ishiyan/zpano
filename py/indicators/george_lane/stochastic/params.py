from dataclasses import dataclass
from enum import IntEnum


class MovingAverageType(IntEnum):
    """Type of moving average for smoothing."""
    SMA = 0
    """Simple Moving Average."""

    EMA = 1
    """Exponential Moving Average."""


@dataclass
class StochasticParams:
    """Parameters for the Stochastic Oscillator indicator."""
    fast_k_length: int = 5
    """The lookback period for the raw %K calculation (highest high / lowest low).

    The value should be greater than 0.
    """

    slow_k_length: int = 3
    """The smoothing period for Slow-K (also known as Fast-D).

    The value should be greater than 0.
    """

    slow_d_length: int = 3
    """The smoothing period for Slow-D.

    The value should be greater than 0.
    """

    slow_k_ma_type: MovingAverageType = MovingAverageType.SMA
    """The smoothing period for Slow-K (also known as Fast-D).

    The value should be greater than 0.
    """

    slow_d_ma_type: MovingAverageType = MovingAverageType.SMA
    """The smoothing period for Slow-D.

    The value should be greater than 0.
    """

    first_is_average: bool = False
    """Controls the EMA seeding algorithm.
    When true, the first EMA value is the simple average of the first period values.
    When false (default), the first input value is used directly (Metastock style).
    Only relevant when an MA type is EMA.
    """


def default_params() -> StochasticParams:
    """Returns default parameters."""
    return StochasticParams()
