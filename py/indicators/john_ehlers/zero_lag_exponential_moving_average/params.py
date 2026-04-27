"""Zero-lag exponential moving average parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class ZeroLagExponentialMovingAverageParams:
    """Parameters for the zero-lag exponential moving average (ZEMA)."""

    smoothing_factor: float = 0.25
    velocity_gain_factor: float = 0.5
    velocity_momentum_length: int = 3
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> ZeroLagExponentialMovingAverageParams:
    """Returns default parameters for the ZEMA."""
    return ZeroLagExponentialMovingAverageParams()
