"""Fractional Bands parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class FractionalBandsParams:
    """Parameters to create an instance of the fractional bands indicator."""

    period: int = 30
    """Period is the lookback period for the FGDI computation. The value should be greater than 1."""

    price_scale: float = 1.0
    """PriceScale is the multiplier converting price to a working numeric space. The value should be greater than 0."""

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


def default_params() -> FractionalBandsParams:
    """Returns default parameters for the fractional bands."""
    return FractionalBandsParams()
