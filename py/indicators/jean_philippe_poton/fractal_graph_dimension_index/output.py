"""Fractal Graph Dimension Index output enum."""

from enum import IntEnum


class FractalGraphDimensionIndexOutput(IntEnum):
    """Enumerates the outputs of the fractal graph dimension index indicator."""

    FGDI = 0
    UPPER = 1
    LOWER = 2
    STDDEV = 3
    BAND = 4
