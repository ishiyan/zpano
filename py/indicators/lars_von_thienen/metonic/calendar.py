# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; CC BY 4.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
def fill_calendar_days(records):
    """Yield (date, payload), copying the preceding FULL payload into gaps.

    records must be increasing completed daily observations. Payload can be
    a close, an OHLC tuple, or any immutable caller-owned value. No market
    holiday knowledge is required: every absent calendar date is filled.
    """
    previous_day = previous_payload = None
    for day, payload in records:
        if type(day) is not date:
            raise ValueError("calendar records require datetime.date")
        if previous_day is not None:
            if day <= previous_day:
                raise ValueError("calendar records must increase strictly")
            missing = previous_day + timedelta(days=1)
            while missing < day:
                yield missing, previous_payload
                missing += timedelta(days=1)
        yield day, payload
        previous_day, previous_payload = day, payload

class _Calendar:
    """Dates affect indexing; they are not merely optional plot metadata."""
    def __init__(self):
        self.last = None
        self.dated = None

    def validate(self, time):
        # Undated mode is allowed only under the caller's daily-grid promise.
        if isinstance(time, datetime):
            raise ValueError("use a date, not a datetime with timezone ambiguity")
        if time is not None and type(time) is not date:
            raise ValueError("time must be a date or None")
        if self.dated is not None and self.dated != (time is not None):
            raise ValueError("cannot mix dated and undated accepted observations")
        if self.last is not None and time != self.last + timedelta(days=1):
            raise ValueError("calendar gap: call fill_calendar_days before update")

    def commit(self, time):
        self.last = time
        self.dated = time is not None

class _DelayMean:
    """Bounded calendar-day ring; NaN underlying WARMUP still occupies a day."""
    def __init__(self, delays):
        self.delays = delays
        self.history = deque(maxlen=max(delays)+1)

    def update(self, value):
        self.history.append(value)
        if len(self.history) < self.history.maxlen:
            return math.nan
        values = [self.history[-1-delay] for delay in self.delays]
        if any(math.isnan(item) for item in values):
            return math.nan
        total = 0.0
        for item in values:
            total += item
        return total/len(values)


def calendar_day(value: date | datetime) -> date:
    """Extract a daily entity timestamp, rejecting intraday or non-UTC times."""
    if isinstance(value, datetime):
        if value.hour or value.minute or value.second or value.microsecond or value.utcoffset() not in (None, timedelta(0)):
            raise ValueError('use UTC midnight calendar timestamps')
        return value.date()
    if type(value) is not date:
        raise ValueError('use a date or UTC midnight timestamp')
    return value
