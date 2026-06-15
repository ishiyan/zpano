"""TrendCycleMode parameters."""

from dataclasses import dataclass, field
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent
from ..hilbert_transformer.cycle_estimator_type import CycleEstimatorType
from ..hilbert_transformer.cycle_estimator_params import CycleEstimatorParams


@dataclass
class TrendCycleModeParams:
    """Parameters for the TrendCycleMode indicator."""
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
    """The additional WMA smoothing length used to smooth the trend line.

    The valid values are 2, 3, 4. The default value is 4.
    """

    cycle_part_multiplier: float = 1.0
    """The multiplier to the dominant cycle period used to determine the window length to
    calculate the trend line. Typical values are in [0.5, 1.5].

    The default value is 1.0. Valid range is (0, 10].
    """

    separation_percentage: float = 1.5
    """The threshold (in percent) above which a wide separation between the WMA-smoothed
    price and the instantaneous trend line forces the trend mode.

    The default value is 1.5. Valid range is (0, 100].
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


def default_params() -> TrendCycleModeParams:
    """Returns default TrendCycleMode parameters."""
    return TrendCycleModeParams()
