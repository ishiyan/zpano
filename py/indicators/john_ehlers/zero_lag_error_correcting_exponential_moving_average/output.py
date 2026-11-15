"""Zero-lag error-correcting exponential moving average output enum."""

from enum import IntEnum


class ZeroLagErrorCorrectingExponentialMovingAverageOutput(IntEnum):
    """Enumerates the outputs of the zero-lag error-correcting exponential moving average indicator."""

    VALUE = 0
    """The scalar value of the error-corrected exponential moving average (EC)."""

    EMA = 1
    """The scalar value of the baseline exponential moving average (EMA)."""
