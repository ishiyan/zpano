"""True Strength Index parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class TrueStrengthIndexParams:
    """Parameters to create an instance of the True Strength Index indicator.

    The parameter names ``q``, ``r``, ``s``, ``u`` and ``ul`` are the canonical
    symbols from William Blau's *Momentum, Direction, and Divergence* (Wiley,
    1995), chapter 2. They are kept verbatim for fidelity with the book, the
    MQL5 reference, and the test-data naming.
    """

    q: int = 2
    """The momentum look-back period; momentum is ``C_k - C_(k-(q-1))``.

    The look-back distance is ``q-1`` bars, so ``q=2`` is the one-bar momentum
    Blau uses throughout the book. The value should be greater than 0
    (``q >= 2`` is meaningful). The default value is 2.
    """

    r: int = 20
    """The period of the 1st (innermost) EMA in the smoothing cascade, applied
    to the momentum.

    The value should be greater than 0. The default value is 20.
    """

    s: int = 5
    """The period of the 2nd EMA in the smoothing cascade, applied to the output
    of the 1st EMA.

    The value should be greater than 0. The default value is 5.
    """

    u: int = 3
    """The period of the 3rd (outermost) EMA in the smoothing cascade, applied
    to the output of the 2nd EMA.

    Setting ``u=1`` switches the 3rd stage off (passthrough), yielding the
    book's classic double-smoothed TSI. The value should be greater than 0.
    The default value is 3.
    """

    ul: int = 3
    """The period of the signal-line EMA, applied to the oscillator to produce
    the second output (Blau's Ergodic signal line).

    Setting ``ul=1`` makes the signal a passthrough (signal == tsi every bar).
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


def default_params() -> TrueStrengthIndexParams:
    """Returns default parameters for the True Strength Index indicator."""
    return TrueStrengthIndexParams()
