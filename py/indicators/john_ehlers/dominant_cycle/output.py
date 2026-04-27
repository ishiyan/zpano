"""DominantCycle output enum."""

from enum import IntEnum


class DominantCycleOutput(IntEnum):
    """Output describes the outputs of the DominantCycle indicator."""
    RAW_PERIOD = 0
    PERIOD = 1
    PHASE = 2
