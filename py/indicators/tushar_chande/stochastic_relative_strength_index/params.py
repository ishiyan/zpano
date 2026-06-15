from dataclasses import dataclass
from enum import IntEnum
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


class MovingAverageType(IntEnum):
    """Type of moving average for Fast-D smoothing."""
    SMA = 0
    """Simple Moving Average."""

    EMA = 1
    """Exponential Moving Average."""


@dataclass
class StochasticRelativeStrengthIndexParams:
    """Parameters for the Stochastic RSI indicator."""
    length: int = 14
    """The number of periods for the RSI calculation.

    The value should be greater than 1.
    """

    fast_k_length: int = 5
    """The number of periods for the Fast-K stochastic calculation.

    The value should be greater than 0.
    """

    fast_d_length: int = 3
    """The number of periods for the Fast-D smoothing.

    The value should be greater than 0.
    """

    moving_average_type: MovingAverageType = MovingAverageType.SMA
    """The type of moving average for Fast-D (SMA or EMA).

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


def default_params() -> StochasticRelativeStrengthIndexParams:
    return StochasticRelativeStrengthIndexParams()
