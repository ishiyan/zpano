"""Jurik adaptive zero lag velocity parameters."""

from dataclasses import dataclass
from typing import Optional
from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class JurikAdaptiveZeroLagVelocityParams:
    """Parameters for the Jurik adaptive zero lag velocity indicator."""
    lo_length: int = 5
    hi_length: int = 30
    sensitivity: float = 1.0
    period: float = 3.0
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> JurikAdaptiveZeroLagVelocityParams:
    """Return default parameters."""
    return JurikAdaptiveZeroLagVelocityParams()
