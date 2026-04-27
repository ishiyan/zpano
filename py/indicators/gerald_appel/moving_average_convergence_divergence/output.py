"""Moving Average Convergence Divergence output enum."""

from enum import IntEnum


class MovingAverageConvergenceDivergenceOutput(IntEnum):
    """Describes the outputs of the indicator."""

    MACD = 0
    SIGNAL = 1
    HISTOGRAM = 2
