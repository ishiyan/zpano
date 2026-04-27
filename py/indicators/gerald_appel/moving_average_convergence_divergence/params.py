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
    SMA = 1


@dataclass
class MovingAverageConvergenceDivergenceParams:
    """Parameters to create an instance of the MACD indicator."""

    fast_length: int = 12
    slow_length: int = 26
    signal_length: int = 9
    moving_average_type: MovingAverageType = MovingAverageType.EMA
    signal_moving_average_type: MovingAverageType = MovingAverageType.EMA
    first_is_average: Optional[bool] = None
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> MovingAverageConvergenceDivergenceParams:
    """Returns default parameters for the MACD indicator."""
    return MovingAverageConvergenceDivergenceParams()
