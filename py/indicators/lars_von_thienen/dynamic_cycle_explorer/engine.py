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
def centered_mean(values, center, length):
    """Integer-centered MA; even length uses two half-weight endpoints.

    An even L spans L+1 observations: average the two length-L windows whose
    centers straddle the integer center. Odd L spans L observations.
    Missing support returns NaN; it is never extrapolated with future values.
    """
    _integer(length, 'length')
    half = length // 2
    if center-half < 0 or center+half >= len(values):
        return math.nan
    if length % 2:
        return math.fsum(values[center-half:center+half+1])/length
    middle = math.fsum(values[center-half+1:center+half])
    return (0.5*values[center-half] + middle + 0.5*values[center+half])/length

class ExplorerOutput(NamedTuple):
    value: float
    period: float
    phase: float
    eta_top: float
    eta_bottom: float
    highlighter: float
    highlighter_index: float
    fitted_at_highlighter: float

class DynamicCycleExplorer(RollingScanner):
    def update(self, price):
        output = super().update(price)
        cycle = output.cycles[0]
        if math.isnan(cycle.period):
            return ExplorerOutput(*([math.nan] * 8))
        long_length = int(cycle.period)
        short_length = max(2, math.floor(cycle.period/2.0 + 0.5))
        delay = long_length // 2
        prices = list(self.prices)
        center = len(prices)-1-delay
        highlighter = centered_mean(prices, center, short_length) - centered_mean(prices, center, long_length)
        # Dynamic selection may change the delay/index between updates. The
        # index is mandatory metadata; do not plot this at the current bar.
        index = float(self.samples_seen-1-delay) if math.isfinite(highlighter) else math.nan
        fitted = modeled_value(cycle, -delay) if math.isfinite(highlighter) else math.nan
        return ExplorerOutput(cycle.value, cycle.period, cycle.phase, cycle.eta_top,
                              cycle.eta_bottom, highlighter, index, fitted)

Engine = DynamicCycleExplorer
