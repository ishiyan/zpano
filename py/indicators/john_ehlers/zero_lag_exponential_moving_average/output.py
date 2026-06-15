"""Zero-lag exponential moving average output enum."""

from enum import IntEnum


class ZeroLagExponentialMovingAverageOutput(IntEnum):
    """Enumerates the outputs of the zero-lag exponential moving average indicator."""

    VALUE = 0
    """The calculated value of the zero-lag exponential moving average."""
