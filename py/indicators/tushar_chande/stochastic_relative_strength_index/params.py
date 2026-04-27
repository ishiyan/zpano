from dataclasses import dataclass
from enum import IntEnum
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


class MovingAverageType(IntEnum):
    """Type of moving average for Fast-D smoothing."""
    SMA = 0
    EMA = 1


@dataclass
class StochasticRelativeStrengthIndexParams:
    """Parameters for the Stochastic RSI indicator."""
    length: int = 14
    fast_k_length: int = 5
    fast_d_length: int = 3
    moving_average_type: MovingAverageType = MovingAverageType.SMA
    first_is_average: bool = False
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> StochasticRelativeStrengthIndexParams:
    return StochasticRelativeStrengthIndexParams()
