"""Schaff Trend Cycle parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class SchaffTrendCycleParams:
    """Parameters to create an instance of the Schaff Trend Cycle indicator."""

    fast: int = 23
    """The number of periods for the fast EMA of the MACD line.

    The value should be greater than 0. The default value is 23.
    """

    slow: int = 50
    """The number of periods for the slow EMA of the MACD line. It also sets the
    warm-up gate (barindex > slow).

    The value should be greater than 0. The default value is 50.
    """

    tclen: int = 10
    """The cycle length -- the look-back for both stochastics.

    The value should be greater than 0. The default value is 10.
    """

    factor: float = 0.5
    """The EMA smoothing alpha for both %D stages.

    The value should be in (0, 1]. The default value is 0.5.
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


def default_params() -> SchaffTrendCycleParams:
    """Returns default parameters for the Schaff Trend Cycle indicator."""
    return SchaffTrendCycleParams()
