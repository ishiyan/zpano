"""Configuration for the Corona spectral analysis engine."""

from dataclasses import dataclass


@dataclass
class CoronaParams:
    """Configures a Corona spectral analysis engine.

    All fields have default values following Ehlers' original TASC article
    (November 2008). Zero or negative values mean "use the default".
    """

    high_pass_filter_cutoff: int = 30
    minimal_period: int = 6
    maximal_period: int = 30
    decibels_lower_threshold: float = 6.0
    decibels_upper_threshold: float = 20.0


def default_params() -> CoronaParams:
    """Return a CoronaParams with Ehlers defaults."""
    return CoronaParams()
