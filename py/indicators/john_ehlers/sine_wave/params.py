"""SineWave parameters."""

from dataclasses import dataclass, field
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent
from ..hilbert_transformer.cycle_estimator_type import CycleEstimatorType
from ..hilbert_transformer.cycle_estimator_params import CycleEstimatorParams


@dataclass
class SineWaveParams:
    """Parameters for the SineWave indicator."""
    estimator_type: CycleEstimatorType = CycleEstimatorType.HOMODYNE_DISCRIMINATOR
    estimator_params: CycleEstimatorParams = field(default_factory=CycleEstimatorParams)
    alpha_ema_period_additional: float = 0.33
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None


def default_params() -> SineWaveParams:
    """Returns default SineWave parameters."""
    return SineWaveParams()
