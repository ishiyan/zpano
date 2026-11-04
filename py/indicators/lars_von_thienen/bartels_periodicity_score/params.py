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
class BartelsPeriodicityScoreParams:
    """Parameters for bartels-periodicity-score; defaults match the prepared reference."""
    period: int = 20
    """Configured period; default 20."""
    segments: int = 5
    """Configured segments; default 5."""
    smoothing: float = 0.0
    """Configured smoothing; default 0.0."""
    bar_component: Optional[BarComponent] = None
    """The bar component. None uses ClosePrice and omits it from the mnemonic."""
    quote_component: Optional[QuoteComponent] = None
    """The quote component. None uses MidPrice and omits it from the mnemonic."""
    trade_component: Optional[TradeComponent] = None
    """The trade component. None uses Price and omits it from the mnemonic."""


def default_params() -> BartelsPeriodicityScoreParams:
    """Return the prepared reference defaults."""
    return BartelsPeriodicityScoreParams()
