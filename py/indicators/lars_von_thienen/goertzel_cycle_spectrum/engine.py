# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; CC BY 4.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
from ..harmonic_cycle.validation import _integer, _real, _sample, _finite
from ..harmonic_cycle.model import _fit_window, composite_value, cycle_value, cycle_slope, cycle_strength, cycle_polarity
class Heatmap(NamedTuple):
    time: object
    parameter_first: float
    parameter_last: float
    parameter_resolution: float
    value_min: float
    value_max: float
    values: list[float]

    def is_empty(self):
        return not self.values

class SpectrumOutput(NamedTuple):
    heatmap: Heatmap
    phases: tuple[float, ...]
    trough_offsets: tuple[float, ...]
    dominant_period: float
    dominant_amplitude: float
    dominant_phase: float

class GoertzelCycleSpectrum:
    """Configured period-grid scan; largest amplitude wins, no Bartels gate.

    This is the disclosed GDFT component, not the proprietary full scanner.
    Every update's heatmap owns its list; later updates cannot overwrite it.
    """
    def __init__(self, window=128, min_period=5, max_period=50, step=1, smoothing=1600.0):
        self.window = _integer(window, 'window', 6)
        self.min_period = _integer(min_period, 'min_period', 3)
        self.max_period = _integer(max_period, 'max_period', self.min_period)
        self.step = _integer(step, 'step')
        if (self.max_period-self.min_period)%self.step:
            raise ValueError("step must land exactly on max_period")
        self.smoothing = _real(smoothing, 'smoothing')
        self.periods = tuple(range(self.min_period,self.max_period+1,self.step))
        self.warmup_bars = self.window-1
        self.reset()

    def reset(self):
        self.samples_seen = 0
        self.prices = deque(maxlen=self.window)

    @property
    def is_primed(self):
        return len(self.prices)==self.window

    def _empty(self, time):
        heatmap = Heatmap(time, float(self.min_period), float(self.max_period), float(self.step),
                          math.nan, math.nan, [])
        return SpectrumOutput(heatmap, (), (), math.nan, math.nan, math.nan)

    def update(self, price, time=None):
        price = _sample(price)
        if math.isnan(price):
            return self._empty(time)
        prices = (list(self.prices)+[price])[-self.window:]
        models = _fit_window(prices,self.smoothing,self.periods) if len(prices)==self.window else ()
        self.prices.append(price)
        self.samples_seen += 1
        if not models:
            return self._empty(time)
        amplitudes = [model.amplitude for model in models]
        # Strict ascending scan and first maximum preserve shortest-period ties.
        best = max(range(len(models)),key=lambda i:amplitudes[i])
        model = models[best]
        heatmap = Heatmap(time,float(self.min_period),float(self.max_period),float(self.step),
                          min(amplitudes),max(amplitudes),amplitudes)
        dominant = float(model.period) if model.amplitude>0.0 else math.nan
        return SpectrumOutput(heatmap,tuple(m.phase for m in models),
                              tuple(m.trough_offset for m in models),dominant,model.amplitude,model.phase)

Engine = GoertzelCycleSpectrum
