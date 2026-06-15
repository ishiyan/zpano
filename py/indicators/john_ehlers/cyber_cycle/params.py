"""CyberCycle parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class CyberCycleLengthParams:
    """Parameters for the CyberCycle indicator based on length."""
    length: int = 28
    """Length is the length (the number of time periods, ℓ) of the Cyber Cycle.

    The smoothing factor α is derived as α = 2/(ℓ+1).
    The value should be a positive integer, greater or equal to 1.
    """

    signal_lag: int = 9
    """SignalLag is the signal lag (the number of time periods) for the signal line EMA.

    The signal EMA factor is 1/(signalLag+1).
    The value should be a positive integer, greater or equal to 1.
    The default value used by Ehlers is 9.
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
class CyberCycleSmoothingFactorParams:
    """Parameters for the CyberCycle indicator based on smoothing factor."""
    smoothing_factor: float = 0.07
    """SmoothingFactor is the smoothing factor, α in [0,1], of the Cyber Cycle.

    The equivalent length ℓ is:

        ℓ = round(2/α) - 1, 0<α≤1, 1≤ℓ.

    If α is near zero (< epsilon), ℓ is set to Number.MAX_SAFE_INTEGER.
    The default value used by Ehlers is 0.07 (ℓ = 28).
    """

    signal_lag: int = 9
    """SignalLag is the signal lag (the number of time periods) for the signal line EMA.

    The signal EMA factor is 1/(signalLag+1).
    The value should be a positive integer, greater or equal to 1.
    The default value used by Ehlers is 9.
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


def default_length_params() -> CyberCycleLengthParams:
    """Returns default length-based parameters."""
    return CyberCycleLengthParams()


def default_smoothing_factor_params() -> CyberCycleSmoothingFactorParams:
    """Returns default smoothing-factor-based parameters."""
    return CyberCycleSmoothingFactorParams()
