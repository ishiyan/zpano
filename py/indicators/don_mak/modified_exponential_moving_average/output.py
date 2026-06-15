"""Modified Exponential Moving Average output enum."""

from enum import IntEnum


class ModifiedExponentialMovingAverageOutput(IntEnum):
    """Describes the outputs of the indicator."""

    VALUE = 0
    """The velocity-corrected EMA value."""
