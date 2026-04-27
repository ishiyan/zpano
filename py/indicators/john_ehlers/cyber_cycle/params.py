"""CyberCycle parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class CyberCycleLengthParams:
    """Parameters for the CyberCycle indicator based on length."""
    length: int = 28
    signal_lag: int = 9
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


@dataclass
class CyberCycleSmoothingFactorParams:
    """Parameters for the CyberCycle indicator based on smoothing factor."""
    smoothing_factor: float = 0.07
    signal_lag: int = 9
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_length_params() -> CyberCycleLengthParams:
    """Returns default length-based parameters."""
    return CyberCycleLengthParams()


def default_smoothing_factor_params() -> CyberCycleSmoothingFactorParams:
    """Returns default smoothing-factor-based parameters."""
    return CyberCycleSmoothingFactorParams()
