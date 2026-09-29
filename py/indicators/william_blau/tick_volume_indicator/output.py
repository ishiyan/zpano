"""Tick Volume Indicator output enum."""

from enum import IntEnum


class TickVolumeIndicatorOutput(IntEnum):
    """Describes the outputs of the indicator."""

    VALUE = 0
    """The Tick Volume Indicator oscillator value (range [-100, +100])."""
