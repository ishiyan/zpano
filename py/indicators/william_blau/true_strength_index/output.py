"""True Strength Index output enum."""

from enum import IntEnum


class TrueStrengthIndexOutput(IntEnum):
    """Describes the outputs of the indicator."""

    TSI = 0
    """The True Strength Index oscillator value (range [-100, +100])."""

    SIGNAL = 1
    """The signal-line value: the ul-period EMA of the oscillator."""
