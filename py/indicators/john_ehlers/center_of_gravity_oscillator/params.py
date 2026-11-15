"""CenterOfGravityOscillator parameters."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Optional

from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT
from ...core.component_triple_mnemonic import component_triple_mnemonic


@dataclass
class CenterOfGravityOscillatorParams:
    """Parameters for the CenterOfGravityOscillator indicator."""

    length: int = 10
    """Observation window length (integer >= 1). Default 10."""

    centered: bool = False
    """Zero-centered 2004 book baseline (-Num/Denom + (length+1)/2) when True; May 2002 article baseline (-Num/Denom) when False."""

    bar_component: Optional[BarComponent] = None
    """Bar component used by update_bar(); None defaults to Close and is omitted from the mnemonic."""

    quote_component: Optional[QuoteComponent] = None
    """Quote component used by update_quote(); None defaults to Mid and is omitted from the mnemonic."""

    trade_component: Optional[TradeComponent] = None
    """Trade component used by update_trade(); None defaults to Price and is omitted from the mnemonic."""


Params = CenterOfGravityOscillatorParams


def default_params() -> CenterOfGravityOscillatorParams:
    """Returns default CenterOfGravityOscillator parameters."""
    return CenterOfGravityOscillatorParams()


def _format_mnemonic(
    prefix: str, params: Optional[CenterOfGravityOscillatorParams] = None
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
    centered_tag = ", centered" if source.centered else ""
    suffix = component_triple_mnemonic(bar_comp, quote_comp, trade_comp)
    return f"{prefix}({source.length}{centered_tag}{suffix})"


def center_of_gravity_oscillator_mnemonic(
    params: Optional[CenterOfGravityOscillatorParams] = None,
) -> str:
    """Build the primary CenterOfGravityOscillator mnemonic string."""
    return _format_mnemonic("cog", params)


def center_of_gravity_oscillator_trigger_mnemonic(
    params: Optional[CenterOfGravityOscillatorParams] = None,
) -> str:
    """Build the CenterOfGravityOscillator trigger mnemonic string."""
    return _format_mnemonic("cogTrig", params)
