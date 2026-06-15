from enum import IntEnum


class StochasticOutput(IntEnum):
    """Output of the Stochastic Oscillator indicator."""
    FAST_K = 0
    """The Fast-K line (raw stochastic)."""

    SLOW_K = 1
    """The Slow-K line (smoothed Fast-K, also known as Fast-D)."""

    SLOW_D = 2
    """The Slow-D line (smoothed Slow-K)."""
