# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; CC BY 4.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
from ..harmonic_cycle.validation import _integer, _real, _sample, _finite
from ..cyclic_smoothed_relative_strength_index.rsi import _WilderRSI
from ..metonic.calendar import _Calendar, _DelayMean
class MetonicRSIOutput(NamedTuple):
    value: float
    linear_rsi: float

class MetonicRSI:
    """Wilder RSI followed by the disclosed three-calendar-delay average."""
    def __init__(self, length=13, delay1=6939, delay2=4016, delay3=2922):
        self.length = _integer(length,'length')

        self.delay1 = _integer(delay1,'delay1')
        self.delay2 = _integer(delay2,'delay2')
        self.delay3 = _integer(delay3,'delay3')
        self.delays = (self.delay1,self.delay2,self.delay3)
        if len(set(self.delays))!=3:
            raise ValueError("three delays must be distinct")

        self.linear_warmup_bars = self.length
        self.warmup_bars = max(self.delays)+self.linear_warmup_bars
        self.reset()

    def reset(self):
        self.samples_seen = 0
        self.rsi = _WilderRSI(self.length)
        self.calendar = _Calendar()
        self.delay = _DelayMean(self.delays)
        self.last_value = math.nan

    @property
    def is_primed(self):
        return not math.isnan(self.last_value)

    def update(self, price, time=None):
        price = _sample(price)
        if math.isnan(price):
            return MetonicRSIOutput(math.nan,math.nan)
        self.calendar.validate(time)
        linear = self.rsi.update(price)
        self.last_value = self.delay.update(linear)
        self.samples_seen += 1
        self.calendar.commit(time)
        return MetonicRSIOutput(self.last_value,linear)

Engine = MetonicRSI
