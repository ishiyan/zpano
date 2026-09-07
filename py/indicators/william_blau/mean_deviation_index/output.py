"""Mean Deviation Index output enum."""

from enum import IntEnum


class MeanDeviationIndexOutput(IntEnum):
    """Describes the outputs of the indicator."""

    MDI = 0
    """The Mean Deviation Index line value, in raw price units (unbounded)."""

    SIGNAL = 1
    """The signal-line value: the ul-period EMA of the index."""
