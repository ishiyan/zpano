"""T2 exponential moving average parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class T2ExponentialMovingAverageLengthParams:
    """Parameters to create a T2 indicator based on length."""

    length: int = 5
    volume_factor: float = 0.7
    first_is_average: bool = False
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


@dataclass
class T2ExponentialMovingAverageSmoothingFactorParams:
    """Parameters to create a T2 indicator based on smoothing factor."""

    smoothing_factor: float = 0.3333
    volume_factor: float = 0.7
    first_is_average: bool = False
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_length_params() -> T2ExponentialMovingAverageLengthParams:
    """Returns default length-based parameters for the T2."""
    return T2ExponentialMovingAverageLengthParams()


def default_smoothing_factor_params() -> T2ExponentialMovingAverageSmoothingFactorParams:
    """Returns default smoothing-factor-based parameters for the T2."""
    return T2ExponentialMovingAverageSmoothingFactorParams()
