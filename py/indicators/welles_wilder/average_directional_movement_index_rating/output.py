"""Output enum for the Average Directional Movement Index Rating indicator."""

from enum import IntEnum


class AverageDirectionalMovementIndexRatingOutput(IntEnum):
    """Describes the outputs of the Average Directional Movement Index Rating indicator."""

    VALUE = 0
    AVERAGE_DIRECTIONAL_MOVEMENT_INDEX = 1
    DIRECTIONAL_MOVEMENT_INDEX = 2
    DIRECTIONAL_INDICATOR_PLUS = 3
    DIRECTIONAL_INDICATOR_MINUS = 4
    DIRECTIONAL_MOVEMENT_PLUS = 5
    DIRECTIONAL_MOVEMENT_MINUS = 6
    AVERAGE_TRUE_RANGE = 7
    TRUE_RANGE = 8
