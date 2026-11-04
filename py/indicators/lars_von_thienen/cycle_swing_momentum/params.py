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
class CycleSwingMomentumParams:
    """Parameters for cycle-swing-momentum; defaults match the prepared reference."""
    window: int = 50
    """Configured window; default 50."""
    fast_smoothing: float = 160.0
    """Configured fast smoothing; default 160.0."""
    slow_smoothing: float = 1600.0
    """Configured slow smoothing; default 1600.0."""
    band_memory: int = 34
    """Configured band memory; default 34."""
    leveling: float = 10.0
    """Configured leveling; default 10.0."""
    bar_component: Optional[BarComponent] = None
    """The bar component. None uses ClosePrice and omits it from the mnemonic."""
    quote_component: Optional[QuoteComponent] = None
    """The quote component. None uses MidPrice and omits it from the mnemonic."""
    trade_component: Optional[TradeComponent] = None
    """The trade component. None uses Price and omits it from the mnemonic."""


def default_params() -> CycleSwingMomentumParams:
    """Return the prepared reference defaults."""
    return CycleSwingMomentumParams()
