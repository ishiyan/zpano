"""TrendCycleMode output enum."""

from enum import IntEnum


class TrendCycleModeOutput(IntEnum):
    """Output indices for the TrendCycleMode indicator."""
    VALUE = 0
    """+1 in trend mode, -1 in cycle mode."""

    IS_TREND_MODE = 1
    """1 if the trend mode is declared, 0 otherwise."""

    IS_CYCLE_MODE = 2
    """1 if the cycle mode is declared, 0 otherwise (= 1 − IsTrendMode)."""

    INSTANTANEOUS_TREND_LINE = 3
    """The WMA-smoothed instantaneous trend line."""

    SINE_WAVE = 4
    """The sine wave value, sin(phase·Deg2Rad)."""

    SINE_WAVE_LEAD = 5
    """The sine wave lead value, sin((phase+45)·Deg2Rad)."""

    DOMINANT_CYCLE_PERIOD = 6
    """The smoothed dominant cycle period."""

    DOMINANT_CYCLE_PHASE = 7
    """The dominant cycle phase, in degrees."""
