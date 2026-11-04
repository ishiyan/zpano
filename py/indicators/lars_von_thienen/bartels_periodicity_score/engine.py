# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; CC BY 4.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
from ..harmonic_cycle.validation import _integer, _real, _sample, _finite
from .scorer import BartelsOutput, bartels_score
from ..hodrick_prescott_smoothing_and_detrending.solver import _cycle_data
class BartelsPeriodicityScore:
    """Rolling period*segments buffer; full configured segments prime output."""
    def __init__(self, period=20, segments=5, smoothing=0.0):
        self.period = _integer(period,'period',3)
        self.segments = _integer(segments,'segments',2)
        self.smoothing = _real(smoothing,'smoothing')
        self.window = self.period*self.segments
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
            return BartelsOutput(math.nan,math.nan,math.nan)
        prices = (list(self.prices)+[price])[-self.window:]
        output = (bartels_score(_cycle_data(prices,self.smoothing),self.period,
                                self.segments,self.segments)
                  if len(prices)==self.window else BartelsOutput(math.nan,math.nan,math.nan))
        self.prices.append(price)
        self.samples_seen += 1
        return output

Engine = BartelsPeriodicityScore
