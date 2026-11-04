# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; CC BY 4.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
from dataclasses import dataclass
from typing import Optional
from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent

@dataclass
class PhasingOscillatorParams:
    """Parameters for phasing-oscillator; defaults match the prepared reference."""
    window: int = 200
    """Configured window; default 200."""
    short_min: int = 5
    """Configured short min; default 5."""
    short_max: int = 30
    """Configured short max; default 30."""
    long_min: int = 31
    """Configured long min; default 31."""
    long_max: int = 80
    """Configured long max; default 80."""
    smoothing: float = 1600.0
    """Configured smoothing; default 1600.0."""
    genuine_threshold: float = 49.0
    """Configured genuine threshold; default 49.0."""
    phase_cycles: float = 2.0
    """Configured phase cycles; default 2.0."""
    bar_component: Optional[BarComponent] = None
    """The bar component. None uses ClosePrice and omits it from the mnemonic."""
    quote_component: Optional[QuoteComponent] = None
    """The quote component. None uses MidPrice and omits it from the mnemonic."""
    trade_component: Optional[TradeComponent] = None
    """The trade component. None uses Price and omits it from the mnemonic."""


def default_params() -> PhasingOscillatorParams:
    """Return the prepared reference defaults."""
    return PhasingOscillatorParams()
