"""Output enum for the Directional Movement Index indicator."""

from enum import IntEnum


class DirectionalMovementIndexOutput(IntEnum):
    """Describes the outputs of the Directional Movement Index indicator."""

    VALUE = 0
    """The Directional Movement Index (DX) value."""

    DIRECTIONAL_INDICATOR_PLUS = 1
    """The Directional Indicator Plus (+DI) value."""

    DIRECTIONAL_INDICATOR_MINUS = 2
    """The Directional Indicator Minus (-DI) value."""

    DIRECTIONAL_MOVEMENT_PLUS = 3
    """The Directional Movement Plus (+DM) value."""

    DIRECTIONAL_MOVEMENT_MINUS = 4
    """The Directional Movement Minus (-DM) value."""

    AVERAGE_TRUE_RANGE = 5
    """The Average True Range (ATR) value."""

    TRUE_RANGE = 6
    """The True Range (TR) value."""
