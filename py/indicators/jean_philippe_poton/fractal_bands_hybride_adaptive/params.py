"""Fractal Bands Hybride Adaptive parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class FractalBandsHybrideAdaptiveParams:
    """Parameters to create an instance of the fractal bands hybride adaptive indicator."""

    period: int = 30
    normal_speed_fallback: int = 30
    alpha: float = 2.0
    nyquist: float = 0.5
    alpha_hp: float = 0.07
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> FractalBandsHybrideAdaptiveParams:
    """Returns default parameters for the fractal bands hybride adaptive."""
    return FractalBandsHybrideAdaptiveParams()
