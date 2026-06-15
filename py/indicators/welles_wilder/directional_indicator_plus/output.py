from enum import IntEnum


class DirectionalIndicatorPlusOutput(IntEnum):
    """Output of the Directional Indicator Plus indicator."""
    VALUE = 0
    """The Directional Indicator Plus (+DI) value."""

    DIRECTIONAL_MOVEMENT_PLUS = 1
    """The Directional Movement Plus (+DM) value."""

    AVERAGE_TRUE_RANGE = 2
    """The Average True Range (ATR) value."""

    TRUE_RANGE = 3
    """The True Range (TR) value."""
