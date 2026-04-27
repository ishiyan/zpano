from enum import IntEnum


class StochasticOutput(IntEnum):
    """Output of the Stochastic Oscillator indicator."""
    FAST_K = 0
    SLOW_K = 1
    SLOW_D = 2
