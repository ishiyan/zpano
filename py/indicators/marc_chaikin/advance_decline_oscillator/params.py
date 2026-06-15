"""Parameters for the Advance-Decline Oscillator."""

from dataclasses import dataclass
from enum import IntEnum


class MovingAverageType(IntEnum):
    """Moving average type for ADOSC."""

    SMA = 0
    """Simple Moving Average."""

    EMA = 1
    """Exponential Moving Average."""


@dataclass
class AdvanceDeclineOscillatorParams:
    """Parameters for Advance-Decline Oscillator.

    fast_length: fast MA period (default 3, must be >= 2).
    slow_length: slow MA period (default 10, must be >= 2).
    moving_average_type: SMA (0) or EMA (1). Default SMA.
    first_is_average: EMA seeding (True = SMA seed, False = first value). Default False.
    """

    fast_length: int = 3
    """The number of periods for the fast moving average.

    The value should be greater than 1.
    """

    slow_length: int = 10
    """The number of periods for the slow moving average.

    The value should be greater than 1.
    """

    moving_average_type: MovingAverageType = MovingAverageType.SMA
    """The type of moving average (SMA or EMA).

    If not set, the Exponential Moving Average is used.
    """

    first_is_average: bool = False
    """Controls the EMA seeding algorithm.
    When true, the first EMA value is the simple average of the first period values.
    When false (default), the first input value is used directly (Metastock style).
    Only relevant when movingAverageType is EMA.
    """


def default_params() -> AdvanceDeclineOscillatorParams:
    """Return default Advance-Decline Oscillator parameters."""
    return AdvanceDeclineOscillatorParams()
