"""Output enum for the Average Directional Movement Index indicator."""

from enum import IntEnum


class AverageDirectionalMovementIndexOutput(IntEnum):
    """Describes the outputs of the Average Directional Movement Index indicator."""

    VALUE = 0
    """The Average Directional Movement Index (ADX) value."""

    DIRECTIONAL_MOVEMENT_INDEX = 1
    """The Directional Movement Index (DX) value."""

    DIRECTIONAL_INDICATOR_PLUS = 2
    """The Directional Indicator Plus (+DI) value."""

    DIRECTIONAL_INDICATOR_MINUS = 3
    """The Directional Indicator Minus (-DI) value."""

    DIRECTIONAL_MOVEMENT_PLUS = 4
    """The Directional Movement Plus (+DM) value."""

    DIRECTIONAL_MOVEMENT_MINUS = 5
    """The Directional Movement Minus (-DM) value."""

    AVERAGE_TRUE_RANGE = 6
    """The Average True Range (ATR) value."""

    TRUE_RANGE = 7
    """The True Range (TR) value."""
