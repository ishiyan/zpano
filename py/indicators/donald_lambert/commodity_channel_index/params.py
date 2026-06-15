"""Commodity Channel Index parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent

DEFAULT_INVERSE_SCALING_FACTOR = 0.015


@dataclass
class CommodityChannelIndexParams:
    """Parameters to create an instance of the commodity channel index indicator."""

    length: int = 20
    """The number of time periods of the commodity channel index.

    The value should be greater than 1.
    """

    inverse_scaling_factor: float = DEFAULT_INVERSE_SCALING_FACTOR
    """The inverse scaling factor to provide more readable value numbers.
    The default value of 0.015 ensures that approximately 70 to 80 percent
    of CCI values would fall between -100 and +100.

    If not set, the default (0.015) is used.
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


def default_params() -> CommodityChannelIndexParams:
    """Returns default parameters for the commodity channel index indicator."""
    return CommodityChannelIndexParams()
