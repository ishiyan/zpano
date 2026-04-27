"""Bollinger Bands parameters."""

from dataclasses import dataclass
from enum import IntEnum
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


class MovingAverageType(IntEnum):
    """Specifies the type of moving average to use."""

    SMA = 0
    EMA = 1


@dataclass
class BollingerBandsParams:
    """Parameters to create an instance of the Bollinger Bands indicator."""

    length: int = 5
    upper_multiplier: float = 2.0
    lower_multiplier: float = 2.0
    is_unbiased: Optional[bool] = None
    moving_average_type: MovingAverageType = MovingAverageType.SMA
    first_is_average: bool = False
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> BollingerBandsParams:
    """Returns default parameters for the Bollinger Bands indicator."""
    return BollingerBandsParams()
