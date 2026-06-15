"""Parameters for the CoronaSpectrum indicator."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class Params:
    """Configuration for the Corona Spectrum indicator."""
    min_raster_value: float = 0.0
    """Minimal raster value (z) of the heatmap, in decibels. Corresponds to the
    CoronaLowerDecibels threshold.

    The default value is 6. A zero value is treated as "use default".
    """

    max_raster_value: float = 0.0
    """Maximal raster value (z) of the heatmap, in decibels. Corresponds to the
    CoronaUpperDecibels threshold.

    The default value is 20. A zero value is treated as "use default".
    """

    min_parameter_value: float = 0.0
    """Minimal ordinate (y) value of the heatmap, representing the minimal cycle
    period covered by the filter bank. Rounded up to the nearest integer.

    The default value is 6. A zero value is treated as "use default".
    """

    max_parameter_value: float = 0.0
    """Maximal ordinate (y) value of the heatmap, representing the maximal cycle
    period covered by the filter bank. Rounded down to the nearest integer.

    The default value is 30. A zero value is treated as "use default".
    """

    high_pass_filter_cutoff: int = 0
    """High-pass filter cutoff (de-trending period) used by the inner Corona engine.
    Suggested values are 20, 30, 100.

    The default value is 30. A zero value is treated as "use default".
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


def default_params() -> Params:
    """Return default Ehlers parameters."""
    return Params(
        min_raster_value=6.0,
        max_raster_value=20.0,
        min_parameter_value=6.0,
        max_parameter_value=30.0,
        high_pass_filter_cutoff=30,
    )
