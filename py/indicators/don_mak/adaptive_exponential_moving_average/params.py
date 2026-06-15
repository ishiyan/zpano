"""Adaptive Exponential Moving Average parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class AdaptiveExponentialMovingAverageParams:
    """Parameters to create an instance of the Adaptive Exponential Moving Average indicator."""

    alpha_max: float = 0.5
    """The smoothing factor for trending data (low frequency).

    The value should be in (0, 1] and greater than alpha_min. The default value is 0.5.
    """

    alpha_min: float = 0.05
    """The smoothing factor for noisy data (high frequency).

    The value should be in (0, alpha_max). The default value is 0.05.
    """

    omega_0: float = 1.0
    """The crossover frequency in radians/bar. Below this, alpha = alpha_max.

    The value should be in (0, pi). The default value is 1.0.
    """

    smoothing: int = 3
    """The embedded ISWP internal smoothing parameter.

    The value should be >= 0. The default value is 3.
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


def default_params() -> AdaptiveExponentialMovingAverageParams:
    """Returns default parameters for the Adaptive Exponential Moving Average indicator."""
    return AdaptiveExponentialMovingAverageParams()
