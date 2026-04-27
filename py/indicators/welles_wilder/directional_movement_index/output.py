"""Output enum for the Directional Movement Index indicator."""

from enum import IntEnum


class DirectionalMovementIndexOutput(IntEnum):
    """Describes the outputs of the Directional Movement Index indicator."""

    VALUE = 0
    DIRECTIONAL_INDICATOR_PLUS = 1
    DIRECTIONAL_INDICATOR_MINUS = 2
    DIRECTIONAL_MOVEMENT_PLUS = 3
    DIRECTIONAL_MOVEMENT_MINUS = 4
    AVERAGE_TRUE_RANGE = 5
    TRUE_RANGE = 6
