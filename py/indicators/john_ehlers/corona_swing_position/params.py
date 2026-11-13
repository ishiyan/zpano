"""Parameters for the CoronaSwingPosition indicator."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ..corona.corona import format_number


@dataclass
class CoronaSwingPositionParams:
    """Configuration for the Corona Swing Position indicator.

    All fields default to Ehlers' TASC settings. Explicit zero values are NOT
    substituted with defaults.
    """

    raster_length: int = 50
    """Length of the heatmap raster (`>= 2`). Default 50."""

    max_raster_value: float = 20.0
    """Maximum raster intensity value (`> 6`). Default 20."""

    min_parameter_value: float = -5.0
    """Minimum ordinate (y) value. Default -5."""

    max_parameter_value: float = 5.0
    """Maximum ordinate (y) value (`> min_parameter_value`). Default 5."""

    high_pass_filter_cutoff: int = 30
    """High-pass filter cutoff (`>= 5`). Default 30."""

    minimal_period: int = 6
    """Minimal cycle period (`>= 6`). Default 6."""

    maximal_period: int = 30
    """Maximal cycle period (`> minimal_period`). Default 30."""

    lead_window: int = 51
    """Window length for 60-degree leading signal normalization (`>= 2`). Default 51."""

    position_window: int = 21
    """Window length for position span fuzz/suppression (`>= 2`). Default 21."""

    bar_component: Optional[BarComponent] = None
    """Bar component used by update_bar(); None defaults to Close."""

    quote_component: Optional[QuoteComponent] = None
    """Quote component used by update_quote(); None defaults to Mid."""

    trade_component: Optional[TradeComponent] = None
    """Trade component used by update_trade(); None defaults to Price."""


Params = CoronaSwingPositionParams


def default_params() -> CoronaSwingPositionParams:
    """Return default Ehlers parameters."""
    return CoronaSwingPositionParams()


def swing_flags(params: CoronaSwingPositionParams) -> str:
    """Build optional override flags when period or window parameters differ from defaults."""
    parts: list[str] = []
    if params.minimal_period != 6:
        parts.append(f", pmin={params.minimal_period}")
    if params.maximal_period != 30:
        parts.append(f", pmax={params.maximal_period}")
    if params.lead_window != 51:
        parts.append(f", lead={params.lead_window}")
    if params.position_window != 21:
        parts.append(f", pos={params.position_window}")
    return "".join(parts)


def corona_swing_position_mnemonic(params: CoronaSwingPositionParams) -> str:
    """Build the primary CoronaSwingPosition mnemonic."""
    bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
    qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
    tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT
    comp = component_triple_mnemonic(bc, qc, tc)
    flags = swing_flags(params)
    return (
        f"cswing({params.raster_length}, {format_number(params.max_raster_value)}, "
        f"{format_number(params.min_parameter_value)}, "
        f"{format_number(params.max_parameter_value)}, "
        f"{params.high_pass_filter_cutoff}{flags}{comp})"
    )
