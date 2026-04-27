"""Simple moving average parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class SimpleMovingAverageParams:
    """Parameters to create an instance of the simple moving average indicator."""

    length: int = 20
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> SimpleMovingAverageParams:
    """Returns default parameters for the simple moving average."""
    return SimpleMovingAverageParams()
