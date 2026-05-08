"""Arnaud Legoux moving average parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class ArnaudLegouxMovingAverageParams:
    """Parameters to create an instance of the Arnaud Legoux moving average indicator."""

    window: int = 9
    sigma: float = 6.0
    offset: float = 0.85
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> ArnaudLegouxMovingAverageParams:
    """Returns default parameters for the Arnaud Legoux moving average."""
    return ArnaudLegouxMovingAverageParams()
