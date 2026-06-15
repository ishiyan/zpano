"""Output enum for the Aroon indicator."""

from enum import IntEnum


class AroonOutput(IntEnum):
    """Describes the outputs of the Aroon indicator."""

    UP = 0
    """The Aroon Up line."""

    DOWN = 1
    """The Aroon Down line."""

    OSC = 2
    """The Aroon Oscillator (AroonUp - AroonDown)."""
