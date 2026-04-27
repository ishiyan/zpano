"""Parameters for the AutoCorrelation Periodogram."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class Params:
    """Parameters for creating an AutoCorrelationPeriodogram instance."""

    min_period: int = 0
    max_period: int = 0
    averaging_length: int = 0
    disable_spectral_squaring: bool = False
    disable_smoothing: bool = False
    disable_automatic_gain_control: bool = False
    automatic_gain_control_decay_factor: float = 0.0
    fixed_normalization: bool = False
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> Params:
    """Returns a Params with Ehlers defaults."""
    return Params(
        min_period=10,
        max_period=48,
        averaging_length=3,
        automatic_gain_control_decay_factor=0.995,
    )
