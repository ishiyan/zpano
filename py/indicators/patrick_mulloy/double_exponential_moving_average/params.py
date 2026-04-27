"""Double exponential moving average parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class DoubleExponentialMovingAverageLengthParams:
    """Parameters to create a DEMA indicator based on length."""

    length: int = 20
    first_is_average: bool = False
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


@dataclass
class DoubleExponentialMovingAverageSmoothingFactorParams:
    """Parameters to create a DEMA indicator based on smoothing factor."""

    smoothing_factor: float = 0.0952
    first_is_average: bool = False
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_length_params() -> DoubleExponentialMovingAverageLengthParams:
    """Returns default length-based parameters for the DEMA."""
    return DoubleExponentialMovingAverageLengthParams()


def default_smoothing_factor_params() -> DoubleExponentialMovingAverageSmoothingFactorParams:
    """Returns default smoothing-factor-based parameters for the DEMA."""
    return DoubleExponentialMovingAverageSmoothingFactorParams()
