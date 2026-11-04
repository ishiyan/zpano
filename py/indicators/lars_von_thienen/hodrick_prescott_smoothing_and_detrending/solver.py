# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; CC BY 4.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
from ..harmonic_cycle.validation import _integer, _real, _sample, _finite
def hp_trend(values, smoothing):
    """Complete source pentadiagonal HP solver, oldest observation first.

    Source: Lars von Thienen, hp_filter.cpp, CC BY 4.0 (2017).
    The forward loop is retained in its published operation order. The
    backward loop returns all trends; the last forward value is its endpoint.
    Unlike a streaming wrapper, this BATCH helper can revise old estimates.
    """
    n = len(values)
    if n < 6:
        raise ValueError("HP source solver requires at least six observations")
    smoothing = _real(smoothing, 'smoothing')
    data = [_sample(value) for value in values]
    if any(math.isnan(value) for value in data):
        raise ValueError("HP batch input must be finite")
    # Diagonals of I + lambda * D^T D, with natural endpoint conditions.
    a = [6.0*smoothing+1.0] * n
    b = [-4.0*smoothing] * n
    c = [smoothing] * n
    a[0] = a[-1] = 1.0+smoothing
    a[1] = a[-2] = 1.0+5.0*smoothing
    b[0] = b[-2] = -2.0*smoothing
    b[-1] = c[-1] = c[-2] = 0.0
    h1 = h2 = h3 = h4 = h5 = 0.0
    hh2 = hh3 = hh5 = 0.0
    for i in range(n):
        # In-place arrays become elimination coefficients, not price history.
        z = _finite(a[i]-h4*h1-hh5*hh2)
        if z <= 0.0:
            raise ValueError("HP elimination lost positive definiteness")
        hb = b[i]
        hh1 = h1
        h1 = (hb-h4*h2)/z
        b[i] = h1
        hc = c[i]
        hh2 = h2
        h2 = hc/z
        c[i] = h2
        a[i] = _finite((data[i]-hh3*hh5-h3*h4)/z)
        hh3 = h3
        h3 = a[i]
        h4 = hb-h5*hh1
        hh5 = h5
        h5 = hc
    output = [0.0]*n
    h2 = 0.0
    h1 = a[-1]
    output[-1] = h1
    for i in range(n-1, 0, -1):
        output[i-1] = _finite(a[i-1]-b[i-1]*h1-c[i-1]*h2)
        h2, h1 = h1, output[i-1]
    return output

def _cycle_data(values, smoothing):
    # smoothing == 0 intentionally disables detrending; it does NOT request
    # the HP residual at lambda=0, which would be identically zero.
    if smoothing == 0.0:
        return list(values)
    trend = hp_trend(values, smoothing)
    return [_finite(value-base) for value, base in zip(values, trend)]
