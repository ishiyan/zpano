# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; CC BY 4.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
from ..harmonic_cycle.validation import _integer, _real, _sample, _finite
class BartelsOutput(NamedTuple):
    score: float
    genuine_percent: float
    segments: float

def bartels_score(values, period, max_segments=5, min_segments=2):
    """Harmonic-segment consistency score from BartelsScore.cpp, CC BY 4.0.

    Full-precision pi replaces 3.1415926535. Invalid/no-energy evidence is
    NaN, not the source's misleading no-segment raw zero (100% genuine).
    period is restricted to an integer so phase and segmentation agree.
    """
    period = _integer(period, 'period', 3)
    max_segments = _integer(max_segments, 'max_segments', 2)
    min_segments = _integer(min_segments, 'min_segments', 2)
    if min_segments>max_segments:
        raise ValueError("min_segments must not exceed max_segments")
    data = [_sample(value) for value in values]
    if any(math.isnan(value) for value in data):
        raise ValueError("Bartels batch data must be finite")
    segments = min(len(data)//period,max_segments)
    if segments < min_segments:
        return BartelsOutput(math.nan,math.nan,float(segments))
    total_a = total_b = total_energy = 0.0
    for segment in range(segments):
        a = b = 0.0
        for j in range(period):
            x = segment*period+j
            # Source phase starts at x+1; preserve that convention and order.
            angle = (360.0/period*(x+1))/180.0*math.pi
            a += math.sin(angle)*data[x]
            b += math.cos(angle)*data[x]
        total_a += a
        total_b += b
        total_energy = _finite(total_energy+a*a+b*b)
    if total_energy == 0.0:
        return BartelsOutput(math.nan,math.nan,float(segments))
    average_a, average_b = total_a/segments,total_b/segments
    amplitude = math.sqrt(average_a*average_a+average_b*average_b)
    rms_amplitude = math.sqrt(total_energy/segments)
    reference = rms_amplitude/math.sqrt(segments)
    ratio = amplitude/reference
    # exp(-r^2) avoids overflow in the source's equivalent 1/exp(r^2).
    score = math.exp(-ratio*ratio)
    return BartelsOutput(score,100.0*(1.0-score),float(segments))
