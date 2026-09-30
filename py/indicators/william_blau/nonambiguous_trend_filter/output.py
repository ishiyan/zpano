"""Nonambiguous Trend Filter output enum."""

from enum import IntEnum


class NonambiguousTrendFilterOutput(IntEnum):
    """Describes the outputs of the indicator."""

    VALUE = 0
    """The Nonambiguous Trend Filter value (the base oscillator value or 0)."""
