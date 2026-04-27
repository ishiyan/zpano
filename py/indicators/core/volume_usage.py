"""Classifies how an indicator uses volume information."""

from enum import IntEnum


class VolumeUsage(IntEnum):
    """Classifies how an indicator uses volume information."""

    # Does not use volume.
    NO_VOLUME = 0

    # Consumes per-bar aggregated volume.
    AGGREGATE_BAR_VOLUME = 1

    # Consumes per-trade volume.
    PER_TRADE_VOLUME = 2

    # Consumes quote-side liquidity (bid/ask sizes).
    QUOTE_LIQUIDITY_VOLUME = 3
