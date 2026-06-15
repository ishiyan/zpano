"""CenterOfGravityOscillator output enum."""

from enum import IntEnum


class CenterOfGravityOscillatorOutput(IntEnum):
    """Describes the outputs of the CenterOfGravityOscillator indicator."""
    VALUE = 0
    """The scalar value of the center of gravity oscillator."""

    TRIGGER = 1
    """The scalar value of the trigger line (previous value of the oscillator)."""
