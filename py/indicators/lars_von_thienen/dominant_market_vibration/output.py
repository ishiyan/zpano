from enum import IntEnum


class DominantMarketVibrationOutput(IntEnum):
    """Registered plots; all twelve diagnostics remain in the snapshot."""
    PERIOD = 0
    AMPLITUDE = 1
    PHASE = 2
    STRENGTH = 3
    GENUINE_PERCENT = 4
    VALUE = 5
    SLOPE = 6
    ETA_TOP = 7
    ETA_BOTTOM = 8
