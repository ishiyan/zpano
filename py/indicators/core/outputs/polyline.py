"""Polyline output type: a time stamp and an ordered sequence of (offset, value) points."""

import datetime


class Point:
    """A single vertex of a Polyline, expressed as (offset, value)."""

    __slots__ = ('offset', 'value')

    def __init__(self, offset: int, value: float) -> None:
        self.offset = offset
        self.value = value

    def __repr__(self) -> str:
        return f"Point({self.offset}, {self.value})"


class Polyline:
    """Holds a time stamp and an ordered sequence of points describing a polyline."""

    __slots__ = ('time', 'points')

    def __init__(self, time: datetime.datetime, points: list[Point]) -> None:
        self.time = time
        self.points = points

    def is_empty(self) -> bool:
        """Indicates whether this polyline has no points."""
        return len(self.points) == 0

    def __repr__(self) -> str:
        pts = ' '.join(f'({p.offset}, {p.value})' for p in self.points)
        return f"Polyline({self.time}, [{pts}])"

    @staticmethod
    def empty(time: datetime.datetime) -> 'Polyline':
        """Creates a new empty polyline with no points."""
        return Polyline(time, [])
