from enum import IntEnum


class CycleSwingMomentumOutput(IntEnum):
    """Describes the output components."""
    VALUE = 0
    """Value."""
    LOWER = 1
    """Lower."""
    UPPER = 2
    """Upper."""
    SCORE = 3
    """Score."""
    BAND = 4
    """Band."""
