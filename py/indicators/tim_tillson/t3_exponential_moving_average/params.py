"""T3 exponential moving average parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class T3ExponentialMovingAverageLengthParams:
    """Parameters to create a T3 indicator based on length."""

    length: int = 5
    volume_factor: float = 0.7
    first_is_average: bool = False
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


@dataclass
class T3ExponentialMovingAverageSmoothingFactorParams:
    """Parameters to create a T3 indicator based on smoothing factor."""

    smoothing_factor: float = 0.3333
    volume_factor: float = 0.7
    first_is_average: bool = False
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_length_params() -> T3ExponentialMovingAverageLengthParams:
    """Returns default length-based parameters for the T3."""
    return T3ExponentialMovingAverageLengthParams()


def default_smoothing_factor_params() -> T3ExponentialMovingAverageSmoothingFactorParams:
    """Returns default smoothing-factor-based parameters for the T3."""
    return T3ExponentialMovingAverageSmoothingFactorParams()
