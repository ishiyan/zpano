"""Moving Mini-Max parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class MovingMiniMaxParams:
    """Parameters to create an instance of the Moving Mini-Max indicator."""

    m: int = 5
    """The smoothing window width controlling the quantum tunnelling ability.

    Larger values produce smoother output, suppressing smaller peaks. The value should be
    >= 1. The default value is 5.
    """

    n: int = 50
    """The lookback window size: the number of price bars over which the indicator is computed.

    Priming requires n prices. The value should be > 2*m. The default value is 50.
    """

    num_extrema: int = 3
    """The number of distinct support/resistance levels to detect and return.

    The value should be >= 1. The default value is 3.
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


def default_params() -> MovingMiniMaxParams:
    """Returns default parameters for the Moving Mini-Max indicator."""
    return MovingMiniMaxParams()
