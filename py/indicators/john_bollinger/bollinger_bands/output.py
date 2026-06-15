"""Bollinger Bands output enum."""

from enum import IntEnum


class BollingerBandsOutput(IntEnum):
    """Describes the outputs of the Bollinger Bands indicator."""

    LOWER = 0
    """The lower band value."""

    MIDDLE = 1
    """The middle band (moving average) value."""

    UPPER = 2
    """The upper band value."""

    BAND_WIDTH = 3
    """The band width value."""

    PERCENT_BAND = 4
    """The percent band (%B) value."""

    BAND = 5
    """The lower/upper band."""
