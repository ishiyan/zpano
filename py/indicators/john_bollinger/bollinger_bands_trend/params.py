"""Bollinger Bands Trend parameters."""

from dataclasses import dataclass
from enum import IntEnum
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


class MovingAverageType(IntEnum):
    """Specifies the type of moving average to use."""

    SMA = 0
    """Simple Moving Average."""

    EMA = 1
    """Exponential Moving Average."""


@dataclass
class BollingerBandsTrendParams:
    """Parameters to create an instance of the Bollinger Bands Trend indicator."""

    fast_length: int = 20
    """The number of periods for the fast (shorter) Bollinger Band.

    The value should be greater than 1. The default value is 20.
    """

    slow_length: int = 50
    """The number of periods for the slow (longer) Bollinger Band.

    The value should be greater than 1 and greater than fastLength. The default value is 50.
    """

    upper_multiplier: float = 2.0
    """The number of standard deviations above the middle band.

    The default value is 2.0.
    """

    lower_multiplier: float = 2.0
    """The number of standard deviations below the middle band.

    The default value is 2.0.
    """

    is_unbiased: Optional[bool] = None
    """Indicates whether to use the unbiased sample standard deviation (true)
    or the population standard deviation (false).

    If not set, defaults to true (unbiased sample standard deviation).
    """

    moving_average_type: MovingAverageType = MovingAverageType.SMA
    """The type of moving average (SMA or EMA).

    If not set, the Simple Moving Average is used.
    """

    first_is_average: bool = False
    """Controls the EMA seeding algorithm.
    When true, the first EMA value is the simple average of the first period values.
    When false (default), the first input value is used directly (Metastock style).
    Only relevant when movingAverageType is EMA.
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


def default_params() -> BollingerBandsTrendParams:
    """Returns default parameters for the Bollinger Bands Trend indicator."""
    return BollingerBandsTrendParams()
