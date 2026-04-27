"""Bollinger Bands output enum."""

from enum import IntEnum


class BollingerBandsOutput(IntEnum):
    """Describes the outputs of the Bollinger Bands indicator."""

    LOWER = 0
    MIDDLE = 1
    UPPER = 2
    BAND_WIDTH = 3
    PERCENT_BAND = 4
    BAND = 5
