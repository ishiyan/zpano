"""Hurst difference output enum."""

from enum import IntEnum


class HurstDifferenceOutput(IntEnum):
    """Enumerates the outputs of the hurst difference indicator."""

    HURST_DIFF = 0
    FGDI = 1
