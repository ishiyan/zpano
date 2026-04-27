"""Band output type: two values (upper/lower) and a time stamp."""

import datetime
import math


class Band:
    """Represents two band values and a time stamp."""

    __slots__ = ('time', 'lower', 'upper')

    def __init__(self, time: datetime.datetime, lower: float, upper: float) -> None:
        self.time = time
        if lower < upper:
            self.lower = lower
            self.upper = upper
        else:
            self.lower = upper
            self.upper = lower

    def is_empty(self) -> bool:
        """Indicates whether this band is not initialized."""
        return math.isnan(self.lower) or math.isnan(self.upper)

    def __repr__(self) -> str:
        return f"Band({self.time}, {self.lower}, {self.upper})"

    @staticmethod
    def empty(time: datetime.datetime) -> 'Band':
        """Creates a new empty band with NaN values."""
        nan = math.nan
        b = object.__new__(Band)
        b.time = time
        b.lower = nan
        b.upper = nan
        return b
