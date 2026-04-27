"""Standard deviation parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class StandardDeviationParams:
    """Parameters to create an instance of the standard deviation indicator."""

    length: int = 20
    is_unbiased: bool = True
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> StandardDeviationParams:
    """Returns default parameters for the standard deviation indicator."""
    return StandardDeviationParams()
