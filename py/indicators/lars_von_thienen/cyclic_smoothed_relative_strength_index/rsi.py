# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; CC BY 4.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
from ..harmonic_cycle.validation import _integer, _real, _sample, _finite
class _WilderRSI:
    """SMA seed over L actual close changes, followed by Wilder's RMA."""
    def __init__(self, length):
        self.length = length
        self.previous = math.nan
        self.count = 0
        self.up = self.down = 0.0

    def update(self, price):
        # No fabricated zero change is counted before the first input price.
        if math.isnan(self.previous):
            self.previous = price
            return math.nan
        change = _finite(price-self.previous)
        gain, loss = max(change, 0.0), max(-change, 0.0)
        if self.count < self.length:
            self.up = _finite(self.up+gain)
            self.down = _finite(self.down+loss)
            self.count += 1
            self.previous = price
            if self.count < self.length:
                return math.nan
            self.up /= self.length
            self.down /= self.length
        else:
            # Preserve the source's multiply-add-divide operation order.
            self.up = _finite((self.up*(self.length-1) + gain)/self.length)
            self.down = _finite((self.down*(self.length-1) + loss)/self.length)
            self.previous = price
        # The Pine conditional tests down == 0 FIRST: a flat seed is 100.
        if self.down == 0.0:
            return 100.0
        if self.up == 0.0:
            return 0.0
        return 100.0 - 100.0/(1.0+self.up/self.down)
