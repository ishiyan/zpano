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
    """The type of cycle estimator to use.

    The default value is HilbertTransformerCycleEstimatorType.HomodyneDiscriminator.
    """

    estimator_params: CycleEstimatorParams = field(default_factory=CycleEstimatorParams)
    """Parameters to create an instance of the Hilbert transformer cycle estimator."""

    alpha_ema_period_additional: float = 0.33
    """The value of α (0 < α ≤ 1) used in EMA for additional smoothing of the instantaneous period.

    The default value is 0.33.
    """

    trend_line_smoothing_length: int = 4
    """The trend line smoothing length, must be 2, 3, or 4.

    The default value is 4.
    """

    cycle_part_multiplier: float = 1.0
    """The cycle part multiplier (0 < m ≤ 10) applied to the smoothed dominant cycle period
    when computing the averaging window length of the trend line.

    The default value is 1.0.
    """

    bar_component: Optional[BarComponent] = None
    """A component of a bar to use when updating the indicator with a bar sample.

    If not set, the bar component defaults to ClosePrice and is not shown in the indicator mnemonic.
    """

    quote_component: Optional[QuoteComponent] = None
    """A component of a quote to use when updating the indicator with a quote sample.

    If not set, the quote component defaults to MidPrice and is not shown in the indicator mnemonic.
    """

    trade_component: Optional[TradeComponent] = None
    """A component of a trade to use when updating the indicator with a trade sample.

    If not set, the trade component defaults to Price and is not shown in the indicator mnemonic.
    """


def default_params() -> HilbertTransformerInstantaneousTrendLineParams:
    """Returns default HilbertTransformerInstantaneousTrendLine parameters."""
    return HilbertTransformerInstantaneousTrendLineParams()
