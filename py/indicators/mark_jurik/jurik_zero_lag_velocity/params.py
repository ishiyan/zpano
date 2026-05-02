"""Jurik zero lag velocity parameters."""

from dataclasses import dataclass
from typing import Optional
from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class JurikZeroLagVelocityParams:
    """Parameters for the Jurik zero lag velocity indicator."""
    depth: int = 10
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> JurikZeroLagVelocityParams:
    """Return default parameters."""
    return JurikZeroLagVelocityParams()
