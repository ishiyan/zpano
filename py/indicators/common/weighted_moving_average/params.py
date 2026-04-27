"""Weighted moving average parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class WeightedMovingAverageParams:
    """Parameters to create an instance of the weighted moving average indicator."""

    length: int = 20
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> WeightedMovingAverageParams:
    """Returns default parameters for the weighted moving average."""
    return WeightedMovingAverageParams()
