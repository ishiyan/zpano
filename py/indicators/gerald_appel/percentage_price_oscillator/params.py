"""Percentage Price Oscillator parameters."""

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
class PercentagePriceOscillatorParams:
    """Parameters to create an instance of the percentage price oscillator indicator."""

    fast_length: int = 12
    """The number of periods for the fast moving average.

    The value should be greater than 1.
    """

    slow_length: int = 26
    """The number of periods for the slow moving average.

    The value should be greater than 1.
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


def default_params() -> PercentagePriceOscillatorParams:
    """Returns default parameters for the percentage price oscillator indicator."""
    return PercentagePriceOscillatorParams()
