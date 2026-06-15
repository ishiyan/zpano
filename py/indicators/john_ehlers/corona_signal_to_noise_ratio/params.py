"""Parameters for the CoronaSignalToNoiseRatio indicator."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class Params:
    """Configuration for the Corona Signal-to-Noise Ratio indicator."""
    raster_length: int = 0
    """Length of the heatmap raster (number of intensity bins). Default 50. A zero value is treated as "use default"."""

    max_raster_value: float = 0.0
    """Maximum raster intensity value. Default 20. A zero value is treated as "use default"."""

    min_parameter_value: float = 0.0
    """Minimum ordinate (y) value of the heatmap — lower bound of the mapped SNR. Default 1. A zero value is treated as "use default"."""

    max_parameter_value: float = 0.0
    """Maximum ordinate (y) value of the heatmap — upper bound of the mapped SNR. Default 11. A zero value is treated as "use default"."""

    high_pass_filter_cutoff: int = 0
    """High-pass filter cutoff used by the inner Corona engine. Default 30. A zero value is treated as "use default"."""

    minimal_period: int = 0
    """Minimal cycle period covered by the filter bank. Default 6. A zero value is treated as "use default"."""

    maximal_period: int = 0
    """Maximal cycle period covered by the filter bank. Default 30. A zero value is treated as "use default"."""

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
        min_parameter_value=1.0,
        max_parameter_value=11.0,
        high_pass_filter_cutoff=30,
        minimal_period=6,
        maximal_period=30,
    )
