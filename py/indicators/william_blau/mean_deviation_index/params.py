"""Mean Deviation Index parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class MeanDeviationIndexParams:
    """Parameters to create an instance of the Mean Deviation Index indicator.

    The parameter names ``r``, ``s``, ``u`` and ``ul`` are the canonical symbols
    from William Blau's *Momentum, Direction, and Divergence* (Wiley, 1995),
    chapter 5. They are kept verbatim for fidelity with the book, the MQL5
    reference, and the test-data naming.
    """

    r: int = 20
    """The period of the baseline (detrending) EMA subtracted from the price.

    The mean deviation is ``md_k = price_k - EMA(price, r)_k``. Setting ``r=1``
    makes the baseline a passthrough, so the deviation is 0 on every bar and the
    index is identically 0. The value should be greater than 0. The default
    value is 20.
    """

    s: int = 5
    """The period of the 1st smoothing EMA, applied to the mean deviation.

    The value should be greater than 0. The default value is 5.
    """

    u: int = 3
    """The period of the 2nd smoothing EMA, applied to the output of the 1st
    smoothing EMA.

    Setting ``u=1`` switches the 2nd stage off (passthrough), yielding the
    book's pure double-smoothed form ``EMA(price - EMA(price, r), s)``. The
    value should be greater than 0. The default value is 3.
    """

    ul: int = 3
    """The period of the signal-line EMA, applied to the index to produce the
    second output (Blau's Ergodic signal line).

    Setting ``ul=1`` makes the signal a passthrough (signal == mdi every bar).
    The value should be greater than 0. This parameter is not shown in the
    indicator mnemonic. The default value is 3.
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


def default_params() -> MeanDeviationIndexParams:
    """Returns default parameters for the Mean Deviation Index indicator."""
    return MeanDeviationIndexParams()
