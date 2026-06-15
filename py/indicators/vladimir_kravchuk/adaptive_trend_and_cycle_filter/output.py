"""Adaptive Trend and Cycle Filter output enum."""

from enum import IntEnum


class AdaptiveTrendAndCycleFilterOutput(IntEnum):
    """Describes the outputs of the Adaptive Trend and Cycle Filter."""
    FATL = 0
    """Fast Adaptive Trend Line (39-tap FIR)."""

    SATL = 1
    """Slow Adaptive Trend Line (65-tap FIR)."""

    RFTL = 2
    """Reference Fast Trend Line (44-tap FIR)."""

    RSTL = 3
    """Reference Slow Trend Line (91-tap FIR)."""

    RBCI = 4
    """Range Bound Channel Index (56-tap FIR)."""

    FTLM = 5
    """Fast Trend Line Momentum (FATL − RFTL)."""

    STLM = 6
    """Slow Trend Line Momentum (SATL − RSTL)."""

    PCCI = 7
    """Perfect Commodity Channel Index (sample − FATL)."""
