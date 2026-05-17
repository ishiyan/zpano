"""Rescaled fractal adaptive simple moving average parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class RescaledFractalAdaptiveSimpleMovingAverageParams:
    """Parameters to create an instance of the rescaled fractal adaptive simple moving average indicator."""

    period: int = 64
    normal_speed: int = 30
    price_scale: float = 1.0
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> RescaledFractalAdaptiveSimpleMovingAverageParams:
    """Returns default parameters for the rescaled fractal adaptive simple moving average."""
    return RescaledFractalAdaptiveSimpleMovingAverageParams()
