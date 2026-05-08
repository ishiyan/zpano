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
    EMA = 1
    SMMA = 2
    LWMA = 3


@dataclass
class NewMovingAverageParams:
    """Parameters to create an instance of the new moving average indicator."""

    primary_period: int = 0
    secondary_period: int = 8
    ma_type: MAType = MAType.LWMA
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> NewMovingAverageParams:
    """Returns default parameters for the new moving average."""
    return NewMovingAverageParams()
