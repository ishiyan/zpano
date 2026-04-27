"""TrendCycleMode output enum."""

from enum import IntEnum


class TrendCycleModeOutput(IntEnum):
    """Output indices for the TrendCycleMode indicator."""
    VALUE = 0
    IS_TREND_MODE = 1
    IS_CYCLE_MODE = 2
    INSTANTANEOUS_TREND_LINE = 3
    SINE_WAVE = 4
    SINE_WAVE_LEAD = 5
    DOMINANT_CYCLE_PERIOD = 6
    DOMINANT_CYCLE_PHASE = 7
