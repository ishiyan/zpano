"""DominantCycle output enum."""

from enum import IntEnum


class DominantCycleOutput(IntEnum):
    """Output describes the outputs of the DominantCycle indicator."""
    RAW_PERIOD = 0
    """The raw instantaneous cycle period produced by the Hilbert transformer estimator."""

    PERIOD = 1
    """The dominant cycle period obtained by additional EMA smoothing of the raw period."""

    PHASE = 2
    """The dominant cycle phase, in degrees."""
