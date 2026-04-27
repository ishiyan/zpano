"""SuperSmoother parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class SuperSmootherParams:
    """Parameters for the SuperSmoother indicator."""
    shortest_cycle_period: int = 10
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> SuperSmootherParams:
    """Returns default SuperSmoother parameters."""
    return SuperSmootherParams()
