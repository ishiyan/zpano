"""Parameters for the Parabolic Stop And Reverse indicator."""

from dataclasses import dataclass


@dataclass
class ParabolicStopAndReverseParams:
    """Parameters for the Parabolic Stop And Reverse indicator.

    The Parabolic SAR Extended supports separate acceleration factor parameters for long
    and short directions.
    """

    start_value: float = 0.0
    offset_on_reverse: float = 0.0
    acceleration_init_long: float = 0.02
    acceleration_long: float = 0.02
    acceleration_max_long: float = 0.20
    acceleration_init_short: float = 0.02
    acceleration_short: float = 0.02
    acceleration_max_short: float = 0.20


def default_params() -> ParabolicStopAndReverseParams:
    """Returns default parameters."""
    return ParabolicStopAndReverseParams()
