"""Fractal Graph Dimension Index parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class FractalGraphDimensionIndexParams:
    """Parameters to create an instance of the fractal graph dimension index indicator."""

    period: int = 30
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> FractalGraphDimensionIndexParams:
    """Returns default parameters for the fractal graph dimension index."""
    return FractalGraphDimensionIndexParams()
