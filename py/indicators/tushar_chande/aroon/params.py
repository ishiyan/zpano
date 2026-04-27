"""Parameters for the Aroon indicator."""

from dataclasses import dataclass


@dataclass
class AroonParams:
    """Parameters for Aroon indicator.

    length: lookback period (must be >= 2, default 14).
    """

    length: int = 14


def default_params() -> AroonParams:
    """Return default Aroon parameters."""
    return AroonParams()
