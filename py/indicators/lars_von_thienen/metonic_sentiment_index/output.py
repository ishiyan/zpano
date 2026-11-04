from enum import IntEnum


class MetonicSentimentIndexOutput(IntEnum):
    """Describes the output components."""
    VALUE = 0
    """Value."""
    DAILY_SENTIMENT = 1
    """Daily sentiment."""
    SMOOTHED_CLOSE = 2
    """Smoothed close."""
    SMOOTHED_OPEN = 3
    """Smoothed open."""
