"""Fractal dimension index parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class FractalDimensionIndexParams:
    """Parameters to create an instance of the fractal dimension index indicator."""

    period: int = 30
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> FractalDimensionIndexParams:
    """Returns default parameters for the fractal dimension index."""
    return FractalDimensionIndexParams()
