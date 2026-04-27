"""Pearson's correlation coefficient parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class PearsonsCorrelationCoefficientParams:
    """Parameters for the Pearson's correlation coefficient indicator."""

    length: int = 20
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> PearsonsCorrelationCoefficientParams:
    """Return default parameters."""
    return PearsonsCorrelationCoefficientParams()
