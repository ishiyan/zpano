"""Fractal Graph Dimension Index output enum."""

from enum import IntEnum


class FractalGraphDimensionIndexOutput(IntEnum):
    """Enumerates the outputs of the fractal graph dimension index indicator."""

    FGDI = 0
    """The fractal graph dimension value."""

    UPPER = 1
    """The upper band (fgdi + stddev)."""

    LOWER = 2
    """The lower band (fgdi - stddev)."""

    STDDEV = 3
    """The standard deviation of the dimension estimate."""

    BAND = 4
    """The lower/upper band pair."""
