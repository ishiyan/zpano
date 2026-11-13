"""RoofingFilter parameters."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT
from ...core.component_triple_mnemonic import component_triple_mnemonic


@dataclass
class RoofingFilterParams:
    """Parameters for the RoofingFilter indicator."""

    shortest_cycle_period: float = 10.0
    """The shortest (SuperSmoother low-pass) cycle period in bars (>= 2.0). Default 10.0."""

    longest_cycle_period: float = 48.0
    """The longest (high-pass) cycle period in bars (> shortest_cycle_period). Default 48.0."""

    highpass_poles: int = 2
    """Number of poles in the initial high-pass stage (1 or 2). Default 2."""

    zero_mean: bool = False
    """Whether to apply a post-smoothing 1-pole high-pass stage (requires highpass_poles == 1). Default False."""

    bar_component: Optional[BarComponent] = None
    """Bar component used by update_bar(); None defaults to Close and is omitted from the mnemonic."""

    quote_component: Optional[QuoteComponent] = None
    """Quote component used by update_quote(); None defaults to Mid and is omitted from the mnemonic."""

    trade_component: Optional[TradeComponent] = None
    """Trade component used by update_trade(); None defaults to Price and is omitted from the mnemonic."""


Params = RoofingFilterParams


def default_params() -> RoofingFilterParams:
    """Return default Ehlers RoofingFilter parameters."""
    return RoofingFilterParams()


def format_number(value: float) -> str:
    """Format a float parameter cleanly for mnemonics without '-0'."""
    normalized = 0.0 if value == 0.0 else float(value)
    return f"{normalized:g}"


def roofing_filter_mnemonic(params: Optional[RoofingFilterParams] = None) -> str:
    """Build the RoofingFilter mnemonic string."""
    source = default_params() if params is None else params
    bar_comp = source.bar_component if source.bar_component is not None else DEFAULT_BAR_COMPONENT
    quote_comp = source.quote_component if source.quote_component is not None else DEFAULT_QUOTE_COMPONENT
    trade_comp = source.trade_component if source.trade_component is not None else DEFAULT_TRADE_COMPONENT
    suffix = component_triple_mnemonic(bar_comp, quote_comp, trade_comp)
    zm_tag = "zm" if source.zero_mean else ""
    shortest_text = format_number(source.shortest_cycle_period)
    longest_text = format_number(source.longest_cycle_period)
    return f"roof{source.highpass_poles}hp{zm_tag}({shortest_text}, {longest_text}{suffix})"
