"""Velocity-Corrected Exponential Moving Average output enum."""

from enum import IntEnum


class VelocityCorrectedExponentialMovingAverageOutput(IntEnum):
    """Describes the outputs of the indicator."""

    VALUE = 0
    """The velocity-corrected EMA value."""
