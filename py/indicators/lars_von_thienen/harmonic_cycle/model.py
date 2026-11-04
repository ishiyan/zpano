# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; CC BY 4.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
from .validation import _integer, _real, _sample, _finite
from ..hodrick_prescott_smoothing_and_detrending.solver import _cycle_data
class CycleModel(NamedTuple):
    """phase is a SINE phase at the FIRST sample of the analysis window."""
    period: float
    amplitude: float
    phase: float
    trough_offset: float

def goertzel(values, period):
    """Published single-period GDFT; no taper, centering, or peak selection.

    Source: Lars von Thienen, goertzel.cpp, CC BY 4.0 (2017).
    Standard-library pi replaces PI; every calculation is binary64.
    """
    period = _real(period, 'period', 2.0000000000000004)
    n = len(values)
    if not n:
        raise ValueError("Goertzel needs observations")
    omega = 2.0*math.pi/period
    sine, cosine = math.sin(omega), math.cos(omega)
    coefficient = 2.0*cosine
    q1 = q2 = 0.0
    for value in values:
        value = _sample(value)
        if math.isnan(value):
            raise ValueError("Goertzel batch input must be finite")
        q0 = _finite(coefficient*q1-q2+value)
        q2, q1 = q1, q0
    r1, i1 = q1-q2*cosine, q2*sine
    cn, sn = math.cos((n-1)*omega), math.sin((n-1)*omega)
    real = _finite(r1*cn+i1*sn)
    imaginary = _finite(i1*cn-r1*sn)
    # Keep published normalization. Using hypot changes operation/rounding.
    amplitude = _finite(2.0*math.sqrt(_finite(real*real+imaginary*imaginary))/n)
    if amplitude == 0.0:
        # A zero vector has no phase. atan2(0,0) would invent a sine phase.
        return CycleModel(period, 0.0, math.nan, math.nan)
    phase = math.pi/2.0 + math.atan2(imaginary, real)
    offset = period*(0.75-phase/(2.0*math.pi))
    return CycleModel(period, amplitude, phase, offset)

def cycle_value(model, offset):
    # A zero-amplitude wave is zero even though its phase is unavailable.
    if model.amplitude == 0.0:
        return 0.0
    return model.amplitude*math.sin(2.0*math.pi*offset/model.period+model.phase)

def cycle_slope(model, offset):
    # Analytic derivative is in price units per bar, distinct from A/P.
    if model.amplitude == 0.0:
        return 0.0
    omega = 2.0*math.pi/model.period
    return model.amplitude*omega*math.cos(omega*offset+model.phase)

def phase_state(phase):
    """Continuous reconstruction of the -100..100 phase state.

    This smooth sine mapping reproduces extrema/range, not proprietary text-bin
    values. The books supply examples but no complete phase-region boundaries.
    """
    phase = _sample(phase)
    return math.nan if math.isnan(phase) else 100.0 * math.sin(phase)
def _fit_window(prices, smoothing, periods):
    values = _cycle_data(prices, smoothing)
    return tuple(goertzel(values, period) for period in periods)

def composite_value(models, offset):
    """Linear superposition of disclosed configured sine-cycle models."""
    total = 0.0
    for model in models:
        total += cycle_value(model,offset)
    return _finite(total)
def cycle_strength(amplitude, period):
    """Disclosed A/P ranking metric; it is not the maximum sine derivative."""
    amplitude = _real(amplitude,'amplitude')
    period = _real(period,'period',2.0000000000000004)
    return amplitude/period
def cycle_polarity(short_slope, long_slope, tolerance=0.0):
    """Both rising: +1; both falling: -1; disagreement/flat: 0."""
    short_slope, long_slope = _sample(short_slope),_sample(long_slope)
    tolerance = _real(tolerance,'tolerance')
    if math.isnan(short_slope) or math.isnan(long_slope):
        return math.nan
    if short_slope>tolerance and long_slope>tolerance:
        return 1.0
    if short_slope<-tolerance and long_slope<-tolerance:
        return -1.0
    return 0.0
