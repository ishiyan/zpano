"""FractalAdaptiveMovingAverage parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class FractalAdaptiveMovingAverageParams:
    """Parameters for the FractalAdaptiveMovingAverage indicator."""
    length: int = 16
    slowest_smoothing_factor: float = 0.01
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> FractalAdaptiveMovingAverageParams:
    """Returns default FractalAdaptiveMovingAverage parameters."""
    return FractalAdaptiveMovingAverageParams()
