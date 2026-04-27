"""Rate of Change Ratio parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class RateOfChangeRatioParams:
    """Parameters to create an instance of the rate of change ratio indicator."""

    length: int = 10
    hundred_scale: bool = False
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> RateOfChangeRatioParams:
    """Returns default parameters for the rate of change ratio indicator."""
    return RateOfChangeRatioParams()
