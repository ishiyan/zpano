from enum import IntEnum


class CyclicSmoothedRelativeStrengthIndexOutput(IntEnum):
    """Describes the output components."""
    VALUE = 0
    """Value."""
    LOWER = 1
    """Lower."""
    UPPER = 2
    """Upper."""
    BAND = 3
    """Band."""
