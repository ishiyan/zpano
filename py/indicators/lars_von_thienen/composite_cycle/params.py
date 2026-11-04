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
class CompositeCycleParams:
    """Parameters for composite-cycle; defaults match the prepared reference."""
    window: int = 100
    """Configured window; default 100."""
    periods: tuple[float, ...] = (20.0, 40.0)
    """Configured periods; default (20.0, 40.0)."""
    smoothing: float = 1600.0
    """Configured smoothing; default 1600.0."""
    bar_component: Optional[BarComponent] = None
    """The bar component. None uses ClosePrice and omits it from the mnemonic."""
    quote_component: Optional[QuoteComponent] = None
    """The quote component. None uses MidPrice and omits it from the mnemonic."""
    trade_component: Optional[TradeComponent] = None
    """The trade component. None uses Price and omits it from the mnemonic."""


def default_params() -> CompositeCycleParams:
    """Return the prepared reference defaults."""
    return CompositeCycleParams()
