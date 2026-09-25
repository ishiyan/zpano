"""Double Smoothed Stochastic output enum."""

from enum import IntEnum


class DoubleSmoothedStochasticOutput(IntEnum):
    """Describes the outputs of the indicator."""

    DSS = 0
    """The Double Smoothed Stochastic oscillator value (range [0, 100])."""

    SIGNAL = 1
    """The signal-line value: the g-period SMA of the oscillator."""
