"""Parameters for the MaximumEntropySpectrum indicator."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class Params:
    """Configuration for the Maximum Entropy Spectrum indicator."""
    length: int = 0
    """Number of time periods in the spectrum window. Must be >= 2. The default value is 60.
    A zero value is treated as "use default".
    """

    degree: int = 0
    """The degree (order) of auto-regression. Must be > 0 and < length. The default value
    is 30. A zero value is treated as "use default".
    """

    min_period: float = 0.0
    """Minimum cycle period covered by the spectrum. Must be >= 2 (2 corresponds to the Nyquist
    frequency). The default value is 2. A zero value is treated as "use default".
    """

    max_period: float = 0.0
    """Maximum cycle period covered by the spectrum. Must be > minPeriod and <= 2 * length.
    The default value is 59. A zero value is treated as "use default".
    """

    spectrum_resolution: int = 0
    """Spectrum resolution (positive integer). A value of 10 means the spectrum is evaluated at
    every 0.1 of period amplitude. Must be >= 1. The default value is 1. A zero value is
    treated as "use default".
    """

    disable_automatic_gain_control: bool = False
    """Disables the fast-attack slow-decay automatic gain control when true. MBST default
    behavior is enabled, so the default value is false (AGC on).
    """

    automatic_gain_control_decay_factor: float = 0.0
    """Decay factor used by the fast-attack slow-decay automatic gain control. Must be in the
    open interval (0, 1) when AGC is enabled. The default value is 0.995. A zero value is
    treated as "use default".
    """

    fixed_normalization: bool = False
    """Selects fixed (min clamped to 0) normalization when true. MBST default is floating
    normalization, so the default value is false (floating normalization).
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


def default_params() -> Params:
    """Return default parameters."""
    return Params(
        length=60,
        degree=30,
        min_period=2.0,
        max_period=59.0,
        spectrum_resolution=1,
        automatic_gain_control_decay_factor=0.995,
    )
