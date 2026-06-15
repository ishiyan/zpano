"""Configuration for the Corona spectral analysis engine."""

from dataclasses import dataclass


@dataclass
class CoronaParams:
    """Configures a Corona spectral analysis engine.

    All fields have default values following Ehlers' original TASC article
    (November 2008). Zero or negative values mean "use the default".
    """

    high_pass_filter_cutoff: int = 30
    """High-pass filter cutoff period (de-trending period), in bars. Must be >= 2.

    The default value is 30.
    """

    minimal_period: int = 6
    """Minimum cycle period (in bars) covered by the bandpass filter bank. Must be >= 2.

    The default value is 6.
    """

    maximal_period: int = 30
    """Maximum cycle period (in bars) covered by the bandpass filter bank. Must be > minimalPeriod.

    The default value is 30.
    """

    decibels_lower_threshold: float = 6.0
    """Filter bins with smoothed dB value at or below this threshold contribute to the
    weighted dominant-cycle estimate.

    The default value is 6.
    """

    decibels_upper_threshold: float = 20.0
    """Upper clamp on the smoothed dB value and reference value for the dominant-cycle
    weighting (weight = upper − dB).

    The default value is 20.
    """


def default_params() -> CoronaParams:
    """Return a CoronaParams with Ehlers defaults."""
    return CoronaParams()
