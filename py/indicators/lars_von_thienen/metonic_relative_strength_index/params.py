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
class MetonicRelativeStrengthIndexParams:
    """Parameters for metonic-relative-strength-index; defaults match the prepared reference."""
    length: int = 13
    """Configured length; default 13."""
    delay1: int = 6939
    """Configured delay1; default 6939."""
    delay2: int = 4016
    """Configured delay2; default 4016."""
    delay3: int = 2922
    """Configured delay3; default 2922."""
    bar_component: Optional[BarComponent] = None
    """The bar component. None uses ClosePrice and omits it from the mnemonic."""
    quote_component: Optional[QuoteComponent] = None
    """The quote component. None uses MidPrice and omits it from the mnemonic."""
    trade_component: Optional[TradeComponent] = None
    """The trade component. None uses Price and omits it from the mnemonic."""


def default_params() -> MetonicRelativeStrengthIndexParams:
    """Return the prepared reference defaults."""
    return MetonicRelativeStrengthIndexParams()
