"""Triple Exponential Moving Average Oscillator parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class TripleExponentialMovingAverageOscillatorParams:
    """Parameters to create an instance of the TRIX indicator."""

    length: int = 30
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> TripleExponentialMovingAverageOscillatorParams:
    """Returns default parameters for the TRIX indicator."""
    return TripleExponentialMovingAverageOscillatorParams()
