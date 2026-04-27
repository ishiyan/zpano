"""Parameters for the Comb Band-Pass Spectrum."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class Params:
    """Parameters for creating a CombBandPassSpectrum instance."""

    min_period: int = 0
    max_period: int = 0
    bandwidth: float = 0.0
    disable_spectral_dilation_compensation: bool = False
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
        bandwidth=0.3,
        automatic_gain_control_decay_factor=0.995,
    )
