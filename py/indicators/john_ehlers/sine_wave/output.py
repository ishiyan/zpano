"""SineWave output enum."""

from enum import IntEnum


class SineWaveOutput(IntEnum):
    """Output describes the outputs of the SineWave indicator."""
    VALUE = 0
    LEAD = 1
    BAND = 2
    DOMINANT_CYCLE_PERIOD = 3
    DOMINANT_CYCLE_PHASE = 4
