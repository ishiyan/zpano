"""Zero-lag error-correcting exponential moving average parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class ZeroLagErrorCorrectingExponentialMovingAverageParams:
    """Parameters for the zero-lag error-correcting exponential moving average (ZECEMA)."""

    smoothing_factor: float = 0.095
    gain_limit: float = 5.0
    gain_step: float = 0.1
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> ZeroLagErrorCorrectingExponentialMovingAverageParams:
    """Returns default parameters for the ZECEMA."""
    return ZeroLagErrorCorrectingExponentialMovingAverageParams()
