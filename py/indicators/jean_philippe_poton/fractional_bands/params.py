"""Fractional Bands parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class FractionalBandsParams:
    """Parameters to create an instance of the fractional bands indicator."""

    period: int = 30
    price_scale: float = 1.0
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> FractionalBandsParams:
    """Returns default parameters for the fractional bands."""
    return FractionalBandsParams()
