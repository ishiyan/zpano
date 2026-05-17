"""Fractal adaptive simple moving average parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class FractalAdaptiveSimpleMovingAverageParams:
    """Parameters to create an instance of the fractal adaptive simple moving average indicator."""

    period: int = 30
    normal_speed: int = 20
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> FractalAdaptiveSimpleMovingAverageParams:
    """Returns default parameters for the fractal adaptive simple moving average."""
    return FractalAdaptiveSimpleMovingAverageParams()
