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
    """FastLimitLength is the fastest boundary length, ℓf.
    The equivalent smoothing factor αf is

      αf = 2/(ℓf + 1), 2 ≤ ℓ

    The value should be greater than 1.
    The default value is 3 (αf=0.5).
    """

    slow_limit_length: int = 39
    """SlowLimitLength is the slowest boundary length, ℓs.
    The equivalent smoothing factor αs is

      αs = 2/(ℓs + 1), 2 ≤ ℓ

    The value should be greater than 1.
    The default value is 39 (αs=0.05).
    """

    estimator_type: CycleEstimatorType = CycleEstimatorType.HOMODYNE_DISCRIMINATOR
    """The type of cycle estimator to use.

    The default value is HilbertTransformerCycleEstimatorType.HomodyneDiscriminator.
    """

    estimator_params: CycleEstimatorParams = field(default_factory=lambda: CycleEstimatorParams())
    """Parameters to create an instance of the Hilbert transformer cycle estimator."""

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


@dataclass
class MesaAdaptiveMovingAverageSmoothingFactorParams:
    """Parameters for creating MAMA from smoothing factors."""
    fast_limit_smoothing_factor: float = 0.5
    """FastLimitSmoothingFactor is the fastest boundary smoothing factor, αf in (0, 1).
    The equivalent length ℓf is

      ℓf = 2/αf - 1, 0 < αf < 1, 1 < ℓf

    The default value is 0.5 (ℓf=3).
    """

    slow_limit_smoothing_factor: float = 0.05
    """SlowLimitSmoothingFactor is the slowest boundary smoothing factor, αs in (0, 1).
    The equivalent length ℓs is

      ℓs = 2/αs - 1, 0 < αs < 1, 1 < ℓs

    The default value is 0.05 (ℓs=39).
    """

    estimator_type: CycleEstimatorType = CycleEstimatorType.HOMODYNE_DISCRIMINATOR
    """The type of cycle estimator to use.

    The default value is HilbertTransformerCycleEstimatorType.HomodyneDiscriminator.
    """

    estimator_params: CycleEstimatorParams = field(default_factory=lambda: CycleEstimatorParams())
    """Parameters to create an instance of the Hilbert transformer cycle estimator."""

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
