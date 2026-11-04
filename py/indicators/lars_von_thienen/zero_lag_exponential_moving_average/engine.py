# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; CC BY 4.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
from ..harmonic_cycle.validation import _integer, _real, _sample, _finite
from ..zero_lag_exponential_moving_average.smoother import _ZLEMA
class ZeroLagEMA:
    """Part 2 section 10.3 ZLEMA; not Ehlers' error-correcting gain search."""
    def __init__(self, length=13):
        self.length = _integer(length, 'length')
        self.lag = (self.length-1)//2
        # The explicit first-price prehistory makes output available immediately.
        self.warmup_bars = 0
        self.reset()

    def reset(self):
        self.samples_seen = 0
        self.core = _ZLEMA(self.length)

    @property
    def is_primed(self):
        return self.samples_seen > 0

    def update(self, price):
        price = _sample(price)
        if math.isnan(price):
            return math.nan
        value = self.core.update(price)
        self.samples_seen += 1
        return value

Engine = ZeroLagEMA
