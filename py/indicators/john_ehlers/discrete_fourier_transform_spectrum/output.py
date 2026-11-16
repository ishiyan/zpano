"""Output enumeration for the Discrete Fourier Transform Spectrum."""

from enum import IntEnum


class DiscreteFourierTransformSpectrumOutput(IntEnum):
    """Describes the outputs of the Discrete Fourier Transform Spectrum indicator."""

    VALUE = 0
    """The discrete Fourier transform spectrum heatmap column."""

    DOMINANT_CYCLE = 1
    """The center-of-gravity dominant cycle period estimate."""


Output = DiscreteFourierTransformSpectrumOutput
