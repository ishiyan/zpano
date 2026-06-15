"""Polynomial Fit Derivative parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class PolynomialFitDerivativeParams:
    """Parameters to create an instance of the Polynomial Fit Derivative indicator."""

    degree: int = 3
    """The polynomial degree. The number of data points used is degree + 1.

    The value should be >= 2. The default value is 3 (cubic).
    """

    order: int = 1
    """The derivative order (1 = velocity, 2 = acceleration).

    The value should be >= 1 and <= degree. The default value is 1.
    """

    smoothing: int = 6
    """The EMA pre-smoothing length applied before the FIR filter.

    The value should be >= 0 (0 means no smoothing). The default value is 6.
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


def default_params() -> PolynomialFitDerivativeParams:
    """Returns default parameters for the Polynomial Fit Derivative indicator."""
    return PolynomialFitDerivativeParams()
