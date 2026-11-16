"""Parameters for Ehlers' Discrete Fourier Transform Spectrum."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT
from ...core.component_triple_mnemonic import component_triple_mnemonic


@dataclass
class DiscreteFourierTransformSpectrumParams:
    """Parameters for creating a DiscreteFourierTransformSpectrum instance.

    All fields default to Ehlers' 2013 Book Listing 9-1 settings.
    Explicit values (including 0 or False) are validated directly without silent substitution.
    """

    min_period: int = 10
    """Shortest evaluated integer cycle period in bars (>= 3). Default 10."""

    max_period: int = 48
    """Longest evaluated integer cycle period in bars (>= min_period). Default 48."""

    window_length: int = 48
    """Sliding roofed DFT window length in bars (>= 2, independent of max_period). Default 48."""

    highpass_period: float = 48.0
    """Two-pole roofing high-pass cutoff period in bars (>= 5.0). Default 48.0."""

    smoothing_period: float = 10.0
    """Two-pole SuperSmoother cutoff period in bars (>= 2.0). Default 10.0."""

    spectral_dilation_compensation: bool = True
    """Divide each cosine/sine Fourier contribution by period before summing/squaring. Default True."""

    enhance_resolution: bool = False
    """Cube normalized display heatmap intensities after centroid calculation. Default False."""

    bar_component: Optional[BarComponent] = None
    """Bar component used by update_bar(); None defaults to Close and is omitted from the mnemonic."""

    quote_component: Optional[QuoteComponent] = None
    """Quote component used by update_quote(); None defaults to Mid and is omitted from the mnemonic."""

    trade_component: Optional[TradeComponent] = None
    """Trade component used by update_trade(); None defaults to Price and is omitted from the mnemonic."""


Params = DiscreteFourierTransformSpectrumParams


def default_params() -> DiscreteFourierTransformSpectrumParams:
    """Return default Ehlers Listing 9-1 parameters."""
    return DiscreteFourierTransformSpectrumParams()


def format_number(value: float) -> str:
    """Format a float parameter cleanly for mnemonics without '-0'."""
    normalized = 0.0 if value == 0.0 else float(value)
    return f"{normalized:g}"


def _format_mnemonic(
    prefix: str,
    include_enhance: bool,
    params: Optional[DiscreteFourierTransformSpectrumParams] = None,
) -> str:
    source = default_params() if params is None else params
    bar_comp = (
        source.bar_component if source.bar_component is not None else DEFAULT_BAR_COMPONENT
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
    sdc_suffix = "" if source.spectral_dilation_compensation else ", no-sdc"
    enh_suffix = ", enh" if (include_enhance and source.enhance_resolution) else ""
    hp_str = format_number(source.highpass_period)
    lp_str = format_number(source.smoothing_period)
    return (
        f"{prefix}({source.min_period}, {source.max_period}, {source.window_length}, "
        f"{hp_str}, {lp_str}{sdc_suffix}{enh_suffix}{comp_suffix})"
    )


def discrete_fourier_transform_spectrum_mnemonic(
    params: Optional[DiscreteFourierTransformSpectrumParams] = None,
) -> str:
    """Build the primary heatmap mnemonic for DiscreteFourierTransformSpectrum."""
    return _format_mnemonic("dftps", True, params)


def discrete_fourier_transform_spectrum_dominant_cycle_mnemonic(
    params: Optional[DiscreteFourierTransformSpectrumParams] = None,
) -> str:
    """Build the dominant-cycle mnemonic for DiscreteFourierTransformSpectrum."""
    return _format_mnemonic("dftps-dc", False, params)
