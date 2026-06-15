"""Sinc Wavelet Band-Pass parameters."""

from dataclasses import dataclass
from enum import IntEnum
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


class Band(IntEnum):
    """Frequency band selection for the sinc wavelet band-pass filter."""

    HIGH = 0
    """Extracts periods 8-16 bars."""

    MID = 1
    """Extracts periods 16-32 bars."""

    LOW = 2
    """Extracts periods 32-64 bars."""

    FULL = 3
    """Extracts periods 8-64 bars (sum of HIGH + MID + LOW)."""


@dataclass
class SincWaveletBandpassParams:
    """Parameters to create an instance of the Sinc Wavelet Band-Pass indicator."""

    band: Band = Band.MID
    """Selects the frequency band (HIGH, MID, LOW, FULL).

    The default value is MID.
    """

    velocity: bool = False
    """Controls whether a cubic velocity kernel is applied to the band-pass output.

    The default value is False.
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


def default_params() -> SincWaveletBandpassParams:
    """Returns default parameters for the Sinc Wavelet Band-Pass indicator."""
    return SincWaveletBandpassParams()
