"""Parameters for the Discrete Fourier Transform Spectrum."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class Params:
    """Parameters for creating a DiscreteFourierTransformSpectrum instance."""

    length: int = 0
    min_period: float = 0.0
    max_period: float = 0.0
    spectrum_resolution: int = 0
    disable_spectral_dilation_compensation: bool = False
    disable_automatic_gain_control: bool = False
    automatic_gain_control_decay_factor: float = 0.0
    fixed_normalization: bool = False
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> Params:
    """Returns a Params with MBST defaults."""
    return Params(
        length=48,
        min_period=10.0,
        max_period=48.0,
        spectrum_resolution=1,
        automatic_gain_control_decay_factor=0.995,
    )
