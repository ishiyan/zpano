"""Signal composition utilities.

Combine multiple fuzzy signals using t-norms, s-norms, and negation.
These are thin wrappers over ``py.fuzzy.operators`` with signal-domain
naming for readability.
"""
from __future__ import annotations

from ..fuzzy import t_product_all, s_probabilistic, f_not


def signal_and(*signals: float) -> float:
    """Combine signals with product t-norm (fuzzy AND).

    All signals must be high for the result to be high.  Each weak
    signal drags the combined confidence down proportionally.

    Args:
        *signals: Two or more membership degrees ∈ [0, 1].

    Returns:
        Combined membership degree ∈ [0, 1].
    """
    return t_product_all(*signals)


def signal_or(a: float, b: float) -> float:
    """Combine two signals with probabilistic s-norm (fuzzy OR).

    Result is high when either signal is high.  Equivalent to
    ``a + b - a*b``.

    Args:
        a: First membership degree ∈ [0, 1].
        b: Second membership degree ∈ [0, 1].

    Returns:
        Combined membership degree ∈ [0, 1].
    """
    return s_probabilistic(a, b)


def signal_not(signal: float) -> float:
    """Negate a signal (fuzzy complement).

    Returns ``1 - signal``.

    Args:
        signal: Membership degree ∈ [0, 1].

    Returns:
        Negated membership degree ∈ [0, 1].
    """
    return f_not(signal)


def signal_strength(signal: float, min_strength: float = 0.5) -> float:
    """Filter weak signals below *min_strength* to zero.

    Signals at or above the threshold pass through unchanged.
    This is analogous to an alpha-cut but preserves the continuous
    value for strong signals rather than discretising.

    Args:
        signal: Membership degree ∈ [0, 1].
        min_strength: Minimum acceptable confidence.

    Returns:
        *signal* if ``signal >= min_strength``, else 0.0.
    """
    return signal if signal >= min_strength else 0.0
