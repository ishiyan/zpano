"""FractalAdaptiveMovingAverage output enum."""

from enum import IntEnum


class FractalAdaptiveMovingAverageOutput(IntEnum):
    """Output indices for the FractalAdaptiveMovingAverage indicator."""
    VALUE = 0
    """The scalar value of the fractal adaptive moving average."""

    FDIM = 1
    """The scalar value of the estimated fractal dimension."""
