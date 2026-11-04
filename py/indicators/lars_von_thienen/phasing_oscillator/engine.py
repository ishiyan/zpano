# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; CC BY 4.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
from ..harmonic_cycle.validation import _integer, _real, _sample, _finite
from ..cycle_scanner.scanner import RollingScanner, modeled_value
from ..harmonic_cycle.model import phase_state
from ..cycle_phase_score.engine import CyclePhaseScore
class PhasingOutput(NamedTuple):
    value: float
    short_score: float
    long_score: float
    short_period: float
    long_period: float

class PhasingOscillator:
    """Independently detect in disjoint short/long ranges, then sum states."""
    lower, upper = -150.0, 150.0

    def __init__(self, window=200, short_min=5, short_max=30, long_min=31,
                 long_max=80, smoothing=1600.0, genuine_threshold=49.0,
                 phase_cycles=2.0):
        _integer(short_max, 'short_max', 3)
        _integer(long_min, 'long_min', 3)
        if short_max >= long_min:
            raise ValueError('short and long search ranges must be disjoint and ordered')
        self.short = CyclePhaseScore(window, short_min, short_max, smoothing,
                                     genuine_threshold, phase_cycles, 1)
        self.long = CyclePhaseScore(window, long_min, long_max, smoothing,
                                    genuine_threshold, phase_cycles, 1)
        self.warmup_bars = window-1

    def reset(self):
        self.short.reset()
        self.long.reset()

    @property
    def is_primed(self):
        return self.short.is_primed and self.long.is_primed

    def update(self, price):
        # Validate before either scanner sees the sample. Both have the same
        # window; an ordinary missing sample does not advance either history.
        price = _sample(price)
        short, long = self.short.update(price), self.long.update(price)
        return PhasingOutput(short.value+long.value, short.value, long.value,
                             short.period, long.period)

Engine = PhasingOscillator
