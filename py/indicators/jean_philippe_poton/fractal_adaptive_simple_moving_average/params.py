"""Fractal adaptive simple moving average parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class FractalAdaptiveSimpleMovingAverageParams:
    """Parameters to create an instance of the fractal adaptive simple moving average indicator."""

    period: int = 30
    """Period is the lookback period N for the FDI computation.

    The value should be greater than 1.
    """

    normal_speed: int = 20
    """NormalSpeed is the base SMA period before fractal adaptation.

    The value should be greater than 0.
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


def default_params() -> FractalAdaptiveSimpleMovingAverageParams:
    """Returns default parameters for the fractal adaptive simple moving average."""
    return FractalAdaptiveSimpleMovingAverageParams()
