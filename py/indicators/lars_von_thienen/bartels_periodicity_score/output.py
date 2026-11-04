from enum import IntEnum


class BartelsPeriodicityScoreOutput(IntEnum):
    """Describes the output components."""
    SCORE = 0
    """Score."""
    GENUINE_PERCENT = 1
    """Genuine percent."""
    SEGMENTS = 2
    """Segments."""
