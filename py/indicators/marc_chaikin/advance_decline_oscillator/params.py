"""Parameters for the Advance-Decline Oscillator."""

from dataclasses import dataclass
from enum import IntEnum


class MovingAverageType(IntEnum):
    """Moving average type for ADOSC."""

    SMA = 0
    EMA = 1


@dataclass
class AdvanceDeclineOscillatorParams:
    """Parameters for Advance-Decline Oscillator.

    fast_length: fast MA period (default 3, must be >= 2).
    slow_length: slow MA period (default 10, must be >= 2).
    moving_average_type: SMA (0) or EMA (1). Default SMA.
    first_is_average: EMA seeding (True = SMA seed, False = first value). Default False.
    """

    fast_length: int = 3
    slow_length: int = 10
    moving_average_type: MovingAverageType = MovingAverageType.SMA
    first_is_average: bool = False


def default_params() -> AdvanceDeclineOscillatorParams:
    """Return default Advance-Decline Oscillator parameters."""
    return AdvanceDeclineOscillatorParams()
