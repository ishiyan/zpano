"""Output enum for the Average Directional Movement Index indicator."""

from enum import IntEnum


class AverageDirectionalMovementIndexOutput(IntEnum):
    """Describes the outputs of the Average Directional Movement Index indicator."""

    VALUE = 0
    DIRECTIONAL_MOVEMENT_INDEX = 1
    DIRECTIONAL_INDICATOR_PLUS = 2
    DIRECTIONAL_INDICATOR_MINUS = 3
    DIRECTIONAL_MOVEMENT_PLUS = 4
    DIRECTIONAL_MOVEMENT_MINUS = 5
    AVERAGE_TRUE_RANGE = 6
    TRUE_RANGE = 7
