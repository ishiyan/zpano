from enum import IntEnum


class RangeEntity(IntEnum):
    """The entities of range that can be considered when comparing a part of a candlestick to other candlesticks."""

    REAL_BODY = 0
    """Identifies the length of the real body of a candlestick."""

    HIGH_LOW = 1
    """Identifies the length of the high-low range of a candlestick."""

    SHADOWS = 2
    """Identifies the length of the shadows of a candlestick."""
