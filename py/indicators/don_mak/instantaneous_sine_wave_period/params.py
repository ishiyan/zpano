"""Instantaneous Sine Wave Period parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class InstantaneousSineWavePeriodParams:
    """Parameters to create an instance of the Instantaneous Sine Wave Period indicator."""

    smoothing: int = 0
    """The EMA smoothing length applied to input prices before frequency estimation.

    The value should be >= 0 (0 means no smoothing). The default value is 0.
    """

    min_period: float = 4.0
    """The minimum allowed period in bars. Estimates below this are rejected.

    The value should be > 0. The default value is 4.0.
    """

    max_period: float = 50.0
    """The maximum allowed period in bars. Estimates above this are rejected.

    The value should be > min_period. The default value is 50.0.
    """

    error_threshold: float = 20.0
    """The maximum tolerated error for the omega estimate. If both methods exceed this,
    the output is NaN.

    The value should be > 0. The default value is 20.0.
    """

    dx: float = 0.01
    """The assumed measurement error for each price point (used in error propagation).

    The value should be > 0. The default value is 0.01.
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


def default_params() -> InstantaneousSineWavePeriodParams:
    """Returns default parameters for the Instantaneous Sine Wave Period indicator."""
    return InstantaneousSineWavePeriodParams()
