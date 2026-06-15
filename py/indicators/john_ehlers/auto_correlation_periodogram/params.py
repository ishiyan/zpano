"""Parameters for the AutoCorrelation Periodogram."""

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent


@dataclass
class Params:
    """Parameters for creating an AutoCorrelationPeriodogram instance."""

    min_period: int = 0
    """Minimum (shortest) cycle period covered by the periodogram. Must be >= 2 (Nyquist).
    Also drives the cutoff of the Super Smoother pre-filter. The default value is 10.
    A zero value is treated as "use default".
    """

    max_period: int = 0
    """Maximum (longest) cycle period covered by the periodogram. Must be > minPeriod. Also
    drives the cutoff of the Butterworth highpass pre-filter, the autocorrelation lag
    range, and the DFT basis length. The default value is 48. A zero value is treated
    as "use default".
    """

    averaging_length: int = 0
    """Number of samples (M) used in each Pearson correlation accumulation. Must be >= 1.
    The default value is 3 (matching Ehlers' EasyLanguage listing 8-3, which hardcodes 3).
    A zero value is treated as "use default".
    """

    disable_spectral_squaring: bool = False
    """Disables squaring the Fourier magnitude before smoothing when true. Ehlers' default
    EasyLanguage listing 8-3 squares SqSum (R[P] = 0.2·SqSum² + 0.8·R_previous[P]); the default
    value is false (squaring on).
    """

    disable_smoothing: bool = False
    """Disables the per-bin exponential smoothing when true. Ehlers' default is enabled,
    so the default value is false (smoothing on).
    """

    disable_automatic_gain_control: bool = False
    """Disables the fast-attack slow-decay automatic gain control when true. Ehlers'
    default is enabled, so the default value is false (AGC on).
    """

    automatic_gain_control_decay_factor: float = 0.0
    """Decay factor used by the fast-attack slow-decay automatic gain control. Must be in
    the open interval (0, 1) when AGC is enabled. The default value is 0.995. A zero
    value is treated as "use default".
    """

    fixed_normalization: bool = False
    """Selects fixed (min clamped to 0) normalization when true. The default is floating
    normalization, consistent with the other zpano spectrum heatmaps.
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
        averaging_length=3,
        automatic_gain_control_decay_factor=0.995,
    )
