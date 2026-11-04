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
class MetonicWilliamsPercentRParams:
    """Parameters for metonic-williams-percent-r; defaults match the prepared reference."""
    length: int = 13
    """Configured length; default 13."""
    delay1: int = 6939
    """Configured delay1; default 6939."""
    delay2: int = 4016
    """Configured delay2; default 4016."""
    delay3: int = 2922
    """Configured delay3; default 2922."""


def default_params() -> MetonicWilliamsPercentRParams:
    """Return the prepared reference defaults."""
    return MetonicWilliamsPercentRParams()
