"""Parameters for the Comb Band-Pass Spectrum."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class Params:
    """Parameters for creating a CombBandPassSpectrum instance."""

    min_period: int = 0
    """Minimum (shortest) cycle period covered by the spectrum. Must be >= 2 (Nyquist).
    Also drives the cutoff of the Super Smoother pre-filter. The default value is 10.
    A zero value is treated as "use default".
    """

    max_period: int = 0
    """Maximum (longest) cycle period covered by the spectrum. Must be > minPeriod. Also
    drives the cutoff of the Butterworth highpass pre-filter and the band-pass output
    history length per bin. The default value is 48. A zero value is treated as "use
    default".
    """

    bandwidth: float = 0.0
    """Fractional bandwidth of each band-pass filter in the comb. Must be in (0, 1).
    Typical Ehlers values are around 0.3 (default) for medium selectivity. A zero
    value is treated as "use default".
    """

    disable_spectral_dilation_compensation: bool = False
    """Disables the spectral dilation compensation (division of each band-pass output by
    its evaluated period before squaring) when true. Ehlers' default is enabled, so the
    default value is false (SDC on).
    """

    disable_automatic_gain_control: bool = False
    """Disables the fast-attack slow-decay automatic gain control when true. Ehlers'
    default is enabled, so the default value is false (AGC on).
    """

    automatic_gain_control_decay_factor: float = 0.0
    """Decay factor used by the fast-attack slow-decay automatic gain control. Must be in
    the open interval (0, 1) when AGC is enabled. The default value is 0.995 (matching
    Ehlers' EasyLanguage listing 10-1). A zero value is treated as "use default".
    """

    fixed_normalization: bool = False
    """Selects fixed (min clamped to 0) normalization when true. The default is floating
    normalization, consistent with the other zpano spectrum heatmaps. Note that Ehlers'
    listing 10-1 uses fixed normalization (MaxPwr only); set this to true for exact
    EL-faithful behavior.
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
    """Returns a Params with Ehlers defaults."""
    return Params(
        min_period=10,
        max_period=48,
        bandwidth=0.3,
        automatic_gain_control_decay_factor=0.995,
    )
