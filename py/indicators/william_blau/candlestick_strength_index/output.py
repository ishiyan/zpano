"""Candlestick Strength Index output enum."""

from enum import IntEnum


class CandlestickStrengthIndexOutput(IntEnum):
    """Describes the outputs of the indicator."""

    CSI = 0
    """The Candlestick Strength Index oscillator value (range [-100, +100])."""

    SIGNAL = 1
    """The signal-line value: the ul-period EMA of the oscillator."""
