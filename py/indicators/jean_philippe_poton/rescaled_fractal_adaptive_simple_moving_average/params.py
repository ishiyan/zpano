"""Rescaled fractal adaptive simple moving average parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class RescaledFractalAdaptiveSimpleMovingAverageParams:
    """Parameters to create an instance of the rescaled fractal adaptive simple moving average indicator."""

    period: int = 64
    """Period is the lookback window for R/S analysis. Must be a power of 2, >= 4."""

    normal_speed: int = 30
    """NormalSpeed is the base SMA period before fractal adaptation. Must be >= 1."""

    price_scale: float = 1.0
    """PriceScale is the multiplier applied to prices before R/S calculation. Default is 1.0."""

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


def default_params() -> RescaledFractalAdaptiveSimpleMovingAverageParams:
    """Returns default parameters for the rescaled fractal adaptive simple moving average."""
    return RescaledFractalAdaptiveSimpleMovingAverageParams()
