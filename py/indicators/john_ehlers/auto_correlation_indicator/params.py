"""Parameters for the AutoCorrelation Indicator."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class Params:
    """Parameters for creating an AutoCorrelationIndicator instance."""

    min_lag: int = 0
    """Minimum (shortest) correlation lag shown on the heatmap axis. Must be >= 1.
    The default value is 3 (matching Ehlers' EasyLanguage listing 8-2, which plots
    lags 3..48). A zero value is treated as "use default".
    """

    max_lag: int = 0
    """Maximum (longest) correlation lag shown on the heatmap axis. Must be > minLag.
    Also drives the cutoff of the 2-pole Butterworth highpass pre-filter. The default
    value is 48. A zero value is treated as "use default".
    """

    smoothing_period: int = 0
    """Cutoff period of the 2-pole Super Smoother pre-filter applied after the highpass.
    Must be >= 2. The default value is 10 (matching Ehlers' EasyLanguage listing 8-2,
    which hardcodes 10). A zero value is treated as "use default".
    """

    averaging_length: int = 0
    """Number of samples (M) used in each Pearson correlation accumulation. When zero
    (the Ehlers default), M equals the current lag, making each correlation use the
    same number of samples as its lag distance. When positive, the same M is used
    for all lags. Must be >= 0.
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
    """Returns a Params with Ehlers defaults."""
    return Params(
        min_lag=3,
        max_lag=48,
        smoothing_period=10,
        averaging_length=0,
    )
