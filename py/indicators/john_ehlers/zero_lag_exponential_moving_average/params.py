"""Zero-lag exponential moving average parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class ZeroLagExponentialMovingAverageParams:
    """Parameters for the zero-lag exponential moving average (ZEMA)."""

    smoothing_factor: float = 0.25
    """The smoothing factor (alpha) of the EMA.

    alpha = 2/(length + 1), 0 < alpha <= 1, 1 <= length.
    The default value is 0.25.
    """

    velocity_gain_factor: float = 0.5
    """The gain factor used to estimate the velocity.

    The default value is 0.5.
    """

    velocity_momentum_length: int = 3
    """The length of the momentum used to estimate the velocity.

    The value should be positive. The default value is 3.
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


def default_params() -> ZeroLagExponentialMovingAverageParams:
    """Returns default parameters for the ZEMA."""
    return ZeroLagExponentialMovingAverageParams()
