"""Chande Momentum Oscillator parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class ChandeMomentumOscillatorParams:
    """Parameters to create an instance of the Chande momentum oscillator indicator."""

    length: int = 14
    """Length is the length (the number of time periods, ℓ) which defines the window of price changes.

    The value should be greater than 0.
    """

    bar_component: Optional[BarComponent] = None
    """A component of a bar to use when updating the indicator with a bar sample.

    If not set, the bar component defaults to ClosePrice and is not shown in the indicator mnemonic.
    """

    quote_component: Optional[QuoteComponent] = None
    """A component of a quote to use when updating the indicator with a quote sample.

    If not set, the quote component defaults to MidPrice and is not shown in the indicator mnemonic.
    """

    trade_component: Optional[TradeComponent] = None
    """A component of a trade to use when updating the indicator with a trade sample.

    If not set, the trade component defaults to Price and is not shown in the indicator mnemonic.
    """


def default_params() -> ChandeMomentumOscillatorParams:
    """Returns default parameters for the Chande momentum oscillator indicator."""
    return ChandeMomentumOscillatorParams()
