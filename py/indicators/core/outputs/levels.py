"""Levels output type: a time stamp and a variable-length set of (value, offset, strength) entries."""

import datetime
import math


class Level:
    """A single entry of a Levels output, expressed as a value with an optional bar
    offset and an optional strength."""

    __slots__ = ('value', 'offset', 'strength')

    def __init__(self, value: float, offset: int = 0, strength: float = math.nan) -> None:
        self.value = value
        self.offset = offset
        self.strength = strength

    def __repr__(self) -> str:
        return f"Level({self.value}, {self.offset}, {self.strength})"


class Levels:
    """Holds a time stamp and a variable-length set of levels (e.g. support/resistance)."""

    __slots__ = ('time', 'levels')

    def __init__(self, time: datetime.datetime, levels: list[Level]) -> None:
        self.time = time
        self.levels = levels

    def is_empty(self) -> bool:
        """Indicates whether this Levels has no entries."""
        return len(self.levels) == 0

    def __repr__(self) -> str:
        items = ' '.join(f'({lv.value}, {lv.offset}, {lv.strength})' for lv in self.levels)
        return f"Levels({self.time}, [{items}])"

    @staticmethod
    def empty(time: datetime.datetime) -> 'Levels':
        """Creates a new empty Levels with no entries."""
        return Levels(time, [])
