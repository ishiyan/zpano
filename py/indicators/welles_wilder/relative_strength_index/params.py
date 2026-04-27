from dataclasses import dataclass
from typing import Optional

from py.entities.bar_component import BarComponent
from py.entities.quote_component import QuoteComponent
from py.entities.trade_component import TradeComponent


@dataclass
class RelativeStrengthIndexParams:
    """Parameters for the Relative Strength Index indicator."""
    length: int = 14
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> RelativeStrengthIndexParams:
    """Returns default parameters."""
    return RelativeStrengthIndexParams()
