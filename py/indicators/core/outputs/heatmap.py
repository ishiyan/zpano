"""Heatmap output type: a time stamp and an array of values for a heatmap column."""

import datetime
import math


class Heatmap:
    """Holds a time stamp (x) and an array of values (z) for a heatmap column."""

    __slots__ = ('time', 'parameter_first', 'parameter_last', 'parameter_resolution',
                 'value_min', 'value_max', 'values')

    def __init__(self, time: datetime.datetime, parameter_first: float, parameter_last: float,
                 parameter_resolution: float, value_min: float, value_max: float,
                 values: list[float]) -> None:
        self.time = time
        self.parameter_first = parameter_first
        self.parameter_last = parameter_last
        self.parameter_resolution = parameter_resolution
        self.value_min = value_min
        self.value_max = value_max
        self.values = values

    def is_empty(self) -> bool:
        """Indicates whether this heatmap is not initialized."""
        return len(self.values) < 1

    def __repr__(self) -> str:
        return (f"Heatmap({self.time}, ({self.parameter_first}, {self.parameter_last}, "
                f"{self.parameter_resolution}), ({self.value_min}, {self.value_max}), "
                f"{self.values})")

    @staticmethod
    def empty(time: datetime.datetime, parameter_first: float, parameter_last: float,
              parameter_resolution: float) -> 'Heatmap':
        """Creates a new empty heatmap with NaN min/max and empty values."""
        nan = math.nan
        return Heatmap(time, parameter_first, parameter_last, parameter_resolution,
                       nan, nan, [])
