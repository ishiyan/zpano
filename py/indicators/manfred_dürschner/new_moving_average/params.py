"""New moving average parameters."""

from dataclasses import dataclass
from enum import IntEnum
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


class MAType(IntEnum):
    """Type of moving average used in the NMA calculation."""

    SMA = 0
    """Simple Moving Average."""

    EMA = 1
    """Exponential Moving Average."""

    SMMA = 2
    """Smoothed Moving Average."""

    LWMA = 3
    """Linear Weighted Moving Average."""


@dataclass
class NewMovingAverageParams:
    """Parameters to create an instance of the new moving average indicator."""

    primary_period: int = 0
    """PrimaryPeriod is the period for the primary (outer) moving average.

    If 0 or too small, it is auto-resolved via Nyquist constraint.
    """

    secondary_period: int = 8
    """SecondaryPeriod is the period for the secondary (inner) moving average.

    The value should be greater than or equal to 2.
    """

    ma_type: MAType = MAType.LWMA
    """MAType selects the moving average type (SMA=0, EMA=1, SMMA=2, LWMA=3)."""

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


def default_params() -> NewMovingAverageParams:
    """Returns default parameters for the new moving average."""
    return NewMovingAverageParams()
