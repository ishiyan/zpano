"""Output enum for the CoronaSignalToNoiseRatio indicator."""

from enum import IntEnum


class CoronaSignalToNoiseRatioOutput(IntEnum):
    """Outputs of the Corona Signal-to-Noise Ratio indicator."""
    VALUE = 0
    SIGNAL_TO_NOISE_RATIO = 1
