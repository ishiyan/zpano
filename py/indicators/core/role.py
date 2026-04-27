"""Classifies the semantic role a single indicator output plays in analysis."""

from enum import IntEnum


class Role(IntEnum):
    """Classifies the semantic role a single indicator output plays in analysis."""

    SMOOTHER = 0
    ENVELOPE = 1
    OVERLAY = 2
    POLYLINE = 3
    OSCILLATOR = 4
    BOUNDED_OSCILLATOR = 5
    VOLATILITY = 6
    VOLUME_FLOW = 7
    DIRECTIONAL = 8
    CYCLE_PERIOD = 9
    CYCLE_PHASE = 10
    FRACTAL_DIMENSION = 11
    SPECTRUM = 12
    SIGNAL = 13
    HISTOGRAM = 14
    REGIME_FLAG = 15
    CORRELATION = 16
