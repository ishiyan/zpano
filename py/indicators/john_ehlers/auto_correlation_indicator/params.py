"""Parameters for the AutoCorrelation Indicator."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class Params:
    """Parameters for creating an AutoCorrelationIndicator instance."""

    min_lag: int = 0
    max_lag: int = 0
    smoothing_period: int = 0
    averaging_length: int = 0
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> Params:
    """Returns a Params with Ehlers defaults."""
    return Params(
        min_lag=3,
        max_lag=48,
        smoothing_period=10,
        averaging_length=0,
    )
