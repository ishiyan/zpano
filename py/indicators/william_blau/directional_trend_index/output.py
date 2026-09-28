"""Directional Trend Index output enum."""

from enum import IntEnum


class DirectionalTrendIndexOutput(IntEnum):
    """Describes the outputs of the indicator."""

    DTI = 0
    """The Directional Trend Index oscillator value (range [-100, +100])."""

    SIGNAL = 1
    """The signal-line value: the ul-period EMA of the oscillator."""
