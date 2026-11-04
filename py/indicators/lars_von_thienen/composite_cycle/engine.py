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

class CompositeOutput(NamedTuple):
    value: float
    slope: float
    next_value: float

class CompositeCycle:
    """Rolling GDFT fits at CONFIGURED periods, then current superposition.

    The sum is disclosed by the book. This price-input fitting wrapper is an
    explicit research adaptation; it does not invent proprietary peak rules.
    """
    def __init__(self, window=100, periods=(20.0,40.0), smoothing=1600.0):
        self.window = _integer(window,'window',6)
        try:
            self.periods = tuple(_real(p,'period',2.0000000000000004) for p in periods)
        except TypeError as error:
            raise ValueError("periods must be a nonempty finite sequence") from error
        if not self.periods or len(set(self.periods)) != len(self.periods):
            raise ValueError("periods must be nonempty and distinct")
        self.smoothing = _real(smoothing,'smoothing')
        self.warmup_bars = self.window-1
        self.reset()

    def reset(self):
        self.samples_seen = 0
        self.prices = deque(maxlen=self.window)
        self.models = ()

    @property
    def is_primed(self):
        return len(self.prices)==self.window

    def project(self, bars_ahead=0):
        # Read-only: forecasts must never be fed back as observed prices.
        ahead = _real(bars_ahead,'bars_ahead')
        return composite_value(self.models,self.window-1+ahead) if self.is_primed else math.nan

    def update(self, price):
        price = _sample(price)
        if math.isnan(price):
            return CompositeOutput(*(math.nan,)*3)
        prices = (list(self.prices)+[price])[-self.window:]
        models = _fit_window(prices,self.smoothing,self.periods) if len(prices)==self.window else ()
        output = CompositeOutput(*(math.nan,)*3)
        if models:
            value = composite_value(models,self.window-1)
            slope = _finite(sum(cycle_slope(m,self.window-1) for m in models))
            output = CompositeOutput(value,slope,composite_value(models,self.window))
        self.prices.append(price)
        self.samples_seen += 1
        self.models = models
        return output

Engine = CompositeCycle
