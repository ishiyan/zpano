from enum import IntEnum


class HodrickPrescottSmoothingAndDetrendingOutput(IntEnum):
    """Describes the output components."""
    TREND = 0
    """Trend."""
    CYCLE = 1
    """Cycle."""
