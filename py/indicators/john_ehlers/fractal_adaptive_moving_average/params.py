"""FractalAdaptiveMovingAverage parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class FractalAdaptiveMovingAverageParams:
    """Parameters for the FractalAdaptiveMovingAverage indicator."""
    length: int = 16
    """Length is the length, l, (the number of time periods) of the Fractal Adaptive Moving Average.

    The value should be an even integer, greater or equal to 2.
    The default value is 16.
    """

    slowest_smoothing_factor: float = 0.01
    """SlowestSmoothingFactor is the slowest boundary smoothing factor, as in [0,1].
    The equivalent length ls is

      ls = 2/as - 1, 0 < as <= 1, 1 <= ls

    The default value is 0.01 (equivalent ls = 199).
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


def default_params() -> FractalAdaptiveMovingAverageParams:
    """Returns default FractalAdaptiveMovingAverage parameters."""
    return FractalAdaptiveMovingAverageParams()
