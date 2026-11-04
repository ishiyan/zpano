# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; CC BY 4.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
from ..harmonic_cycle.validation import _integer, _real, _sample, _finite
from ..metonic.calendar import _Calendar, _DelayMean
class MetonicWilliamsOutput(NamedTuple):
    value: float
    linear_percent_r: float

class MetonicWilliamsR:
    """Conventional high/low/close Williams %R, then three-delay average.

    No close-only extrema substitute is silently created. All three prices
    belong to the same completed calendar-day bar, including filled days.
    """
    def __init__(self, length=13, delay1=6939, delay2=4016, delay3=2922):
        self.length = _integer(length,'length')

        self.delay1 = _integer(delay1,'delay1')
        self.delay2 = _integer(delay2,'delay2')
        self.delay3 = _integer(delay3,'delay3')
        self.delays = (self.delay1,self.delay2,self.delay3)
        if len(set(self.delays))!=3:
            raise ValueError("three delays must be distinct")

        self.linear_warmup_bars = self.length-1
        self.warmup_bars = max(self.delays)+self.linear_warmup_bars
        self.reset()

    def reset(self):
        self.samples_seen = 0
        self.highs,self.lows = deque(maxlen=self.length),deque(maxlen=self.length)
        self.calendar = _Calendar()
        self.delay = _DelayMean(self.delays)
        self.last_value = math.nan

    @property
    def is_primed(self):
        return not math.isnan(self.last_value)

    def update(self, close, high=None, low=None, time=None):
        close = _sample(close)
        if math.isnan(close):
            return MetonicWilliamsOutput(math.nan,math.nan)
        if high is None or low is None:
            raise ValueError("conventional Williams %R requires high and low")
        high,low = _sample(high),_sample(low)
        if math.isnan(high) or math.isnan(low):
            return MetonicWilliamsOutput(math.nan,math.nan)
        if not low<=close<=high:
            raise ValueError("completed HLC bar must satisfy low <= close <= high")
        self.calendar.validate(time)
        self.highs.append(high)
        self.lows.append(low)
        linear = math.nan
        if len(self.highs)==self.length:
            maximum,minimum = max(self.highs),min(self.lows)
            # A zero range is undefined. Preserve its DAY as NaN in the ring.
            if maximum>minimum:
                linear = -100.0*(maximum-close)/(maximum-minimum)
        self.last_value = self.delay.update(linear)
        self.samples_seen += 1
        self.calendar.commit(time)
        return MetonicWilliamsOutput(self.last_value,linear)

Engine = MetonicWilliamsR
