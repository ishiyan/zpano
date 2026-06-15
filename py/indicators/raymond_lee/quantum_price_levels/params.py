"""Quantum Price Levels parameters."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class QuantumPriceLevelsParams:
    """Parameters to create an instance of the Quantum Price Levels indicator."""

    lookback: int = 2048
    """The number of price-return ratios maintained in the sliding window.

    Priming requires lookback+1 prices. The value should be >= 2. The default value is 2048.
    """

    num_levels: int = 21
    """The number of quantum energy levels to compute (n = 0..num_levels-1).

    The value should be >= 1. The default value is 21.
    """

    num_bins: int = 100
    """The number of histogram bins for the wavefunction distribution.

    The value should be >= 2. The default value is 100.
    """

    scale_factor: float = 0.21
    """The empirical scaling constant in the NQPR formula (1 + scale_factor*sigma*QPR).

    The value should be > 0. The default value is 0.21.
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


def default_params() -> QuantumPriceLevelsParams:
    """Returns default parameters for the Quantum Price Levels indicator."""
    return QuantumPriceLevelsParams()
