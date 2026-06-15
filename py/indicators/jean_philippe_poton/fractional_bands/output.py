"""Fractional Bands output enum."""

from enum import IntEnum


class FractionalBandsOutput(IntEnum):
    """Enumerates the outputs of the fractional bands indicator."""

    FRASMA2 = 0
    """FRASMA2 center line."""

    UPPER = 1
    """Upper band."""

    LOWER = 2
    """Lower band."""

    BAND = 3
    """Band (lower/upper pair)."""
