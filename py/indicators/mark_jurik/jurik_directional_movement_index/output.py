"""Jurik directional movement index output enum."""

from enum import IntEnum


class JurikDirectionalMovementIndexOutput(IntEnum):
    """Output of the Jurik directional movement index indicator."""
    BIPOLAR = 0
    """The bipolar value: 100*(Plus-Minus)/(Plus+Minus)."""

    PLUS = 1
    """The plus directional movement."""

    MINUS = 2
    """The minus directional movement."""
