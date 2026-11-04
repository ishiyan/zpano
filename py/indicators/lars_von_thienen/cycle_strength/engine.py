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

class StrengthOutput(NamedTuple):
    strength: float
    amplitude: float
    phase: float

class CycleStrength:
    """Price-driven wrapper of A/P using a CONFIGURED period, not a scanner."""
    def __init__(self, window=100, period=20.0, smoothing=1600.0):
        self.window = _integer(window,'window',6)
        self.period = _real(period,'period',2.0000000000000004)
        self.smoothing = _real(smoothing,'smoothing')
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
            return StrengthOutput(*(math.nan,)*3)
        prices = (list(self.prices)+[price])[-self.window:]
        model = _fit_window(prices,self.smoothing,(self.period,))[0] if len(prices)==self.window else None
        self.prices.append(price)
        self.samples_seen += 1
        if model is None:
            return StrengthOutput(*(math.nan,)*3)
        return StrengthOutput(cycle_strength(model.amplitude,self.period),model.amplitude,model.phase)

Engine = CycleStrength
