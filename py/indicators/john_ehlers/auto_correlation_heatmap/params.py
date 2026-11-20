"""Parameters for Ehlers' Autocorrelation Heatmap."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT
from ....entities.quote_component import DEFAULT_QUOTE_COMPONENT, QuoteComponent
from ....entities.trade_component import DEFAULT_TRADE_COMPONENT, TradeComponent
from ...core.component_triple_mnemonic import component_triple_mnemonic


@dataclass
class AutoCorrelationHeatmapParams:
    """Parameters to create an instance of AutoCorrelationHeatmap."""

    length: int = 20
    """The number of aligned sample pairs per Pearson correlation comparison.

    The value must be an integer >= 2. The default value is 20.
    """

    smoothing_period: float = 20.0
    """The Ultimate Smoother critical period in bars.

    The value must be a finite real number >= 1.0. The default value is 20.0.
    """

    max_lag: int = 99
    """The inclusive upper correlation lag; minimum lag is always zero.

    The value must be an integer in [0, 99]. The default value is 99.
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


Params = AutoCorrelationHeatmapParams


def default_params() -> AutoCorrelationHeatmapParams:
    """Return default Ehlers February 2025 article parameters."""
    return AutoCorrelationHeatmapParams()


def format_number(value: float) -> str:
    """Format a float parameter cleanly for mnemonics without '-0'."""
    normalized = 0.0 if value == 0.0 else float(value)
    return f"{normalized:g}"


def auto_correlation_heatmap_mnemonic(
    params: Optional[AutoCorrelationHeatmapParams] = None,
) -> str:
    """Build the mnemonic for AutoCorrelationHeatmap."""
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
    period_text = format_number(source.smoothing_period)
    return f"ach({source.length}, {period_text}, {source.max_lag}{comp_suffix})"
