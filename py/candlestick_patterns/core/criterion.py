"""Criterion: specifies a threshold based on the average value of a candlestick range entity.

The criteria are based on parts of the candlestick and common words indicating length
(short, long, very long), displacement (near, far), or equality (equal).

For streaming efficiency, the criterion maintains a running total that is updated
incrementally via add() and remove() rather than rescanning the entire history.
"""
from __future__ import annotations

from .range_entity import RangeEntity
from .primitives import candle_range_value


class Criterion:
    """A criterion based on the average value of a certain part of a candlestick multiplied by a factor.

    Args:
        entity: The type of range entity to consider (RealBody, HighLow, or Shadows).
        average_period: Number of previous candlesticks to calculate an average value.
        factor: Coefficient to multiply the average value.
    """

    __slots__ = ('entity', 'average_period', 'factor')

    def __init__(self, entity: RangeEntity, average_period: int, factor: float) -> None:
        self.entity = entity
        self.average_period = average_period
        self.factor = factor

    def copy(self) -> Criterion:
        """Create an independent copy."""
        return Criterion(self.entity, self.average_period, self.factor)

    def average_value_from_total(self, total: float, o: float, h: float, l: float, c: float) -> float:
        """Compute the criterion threshold from a precomputed running total.

        When average_period > 0, uses the running total.
        When average_period == 0, uses the current candle's own range value.

        Args:
            total: The running sum of range values over the averaging window.
            o, h, l, c: OHLC of the reference candle (used when average_period == 0).
        """
        if self.average_period > 0:
            if self.entity == RangeEntity.SHADOWS:
                return self.factor * total / (self.average_period * 2.0)
            return self.factor * total / self.average_period

        # Period == 0: use the candle's own range value directly.
        return self.factor * candle_range_value(self.entity, o, h, l, c)

    def candle_contribution(self, o: float, h: float, l: float, c: float) -> float:
        """Compute the contribution of a single candle to the running total.

        For SHADOWS entity, this returns the full (upper + lower) shadow sum
        (not yet divided by 2 -- the division happens in average_value_from_total).
        """
        if self.entity == RangeEntity.REAL_BODY:
            return c - o if c >= o else o - c
        if self.entity == RangeEntity.HIGH_LOW:
            return h - l
        # SHADOWS: upper + lower shadow sum (division by 2 deferred to average_value_from_total)
        if c >= o:
            return h - c + o - l
        return h - o + c - l
