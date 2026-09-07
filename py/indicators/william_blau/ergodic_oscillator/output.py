"""Ergodic Oscillator output enum."""

from enum import IntEnum


class ErgodicOscillatorOutput(IntEnum):
    """Describes the outputs of the indicator."""

    ERGODIC = 0
    """The Ergodic oscillator value (the True Strength Index, range [-100, +100])."""

    SIGNAL = 1
    """The signal-line value: the ul-period EMA of the oscillator."""
