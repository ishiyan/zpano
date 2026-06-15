"""Output enum for the CoronaSwingPosition indicator."""

from enum import IntEnum


class CoronaSwingPositionOutput(IntEnum):
    """Outputs of the Corona Swing Position indicator."""
    VALUE = 0
    """The Corona swing position heatmap column."""

    SWING_POSITION = 1
    """The current swing position scalar, mapped to [MinParameterValue, MaxParameterValue]."""
