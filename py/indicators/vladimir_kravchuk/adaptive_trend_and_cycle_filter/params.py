"""Adaptive Trend and Cycle Filter parameters."""

from dataclasses import dataclass
from typing import Optional
from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class AdaptiveTrendAndCycleFilterParams:
    """Parameters for the Adaptive Trend and Cycle Filter.

    The ATCF suite has no user-tunable numeric parameters: all five FIR
    filters use fixed coefficient arrays published by Vladimir Kravchuk.
    """
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> AdaptiveTrendAndCycleFilterParams:
    """Return default parameters."""
    return AdaptiveTrendAndCycleFilterParams()
