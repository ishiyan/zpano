# SPDX-License-Identifier: MPL-2.0
# Source: Detrended Rhythm Oscillator, © WhenToTrade / Lars von Thienen.
from enum import IntEnum


class DetrendedRhythmOscillatorsOutput(IntEnum):
    """Describes the output components."""
    VALUE = 0
    """Value."""
    DIRECTION = 1
    """Direction."""
    CYCLE_LENGTH = 2
    """Cycle length."""
    PIVOT_VALUE = 3
    """Pivot value."""
    PIVOT_OFFSET = 4
    """Pivot offset."""
