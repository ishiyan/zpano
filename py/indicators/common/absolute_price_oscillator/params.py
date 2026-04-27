"""Absolute Price Oscillator parameters."""

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
class AbsolutePriceOscillatorParams:
    """Parameters to create an instance of the absolute price oscillator indicator."""

    fast_length: int = 12
    slow_length: int = 26
    moving_average_type: MovingAverageType = MovingAverageType.SMA
    first_is_average: bool = False
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> AbsolutePriceOscillatorParams:
    """Returns default parameters for the absolute price oscillator indicator."""
    return AbsolutePriceOscillatorParams()
