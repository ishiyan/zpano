"""Parameters for the CoronaSpectrum indicator."""

from __future__ import annotations

import math
from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ..corona.corona import format_number


@dataclass
class CoronaSpectrumParams:
    """Configuration for the Corona Spectrum indicator.

    All fields default to Ehlers' November 2008 TASC article settings.
    Explicit zero values are NOT substituted with defaults.
    """

    min_raster_value: float = 6.0
    """Selection threshold in decibels (`0 <= min_raster_value < max_raster_value`)."""

    max_raster_value: float = 20.0
    """Intensity ceiling in decibels (`> 0` and `> min_raster_value`)."""

    min_parameter_value: float = 6.0
    """Minimum ordinate (y) value of the heatmap, rounded up to integer `>= 6`."""

    max_parameter_value: float = 30.0
    """Maximum ordinate (y) value of the heatmap, rounded down to integer `> min_period`."""

    high_pass_filter_cutoff: int = 30
    """High-pass filter cutoff (de-trending period) in bars (`>= 5`)."""

    bar_component: Optional[BarComponent] = None
    """Bar component used by update_bar(); None defaults to Close."""

    quote_component: Optional[QuoteComponent] = None
    """Quote component used by update_quote(); None defaults to Mid."""

    trade_component: Optional[TradeComponent] = None
    """Trade component used by update_trade(); None defaults to Price."""


Params = CoronaSpectrumParams


def default_params() -> CoronaSpectrumParams:
    """Return default Ehlers parameters."""
    return CoronaSpectrumParams()


def corona_spectrum_mnemonic(params: CoronaSpectrumParams) -> str:
    """Build the primary CoronaSpectrum mnemonic."""
    bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
    qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
    tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT
    comp = component_triple_mnemonic(bc, qc, tc)
    min_period = int(math.ceil(params.min_parameter_value))
    max_period = int(math.floor(params.max_parameter_value))
    return (
        f"cspect({format_number(params.min_raster_value)}, "
        f"{format_number(params.max_raster_value)}, "
        f"{min_period}, {max_period}, {params.high_pass_filter_cutoff}{comp})"
    )
