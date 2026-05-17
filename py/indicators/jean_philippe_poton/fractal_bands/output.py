"""Fractal Bands output enum."""

from enum import IntEnum


class FractalBandsOutput(IntEnum):
    """Enumerates the outputs of the fractal bands indicator."""

    FRASMA2 = 0
    UPPER = 1
    LOWER = 2
    BAND = 3
