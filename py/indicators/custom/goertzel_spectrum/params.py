"""Parameters for the GoertzelSpectrum indicator."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class Params:
    """Configuration for the Goertzel Spectrum indicator."""
    length: int = 0
    min_period: float = 0.0
    max_period: float = 0.0
    spectrum_resolution: int = 0
    is_first_order: bool = False
    disable_spectral_dilation_compensation: bool = False
    disable_automatic_gain_control: bool = False
    automatic_gain_control_decay_factor: float = 0.0
    fixed_normalization: bool = False
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> Params:
    """Return default parameters."""
    return Params(
        length=64,
        min_period=2.0,
        max_period=64.0,
        spectrum_resolution=1,
        automatic_gain_control_decay_factor=0.991,
    )
