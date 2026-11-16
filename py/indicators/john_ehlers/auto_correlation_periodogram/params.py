"""Parameters for Ehlers' Autocorrelation Periodogram."""

from __future__ import annotations

from dataclasses import dataclass
from enum import Enum
from typing import Optional

from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT
from ...core.component_triple_mnemonic import component_triple_mnemonic


class AutoCorrelationPeriodogramSource(str, Enum):
    """Selects the Ehlers source profile for AutoCorrelationPeriodogram."""

    ARTICLE = "article"
    """TASC September 2016 'Measuring Market Cycles' sidebar profile (default)."""

    BOOK = "book"
    """Corrected Cycle Analytics for Traders (2013), Listing 8-3 profile."""


Source = AutoCorrelationPeriodogramSource


@dataclass
class AutoCorrelationPeriodogramParams:
    """Parameters for creating an AutoCorrelationPeriodogram instance.

    ``min_period=None`` selects ``8`` for ``"article"`` and ``10`` for ``"book"``.
    ``averaging_length=0`` selects lag-dependent Pearson windows (``M = lag``).
    """

    source: AutoCorrelationPeriodogramSource | str = (
        AutoCorrelationPeriodogramSource.ARTICLE
    )
    """Source profile: ``'article'`` (default) or ``'book'``."""

    min_period: Optional[int] = None
    """Shortest spectrum bin and SuperSmoother cutoff in bars (>= 3); None selects 8 or 10."""

    max_period: int = 48
    """Longest spectrum bin, high-pass cutoff, and largest lag (>= 5, > min_period). Default 48."""

    averaging_length: int = 3
    """Pearson window length: 0 for lag-dependent (M = lag) or >= 2. Default 3."""

    enhance_resolution: bool = False
    """Cube power before centroid in article mode or square display power after centroid in book mode."""

    bar_component: Optional[BarComponent] = None
    """Bar component used by update_bar(); None defaults to Close and is omitted from the mnemonic."""

    quote_component: Optional[QuoteComponent] = None
    """Quote component used by update_quote(); None defaults to Mid and is omitted from the mnemonic."""

    trade_component: Optional[TradeComponent] = None
    """Trade component used by update_trade(); None defaults to Price and is omitted from the mnemonic."""


Params = AutoCorrelationPeriodogramParams


def default_params() -> AutoCorrelationPeriodogramParams:
    """Return default Ehlers 2016 article parameters."""
    return AutoCorrelationPeriodogramParams()


def resolve_min_period(
    source: AutoCorrelationPeriodogramSource | str,
    min_period: Optional[int],
) -> int:
    """Resolve None min_period to 8 for article mode or 10 for book mode."""
    if min_period is not None:
        return min_period
    return 10 if source == AutoCorrelationPeriodogramSource.BOOK.value else 8


def _format_mnemonic(
    prefix: str,
    include_book_enhance: bool,
    params: Optional[AutoCorrelationPeriodogramParams] = None,
) -> str:
    configuration = default_params() if params is None else params
    bar_comp = (
        configuration.bar_component
        if configuration.bar_component is not None
        else DEFAULT_BAR_COMPONENT
    )
    quote_comp = (
        configuration.quote_component
        if configuration.quote_component is not None
        else DEFAULT_QUOTE_COMPONENT
    )
    trade_comp = (
        configuration.trade_component
        if configuration.trade_component is not None
        else DEFAULT_TRADE_COMPONENT
    )
    comp_suffix = component_triple_mnemonic(bar_comp, quote_comp, trade_comp)
    is_book = configuration.source == AutoCorrelationPeriodogramSource.BOOK.value
    resolved_min = resolve_min_period(configuration.source, configuration.min_period)
    source_tag = ", book" if is_book else ""
    enh_tag = (
        ", enh"
        if configuration.enhance_resolution and (not is_book or include_book_enhance)
        else ""
    )
    return (
        f"{prefix}({resolved_min}, {configuration.max_period}, "
        f"{configuration.averaging_length}{source_tag}{enh_tag}{comp_suffix})"
    )


def auto_correlation_periodogram_mnemonic(
    params: Optional[AutoCorrelationPeriodogramParams] = None,
) -> str:
    """Build the primary heatmap mnemonic for AutoCorrelationPeriodogram."""
    return _format_mnemonic("acp", True, params)


def auto_correlation_periodogram_dominant_cycle_mnemonic(
    params: Optional[AutoCorrelationPeriodogramParams] = None,
) -> str:
    """Build the dominant-cycle mnemonic for AutoCorrelationPeriodogram."""
    return _format_mnemonic("acp-dc", False, params)
