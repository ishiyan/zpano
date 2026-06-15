"""Parameters for the CoronaTrendVigor indicator."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class Params:
    """Configuration for the Corona Trend Vigor indicator."""
    raster_length: int = 0
    """Length of the heatmap raster. Default 50. A zero value is treated as "use default"."""

    max_raster_value: float = 0.0
    """Maximum raster intensity value. Default 20. A zero value is treated as "use default"."""

    min_parameter_value: float = 0.0
    """Minimum ordinate (y) value. Default -10. Only substituted when both Min and Max are 0 (unconfigured)."""

    max_parameter_value: float = 0.0
    """Maximum ordinate (y) value. Default 10. Only substituted when both Min and Max are 0 (unconfigured)."""

    high_pass_filter_cutoff: int = 0
    """High-pass filter cutoff. Default 30. A zero value is treated as "use default"."""

    minimal_period: int = 0
    """Minimal cycle period. Default 6. A zero value is treated as "use default"."""

    maximal_period: int = 0
    """Maximal cycle period. Default 30. A zero value is treated as "use default"."""

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


def default_params() -> Params:
    """Return default Ehlers parameters."""
    return Params(
        raster_length=50,
        max_raster_value=20.0,
        min_parameter_value=-10.0,
        max_parameter_value=10.0,
        high_pass_filter_cutoff=30,
        minimal_period=6,
        maximal_period=30,
    )
