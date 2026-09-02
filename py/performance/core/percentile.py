import math
from typing import Iterable

def percentile(window: Iterable[float], q: float) -> float:
    """
    Compute the q-th percentile using NumPy's ``method="linear"``
    definition.

    Args:
        window: Input data.
        q: Percentile in the range [0, 1].

    Returns:
        The interpolated percentile value.
    """
    if not 0 <= q <= 1:
        raise ValueError("q must be between 0 and 1")

    values = sorted(window)
    if not values:
        raise ValueError("window must not be empty")
    n = len(values)

    if n == 1:
        return values[0]

    idx = q * (n - 1)
    lo = int(idx)

    if lo >= n - 1:
        return values[-1]

    hi = lo + 1
    frac = idx - lo

    return values[lo] + frac * (values[hi] - values[lo])
