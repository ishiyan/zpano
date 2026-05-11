"""Fuzzy signal functions for technical indicator interpretation.

Standalone functions that take raw indicator values and return fuzzy
membership degrees ∈ [0, 1].  Zero dependency on the indicators module.

Signal types:
- **Threshold**: overbought/oversold conditions
- **Crossover**: line crossings and threshold crossings
- **Band**: Bollinger-style band touches and containment
- **Histogram**: sign changes in oscillator histograms
- **Compose**: combine multiple signals with fuzzy logic operators
"""
from __future__ import annotations

from .threshold import (
    mu_above,
    mu_below,
    mu_overbought,
    mu_oversold,
)
from .crossover import (
    mu_crosses_above,
    mu_crosses_below,
    mu_line_crosses_above,
    mu_line_crosses_below,
)
from .band import (
    mu_above_band,
    mu_below_band,
    mu_between_bands,
)
from .histogram import (
    mu_turns_positive,
    mu_turns_negative,
)
from .compose import (
    signal_and,
    signal_or,
    signal_not,
    signal_strength,
)

__all__ = [
    'mu_above', 'mu_below', 'mu_overbought', 'mu_oversold',
    'mu_crosses_above', 'mu_crosses_below',
    'mu_line_crosses_above', 'mu_line_crosses_below',
    'mu_above_band', 'mu_below_band', 'mu_between_bands',
    'mu_turns_positive', 'mu_turns_negative',
    'signal_and', 'signal_or', 'signal_not', 'signal_strength',
]
