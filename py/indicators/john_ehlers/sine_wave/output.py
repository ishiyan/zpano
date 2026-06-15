"""SineWave output enum."""

from enum import IntEnum


class SineWaveOutput(IntEnum):
    """Output describes the outputs of the SineWave indicator."""
    VALUE = 0
    """The sine wave value, sin(phase·Deg2Rad)."""

    LEAD = 1
    """The sine wave lead value, sin((phase+45)·Deg2Rad)."""

    BAND = 2
    """The band formed by the sine wave (upper) and the lead sine wave (lower)."""

    DOMINANT_CYCLE_PERIOD = 3
    """The smoothed dominant cycle period."""

    DOMINANT_CYCLE_PHASE = 4
    """The dominant cycle phase, in degrees."""
