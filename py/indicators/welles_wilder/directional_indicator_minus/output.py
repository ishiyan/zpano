from enum import IntEnum


class DirectionalIndicatorMinusOutput(IntEnum):
    """Output of the Directional Indicator Minus indicator."""
    VALUE = 0
    """The Directional Indicator Minus (-DI) value."""

    DIRECTIONAL_MOVEMENT_MINUS = 1
    """The Directional Movement Minus (-DM) value."""

    AVERAGE_TRUE_RANGE = 2
    """The Average True Range (ATR) value."""

    TRUE_RANGE = 3
    """The True Range (TR) value."""
