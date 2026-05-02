"""Jurik composite fractal behavior index parameters."""

from dataclasses import dataclass
from typing import Optional
from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class JurikCompositeFractalBehaviorIndexParams:
    """Parameters for the Jurik composite fractal behavior index indicator."""
    fractal_type: int = 1
    smooth: int = 10
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> JurikCompositeFractalBehaviorIndexParams:
    """Return default parameters."""
    return JurikCompositeFractalBehaviorIndexParams()
