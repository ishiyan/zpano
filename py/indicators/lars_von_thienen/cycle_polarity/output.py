from enum import IntEnum


class CyclePolarityOutput(IntEnum):
    """Describes the output components."""
    POLARITY = 0
    """Polarity."""
    SHORT_VALUE = 1
    """Short value."""
    LONG_VALUE = 2
    """Long value."""
    SHORT_SLOPE = 3
    """Short slope."""
    LONG_SLOPE = 4
    """Long slope."""
