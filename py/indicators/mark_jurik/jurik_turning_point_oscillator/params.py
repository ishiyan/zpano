"""Jurik turning point oscillator parameters."""

from dataclasses import dataclass
from typing import Optional
from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class JurikTurningPointOscillatorParams:
    """Parameters for the Jurik turning point oscillator indicator."""
    length: int = 14
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> JurikTurningPointOscillatorParams:
    """Return default parameters."""
    return JurikTurningPointOscillatorParams()
