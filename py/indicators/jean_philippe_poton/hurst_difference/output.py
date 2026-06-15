"""Hurst difference output enum."""

from enum import IntEnum


class HurstDifferenceOutput(IntEnum):
    """Enumerates the outputs of the hurst difference indicator."""

    HURST_DIFF = 0
    """HurstDiff is the first difference of the FGDI."""

    FGDI = 1
    """Fgdi is the raw FGDI value."""
