"""InstantaneousTrendLine parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class InstantaneousTrendLineLengthParams:
    """Parameters for the InstantaneousTrendLine indicator based on length."""
    length: int = 28
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


@dataclass
class InstantaneousTrendLineSmoothingFactorParams:
    """Parameters for the InstantaneousTrendLine indicator based on smoothing factor."""
    smoothing_factor: float = 0.07
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_length_params() -> InstantaneousTrendLineLengthParams:
    """Returns default length-based parameters."""
    return InstantaneousTrendLineLengthParams()


def default_smoothing_factor_params() -> InstantaneousTrendLineSmoothingFactorParams:
    """Returns default smoothing-factor-based parameters."""
    return InstantaneousTrendLineSmoothingFactorParams()
