"""Candlestick Momentum Index output enum."""

from enum import IntEnum


class CandlestickMomentumIndexOutput(IntEnum):
    """Describes the outputs of the indicator."""

    CMI = 0
    """The Candlestick Momentum Index oscillator value (range [-100, +100])."""

    SIGNAL = 1
    """The signal-line value: the ul-period EMA of the oscillator."""
