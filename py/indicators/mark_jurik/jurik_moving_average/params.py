"""Jurik moving average parameters."""

from dataclasses import dataclass
from typing import Optional
from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class JurikMovingAverageParams:
    """Parameters for the Jurik moving average indicator."""
    length: int = 14
    """Length (the number of time periods, ℓ) determines
    the degree of smoothness and it can be any positive value.

    Small values make the moving average respond rapidly to price change
    and larger values produce smoother, flatter curves.

    The value should be greater than 1. Typical values range from 5 to 20.

    Irrespective from the value, the indicator needs at 30 first values to be primed.
    """

    phase: int = 0
    """Phase affects the amount of lag (delay).

    Lower lag tends to produce larger overshoot during price gaps, so you need
    to consider the trade-off between lag and overshoot and select a value for
    phase that balances your trading system's needs.

    Small values make the moving average respond rapidly to price change
    and larger values produce smoother, flatter curves.

    The phase values should be in [-100, 100].

    - The value of -100 results in maximum lag and no overshoot.
    - The value of 0 results in some lag and some overshoot.
    - The value of 100 results in minimum lag and maximum overshoot.
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


def default_params() -> JurikMovingAverageParams:
    """Return default parameters."""
    return JurikMovingAverageParams()
