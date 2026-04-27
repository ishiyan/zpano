from enum import IntEnum


class DirectionalIndicatorPlusOutput(IntEnum):
    """Output of the Directional Indicator Plus indicator."""
    VALUE = 0
    DIRECTIONAL_MOVEMENT_PLUS = 1
    AVERAGE_TRUE_RANGE = 2
    TRUE_RANGE = 3
