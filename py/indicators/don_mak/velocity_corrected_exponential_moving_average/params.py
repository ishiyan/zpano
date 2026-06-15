"""Velocity-Corrected Exponential Moving Average parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class VelocityCorrectedExponentialMovingAverageParams:
    """Parameters to create an instance of the Velocity-Corrected Exponential Moving Average indicator."""

    period: int = 6
    """The EMA smoothing period.

    The value should be >= 2. The default value is 6.
    """

    degree: int = 3
    """The polynomial degree for the velocity estimation.

    The value should be >= 2. The default value is 3.
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


def default_params() -> VelocityCorrectedExponentialMovingAverageParams:
    """Returns default parameters for the Velocity-Corrected Exponential Moving Average indicator."""
    return VelocityCorrectedExponentialMovingAverageParams()
