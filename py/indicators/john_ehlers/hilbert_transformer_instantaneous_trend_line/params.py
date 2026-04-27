"""HilbertTransformerInstantaneousTrendLine parameters."""

from dataclasses import dataclass, field
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent
from ..hilbert_transformer.cycle_estimator_type import CycleEstimatorType
from ..hilbert_transformer.cycle_estimator_params import CycleEstimatorParams


@dataclass
class HilbertTransformerInstantaneousTrendLineParams:
    """Parameters for the HilbertTransformerInstantaneousTrendLine indicator."""
    estimator_type: CycleEstimatorType = CycleEstimatorType.HOMODYNE_DISCRIMINATOR
    estimator_params: CycleEstimatorParams = field(default_factory=CycleEstimatorParams)
    alpha_ema_period_additional: float = 0.33
    trend_line_smoothing_length: int = 4
    cycle_part_multiplier: float = 1.0
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> HilbertTransformerInstantaneousTrendLineParams:
    """Returns default HilbertTransformerInstantaneousTrendLine parameters."""
    return HilbertTransformerInstantaneousTrendLineParams()
