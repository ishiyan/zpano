"""Mesa Adaptive Moving Average parameters."""

from dataclasses import dataclass, field
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent
from ..hilbert_transformer.cycle_estimator_type import CycleEstimatorType
from ..hilbert_transformer.cycle_estimator_params import CycleEstimatorParams


@dataclass
class MesaAdaptiveMovingAverageLengthParams:
    """Parameters for creating MAMA from lengths."""
    fast_limit_length: int = 3
    slow_limit_length: int = 39
    estimator_type: CycleEstimatorType = CycleEstimatorType.HOMODYNE_DISCRIMINATOR
    estimator_params: CycleEstimatorParams = field(default_factory=lambda: CycleEstimatorParams())
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


@dataclass
class MesaAdaptiveMovingAverageSmoothingFactorParams:
    """Parameters for creating MAMA from smoothing factors."""
    fast_limit_smoothing_factor: float = 0.5
    slow_limit_smoothing_factor: float = 0.05
    estimator_type: CycleEstimatorType = CycleEstimatorType.HOMODYNE_DISCRIMINATOR
    estimator_params: CycleEstimatorParams = field(default_factory=lambda: CycleEstimatorParams())
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None
