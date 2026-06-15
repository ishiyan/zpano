"""Jurik adaptive relative trend strength index parameters."""

from dataclasses import dataclass
from typing import Optional
from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class JurikAdaptiveRelativeTrendStrengthIndexParams:
    """Parameters for the Jurik adaptive relative trend strength index indicator."""
    lo_length: int = 5
    """LoLength is the minimum adaptive RSX length.
    The value should be at least 2.
    """

    hi_length: int = 30
    """HiLength is the maximum adaptive RSX length.
    The value should be at least loLength.
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


def default_params() -> JurikAdaptiveRelativeTrendStrengthIndexParams:
    """Return default parameters."""
    return JurikAdaptiveRelativeTrendStrengthIndexParams()
