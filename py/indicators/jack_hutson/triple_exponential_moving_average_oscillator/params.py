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
    """The number of time periods for the three chained EMA calculations.

    The value should be greater than or equal to 1. The default value is 30.
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


def default_params() -> TripleExponentialMovingAverageOscillatorParams:
    """Returns default parameters for the TRIX indicator."""
    return TripleExponentialMovingAverageOscillatorParams()
