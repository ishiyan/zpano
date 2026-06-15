"""Arnaud Legoux moving average parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class ArnaudLegouxMovingAverageParams:
    """Parameters to create an instance of the Arnaud Legoux moving average indicator."""

    window: int = 9
    """Window is the number of bars in the lookback window.

    The value should be greater than 0.
    """

    sigma: float = 6.0
    """Sigma controls the Gaussian width; larger values produce smoother output.

    The value should be greater than 0.
    """

    offset: float = 0.85
    """Offset shifts the Gaussian peak; 0 = centered, 1 = newest bar.

    The value should be between 0 and 1.
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


def default_params() -> ArnaudLegouxMovingAverageParams:
    """Returns default parameters for the Arnaud Legoux moving average."""
    return ArnaudLegouxMovingAverageParams()
