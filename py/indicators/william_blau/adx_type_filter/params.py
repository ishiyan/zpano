"""ADX-Type Filter parameters."""

from dataclasses import dataclass
from enum import IntEnum
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


class AdxTypeFilterSource(IntEnum):
    """Specifies the bipolar momentum the ADX-Type Filter is applied to."""

    TSI_MOMENTUM = 0
    """The TSI numerator ``C - C[q-1]`` (named instance TSI_ATF)."""

    SMI_MOMENTUM = 1
    """The SMI raw stochastic momentum ``C - 0.5*(HH(q) + LL(q))`` (named instance SMI_ATF)."""

    DTI_MOMENTUM = 2
    """The DTI numerator ``max(H - H[q-1], 0) - max(L[q-1] - L, 0)``."""

    TVI_BALANCE = 3
    """The TVI tick balance ``upticks - downticks`` (``2C - H - L`` for a bar)."""

    TSI_NORMALIZED = 4
    """The single-smoothed normalized ``TSI(q, r, 1, 1)``, replacing the inner EMA."""


@dataclass
class AdxTypeFilterParams:
    """Parameters to create an instance of the ADX-Type Filter indicator.

    The parameter names ``q``, ``r`` and ``s`` are the canonical symbols from
    William Blau's *Momentum, Direction, and Divergence* (Wiley, 1995),
    Appendix B, Figure B-24. They are kept verbatim for fidelity with the book
    and the test-data naming.
    """

    source: AdxTypeFilterSource = AdxTypeFilterSource.TSI_MOMENTUM
    """The bipolar momentum the filter is applied to.

    The default value is TSI_MOMENTUM.
    """

    q: int = 0
    """The momentum look-back period.

    Zero selects the source default: 2 for TSI_MOMENTUM, DTI_MOMENTUM and
    TSI_NORMALIZED, 32 for SMI_MOMENTUM. The look-back is not used by
    TVI_BALANCE. A non-zero value should be greater than 0.
    """

    r: int = 32
    """The period of the inner EMA, applied to the signed momentum.

    For TSI_NORMALIZED this is the smoothing period of the normalized TSI, which
    replaces the inner EMA. The value should be greater than 0. The default
    value is 32.
    """

    s: int = 32
    """The period of the outer EMA, applied to the rectified momentum.

    The value should be greater than 0. The default value is 32.
    """

    bar_component: Optional[BarComponent] = None
    """A component of a bar to use when updating the indicator with a bar sample.

    Used only by the TSI_MOMENTUM and TSI_NORMALIZED sources. If not set, the
    bar component defaults to ClosePrice and is not shown in the indicator mnemonic.
    """

    quote_component: Optional[QuoteComponent] = None
    """A component of a quote to use when updating the indicator with a quote sample.

    Used only by the TSI_MOMENTUM and TSI_NORMALIZED sources. If not set, the
    quote component defaults to MidPrice and is not shown in the indicator mnemonic.
    """

    trade_component: Optional[TradeComponent] = None
    """A component of a trade to use when updating the indicator with a trade sample.

    Used only by the TSI_MOMENTUM and TSI_NORMALIZED sources. If not set, the
    trade component defaults to Price and is not shown in the indicator mnemonic.
    """


def default_params() -> AdxTypeFilterParams:
    """Returns default parameters for the ADX-Type Filter indicator.

    The look-back ``q`` is left at zero so that it resolves to the default of
    whichever source is selected.
    """
    return AdxTypeFilterParams()
