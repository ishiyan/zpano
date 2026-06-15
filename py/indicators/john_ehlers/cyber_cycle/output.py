"""CyberCycle output enum."""

from enum import IntEnum


class CyberCycleOutput(IntEnum):
    """Describes the outputs of the CyberCycle indicator."""
    VALUE = 0
    """The scalar value of the cyber cycle."""

    SIGNAL = 1
    """The scalar value of the signal line (EMA of the cycle value)."""
