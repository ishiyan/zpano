"""Sinc Wavelet Band-Pass output enum."""

from enum import IntEnum


class SincWaveletBandpassOutput(IntEnum):
    """Describes the outputs of the indicator."""

    VALUE = 0
    """The band-passed price component (or its velocity)."""
