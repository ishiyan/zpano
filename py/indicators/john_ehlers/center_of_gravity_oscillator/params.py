"""CenterOfGravityOscillator parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class CenterOfGravityOscillatorParams:
    """Parameters for the CenterOfGravityOscillator indicator."""
    length: int = 10
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> CenterOfGravityOscillatorParams:
    """Returns default CenterOfGravityOscillator parameters."""
    return CenterOfGravityOscillatorParams()
