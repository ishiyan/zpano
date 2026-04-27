"""Jurik moving average parameters."""

from dataclasses import dataclass
from typing import Optional
from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class JurikMovingAverageParams:
    """Parameters for the Jurik moving average indicator."""
    length: int = 14
    phase: int = 0
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> JurikMovingAverageParams:
    """Return default parameters."""
    return JurikMovingAverageParams()
