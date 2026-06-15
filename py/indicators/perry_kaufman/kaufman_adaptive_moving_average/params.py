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
    """Efficiency ratio length is the number of last samples used to calculate the efficiency ratio.

    The value should be greater than 1.
    The default value is 10.
    """

    fastest_length: int = 2
    """Fastest length is the fastest boundary length, ℓf.
    The equivalent smoothing factor αf is

      αf = 2/(ℓf + 1), 2 ≤ ℓ

    The value should be greater than 1.
    The default value is 2.
    """

    slowest_length: int = 30
    """Slowest length is the slowest boundary length, ℓs.
    The equivalent smoothing factor αs is

      αs = 2/(ℓs + 1), 2 ≤ ℓ

    The value should be greater than 1.
    The default value is 30.
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
class KaufmanAdaptiveMovingAverageSmoothingFactorParams:
    """Parameters to create KAMA from smoothing factors."""

    efficiency_ratio_length: int = 10
    """Efficiency ratio length is the number of last samples used to calculate the efficiency ratio.

    The value should be greater than 1.
    The default value is 10.
    """

    fastest_smoothing_factor: float = 2.0 / 3.0
    """Fastest smoothing factor is the fastest boundary smoothing factor, αf in (0,1).
    The equivalent length ℓf is

      ℓf = 2/αf - 1, 0 < αf ≤ 1, 1 ≤ ℓf

    The default value is 2/3 (0.6666...).
    """

    slowest_smoothing_factor: float = 2.0 / 31.0
    """Slowest smoothing factor is the slowest boundary smoothing factor, αs in (0,1).
    The equivalent length ℓs is

      ℓs = 2/αs - 1, 0 < αs ≤ 1, 1 ≤ ℓs

    The default value is 2/31 (0.06451612903225806451612903225806).
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


def default_length_params() -> KaufmanAdaptiveMovingAverageLengthParams:
    """Returns default length-based parameters for KAMA."""
    return KaufmanAdaptiveMovingAverageLengthParams()


def default_smoothing_factor_params() -> KaufmanAdaptiveMovingAverageSmoothingFactorParams:
    """Returns default smoothing-factor-based parameters for KAMA."""
    return KaufmanAdaptiveMovingAverageSmoothingFactorParams()
