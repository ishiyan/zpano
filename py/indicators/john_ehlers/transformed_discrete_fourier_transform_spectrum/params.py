"""Parameters for Ehlers' Transformed Discrete Fourier Transform Spectrum."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT
from ....entities.quote_component import DEFAULT_QUOTE_COMPONENT, QuoteComponent
from ....entities.trade_component import DEFAULT_TRADE_COMPONENT, TradeComponent
from ...core.component_triple_mnemonic import component_triple_mnemonic


@dataclass
class TransformedDiscreteFourierTransformSpectrumParams:
    """Parameters to create an instance of TransformedDiscreteFourierTransformSpectrum."""

    min_period: int = 8
    """The shortest evaluated integer cycle period in bars.

    The value must be an integer >= 8. The default value is 8.
    """

    max_period: int = 50
    """The longest evaluated integer cycle period in bars.

    The value must be an integer >= min_period. The default value is 50.
    """

    window_length: int = 50
    """The moving observation window length in cleaned samples for each Fourier sum.

    The value must be an integer >= 1. The default value is 50.
    """

    highpass_period: float = 40.0
    """The high-pass filter cutoff period in bars.

    The value must be a finite real number > 4.0. The default value is 40.0.
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


Params = TransformedDiscreteFourierTransformSpectrumParams


def default_params() -> TransformedDiscreteFourierTransformSpectrumParams:
    """Return default Ehlers January 2007 article parameters."""
    return TransformedDiscreteFourierTransformSpectrumParams()


def format_number(value: float) -> str:
    """Format a float parameter cleanly for mnemonics without '-0'."""
    normalized = 0.0 if value == 0.0 else float(value)
    return f"{normalized:g}"


def _format_mnemonic(
    prefix: str,
    params: Optional[TransformedDiscreteFourierTransformSpectrumParams] = None,
) -> str:
    source = default_params() if params is None else params
    bar_comp = (
        source.bar_component
        if source.bar_component is not None
        else DEFAULT_BAR_COMPONENT
    )
    quote_comp = (
        source.quote_component
        if source.quote_component is not None
        else DEFAULT_QUOTE_COMPONENT
    )
    trade_comp = (
        source.trade_component
        if source.trade_component is not None
        else DEFAULT_TRADE_COMPONENT
    )
    comp_suffix = component_triple_mnemonic(bar_comp, quote_comp, trade_comp)
    hp_text = format_number(source.highpass_period)
    return (
        f"{prefix}({source.min_period}, {source.max_period}, "
        f"{source.window_length}, {hp_text}{comp_suffix})"
    )


def transformed_discrete_fourier_transform_spectrum_mnemonic(
    params: Optional[TransformedDiscreteFourierTransformSpectrumParams] = None,
) -> str:
    """Build the primary heatmap mnemonic for TransformedDiscreteFourierTransformSpectrum."""
    return _format_mnemonic("tdfts", params)


def transformed_discrete_fourier_transform_spectrum_dominant_cycle_mnemonic(
    params: Optional[TransformedDiscreteFourierTransformSpectrumParams] = None,
) -> str:
    """Build the dominant-cycle mnemonic for TransformedDiscreteFourierTransformSpectrum."""
    return _format_mnemonic("tdfts-dc", params)
