from enum import IntEnum


class CompleteCycleScannerOutput(IntEnum):
    """Describes the output components."""
    SPECTRUM = 0
    """Spectrum."""
    GENUINE_PERCENT = 1
    """Genuine percent."""
    DOMINANT_PERIOD = 2
    """Dominant period."""
    ACCEPTED_COUNT = 3
    """Accepted count."""
