from enum import IntEnum


class PhasingOscillatorOutput(IntEnum):
    """Describes the output components."""
    VALUE = 0
    """Value."""
    SHORT_SCORE = 1
    """Short score."""
    LONG_SCORE = 2
    """Long score."""
    SHORT_PERIOD = 3
    """Short period."""
    LONG_PERIOD = 4
    """Long period."""
