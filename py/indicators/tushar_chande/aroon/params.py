"""Parameters for the Aroon indicator."""

from dataclasses import dataclass


@dataclass
class AroonParams:
    """Parameters for Aroon indicator.

    length: lookback period (must be >= 2, default 14).
    """

    length: int = 14
    """The lookback period for the Aroon calculation.

    The value should be greater than 1. The default value is 14.
    """


def default_params() -> AroonParams:
    """Return default Aroon parameters."""
    return AroonParams()
