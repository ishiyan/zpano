"""Moving Average Convergence Divergence output enum."""

from enum import IntEnum


class MovingAverageConvergenceDivergenceOutput(IntEnum):
    """Describes the outputs of the indicator."""

    MACD = 0
    """The MACD line value (fast MA - slow MA)."""

    SIGNAL = 1
    """The signal line value (MA of MACD line)."""

    HISTOGRAM = 2
    """The histogram value (MACD - signal)."""
