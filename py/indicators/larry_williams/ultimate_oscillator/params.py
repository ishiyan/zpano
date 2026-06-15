"""Parameters for the Ultimate Oscillator."""

from dataclasses import dataclass


@dataclass
class UltimateOscillatorParams:
    """Parameters for Ultimate Oscillator.

    length1: first (shortest) period (default 7, must be >= 2, 0 = default).
    length2: second (medium) period (default 14, must be >= 2, 0 = default).
    length3: third (longest) period (default 28, must be >= 2, 0 = default).
    """

    length1: int = 7
    """First time period (default 7). Minimum 2."""

    length2: int = 14
    """Second time period (default 14). Minimum 2."""

    length3: int = 28
    """Third time period (default 28). Minimum 2."""


def default_params() -> UltimateOscillatorParams:
    """Return default Ultimate Oscillator parameters."""
    return UltimateOscillatorParams()
