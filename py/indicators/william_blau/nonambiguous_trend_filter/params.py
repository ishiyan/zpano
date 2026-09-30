"""Nonambiguous Trend Filter parameters."""

from dataclasses import dataclass
from enum import IntEnum
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


class NonambiguousTrendFilterBase(IntEnum):
    """Specifies the base oscillator the Nonambiguous Trend Filter is applied to."""

    TSI = 0
    """True Strength Index (named instance TSI_Trade); defaults q=2, r=32, s=13, u=3."""

    SMI = 1
    """Stochastic Momentum Index (named instance SMI_Trade); defaults q=32, r=64, s=7, u=1."""

    DTI = 2
    """Directional Trend Index (named instance DTI_Trade); defaults q=2, r=28, s=28, u=5."""

    TVI = 3
    """Tick Volume Indicator (named instance TVI_Trade); defaults r=32, s=32, u=5."""

    MDI = 4
    """Mean Deviation Index (MDI_Trade); defaults r=20, s=5, u=3."""

    CMI = 5
    """Candlestick Momentum Index (CMI_Trade); defaults r=20, s=5, u=3."""

    CSI = 6
    """Candlestick Strength Index (CSI_Trade); defaults r=32, s=32, u=1."""


@dataclass
class NonambiguousTrendFilterParams:
    """Parameters to create an instance of the Nonambiguous Trend Filter indicator.

    The filter itself is parameterless; ``q``, ``r``, ``s`` and ``u`` are the
    canonical symbols of the base oscillator from William Blau's *Momentum,
    Direction, and Divergence* (Wiley, 1995). A zero value selects the book
    default of the selected base (see :class:`NonambiguousTrendFilterBase`).
    """

    base: NonambiguousTrendFilterBase = NonambiguousTrendFilterBase.TSI
    """The base oscillator the filter is applied to.

    The default value is TSI.
    """

    q: int = 0
    """The momentum look-back period of the base (TSI, SMI and DTI only).

    Zero selects the base default. A non-zero value should be greater than 0.
    """

    r: int = 0
    """The period of the 1st EMA of the base smoothing cascade.

    Zero selects the base default. A non-zero value should be greater than 0.
    """

    s: int = 0
    """The period of the 2nd EMA of the base smoothing cascade.

    Zero selects the base default. A non-zero value should be greater than 0.
    """

    u: int = 0
    """The period of the 3rd EMA of the base smoothing cascade.

    Zero selects the base default. A non-zero value should be greater than 0.
    """

    bar_component: Optional[BarComponent] = None
    """A component of a bar to use when updating the indicator with a bar sample.

    Used only by the TSI and MDI bases. If not set, the bar component defaults
    to ClosePrice and is not shown in the indicator mnemonic.
    """

    quote_component: Optional[QuoteComponent] = None
    """A component of a quote to use when updating the indicator with a quote sample.

    Used only by the TSI and MDI bases. If not set, the quote component defaults
    to MidPrice and is not shown in the indicator mnemonic.
    """

    trade_component: Optional[TradeComponent] = None
    """A component of a trade to use when updating the indicator with a trade sample.

    Used only by the TSI and MDI bases. If not set, the trade component defaults
    to Price and is not shown in the indicator mnemonic.
    """


def default_params() -> NonambiguousTrendFilterParams:
    """Returns default parameters for the Nonambiguous Trend Filter indicator.

    The base periods are left at zero so that they resolve to the book defaults
    of whichever base is selected.
    """
    return NonambiguousTrendFilterParams()
