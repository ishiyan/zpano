"""Output enum for the Aroon indicator."""

from enum import IntEnum


class AroonOutput(IntEnum):
    """Describes the outputs of the Aroon indicator."""

    UP = 0
    DOWN = 1
    OSC = 2
