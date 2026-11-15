"""Zero-lag error-correcting exponential moving average parameters."""

from __future__ import annotations

from dataclasses import dataclass
import math
import sys
from typing import Optional, Union

from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT
from ...core.component_triple_mnemonic import component_triple_mnemonic

_ALPHA_EPSILON = 1e-8


@dataclass
class ZeroLagErrorCorrectingExponentialMovingAverageLengthParams:
    """Length-based parameters for ZeroLagErrorCorrectingExponentialMovingAverage."""

    length: int = 20
    """Equivalent EMA length L (integer >= 1), defining alpha = 2 / (L + 1). Default 20."""

    gain_limit: int = 50
    """Nonnegative integer gain bound in tenths (candidates span [-G/10, +G/10]). Default 50."""

    bar_component: Optional[BarComponent] = None
    """Bar component used by update_bar(); None defaults to Close and is omitted from the mnemonic."""

    quote_component: Optional[QuoteComponent] = None
    """Quote component used by update_quote(); None defaults to Mid and is omitted from the mnemonic."""

    trade_component: Optional[TradeComponent] = None
    """Trade component used by update_trade(); None defaults to Price and is omitted from the mnemonic."""


@dataclass
class ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams:
    """Smoothing-factor-based parameters for ZeroLagErrorCorrectingExponentialMovingAverage."""

    smoothing_factor: float = 2.0 / 21.0
    """Baseline EMA smoothing factor alpha in (0, 1]. Default 2/21."""

    gain_limit: int = 50
    """Nonnegative integer gain bound in tenths (candidates span [-G/10, +G/10]). Default 50."""

    bar_component: Optional[BarComponent] = None
    """Bar component used by update_bar(); None defaults to Close and is omitted from the mnemonic."""

    quote_component: Optional[QuoteComponent] = None
    """Quote component used by update_quote(); None defaults to Mid and is omitted from the mnemonic."""

    trade_component: Optional[TradeComponent] = None
    """Trade component used by update_trade(); None defaults to Price and is omitted from the mnemonic."""


ZeroLagErrorCorrectingExponentialMovingAverageParams = Union[
    ZeroLagErrorCorrectingExponentialMovingAverageLengthParams,
    ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams,
]
Params = ZeroLagErrorCorrectingExponentialMovingAverageLengthParams


def default_length_params() -> ZeroLagErrorCorrectingExponentialMovingAverageLengthParams:
    """Return default length-based ZECEMA parameters (length=20, gain_limit=50)."""
    return ZeroLagErrorCorrectingExponentialMovingAverageLengthParams()


def default_smoothing_factor_params() -> ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams:
    """Return default smoothing-factor-based ZECEMA parameters (smoothing_factor=2/21, gain_limit=50)."""
    return ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams()


def default_params() -> ZeroLagErrorCorrectingExponentialMovingAverageLengthParams:
    """Return default ZECEMA parameters."""
    return default_length_params()


def equivalent_length_from_alpha(alpha: float) -> int:
    """Compute equivalent integer length round_half_away_from_zero(2 / alpha) - 1."""
    if not (alpha >= _ALPHA_EPSILON):
        return sys.maxsize
    rounded = math.floor(2.0 / alpha + 0.5)
    if rounded >= sys.maxsize:
        return sys.maxsize
    return int(rounded) - 1


def _format_mnemonic(
    prefix: str,
    params: Optional[ZeroLagErrorCorrectingExponentialMovingAverageParams] = None,
) -> str:
    source = default_length_params() if params is None else params
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
    suffix = component_triple_mnemonic(bar_comp, quote_comp, trade_comp)
    if isinstance(
        source, ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams
    ):
        equiv_length = equivalent_length_from_alpha(source.smoothing_factor)
        return f"{prefix}({equiv_length}, {source.smoothing_factor:.8f}, {source.gain_limit}{suffix})"
    return f"{prefix}({source.length}, {source.gain_limit}{suffix})"


def zero_lag_error_correcting_exponential_moving_average_mnemonic(
    params: Optional[ZeroLagErrorCorrectingExponentialMovingAverageParams] = None,
) -> str:
    """Build the primary EC mnemonic string for ZECEMA."""
    return _format_mnemonic("zecema", params)


def zero_lag_error_correcting_exponential_moving_average_ema_mnemonic(
    params: Optional[ZeroLagErrorCorrectingExponentialMovingAverageParams] = None,
) -> str:
    """Build the baseline EMA mnemonic string for ZECEMA."""
    return _format_mnemonic("zecemaEma", params)
