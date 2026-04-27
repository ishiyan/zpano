"""Adaptive Trend and Cycle Filter output enum."""

from enum import IntEnum


class AdaptiveTrendAndCycleFilterOutput(IntEnum):
    """Describes the outputs of the Adaptive Trend and Cycle Filter."""
    FATL = 0
    SATL = 1
    RFTL = 2
    RSTL = 3
    RBCI = 4
    FTLM = 5
    STLM = 6
    PCCI = 7
