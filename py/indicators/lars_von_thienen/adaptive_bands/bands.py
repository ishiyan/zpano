# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; CC BY 4.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
def _grid_bands(values, leveling):
    """Source's 101-level search, NOT an interpolated quantile function."""
    # Independent min/max fixes the source's else-if extremum bug on ramps.
    low, high = min(values), max(values)
    step = (high-low)/100.0
    fraction = leveling/100.0
    lower, upper = low, high
    for k in range(101):
        test = low + step*k
        below = sum(1 for value in values if value < test)
        if below/len(values) >= fraction:
            lower = test
            break
    for k in range(101):
        test = high - step*k
        above = sum(1 for value in values if value >= test)
        if above/len(values) >= fraction:
            upper = test
            break
    # A constant window returns its sole value even though strict '<' cannot
    # reach a positive fraction in the lower search. No sentinel escapes.
    return lower, upper
