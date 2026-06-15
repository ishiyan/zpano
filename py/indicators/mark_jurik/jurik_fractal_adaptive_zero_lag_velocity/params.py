"""Jurik fractal adaptive zero lag velocity parameters."""

from dataclasses import dataclass
from typing import Optional
from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class JurikFractalAdaptiveZeroLagVelocityParams:
    """Parameters for the Jurik fractal adaptive zero lag velocity indicator."""
    lo_depth: int = 5
    """Minimum depth for the velocity computation. Must be >= 2."""

    hi_depth: int = 30
    """Maximum depth for the velocity computation. Must be >= loDepth."""

    fractal_type: int = 1
    """Selects the scale set (1-4)."""

    smooth: int = 10
    """Smoothing window for CFB channel averages. Must be >= 1."""

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


def default_params() -> JurikFractalAdaptiveZeroLagVelocityParams:
    """Return default parameters."""
    return JurikFractalAdaptiveZeroLagVelocityParams()
