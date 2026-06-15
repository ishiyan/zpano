"""Output enumeration for the AutoCorrelation Indicator."""

from enum import IntEnum


class Output(IntEnum):
    """Describes the outputs of the indicator."""

    VALUE = 0
    """The autocorrelation indicator heatmap column."""
