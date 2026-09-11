"""MACD Index output enum."""

from enum import IntEnum


class MacdIndexOutput(IntEnum):
    """Describes the outputs of the indicator."""

    MACDI = 0
    """The MACD Index line value, in raw price units (unbounded)."""

    SIGNAL = 1
    """The signal-line value: the ul-period EMA of the index."""
