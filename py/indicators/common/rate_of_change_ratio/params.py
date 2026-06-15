"""Rate of Change Ratio parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class RateOfChangeRatioParams:
    """Parameters to create an instance of the rate of change ratio indicator."""

    length: int = 10
    """Length is the length (the number of time periods, ℓ) between today's sample and the sample ℓ periods ago.

    The value should be greater than 0.
    """

    hundred_scale: bool = False
    """Indicates whether to multiply the ratio by 100.

    If false (default), the result is price/previousPrice (centered at 1).
    If true, the result is (price/previousPrice)*100 (centered at 100).
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


def default_params() -> RateOfChangeRatioParams:
    """Returns default parameters for the rate of change ratio indicator."""
    return RateOfChangeRatioParams()
