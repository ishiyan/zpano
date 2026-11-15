"""FractalAdaptiveMovingAverage parameters."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT
from ...core.component_triple_mnemonic import component_triple_mnemonic


@dataclass
class FractalAdaptiveMovingAverageParams:
    """Parameters for the FractalAdaptiveMovingAverage indicator."""

    length: int = 16
    """Lookback length (must be an even integer >= 2). Default 16."""

    slowest_smoothing_factor: float = 0.01
    """Slowest boundary smoothing factor (finite, 0 < A <= 1). Default 0.01."""

    coefficient_mode: int = 1
    """Alpha exponent scale mode: 1 for full-precision ln(A), 0 for literal 2005 -4.6 (requires A == 0.01). Default 1."""

    bar_component: Optional[BarComponent] = None
    """Bar component used by update_bar(); None defaults to Close and is omitted from the mnemonic."""

    quote_component: Optional[QuoteComponent] = None
    """Quote component used by update_quote(); None defaults to Mid and is omitted from the mnemonic."""

    trade_component: Optional[TradeComponent] = None
    """Trade component used by update_trade(); None defaults to Price and is omitted from the mnemonic."""


Params = FractalAdaptiveMovingAverageParams


def default_params() -> FractalAdaptiveMovingAverageParams:
    """Returns default FractalAdaptiveMovingAverage parameters."""
    return FractalAdaptiveMovingAverageParams()


def format_number(value: float) -> str:
    """Format a float parameter cleanly for mnemonics without '-0'."""
    normalized = 0.0 if value == 0.0 else float(value)
    return f"{normalized:g}"


def _format_mnemonic(
    prefix: str, params: Optional[FractalAdaptiveMovingAverageParams] = None
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
    mode_tag = ", mode=0" if source.coefficient_mode == 0 else ""
    suffix = component_triple_mnemonic(bar_comp, quote_comp, trade_comp)
    slowest_text = format_number(source.slowest_smoothing_factor)
    return f"{prefix}({source.length}, {slowest_text}{mode_tag}{suffix})"


def fractal_adaptive_moving_average_mnemonic(
    params: Optional[FractalAdaptiveMovingAverageParams] = None,
) -> str:
    """Build the primary FractalAdaptiveMovingAverage mnemonic string."""
    return _format_mnemonic("frama", params)


def fractal_adaptive_moving_average_dimension_mnemonic(
    params: Optional[FractalAdaptiveMovingAverageParams] = None,
) -> str:
    """Build the FractalAdaptiveMovingAverage dimension mnemonic string."""
    return _format_mnemonic("framaDim", params)
