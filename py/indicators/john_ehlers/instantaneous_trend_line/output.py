"""InstantaneousTrendLine output enum."""

from enum import IntEnum


class InstantaneousTrendLineOutput(IntEnum):
    """Describes the outputs of the InstantaneousTrendLine indicator."""
    VALUE = 0
    """The scalar value of the instantaneous trend line."""

    TRIGGER = 1
    """The scalar value of the trigger line."""
