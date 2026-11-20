"""Output enumeration for Ehlers' Transformed Discrete Fourier Transform Spectrum."""

from enum import IntEnum


class TransformedDiscreteFourierTransformSpectrumOutput(IntEnum):
    """Describes the outputs of the Transformed Discrete Fourier Transform Spectrum indicator."""

    VALUE = 0
    """The transformed discrete Fourier transform spectrum heatmap column in dB."""

    DOMINANT_CYCLE = 1
    """The thresholded centroid dominant cycle period estimate in bars."""


Output = TransformedDiscreteFourierTransformSpectrumOutput
