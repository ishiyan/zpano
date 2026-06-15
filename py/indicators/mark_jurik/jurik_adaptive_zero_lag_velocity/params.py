"""Jurik adaptive zero lag velocity parameters."""

from dataclasses import dataclass
from typing import Optional
from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class JurikAdaptiveZeroLagVelocityParams:
    """Parameters for the Jurik adaptive zero lag velocity indicator."""
    lo_length: int = 5
    """Minimum adaptive depth. Must be >= 2."""

    hi_length: int = 30
    """Maximum adaptive depth. Must be >= loLength."""

    sensitivity: float = 1.0
    """Controls the volatility regime detection sensitivity."""

    period: float = 3.0
    """Controls the adaptive smoother period. Must be > 0."""

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


def default_params() -> JurikAdaptiveZeroLagVelocityParams:
    """Return default parameters."""
    return JurikAdaptiveZeroLagVelocityParams()
