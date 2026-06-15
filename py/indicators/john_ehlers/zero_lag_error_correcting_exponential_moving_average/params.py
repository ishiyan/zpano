"""Zero-lag error-correcting exponential moving average parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class ZeroLagErrorCorrectingExponentialMovingAverageParams:
    """Parameters for the zero-lag error-correcting exponential moving average (ZECEMA)."""

    smoothing_factor: float = 0.095
    """The smoothing factor (alpha) of the EMA.

    alpha = 2/(length + 1), 0 < alpha <= 1, 1 <= length.
    The default value is 0.095 (equivalent to length 20).
    """

    gain_limit: float = 5.0
    """Defines the range [-g, g] for finding the best gain factor.

    The value should be positive. The default value is 5.
    """

    gain_step: float = 0.1
    """Defines the iteration step for finding the best gain factor.

    The value should be positive. The default value is 0.1.
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


def default_params() -> ZeroLagErrorCorrectingExponentialMovingAverageParams:
    """Returns default parameters for the ZECEMA."""
    return ZeroLagErrorCorrectingExponentialMovingAverageParams()
