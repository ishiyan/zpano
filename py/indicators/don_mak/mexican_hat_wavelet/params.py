"""Mexican Hat Wavelet parameters."""

from dataclasses import dataclass
from enum import IntEnum
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


class Band(IntEnum):
    """Frequency band selection for the Mexican Hat Wavelet filter."""

    HIGH = 0
    """The high-frequency band (a_f = 1.483, period ~ 4.6 bars)."""

    MID = 1
    """The mid-frequency band (a_f = 4.048, period ~ 13.5 bars)."""

    LOW = 2
    """The low-frequency band (a_f = 15.97, period ~ 54 bars)."""

    CUSTOM = 3
    """Uses a user-specified dilation or period."""


@dataclass
class MexicanHatWaveletParams:
    """Parameters to create an instance of the Mexican Hat Wavelet indicator."""

    band: Band = Band.MID
    """Selects the frequency band (HIGH, MID, LOW, CUSTOM).

    The default value is MID.
    """

    dilation: float = 0.0
    """The custom dilation parameter a_f, used only when band is CUSTOM.

    The value should be > 0. Zero means unset.
    """

    period: float = 0.0
    """The custom center period in bars, used only when band is CUSTOM.

    The value should be > 2. Zero means unset. Mutually exclusive with dilation.
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


def default_params() -> MexicanHatWaveletParams:
    """Returns default parameters for the Mexican Hat Wavelet indicator."""
    return MexicanHatWaveletParams()
