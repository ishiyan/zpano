from enum import IntEnum


class StochasticRelativeStrengthIndexOutput(IntEnum):
    """Output indices for the Stochastic RSI indicator."""
    FAST_K = 0
    """The Fast-K line of the stochastic RSI."""

    FAST_D = 1
    """The Fast-D line of the stochastic RSI (smoothed Fast-K)."""
