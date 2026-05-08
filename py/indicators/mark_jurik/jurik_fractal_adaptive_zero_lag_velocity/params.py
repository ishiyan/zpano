"""Jurik fractal adaptive zero lag velocity parameters."""

from dataclasses import dataclass
from typing import Optional
from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class JurikFractalAdaptiveZeroLagVelocityParams:
    """Parameters for the Jurik fractal adaptive zero lag velocity indicator."""
    lo_depth: int = 5
    hi_depth: int = 30
    fractal_type: int = 1
    smooth: int = 10
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> JurikFractalAdaptiveZeroLagVelocityParams:
    """Return default parameters."""
    return JurikFractalAdaptiveZeroLagVelocityParams()
