"""SuperSmoother parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class SuperSmootherParams:
    """Parameters for the SuperSmoother indicator."""
    shortest_cycle_period: int = 10
    """The shortest cycle period in bars.
    The Super Smoother attenuates all cycle periods shorter than this one.

    The value should be greater than 1. The default value is 10.
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


def default_params() -> SuperSmootherParams:
    """Returns default SuperSmoother parameters."""
    return SuperSmootherParams()
