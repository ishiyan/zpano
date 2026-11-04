# SPDX-License-Identifier: MPL-2.0
# Source: Detrended Rhythm Oscillator, © WhenToTrade / Lars von Thienen.
# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; MPL 2.0; ZigZag helpers credit LonesomeTheBlue.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
from ..harmonic_cycle.validation import _integer, _real, _sample, _finite
def _median(values):
    # Ordinary median, including the arithmetic mean of two central values.
    ordered = sorted(values)
    n = len(ordered)
    if not n:
        return math.nan
    middle = n // 2
    return ordered[middle] if n % 2 else (ordered[middle-1] + ordered[middle]) / 2.0

class _EMA:
    """First-price seed; recursive output starts on the first finite sample."""
    def __init__(self, length):
        self.alpha = 2.0 / (length + 1.0)
        self.value = math.nan

    def update(self, value):
        # Every term reads the previous accepted observation's state.
        new = value if math.isnan(self.value) else (
            self.alpha * value + (1.0-self.alpha) * self.value)
        self.value = _finite(new)
        return self.value

class RhythmOutput(NamedTuple):
    value: float
    direction: float
    cycle_length: float
    pivot_value: float
    pivot_offset: float

class DetrendedRhythmOscillator:
    """PPO-driven rolling-extremum ZigZag and completed-cycle median.

    Source: dro.pinesctipt, Lars von Thienen (MPL 2.0). This numerical
    extraction removes dependence on chart labels being enabled and fixes
    zero-valued-extremum truthiness. Current pivot outputs are PROVISIONAL.
    """
    def __init__(self, fast_length=3, slow_length=50, zig_period=30,
                 cycle_memory=15, detrend=True):
        self.fast_length = _integer(fast_length, 'fast_length')
        self.slow_length = _integer(slow_length, 'slow_length', 2)
        if self.fast_length >= self.slow_length:
            raise ValueError("fast_length must be smaller than slow_length")
        self.zig_period = _integer(zig_period, 'zig_period', 5)
        self.cycle_memory = _integer(cycle_memory, 'cycle_memory', 3)
        if type(detrend) is not bool:
            raise ValueError("detrend must be bool")
        self.detrend = detrend
        # A finite PPO exists immediately, but hide its initialization period.
        self.warmup_bars = self.slow_length-1 if detrend else 0
        self.reset()

    def reset(self):
        self.samples_seen = 0
        self.fast, self.slow = _EMA(self.fast_length), _EMA(self.slow_length)
        self.window = deque(maxlen=self.zig_period)
        self.pivots = []  # newest first; (value, accepted zero-based index)
        self.direction = 0
        self.last_cycle_length = math.nan

    @property
    def is_primed(self):
        return self.samples_seen > self.warmup_bars

    @property
    def is_cycle_primed(self):
        return self.is_primed and not math.isnan(self.last_cycle_length)

    def update(self, price):
        price = _sample(price)
        if math.isnan(price):
            return RhythmOutput(*(math.nan,)*5)
        if self.detrend:
            fast, slow = self.fast.update(price), self.slow.update(price)
            # PPO is undefined at a zero slow EMA. No made-up ratio is used.
            value = math.nan if slow == 0.0 else _finite(100.0*(fast-slow)/slow)
        else:
            value = price
        index = self.samples_seen
        self.samples_seen += 1
        if math.isnan(value):
            return RhythmOutput(*(math.nan,)*5)
        self.window.append(value)
        high, low = value == max(self.window), value == min(self.window)
        # If both extrema occur (constant window), retain the old direction.
        direction = 1 if high and not low else -1 if low and not high else self.direction
        if direction and (high or low):
            if direction != self.direction or not self.pivots:
                self.pivots.insert(0, (value, index))
                # At most memory completed full-cycle measurements are needed.
                del self.pivots[self.cycle_memory+3:]
            elif (direction == 1 and value > self.pivots[0][0]) or (
                    direction == -1 and value < self.pivots[0][0]):
                self.pivots[0] = (value, index)
        self.direction = direction
        # Pivot 0 is still moving. Use only pairs of CONFIRMED like-type pivots.
        # Five pivots give three completed legs and two full-cycle distances.
        cycles = [float(self.pivots[i][1]-self.pivots[i+2][1])
                  for i in range(1, len(self.pivots)-2)] if len(self.pivots)>=5 else []
        self.last_cycle_length = _median(cycles[-self.cycle_memory:])
        if not self.is_primed:
            return RhythmOutput(*(math.nan,)*5)
        pivot_value = self.pivots[0][0] if self.pivots else math.nan
        offset = float(self.pivots[0][1]-index) if self.pivots else math.nan
        return RhythmOutput(value, float(direction), self.last_cycle_length, pivot_value, offset)

Engine = DetrendedRhythmOscillator
