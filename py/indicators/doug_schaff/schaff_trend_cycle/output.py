"""Schaff Trend Cycle output enum."""

from enum import IntEnum


class SchaffTrendCycleOutput(IntEnum):
    """Describes the outputs of the indicator."""

    STC = 0
    """The Schaff Trend Cycle oscillator value (range [0, 100])."""

    MACD = 1
    """The gated MACD line (XMAC) value."""

    PF = 2
    """The first smoothed %D stage value."""
