"""Output enum for the CoronaTrendVigor indicator."""

from enum import IntEnum


class CoronaTrendVigorOutput(IntEnum):
    """Outputs of the Corona Trend Vigor indicator."""
    VALUE = 0
    """The Corona trend vigor heatmap column."""

    TREND_VIGOR = 1
    """The current trend vigor scalar, mapped to [MinParameterValue, MaxParameterValue]."""
