"""Output enum for the CoronaSignalToNoiseRatio indicator."""

from enum import IntEnum


class CoronaSignalToNoiseRatioOutput(IntEnum):
    """Outputs of the Corona Signal-to-Noise Ratio indicator."""
    VALUE = 0
    """The Corona signal-to-noise ratio heatmap column."""

    SIGNAL_TO_NOISE_RATIO = 1
    """The current signal-to-noise ratio scalar, mapped to [MinParameterValue, MaxParameterValue]."""
