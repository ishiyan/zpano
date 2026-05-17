"""Fractional Bands output enum."""

from enum import IntEnum


class FractionalBandsOutput(IntEnum):
    """Enumerates the outputs of the fractional bands indicator."""

    FRASMA2 = 0
    UPPER = 1
    LOWER = 2
    BAND = 3
