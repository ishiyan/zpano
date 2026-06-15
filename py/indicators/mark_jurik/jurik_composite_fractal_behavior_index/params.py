"""Jurik composite fractal behavior index parameters."""

from dataclasses import dataclass
from typing import Optional
from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class JurikCompositeFractalBehaviorIndexParams:
    """Parameters for the Jurik composite fractal behavior index indicator."""
    fractal_type: int = 1
    """FractalType controls the maximum fractal depth. Valid values are 1–4:
      1 = JCFB24 (8 depths: 2,3,4,6,8,12,16,24)
      2 = JCFB48 (10 depths: +32,48)
      3 = JCFB96 (12 depths: +64,96)
      4 = JCFB192 (14 depths: +128,192)
    """

    smooth: int = 10
    """Smooth is the smoothing window for the running averages.
    The value should be >= 1.
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


def default_params() -> JurikCompositeFractalBehaviorIndexParams:
    """Return default parameters."""
    return JurikCompositeFractalBehaviorIndexParams()
