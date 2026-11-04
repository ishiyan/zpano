# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; CC BY 4.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
from ..harmonic_cycle.validation import _integer, _real, _sample, _finite
from ..adaptive_bands.bands import _grid_bands
from ..cyclic_smoothed_relative_strength_index.rsi import _WilderRSI
class CyclicSmoothedRSIOutput(NamedTuple):
    value: float
    lower: float
    upper: float

class CyclicSmoothedRSI:
    """Pine dominant-cycle convention, explicit lag floor and complete bands.

    Numerical source: RSI_cyclic_smoothed_v2.pinescript, Lars von Thienen,
    CC BY 4.0. Adaptations: deterministic seeds, independent extrema, full
    window band priming, and documented integer coercion. See description.md.
    """
    def __init__(self, dominant_cycle=20, vibration=10, leveling=10.0):
        self.dominant_cycle = _integer(dominant_cycle, 'dominant_cycle', 10)
        self.vibration = _integer(vibration, 'vibration', 2)
        self.leveling = _real(leveling, 'leveling')
        if not 0.0 < self.leveling < 50.0:
            raise ValueError("leveling must be in (0,50)")
        self.length = self.dominant_cycle//2
        self.lag = (self.vibration-1)//2
        self.memory = 2*self.dominant_cycle
        self.torque = 2.0/(self.vibration+1.0)
        self.value_warmup_bars = self.length+self.lag
        self.warmup_bars = self.value_warmup_bars+self.memory-1
        self.reset()

    def reset(self):
        self.samples_seen = 0
        self.rsi = _WilderRSI(self.length)
        self.rsi_history = deque(maxlen=self.lag+1)
        self.values = deque(maxlen=self.memory)
        self.previous_value = 0.0  # source nz(crsi[1]) before first cRSI

    @property
    def is_primed(self):
        return len(self.values) == self.memory

    def update(self, price):
        price = _sample(price)
        if math.isnan(price):
            return CyclicSmoothedRSIOutput(math.nan, math.nan, math.nan)
        self.samples_seen += 1
        rsi = self.rsi.update(price)
        if math.isnan(rsi):
            return CyclicSmoothedRSIOutput(math.nan, math.nan, math.nan)
        self.rsi_history.append(rsi)
        if len(self.rsi_history) < self.lag+1:
            return CyclicSmoothedRSIOutput(math.nan, math.nan, math.nan)
        value = _finite(self.torque*(2.0*rsi-self.rsi_history[0])
                        +(1.0-self.torque)*self.previous_value)
        self.previous_value = value
        self.values.append(value)
        lower = upper = math.nan
        if self.is_primed:
            lower, upper = _grid_bands(self.values, self.leveling)
        return CyclicSmoothedRSIOutput(value, lower, upper)

Engine = CyclicSmoothedRSI
