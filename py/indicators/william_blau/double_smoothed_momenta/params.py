"""Double-Smoothed Momenta parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class DoubleSmoothedMomentaParams:
    """Parameters to create an instance of the Double-Smoothed Momenta indicator.

    The parameter names ``a``, ``y`` and ``z`` are the canonical symbols from
    William Blau's double-smoothed momentum family. They are kept verbatim for
    fidelity with the catalog definition and the test-data naming.
    """

    a: int = 2
    """The highest/lowest close look-back.

    ``a=2`` gives the one-bar momentum of the RSI family; ``a>2`` gives a
    double-smoothed stochastic of the close. The value should be greater than 0.
    The default value is 2.
    """

    y: int = 2
    """The period of the inner (1st) smoothing EMA of the cascade.

    Setting ``y=1`` makes the inner stage a passthrough, so ``DM(2,1,z)`` is the
    EMA-form ``RSI(z)``. The value should be greater than 0. The default value is 2.
    """

    z: int = 14
    """The period of the outer (2nd) smoothing EMA of the cascade.

    This is the dominant smoothing (the RSI-style default 14). The value should
    be greater than 0. The default value is 14.
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


def default_params() -> DoubleSmoothedMomentaParams:
    """Returns default parameters for the Double-Smoothed Momenta indicator."""
    return DoubleSmoothedMomentaParams()
