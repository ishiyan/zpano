"""Jurik wavelet sampler parameters."""

from dataclasses import dataclass
from typing import Optional
from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class JurikWaveletSamplerParams:
    """Parameters for the Jurik wavelet sampler indicator."""
    index: int = 12
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> JurikWaveletSamplerParams:
    """Return default parameters."""
    return JurikWaveletSamplerParams()
