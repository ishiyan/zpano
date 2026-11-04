# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; CC BY 4.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
from ..harmonic_cycle.validation import _integer, _real, _sample, _finite
class _ZLEMA:
    """First-price prehistory makes the book's unspecified startup explicit."""
    def __init__(self, length):
        self.alpha = 2.0/(length+1.0)
        self.lag = (length-1)//2
        self.history = deque(maxlen=self.lag+1)
        self.first = self.value = math.nan

    def update(self, price):
        if math.isnan(self.first):
            self.first = self.value = price
            self.history.append(price)
            return price  # book's explicit CurrentBar=1 branch
        history = list(self.history) + [price]
        # Only absent prehistory is extended; accepted NaNs never reach here.
        lagged = history[-1-self.lag] if len(history)>self.lag else self.first
        adjusted = _finite(2.0*price-lagged)
        new = _finite(self.alpha*adjusted+(1.0-self.alpha)*self.value)
        self.history.append(price)
        self.value = new
        return new
