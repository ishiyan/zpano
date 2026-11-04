# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; CC BY 4.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
from ..harmonic_cycle.validation import _integer, _real, _sample, _finite
from ..harmonic_cycle.model import _fit_window, composite_value, cycle_value, cycle_slope, cycle_strength, cycle_polarity
from ..harmonic_cycle.model import CycleModel

class PolarityOutput(NamedTuple):
    polarity: float
    short_value: float
    long_value: float
    short_slope: float
    long_slope: float

class CyclePolarity:
    """Configured two-cycle rolling fits plus the book's polarity rule."""
    def __init__(self, window=100, short_period=20.0, long_period=40.0,
                 smoothing=1600.0, tolerance=0.0):
        self.window = _integer(window,'window',6)
        self.short_period = _real(short_period,'short_period',2.0000000000000004)
        self.long_period = _real(long_period,'long_period',2.0000000000000004)
        if self.short_period>=self.long_period:
            raise ValueError("short_period must be smaller than long_period")
        self.smoothing = _real(smoothing,'smoothing')
        self.tolerance = _real(tolerance,'tolerance')
        self.warmup_bars = self.window-1
        self.reset()

    def reset(self):
        self.samples_seen = 0
        self.prices = deque(maxlen=self.window)

    @property
    def is_primed(self):
        return len(self.prices)==self.window

    def update(self, price):
        price = _sample(price)
        if math.isnan(price):
            return PolarityOutput(*(math.nan,)*5)
        prices = (list(self.prices)+[price])[-self.window:]
        models = _fit_window(prices,self.smoothing,(self.short_period,self.long_period)) if len(prices)==self.window else ()
        output = PolarityOutput(*(math.nan,)*5)
        if models:
            short,long = models
            s1,s2 = cycle_slope(short,self.window-1),cycle_slope(long,self.window-1)
            output = PolarityOutput(cycle_polarity(s1,s2,self.tolerance),
                                    cycle_value(short,self.window-1),cycle_value(long,self.window-1),s1,s2)
        self.prices.append(price)
        self.samples_seen += 1
        return output

Engine = CyclePolarity
