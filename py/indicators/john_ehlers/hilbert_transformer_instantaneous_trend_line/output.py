"""HilbertTransformerInstantaneousTrendLine output enum."""

from enum import IntEnum


class HilbertTransformerInstantaneousTrendLineOutput(IntEnum):
    """Output indices for the HilbertTransformerInstantaneousTrendLine indicator."""
    VALUE = 0
    """The instantaneous trend line value."""

    DOMINANT_CYCLE_PERIOD = 1
    """The smoothed dominant cycle period."""
