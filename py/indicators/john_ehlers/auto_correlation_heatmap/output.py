"""Output enumeration for Ehlers' Autocorrelation Heatmap."""

from enum import IntEnum


class AutoCorrelationHeatmapOutput(IntEnum):
    """Describes the outputs of the Autocorrelation Heatmap indicator."""

    VALUE = 0
    """The signed Pearson lag-correlation heatmap column across [0, max_lag]."""


Output = AutoCorrelationHeatmapOutput
