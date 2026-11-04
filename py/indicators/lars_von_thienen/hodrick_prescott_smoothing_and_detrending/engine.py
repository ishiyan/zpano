# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; CC BY 4.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
from ..harmonic_cycle.validation import _integer, _real, _sample, _finite
from ..hodrick_prescott_smoothing_and_detrending.solver import hp_trend
class HPOutput(NamedTuple):
    trend: float
    cycle: float

class HodrickPrescott:
    """Fixed-window current endpoint of the source's two-sided batch filter."""
    def __init__(self, window=50, smoothing=1600.0):
        self.window = _integer(window, 'window', 6)
        self.smoothing = _real(smoothing, 'smoothing')
        self.warmup_bars = self.window-1
        self.reset()

    def reset(self):
        self.samples_seen = 0
        self.prices = deque(maxlen=self.window)

    @property
    def is_primed(self):
        return len(self.prices) == self.window

    def update(self, price):
        price = _sample(price)
        if math.isnan(price):
            return HPOutput(math.nan, math.nan)
        prices = (list(self.prices)+[price])[-self.window:]
        trend = hp_trend(prices, self.smoothing)[-1] if len(prices)==self.window else math.nan
        self.prices.append(price)
        self.samples_seen += 1
        return HPOutput(trend, _finite(price-trend) if self.is_primed else math.nan)

Engine = HodrickPrescott
