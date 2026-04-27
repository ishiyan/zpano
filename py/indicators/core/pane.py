"""Identifies the chart pane an indicator output is drawn on."""

from enum import IntEnum


class Pane(IntEnum):
    """Identifies the chart pane an indicator output is drawn on."""

    # The primary price pane.
    PRICE = 0

    # A dedicated sub-pane for this indicator.
    OWN = 1

    # Drawing on the parent indicator's pane.
    OVERLAY_ON_PARENT = 2
