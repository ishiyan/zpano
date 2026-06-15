"""Adaptive Exponential Moving Average output enum."""

from enum import IntEnum


class AdaptiveExponentialMovingAverageOutput(IntEnum):
    """Describes the outputs of the indicator."""

    VALUE = 0
    """The adaptively smoothed price value."""

    OMEGA = 1
    """The instantaneous frequency estimate (may be NaN)."""

    ALPHA = 2
    """The smoothing factor used for the bar."""
