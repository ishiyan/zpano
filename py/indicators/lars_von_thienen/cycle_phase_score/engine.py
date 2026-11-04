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
class PhaseOutput(NamedTuple):
    value: float
    period: float
    phase: float

class CyclePhaseScore(RollingScanner):
    def update(self, price):
        cycle = super().update(price).cycles[0]
        return PhaseOutput(phase_state(cycle.phase), cycle.period, cycle.phase)

Engine = CyclePhaseScore
