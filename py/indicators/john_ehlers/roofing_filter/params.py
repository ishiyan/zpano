"""RoofingFilter parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class RoofingFilterParams:
    """Parameters for the RoofingFilter indicator."""
    shortest_cycle_period: int = 10
    longest_cycle_period: int = 48
    has_two_pole_highpass_filter: bool = False
    has_zero_mean: bool = False
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> RoofingFilterParams:
    """Returns default RoofingFilter parameters."""
    return RoofingFilterParams()
