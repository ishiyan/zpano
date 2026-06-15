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
    """Length is the length, l, (the number of time periods) of the Center of Gravity oscillator.

    The value should be a positive integer, greater or equal to 1.
    The default value used by Ehlers is 10.
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


def default_params() -> CenterOfGravityOscillatorParams:
    """Returns default CenterOfGravityOscillator parameters."""
    return CenterOfGravityOscillatorParams()
