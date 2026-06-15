from dataclasses import dataclass
from typing import Optional

from py.entities.bar_component import BarComponent
from py.entities.quote_component import QuoteComponent
from py.entities.trade_component import TradeComponent


@dataclass
class RelativeStrengthIndexParams:
    """Parameters for the Relative Strength Index indicator."""
    length: int = 14
    """The number of periods for the RSI calculation.

    The value should be greater than 1. The default value is 14.
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


def default_params() -> RelativeStrengthIndexParams:
    """Returns default parameters."""
    return RelativeStrengthIndexParams()
