from enum import IntEnum


class DirectionalIndicatorMinusOutput(IntEnum):
    """Output of the Directional Indicator Minus indicator."""
    VALUE = 0
    DIRECTIONAL_MOVEMENT_MINUS = 1
    AVERAGE_TRUE_RANGE = 2
    TRUE_RANGE = 3
