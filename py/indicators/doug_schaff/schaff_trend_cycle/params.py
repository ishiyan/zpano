"""Schaff Trend Cycle parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class SchaffTrendCycleParams:
    """Parameters to create an instance of the Schaff Trend Cycle indicator."""

    fast: int = 23
    slow: int = 50
    tclen: int = 10
    factor: float = 0.5
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> SchaffTrendCycleParams:
    """Returns default parameters for the Schaff Trend Cycle indicator."""
    return SchaffTrendCycleParams()
