"""ADX-Type Filter output enum."""

from enum import IntEnum


class AdxTypeFilterOutput(IntEnum):
    """Describes the outputs of the indicator."""

    VALUE = 0
    """The ADX-Type Filter value (non-negative)."""
