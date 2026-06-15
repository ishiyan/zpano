"""InstantaneousTrendLine parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class InstantaneousTrendLineLengthParams:
    """Parameters for the InstantaneousTrendLine indicator based on length."""
    length: int = 28
    """Length is the length (the number of time periods, \u2113) of the Instantaneous Trend Line.

    The smoothing factor \u03b1 is derived as \u03b1 = 2/(\u2113+1).
    The value should be a positive integer, greater or equal to 1.
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


@dataclass
class InstantaneousTrendLineSmoothingFactorParams:
    """Parameters for the InstantaneousTrendLine indicator based on smoothing factor."""
    smoothing_factor: float = 0.07
    """SmoothingFactor is the smoothing factor, \u03b1 in [0,1], of the Instantaneous Trend Line.

    The equivalent length \u2113 is:

        \u2113 = round(2/\u03b1) - 1, 0<\u03b1\u22641, 1\u2264\u2113.

    If \u03b1 is near zero (< epsilon), \u2113 is set to Number.MAX_SAFE_INTEGER.
    The default value used by Ehlers is 0.07 (\u2113 = 28).
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


def default_length_params() -> InstantaneousTrendLineLengthParams:
    """Returns default length-based parameters."""
    return InstantaneousTrendLineLengthParams()


def default_smoothing_factor_params() -> InstantaneousTrendLineSmoothingFactorParams:
    """Returns default smoothing-factor-based parameters."""
    return InstantaneousTrendLineSmoothingFactorParams()
