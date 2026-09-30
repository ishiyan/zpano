"""Slope Divergence TSI Filter output enum."""

from enum import IntEnum


class SlopeDivergenceTsiFilterOutput(IntEnum):
    """Describes the outputs of the indicator."""

    VALUE = 0
    """The Slope Divergence TSI Filter value (range [-100, +100])."""
