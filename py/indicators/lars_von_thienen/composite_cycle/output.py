from enum import IntEnum


class CompositeCycleOutput(IntEnum):
    """Describes the output components."""
    VALUE = 0
    """Value."""
    SLOPE = 1
    """Slope."""
    NEXT_VALUE = 2
    """Next value."""
