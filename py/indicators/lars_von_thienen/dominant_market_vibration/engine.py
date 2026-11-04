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
class DominantMarketVibration(RollingScanner):
    def update(self, price):
        # The dominant vibration is the first accepted A/P rank, not maximum A.
        return super().update(price).cycles[0]

Engine = DominantMarketVibration
