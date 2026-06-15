"""Output enum for the Average Directional Movement Index Rating indicator."""

from enum import IntEnum


class AverageDirectionalMovementIndexRatingOutput(IntEnum):
    """Describes the outputs of the Average Directional Movement Index Rating indicator."""

    VALUE = 0
    """The Average Directional Movement Index Rating (ADXR) value."""

    AVERAGE_DIRECTIONAL_MOVEMENT_INDEX = 1
    """The Average Directional Movement Index (ADX) value."""

    DIRECTIONAL_MOVEMENT_INDEX = 2
    """The Directional Movement Index (DX) value."""

    DIRECTIONAL_INDICATOR_PLUS = 3
    """The Directional Indicator Plus (+DI) value."""

    DIRECTIONAL_INDICATOR_MINUS = 4
    """The Directional Indicator Minus (-DI) value."""

    DIRECTIONAL_MOVEMENT_PLUS = 5
    """The Directional Movement Plus (+DM) value."""

    DIRECTIONAL_MOVEMENT_MINUS = 6
    """The Directional Movement Minus (-DM) value."""

    AVERAGE_TRUE_RANGE = 7
    """The Average True Range (ATR) value."""

    TRUE_RANGE = 8
    """The True Range (TR) value."""
