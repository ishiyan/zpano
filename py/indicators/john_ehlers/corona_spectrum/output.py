"""Output enum for the CoronaSpectrum indicator."""

from enum import IntEnum


class CoronaSpectrumOutput(IntEnum):
    """Outputs of the Corona Spectrum indicator."""
    VALUE = 0
    """The Corona spectrum heatmap column (decibels across the filter bank)."""

    DOMINANT_CYCLE = 1
    """The weighted-center-of-gravity dominant cycle estimate."""

    DOMINANT_CYCLE_MEDIAN = 2
    """The 5-sample median of DominantCycle."""
