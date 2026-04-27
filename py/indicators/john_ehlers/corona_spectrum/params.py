"""Parameters for the CoronaSpectrum indicator."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class Params:
    """Configuration for the Corona Spectrum indicator."""
    min_raster_value: float = 0.0
    max_raster_value: float = 0.0
    min_parameter_value: float = 0.0
    max_parameter_value: float = 0.0
    high_pass_filter_cutoff: int = 0
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> Params:
    """Return default Ehlers parameters."""
    return Params(
        min_raster_value=6.0,
        max_raster_value=20.0,
        min_parameter_value=6.0,
        max_parameter_value=30.0,
        high_pass_filter_cutoff=30,
    )
