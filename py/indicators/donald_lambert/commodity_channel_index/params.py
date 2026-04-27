"""Commodity Channel Index parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent

DEFAULT_INVERSE_SCALING_FACTOR = 0.015


@dataclass
class CommodityChannelIndexParams:
    """Parameters to create an instance of the commodity channel index indicator."""

    length: int = 20
    inverse_scaling_factor: float = DEFAULT_INVERSE_SCALING_FACTOR
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> CommodityChannelIndexParams:
    """Returns default parameters for the commodity channel index indicator."""
    return CommodityChannelIndexParams()
