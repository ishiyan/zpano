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
class DominantMarketVibrationParams:
    """Parameters for dominant-market-vibration; defaults match the prepared reference."""
    window: int = 200
    """Configured window; default 200."""
    min_period: int = 5
    """Configured min period; default 5."""
    max_period: int = 80
    """Configured max period; default 80."""
    smoothing: float = 1600.0
    """Configured smoothing; default 1600.0."""
    genuine_threshold: float = 49.0
    """Configured genuine threshold; default 49.0."""
    phase_cycles: float = 2.0
    """Configured phase cycles; default 2.0."""
    top_k: int = 3
    """Configured top k; default 3."""
    peak_distance: int = 2
    """Configured peak distance; default 2."""
    max_segments: int = 5
    """Configured max segments; default 5."""
    bar_component: Optional[BarComponent] = None
    """The bar component. None uses ClosePrice and omits it from the mnemonic."""
    quote_component: Optional[QuoteComponent] = None
    """The quote component. None uses MidPrice and omits it from the mnemonic."""
    trade_component: Optional[TradeComponent] = None
    """The trade component. None uses Price and omits it from the mnemonic."""


def default_params() -> DominantMarketVibrationParams:
    """Return the prepared reference defaults."""
    return DominantMarketVibrationParams()
