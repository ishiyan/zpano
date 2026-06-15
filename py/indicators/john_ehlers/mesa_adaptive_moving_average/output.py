"""Mesa Adaptive Moving Average output enum."""

from enum import IntEnum


class MesaAdaptiveMovingAverageOutput(IntEnum):
    """Output indices for the Mesa Adaptive Moving Average indicator."""
    VALUE = 0
    """The scalar value of the MAMA (Mesa Adaptive Moving Average)."""

    FAMA = 1
    """The scalar value of the FAMA (Following Adaptive Moving Average)."""

    BAND = 2
    """The band output, with MAMA as the upper line and FAMA as the lower line."""
