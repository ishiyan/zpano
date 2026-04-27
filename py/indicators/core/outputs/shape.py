"""Enumerates the data shapes an indicator output can take."""

from enum import IntEnum


class Shape(IntEnum):
    """Identifies the data shape of an indicator output."""

    # Scalar holds a time stamp and a value.
    SCALAR = 0

    # Band holds a time stamp and two values representing upper and lower lines of a band.
    BAND = 1

    # Heatmap holds a time stamp and an array of values representing a heat-map column.
    HEATMAP = 2

    # Polyline holds a time stamp and an ordered, variable-length sequence of (offset, value) points.
    POLYLINE = 3
