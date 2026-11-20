"""Output enumeration for Ehlers' Roofed Autocorrelation Heatmap."""

from enum import IntEnum


class RoofedAutoCorrelationHeatmapOutput(IntEnum):
    """Describes the outputs of the Roofed Autocorrelation Heatmap indicator."""

    VALUE = 0
    """The roofed Pearson lag-correlation heatmap column in [0, 1] across [min_lag, max_lag]."""


Output = RoofedAutoCorrelationHeatmapOutput
