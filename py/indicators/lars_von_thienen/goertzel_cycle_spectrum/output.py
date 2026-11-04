from enum import IntEnum


class GoertzelCycleSpectrumOutput(IntEnum):
    """Describes the output components."""
    AMPLITUDE = 0
    """Amplitude."""
    PHASES = 1
    """Phases."""
    TROUGH_OFFSETS = 2
    """Trough offsets."""
    DOMINANT_PERIOD = 3
    """Dominant period."""
    DOMINANT_AMPLITUDE = 4
    """Dominant amplitude."""
    DOMINANT_PHASE = 5
    """Dominant phase."""
