"""Output enumeration for the AutoCorrelationPeriodogram indicator."""

from enum import IntEnum


class AutoCorrelationPeriodogramOutput(IntEnum):
    """Describes the outputs of the AutoCorrelationPeriodogram indicator."""

    VALUE = 0
    """The autocorrelation periodogram heatmap column."""

    DOMINANT_CYCLE = 1
    """The thresholded center-of-gravity dominant cycle period estimate."""


Output = AutoCorrelationPeriodogramOutput
