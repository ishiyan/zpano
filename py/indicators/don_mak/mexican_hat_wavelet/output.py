"""Mexican Hat Wavelet output enum."""

from enum import IntEnum


class MexicanHatWaveletOutput(IntEnum):
    """Describes the outputs of the indicator."""

    VALUE = 0
    """The bandpass-filtered price component."""
