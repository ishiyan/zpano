"""Output enum for the Parabolic Stop And Reverse indicator."""

from enum import IntEnum


class ParabolicStopAndReverseOutput(IntEnum):
    """Describes the outputs of the Parabolic Stop And Reverse indicator."""

    VALUE = 0
    """The scalar value of the Parabolic Stop And Reverse.
    Positive values indicate a long position; negative values indicate a short position.
    """
