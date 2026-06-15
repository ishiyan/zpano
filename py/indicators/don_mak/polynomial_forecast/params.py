"""Polynomial Forecast parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class PolynomialForecastParams:
    """Parameters to create an instance of the Polynomial Forecast indicator."""

    degree: int = 3
    """The polynomial degree for the local fit (uses degree+1 bars).

    The value should be >= 2. The default value is 3.
    """

    order: int = 1
    """The Taylor expansion order: 1 = velocity only (F1V), 2 = velocity + acceleration (F1VA).

    The value should be 1 or 2. The default value is 1.
    """

    smoothing: int = 0
    """The EMA pre-smoothing period applied to price before fitting (0 = none).

    The value should be >= 0. The default value is 0.
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


def default_params() -> PolynomialForecastParams:
    """Returns default parameters for the Polynomial Forecast indicator."""
    return PolynomialForecastParams()
