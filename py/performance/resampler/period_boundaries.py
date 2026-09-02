from __future__ import annotations

import math
from abc import ABC, abstractmethod
from collections.abc import Callable
from dataclasses import dataclass
from datetime import datetime, timedelta
from typing import Optional


class PeriodBoundary(ABC):
    """
    Define the boundaries of regular calendar periods.

    A ``PeriodBoundary`` maps a timestamp to the start of its containing
    period and determines the start of the following period.
    """

    @abstractmethod
    def period_start(self, timestamp: datetime) -> datetime:
        """
        Return the start of the period containing ``timestamp``.
        """
        raise NotImplementedError

    @abstractmethod
    def next_period_start(self, period_start: datetime) -> datetime:
        """
        Return the start of the period immediately following ``period_start``.
        """
        raise NotImplementedError


@dataclass(frozen=True)
class MinuteBoundary(PeriodBoundary):
    """
    Calendar-aligned minute boundary.

    Args:
        minutes:
            Number of minutes per period.

            Examples::

                MinuteBoundary(1)
                MinuteBoundary(5)
                MinuteBoundary(15)
                MinuteBoundary(30)
    """
    minutes: int = 1

    def __post_init__(self) -> None:
        if self.minutes <= 0:
            raise ValueError("minutes must be positive")
        if 60 % self.minutes != 0:
            raise ValueError("minutes must divide evenly into one hour")

    def period_start(self, timestamp: datetime) -> datetime:
        minute = (timestamp.minute // self.minutes) * self.minutes
        return timestamp.replace(minute=minute, second=0, microsecond=0)

    def next_period_start(self, period_start: datetime) -> datetime:
        return period_start + timedelta(minutes=self.minutes)

@dataclass(frozen=True)
class HourBoundary(PeriodBoundary):
    """
    Calendar-aligned hour boundary.

    Args:
        hours:
            Number of hours per period.

            Examples::

                HourBoundary(1)
                HourBoundary(2)
                HourBoundary(4)
                HourBoundary(6)
                HourBoundary(12)
    """
    hours: int = 1

    def __post_init__(self) -> None:
        if self.hours <= 0:
            raise ValueError("hours must be positive")
        if 24 % self.hours != 0:
            raise ValueError("hours must divide evenly into one day")

    def period_start(self, timestamp: datetime) -> datetime:
        hour = (timestamp.hour // self.hours) * self.hours
        return timestamp.replace(hour=hour, minute=0, second=0, microsecond=0)

    def next_period_start(self, period_start: datetime) -> datetime:
        return period_start + timedelta(hours=self.hours)

@dataclass(frozen=True)
class DayBoundary(PeriodBoundary):
    """
    Calendar-day boundary.

    Each period starts at midnight in the timezone of the supplied
    ``datetime``.

    Args:
        days:
            Number of calendar days per period.

    Notes:
        ``DayBoundary(1)`` is the normal daily boundary.

        For financial-market sessions, where a trading day does not
        necessarily correspond to a calendar day, a dedicated exchange
        calendar boundary should be implemented instead.
    """

    days: int = 1

    def __post_init__(self) -> None:
        if self.days <= 0:
            raise ValueError("days must be positive")

    def period_start(self, timestamp: datetime) -> datetime:
        if self.days == 1:
            return timestamp.replace(hour=0, minute=0, second=0, microsecond=0)

        day = ((timestamp.day - 1) // self.days) * self.days + 1

        return timestamp.replace(day=day, hour=0, minute=0, second=0, microsecond=0)

    def next_period_start(self, period_start: datetime) -> datetime:
        return period_start + timedelta(days=self.days)

