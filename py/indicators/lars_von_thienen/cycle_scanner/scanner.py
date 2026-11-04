# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; CC BY 4.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
from ..harmonic_cycle.validation import _integer, _real, _sample, _finite
from dataclasses import dataclass, asdict
from ..hodrick_prescott_smoothing_and_detrending.solver import _cycle_data
from ..harmonic_cycle.model import goertzel
from ..bartels_periodicity_score.scorer import bartels_score
@dataclass(frozen=True)
class ScanConfig:
    """Explicit reconstruction parameters, not undisclosed WTT settings.

    Integer wavelengths keep Bartels segmentation and spectral periods equal.
    No dynamic HP adjustment or taper is inferred from the books.
    """
    window: int = 200
    min_period: int = 5
    max_period: int = 80
    smoothing: float = 1600.0
    genuine_threshold: float = 49.0
    phase_cycles: float = 2.0
    top_k: int = 3
    peak_distance: int = 2
    max_segments: int = 5

    def __post_init__(self):
        for name, minimum in (('window', 6), ('min_period', 3),
                              ('max_period', 3), ('top_k', 1),
                              ('peak_distance', 1), ('max_segments', 2)):
            _integer(getattr(self, name), name, minimum)
        for name in ('smoothing', 'genuine_threshold', 'phase_cycles'):
            object.__setattr__(self, name, _real(getattr(self, name), name))
        if self.min_period > self.max_period:
            raise ValueError('min_period must not exceed max_period')
        if self.genuine_threshold > 100:
            raise ValueError('genuine_threshold must be in [0, 100]')
        if self.phase_cycles != 0 and self.phase_cycles < 1:
            raise ValueError('phase_cycles is zero (full window) or >= 1')

class RankedCycle(NamedTuple):
    # phase is normalized SINE phase AT THE CURRENT ANALYSIS ENDPOINT.
    # amplitude/strength describe full-window detection, fit_amplitude describes
    # the separate recent-window waveform. Never rank by the latter by accident.
    period: float
    amplitude: float
    phase: float
    strength: float
    genuine_percent: float
    segments: float
    fit_amplitude: float
    fit_samples: float
    value: float
    slope: float
    eta_top: float
    eta_bottom: float

CYCLE_FIELDS = RankedCycle._fields
EMPTY_CYCLE = RankedCycle(*([math.nan] * len(CYCLE_FIELDS)))

class Heatmap(NamedTuple):
    # Structural match to zpano Heatmap; parameter axis is period in bars.
    time: object
    parameter_first: float
    parameter_last: float
    parameter_resolution: float
    value_min: float
    value_max: float
    values: list[float]

    def is_empty(self):
        return not self.values

class ScanOutput(NamedTuple):
    spectrum: Heatmap
    genuine_percent: tuple[float, ...]
    cycles: tuple[RankedCycle, ...]
    accepted_count: float

def empty_scan(config):
    count = config.max_period - config.min_period + 1
    return ScanOutput(Heatmap(None, float(config.min_period),
                             float(config.max_period), 1.0, math.nan, math.nan,
                             [math.nan] * count),
                      tuple([math.nan] * count),
                      tuple([EMPTY_CYCLE] * config.top_k), math.nan)

def modeled_value(cycle, offset=0.0):
    # offset=0 is the endpoint; negative offsets reconstruct earlier samples.
    return cycle.fit_amplitude * math.sin(cycle.phase + 2.0*math.pi*offset/cycle.period)

def scan_cycles(prices, config):
    """Batch detection/validation/ranking, reusable by rolling and annual tools.

    Window length need not equal config.window here: batch CLI analyzes exactly
    the requested range. RollingScanner alone imposes the fixed-size window.
    """
    prices = [_sample(value) for value in prices]
    if len(prices) < 6 or any(math.isnan(value) for value in prices):
        raise ValueError('scan requires >= 6 finite, oldest-first prices')
    residual = _cycle_data(prices, config.smoothing)
    periods = list(range(config.min_period, config.max_period + 1))
    models = [goertzel(residual, period) for period in periods]
    # Validate ORIGINAL prices, not HP residuals that may contain artificial
    # cycles. Take the most recent m whole cycles, with no zero padding.
    evidence = []
    for period in periods:
        segments = min(len(prices) // period, config.max_segments)
        evidence.append(bartels_score(prices[-segments*period:] if segments else [],
                                      period, config.max_segments, 2))
    amplitudes = [model.amplitude for model in models]
    candidates = []
    for i, (model, score) in enumerate(zip(models, evidence)):
        # Plateau tie: the shortest member wins. Boundary peaks are one-sided;
        # a one-period grid has no neighbors and is a valid candidate.
        left = amplitudes[i-1] if i else -math.inf
        right = amplitudes[i+1] if i+1 < len(amplitudes) else -math.inf
        peak = model.amplitude > 0 and model.amplitude > left and model.amplitude >= right
        if not peak or math.isnan(score.genuine_percent) or score.genuine_percent <= config.genuine_threshold:
            continue
        candidates.append((model, score))
    # Rank the price influence per bar A/P, as disclosed in the whitepaper.
    # Suppression follows validation/ranking: an invalid peak cannot suppress
    # a valid neighbor. Exact strength ties prefer the shorter wavelength.
    candidates.sort(key=lambda item: (-item[0].amplitude/item[0].period, item[0].period))
    accepted = []
    for model, score in candidates:
        if any(abs(model.period - prior.period) < config.peak_distance for prior in accepted):
            continue
        fit_n = (len(prices) if config.phase_cycles == 0 else
                 min(len(prices), max(6, math.ceil(config.phase_cycles * model.period))))
        # Re-detrend the recent raw subset, rather than slicing a full-window
        # residual with different HP endpoint conditions. This is a declared
        # reconstruction of the undisclosed recent-phase procedure.
        fit = model if fit_n == len(prices) else goertzel(
            _cycle_data(prices[-fit_n:], config.smoothing), model.period)
        phase = (fit.phase + 2.0*math.pi*(fit_n-1)/model.period) % (2.0*math.pi)
        omega = 2.0*math.pi/model.period
        value = fit.amplitude * math.sin(phase)
        slope = fit.amplitude * omega * math.cos(phase)
        eta_top = ((math.pi/2.0 - phase) % (2.0*math.pi)) / omega
        eta_bottom = ((3.0*math.pi/2.0 - phase) % (2.0*math.pi)) / omega
        accepted.append(RankedCycle(model.period, model.amplitude, phase,
                                    model.amplitude/model.period,
                                    score.genuine_percent, score.segments,
                                    fit.amplitude, float(fit_n), value, slope,
                                    eta_top, eta_bottom))
    # Unused ranks are NaN; count can legitimately be zero after full priming.
    top = tuple(accepted[:config.top_k])
    top += tuple([EMPTY_CYCLE] * (config.top_k-len(top)))
    spectrum = Heatmap(None, float(config.min_period), float(config.max_period),
                       1.0, min(amplitudes), max(amplitudes), list(amplitudes))
    return ScanOutput(spectrum, tuple(item.genuine_percent for item in evidence),
                      top, float(len(accepted)))

class RollingScanner:
    """Causal rolling pipeline; full N prices prime the analysis, not a cycle.

    Lack of an accepted cycle after priming is a legitimate missing output.
    Validation and calculation happen before the history is committed.
    """
    def __init__(self, window=200, min_period=5, max_period=80, smoothing=1600.0,
                 genuine_threshold=49.0, phase_cycles=2.0, top_k=3,
                 peak_distance=2, max_segments=5):
        self.config = ScanConfig(window, min_period, max_period, smoothing,
                                 genuine_threshold, phase_cycles, top_k,
                                 peak_distance, max_segments)
        self.warmup_bars = window-1
        self.reset()

    def reset(self):
        self.prices = deque(maxlen=self.config.window)
        self.samples_seen = 0

    @property
    def is_primed(self):
        return len(self.prices) == self.config.window

    def update(self, price):
        price = _sample(price)
        if math.isnan(price):
            return empty_scan(self.config)
        prices = (list(self.prices) + [price])[-self.config.window:]
        result = scan_cycles(prices, self.config) if len(prices) == self.config.window else empty_scan(self.config)
        self.prices.append(price)
        self.samples_seen += 1
        return result
