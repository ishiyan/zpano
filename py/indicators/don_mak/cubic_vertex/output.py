"""Cubic Vertex output enum."""

from enum import IntEnum


class CubicVertexOutput(IntEnum):
    """Describes the outputs of the indicator."""

    BARS_TO_NEAR_TURN = 0
    """The number of bars to the more imminent turning point."""

    BARS_TO_FAR_TURN = 1
    """The number of bars to the more distant turning point."""
