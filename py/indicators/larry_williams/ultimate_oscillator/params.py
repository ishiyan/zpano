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
    length2: int = 14
    length3: int = 28


def default_params() -> UltimateOscillatorParams:
    """Return default Ultimate Oscillator parameters."""
    return UltimateOscillatorParams()
