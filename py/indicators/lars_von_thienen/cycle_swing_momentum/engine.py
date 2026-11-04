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
from ..adaptive_bands.bands import _grid_bands
class CycleSwingOutput(NamedTuple):
    value: float
    lower: float
    upper: float
    score: float

class CycleSwing:
    """Difference of two HP ENDPOINTS, as in the available CSI source.

    Source: Cycle_Swing_Momentum.pinescript, Lars von Thienen, CC BY 4.0.
    No inferred cycle scanner or second price derivative is substituted.
    Band priming and source extrema repair are documented adaptations.
    """
    def __init__(self, window=50, fast_smoothing=160.0, slow_smoothing=1600.0,
                 band_memory=34, leveling=10.0):
        self.window = _integer(window, 'window', 6)
        self.fast_smoothing = _real(fast_smoothing, 'fast_smoothing')
        self.slow_smoothing = _real(slow_smoothing, 'slow_smoothing')
        if self.fast_smoothing >= self.slow_smoothing:
            raise ValueError("fast_smoothing must be smaller than slow_smoothing")
        self.band_memory = _integer(band_memory, 'band_memory', 2)
        self.leveling = _real(leveling, 'leveling')
        if not 0.0 < self.leveling < 50.0:
            raise ValueError("leveling must be in (0,50)")
        self.value_warmup_bars = self.window-1
        self.warmup_bars = self.window+self.band_memory-2
        self.reset()

    def reset(self):
        self.samples_seen = 0
        self.prices = deque(maxlen=self.window)
        self.values = deque(maxlen=self.band_memory)

    @property
    def is_primed(self):
        return len(self.values) == self.band_memory

    def update(self, price):
        price = _sample(price)
        if math.isnan(price):
            return CycleSwingOutput(*(math.nan,)*4)
        # Candidate window allows HP arithmetic rejection before this commit.
        prices = (list(self.prices)+[price])[-self.window:]
        value = math.nan
        if len(prices) == self.window:
            fast = hp_trend(prices, self.fast_smoothing)[-1]
            slow = hp_trend(prices, self.slow_smoothing)[-1]
            value = _finite(fast-slow)
        self.prices.append(price)
        self.samples_seen += 1
        if math.isnan(value):
            return CycleSwingOutput(*(math.nan,)*4)
        self.values.append(value)
        if not self.is_primed:
            return CycleSwingOutput(value, math.nan, math.nan, math.nan)
        lower, upper = _grid_bands(self.values, self.leveling)
        # Preserve source order: upper test wins if both bands coincide.
        score = 1.0 if value >= upper else -1.0 if value <= lower else 0.0
        return CycleSwingOutput(value, lower, upper, score)

Engine = CycleSwing
