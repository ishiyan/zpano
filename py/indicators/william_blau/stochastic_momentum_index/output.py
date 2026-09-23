"""Stochastic Momentum Index output enum."""

from enum import IntEnum


class StochasticMomentumIndexOutput(IntEnum):
    """Describes the outputs of the indicator."""

    SMI = 0
    """The Stochastic Momentum Index oscillator value (range [-100, +100])."""

    SIGNAL = 1
    """The signal-line value: the ul-period EMA of the oscillator."""
