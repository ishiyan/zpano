# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; CC BY 4.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
def _integer(value, name, minimum=1):
    # A Python bool is an int subclass, but is not a meaningful period.
    if type(value) is not int or value < minimum:
        raise ValueError(f"{name} must be an integer >= {minimum}")
    return value

def _real(value, name, minimum=0.0):
    # Reject booleans and non-finite parameters before allocating state.
    if isinstance(value, bool):
        raise ValueError(f"{name} must be a finite real >= {minimum}")
    try:
        value = float(value)
    except (TypeError, ValueError, OverflowError) as error:
        raise ValueError(f"invalid {name}") from error
    if not math.isfinite(value) or value < minimum:
        raise ValueError(f"{name} must be a finite real >= {minimum}")
    return value

def _sample(value):
    # NaN denotes a skipped observation. It must never enter feedback state.
    # Infinity and booleans are invalid; validation itself changes no state.
    if isinstance(value, bool):
        raise ValueError("price must be finite or NaN")
    try:
        value = float(value)
    except (TypeError, ValueError, OverflowError) as error:
        raise ValueError("price must be finite or NaN") from error
    if math.isinf(value):
        raise ValueError("price must be finite or NaN")
    return value

def _finite(value):
    # Do not serialize or feed back an infinity produced by overflow.
    if not math.isfinite(value):
        raise ValueError("numerical overflow; reset the indicator before reuse")
    return value
