"""Output enumeration for the Comb Band-Pass Spectrum."""

from enum import IntEnum


class CombBandPassSpectrumOutput(IntEnum):
    """Describes the outputs of the Comb Band-Pass Spectrum indicator."""

    VALUE = 0
    """The comb band-pass spectrum heatmap column."""

    DOMINANT_CYCLE = 1
    """The power-weighted center-of-gravity dominant cycle period estimate."""


Output = CombBandPassSpectrumOutput
