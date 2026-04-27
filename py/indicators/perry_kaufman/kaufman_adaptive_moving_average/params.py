"""Kaufman Adaptive Moving Average parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class KaufmanAdaptiveMovingAverageLengthParams:
    """Parameters to create KAMA from lengths."""

    efficiency_ratio_length: int = 10
    fastest_length: int = 2
    slowest_length: int = 30
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


@dataclass
class KaufmanAdaptiveMovingAverageSmoothingFactorParams:
    """Parameters to create KAMA from smoothing factors."""

    efficiency_ratio_length: int = 10
    fastest_smoothing_factor: float = 2.0 / 3.0
    slowest_smoothing_factor: float = 2.0 / 31.0
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_length_params() -> KaufmanAdaptiveMovingAverageLengthParams:
    """Returns default length-based parameters for KAMA."""
    return KaufmanAdaptiveMovingAverageLengthParams()


def default_smoothing_factor_params() -> KaufmanAdaptiveMovingAverageSmoothingFactorParams:
    """Returns default smoothing-factor-based parameters for KAMA."""
    return KaufmanAdaptiveMovingAverageSmoothingFactorParams()
