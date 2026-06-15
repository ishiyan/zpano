"""Instantaneous Sine Wave Period output enum."""

from enum import IntEnum


class InstantaneousSineWavePeriodOutput(IntEnum):
    """Describes the outputs of the indicator."""

    PERIOD = 0
    """The estimated cycle period in bars (may be NaN)."""

    OMEGA = 1
    """The circular frequency in radians/bar (may be NaN)."""

    VELOCITY = 2
    """The wave velocity (may be NaN)."""

    ACCELERATION = 3
    """The wave acceleration (may be NaN)."""

    AMPLITUDE = 4
    """The estimated sine wave amplitude (may be NaN)."""

    PHASE = 5
    """The phase angle in radians (may be NaN)."""

    DC_LEVEL = 6
    """The constant level D (may be NaN)."""
