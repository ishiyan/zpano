from enum import IntEnum


class DynamicCycleExplorerOutput(IntEnum):
    """Describes the output components."""
    VALUE = 0
    """Value."""
    PERIOD = 1
    """Period."""
    PHASE = 2
    """Phase."""
    ETA_TOP = 3
    """Eta top."""
    ETA_BOTTOM = 4
    """Eta bottom."""
    HIGHLIGHTER = 5
    """Highlighter."""
    HIGHLIGHTER_INDEX = 6
    """Highlighter index."""
    FITTED_AT_HIGHLIGHTER = 7
    """Fitted at highlighter."""
    DELAYED_HIGHLIGHTER = 8
    """Delayed highlighter."""
