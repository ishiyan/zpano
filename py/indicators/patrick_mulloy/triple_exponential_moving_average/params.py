"""Triple exponential moving average parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class TripleExponentialMovingAverageLengthParams:
    """Parameters to create a TEMA indicator based on length."""

    length: int = 20
    """Length is the length (the number of time periods, ℓ) of the moving window to calculate the average.

    The value should be greater than 1.
    """

    first_is_average: bool = False
    """FirstIsAverage indicates whether the very first triple exponential moving average value is
    a simple average of the first 'period' (the most widely documented approach) or
    the first input value (used in Metastock).
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


@dataclass
class TripleExponentialMovingAverageSmoothingFactorParams:
    """Parameters to create a TEMA indicator based on smoothing factor."""

    smoothing_factor: float = 0.0952
    """SmoothingFactor is the smoothing factor, α in (0,1), of the triple exponential moving average.

    The equivalent length ℓ is:

        ℓ = 2/α - 1, 0<α≤1, 1≤ℓ.
    """

    first_is_average: bool = False
    """FirstIsAverage indicates whether the very first triple exponential moving average value is
    a simple average of the first 'period' (the most widely documented approach) or
    the first input value (used in Metastock).
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


def default_length_params() -> TripleExponentialMovingAverageLengthParams:
    """Returns default length-based parameters for the TEMA."""
    return TripleExponentialMovingAverageLengthParams()


def default_smoothing_factor_params() -> TripleExponentialMovingAverageSmoothingFactorParams:
    """Returns default smoothing-factor-based parameters for the TEMA."""
    return TripleExponentialMovingAverageSmoothingFactorParams()
