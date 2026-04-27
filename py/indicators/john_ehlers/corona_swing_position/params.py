"""Parameters for the CoronaSwingPosition indicator."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class Params:
    """Configuration for the Corona Swing Position indicator."""
    raster_length: int = 0
    max_raster_value: float = 0.0
    min_parameter_value: float = 0.0
    max_parameter_value: float = 0.0
    high_pass_filter_cutoff: int = 0
    minimal_period: int = 0
    maximal_period: int = 0
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> Params:
    """Return default Ehlers parameters."""
    return Params(
        raster_length=50,
        max_raster_value=20.0,
        min_parameter_value=-5.0,
        max_parameter_value=5.0,
        high_pass_filter_cutoff=30,
        minimal_period=6,
        maximal_period=30,
    )
