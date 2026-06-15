"""Moving Average Convergence Divergence parameters."""

from dataclasses import dataclass
from enum import IntEnum
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


class MovingAverageType(IntEnum):
    """Specifies the type of moving average to use."""

    EMA = 0
    """Exponential Moving Average (default for classic MACD)."""

    SMA = 1
    """Simple Moving Average."""


@dataclass
class MovingAverageConvergenceDivergenceParams:
    """Parameters to create an instance of the MACD indicator."""

    fast_length: int = 12
    """The number of periods for the fast moving average.

    The value should be greater than 1. The default value is 12.
    """

    slow_length: int = 26
    """The number of periods for the slow moving average.

    The value should be greater than 1. The default value is 26.
    """

    signal_length: int = 9
    """The number of periods for the signal line moving average.

    The value should be greater than 0. The default value is 9.
    """

    moving_average_type: MovingAverageType = MovingAverageType.EMA
    """The type of moving average for the fast and slow lines (EMA or SMA).

    If not set, the Exponential Moving Average is used.
    """

    signal_moving_average_type: MovingAverageType = MovingAverageType.EMA
    """The type of moving average for the signal line (EMA or SMA).

    If not set, the Exponential Moving Average is used.
    """

    first_is_average: Optional[bool] = None
    """Controls the EMA seeding algorithm.
    When true (default), the first EMA value is the simple average of the first period values
    (TA-Lib compatible). When false, the first input value is used directly (Metastock style).
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


def default_params() -> MovingAverageConvergenceDivergenceParams:
    """Returns default parameters for the MACD indicator."""
    return MovingAverageConvergenceDivergenceParams()
