"""Parameters for the Parabolic Stop And Reverse indicator."""

from dataclasses import dataclass


@dataclass
class ParabolicStopAndReverseParams:
    """Parameters for the Parabolic Stop And Reverse indicator.

    The Parabolic SAR Extended supports separate acceleration factor parameters for long
    and short directions.
    """

    start_value: float = 0.0
    """Controls the initial direction and SAR value.

     0  = Auto-detect direction using the first two bars (default).
     >0 = Force long at the specified SAR value.
     <0 = Force short at abs(startValue) as the initial SAR value.

    Default is 0.0.
    """

    offset_on_reverse: float = 0.0
    """A percent offset added/removed to the initial stop on short/long reversal.

    Default is 0.0.
    """

    acceleration_init_long: float = 0.02
    """The initial acceleration factor for the long direction.

    Default is 0.02.
    """

    acceleration_long: float = 0.02
    """The acceleration factor increment for the long direction.

    Default is 0.02.
    """

    acceleration_max_long: float = 0.20
    """The maximum acceleration factor for the long direction.

    Default is 0.20.
    """

    acceleration_init_short: float = 0.02
    """The initial acceleration factor for the short direction.

    Default is 0.02.
    """

    acceleration_short: float = 0.02
    """The acceleration factor increment for the short direction.

    Default is 0.02.
    """

    acceleration_max_short: float = 0.20
    """The maximum acceleration factor for the short direction.

    Default is 0.20.
    """


def default_params() -> ParabolicStopAndReverseParams:
    """Returns default parameters."""
    return ParabolicStopAndReverseParams()
