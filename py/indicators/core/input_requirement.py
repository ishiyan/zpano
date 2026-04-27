"""Classifies the minimum input data type an indicator consumes."""

from enum import IntEnum


class InputRequirement(IntEnum):
    """Classifies the minimum input data type an indicator consumes."""

    # Consumes a scalar time series (e.g., prices).
    SCALAR_INPUT = 0

    # Consumes level-1 quotes.
    QUOTE_INPUT = 1

    # Consumes OHLCV bars.
    BAR_INPUT = 2

    # Consumes individual trades.
    TRADE_INPUT = 3
