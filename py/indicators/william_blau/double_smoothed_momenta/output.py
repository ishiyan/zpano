"""Double-Smoothed Momenta output enum."""

from enum import IntEnum


class DoubleSmoothedMomentaOutput(IntEnum):
    """Describes the outputs of the indicator."""

    VALUE = 0
    """The Double-Smoothed Momenta oscillator value (range [0, 100])."""
