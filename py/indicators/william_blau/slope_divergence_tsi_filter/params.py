"""Slope Divergence TSI Filter parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class SlopeDivergenceTsiFilterParams:
    """Parameters to create an instance of the Slope Divergence TSI Filter indicator.

    The parameter names ``q``, ``r``, ``s``, ``u``, ``x`` and ``y`` are the
    canonical symbols from William Blau's *Momentum, Direction, and Divergence*
    (Wiley, 1995), chapter 12 and Appendix B, Figure B-25. They are kept
    verbatim for fidelity with the book and the test-data naming.
    """

    q: int = 2
    """The TSI momentum look-back period; momentum is ``C_k - C_(k-(q-1))``.

    The look-back distance is ``q-1`` bars, so ``q=2`` is the one-bar momentum
    Blau uses throughout the book. The value should be greater than 0.
    The default value is 2.
    """

    r: int = 32
    """The period of the 1st (innermost) EMA of the TSI smoothing cascade.

    The value should be greater than 0. The default value is 32.
    """

    s: int = 32
    """The period of the 2nd EMA of the TSI smoothing cascade.

    The value should be greater than 0. The default value is 32.
    """

    u: int = 7
    """The period of the 3rd (outermost) EMA of the TSI smoothing cascade.

    Setting ``u=1`` switches the 3rd stage off (passthrough), yielding the
    double-smoothed TSI of the book's raw form (Fig. 12-1). The value should be
    greater than 0. The default value is 7.
    """

    x: int = 32
    """The period of the 1st EMA of the price reference ``DEMA(close, x, y)``.

    The value should be greater than 0. The default value is 32.
    """

    y: int = 7
    """The period of the 2nd EMA of the price reference ``DEMA(close, x, y)``.

    Setting ``y=1`` makes the price reference a single EMA. The value should be
    greater than 0. The default value is 7.
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


def default_params() -> SlopeDivergenceTsiFilterParams:
    """Returns default parameters for the Slope Divergence TSI Filter indicator."""
    return SlopeDivergenceTsiFilterParams()
