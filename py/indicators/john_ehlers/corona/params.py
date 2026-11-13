"""Configuration for the Corona spectral analysis engine."""

from __future__ import annotations

from dataclasses import dataclass


@dataclass
class CoronaParams:
    """Configures a Corona spectral analysis engine.

    All fields default to Ehlers' November 2008 TASC article settings.
    Explicit zero values are NOT substituted with defaults.
    """

    high_pass_filter_cutoff: int = 30
    """High-pass filter cutoff period (de-trending period), in bars. Must be >= 5."""

    minimal_period: int = 6
    """Minimum cycle period (in bars) covered by the bandpass filter bank. Must be >= 6."""

    maximal_period: int = 30
    """Maximum cycle period (in bars) covered by the bandpass filter bank. Must be > minimal_period."""

    decibels_lower_threshold: float = 6.0
    """Filter bins with smoothed dB value at or below this threshold contribute to the
    weighted dominant-cycle estimate. Must satisfy 0 <= decibels_lower_threshold < decibels_upper_threshold.
    """

    decibels_upper_threshold: float = 20.0
    """Upper clamp on the smoothed dB value and reference value for the dominant-cycle
    weighting (weight = upper - dB). Must be > 0 and > decibels_lower_threshold.
    """


def default_params() -> CoronaParams:
    """Return a CoronaParams with Ehlers defaults."""
    return CoronaParams()
