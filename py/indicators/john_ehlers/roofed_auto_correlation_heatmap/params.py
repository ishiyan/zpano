"""Parameters for Ehlers' Roofed Autocorrelation Heatmap."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT
from ....entities.quote_component import DEFAULT_QUOTE_COMPONENT, QuoteComponent
from ....entities.trade_component import DEFAULT_TRADE_COMPONENT, TradeComponent
from ...core.component_triple_mnemonic import component_triple_mnemonic


@dataclass
class RoofedAutoCorrelationHeatmapParams:
    """Parameters to create an instance of RoofedAutoCorrelationHeatmap."""

    min_lag: int = 3
    """The first displayed correlation lag on the heatmap axis.

    The value must be an integer >= 0. The default value is 3.
    """

    max_lag: int = 48
    """The last displayed correlation lag on the heatmap axis.

    The value must be an integer >= min_lag. The default value is 48.
    """

    averaging_length: int = 0
    """The number of sample pairs per Pearson correlation comparison.

    Zero selects lag-dependent windows (M = lag); a positive integer selects a fixed M >= 1. The default value is 0.
    """

    highpass_period: float = 48.0
    """The two-pole roofing high-pass cutoff period in bars.

    The value must be a finite real number >= 5.0. The default value is 48.0.
    """

    smoothing_period: float = 10.0
    """The two-pole SuperSmoother cutoff period in bars.

    The value must be a finite real number >= 2.0. The default value is 10.0.
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


Params = RoofedAutoCorrelationHeatmapParams


def default_params() -> RoofedAutoCorrelationHeatmapParams:
    """Return default Ehlers 2013 Listing 8-2 parameters."""
    return RoofedAutoCorrelationHeatmapParams()


def format_number(value: float) -> str:
    """Format a float parameter cleanly for mnemonics without '-0'."""
    normalized = 0.0 if value == 0.0 else float(value)
    return f"{normalized:g}"


def roofed_auto_correlation_heatmap_mnemonic(
    params: Optional[RoofedAutoCorrelationHeatmapParams] = None,
) -> str:
    """Build the mnemonic for RoofedAutoCorrelationHeatmap."""
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
    lp_text = format_number(source.smoothing_period)
    return (
        f"rach({source.min_lag}, {source.max_lag}, {source.averaging_length}, "
        f"{hp_text}, {lp_text}{comp_suffix})"
    )
