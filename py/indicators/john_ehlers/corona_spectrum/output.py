"""Output enum for the CoronaSpectrum indicator."""

from enum import IntEnum


class CoronaSpectrumOutput(IntEnum):
    """Outputs of the Corona Spectrum indicator."""
    VALUE = 0
    DOMINANT_CYCLE = 1
    DOMINANT_CYCLE_MEDIAN = 2
