"""Mesa Adaptive Moving Average output enum."""

from enum import IntEnum


class MesaAdaptiveMovingAverageOutput(IntEnum):
    """Output indices for the Mesa Adaptive Moving Average indicator."""
    VALUE = 0
    FAMA = 1
    BAND = 2
