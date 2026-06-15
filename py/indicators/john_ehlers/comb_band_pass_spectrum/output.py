"""Output enumeration for the Comb Band-Pass Spectrum."""

from enum import IntEnum


class Output(IntEnum):
    """Describes the outputs of the indicator."""

    VALUE = 0
    """The comb band-pass spectrum heatmap column."""
