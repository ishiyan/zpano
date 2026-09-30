"""Nonambiguous Trend Filter (_Trade) -- William Blau.

A post-processing transform applied to a normalized, signed base oscillator X
(TSI, SMI, DTI, TVI, MDI, CMI, CSI). It keeps X only where its sign and slope
agree and zeroes every ambiguous bar (book Ch. 8, Appendix B Figs. B-20..B-23):

    X_Trade[k] = X[k]   if  X[k] > 0  and  X[k] - X[k-1] > 0    (positive and rising)
               = X[k]   if  X[k] < 0  and  X[k] - X[k-1] < 0    (negative and falling)
               = 0      otherwise                               (ambiguous)

The nonzero stretches correspond one-to-one with genuine up/down trends;
congestion and flat regions are blanked to zero.

Conventions (Option B):
    * A NaN base value (the base's own look-back warm-up) yields NaN and leaves
      the filter state untouched.
    * The first finite base value has no prior slope, so the output is 0.0.
    * Strict slope: a flat step (delta == 0) is neither rising nor falling -> 0.

Reference:

Blau, William (1995). Momentum, Direction, and Divergence, ch. 8, Appendix B
Figs. B-20..B-23. Wiley.
"""

import math
from datetime import datetime
from typing import List, Any

from ...core.indicator import Indicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ....entities.bar_component import DEFAULT_BAR_COMPONENT
from ....entities.quote_component import DEFAULT_QUOTE_COMPONENT
from ....entities.trade_component import DEFAULT_TRADE_COMPONENT
from ..true_strength_index.true_strength_index import TrueStrengthIndex
from ..true_strength_index.params import TrueStrengthIndexParams
from ..stochastic_momentum_index.stochastic_momentum_index import StochasticMomentumIndex
from ..stochastic_momentum_index.params import StochasticMomentumIndexParams
from ..directional_trend_index.directional_trend_index import DirectionalTrendIndex
from ..directional_trend_index.params import DirectionalTrendIndexParams
from ..tick_volume_indicator.tick_volume_indicator import TickVolumeIndicator
from ..tick_volume_indicator.params import TickVolumeIndicatorParams
from ..mean_deviation_index.mean_deviation_index import MeanDeviationIndex
from ..mean_deviation_index.params import MeanDeviationIndexParams
from ..candlestick_momentum_index.candlestick_momentum_index import CandlestickMomentumIndex
from ..candlestick_momentum_index.params import CandlestickMomentumIndexParams
from ..candlestick_strength_index.candlestick_strength_index import CandlestickStrengthIndex
from ..candlestick_strength_index.params import CandlestickStrengthIndexParams
from .params import NonambiguousTrendFilterBase, NonambiguousTrendFilterParams

# Book named-instance defaults (q, r, s, u) per base; q is unused by TVI, MDI, CMI, CSI.
_DEFAULTS = {
    NonambiguousTrendFilterBase.TSI: (2, 32, 13, 3),
    NonambiguousTrendFilterBase.SMI: (32, 64, 7, 1),
    NonambiguousTrendFilterBase.DTI: (2, 28, 28, 5),
    NonambiguousTrendFilterBase.TVI: (0, 32, 32, 5),
    NonambiguousTrendFilterBase.MDI: (0, 20, 5, 3),
    NonambiguousTrendFilterBase.CMI: (0, 20, 5, 3),
    NonambiguousTrendFilterBase.CSI: (0, 32, 32, 1),
}

_BASE_MNEMONICS = {
    NonambiguousTrendFilterBase.TSI: "tsi",
    NonambiguousTrendFilterBase.SMI: "smi",
    NonambiguousTrendFilterBase.DTI: "dti",
    NonambiguousTrendFilterBase.TVI: "tvi",
    NonambiguousTrendFilterBase.MDI: "mdi",
    NonambiguousTrendFilterBase.CMI: "cmi",
    NonambiguousTrendFilterBase.CSI: "csi",
}

# The time stamp of the scalar fed to the base by the plain update(sample).
_EPOCH = datetime(1970, 1, 1)


class NonambiguousTrendFilter(Indicator):
    """William Blau's Nonambiguous Trend Filter (_Trade)."""

    def __init__(self, params: NonambiguousTrendFilterParams) -> None:
        try:
            base = NonambiguousTrendFilterBase(params.base)
        except ValueError:
            raise ValueError(
                "invalid nonambiguous trend filter parameters: "
                f"unknown base {params.base}") from None

        dq, dr, ds, du = _DEFAULTS[base]
        q = params.q if params.q != 0 else dq
        r = params.r if params.r != 0 else dr
        s = params.s if params.s != 0 else ds
        u = params.u if params.u != 0 else du

        uses_q = base in (NonambiguousTrendFilterBase.TSI,
                          NonambiguousTrendFilterBase.SMI,
                          NonambiguousTrendFilterBase.DTI)

        if uses_q and q < 1:
            raise ValueError(
                "invalid nonambiguous trend filter parameters: "
                "q should be greater than 0")
        if r < 1:
            raise ValueError(
                "invalid nonambiguous trend filter parameters: "
                "r should be greater than 0")
        if s < 1:
            raise ValueError(
                "invalid nonambiguous trend filter parameters: "
                "s should be greater than 0")
        if u < 1:
            raise ValueError(
                "invalid nonambiguous trend filter parameters: "
                "u should be greater than 0")

        # Price components are meaningful only for the single-price TSI and MDI bases.
        uses_components = base in (NonambiguousTrendFilterBase.TSI,
                                   NonambiguousTrendFilterBase.MDI)
        bc = params.bar_component if uses_components else None
        qc = params.quote_component if uses_components else None
        tc = params.trade_component if uses_components else None

        # The base signal line is unused, so its period is 1.
        if base == NonambiguousTrendFilterBase.TSI:
            self._base = TrueStrengthIndex(TrueStrengthIndexParams(
                q=q, r=r, s=s, u=u, ul=1,
                bar_component=bc, quote_component=qc, trade_component=tc))
        elif base == NonambiguousTrendFilterBase.SMI:
            self._base = StochasticMomentumIndex(StochasticMomentumIndexParams(q=q, r=r, s=s, u=u, ul=1))
        elif base == NonambiguousTrendFilterBase.DTI:
            self._base = DirectionalTrendIndex(DirectionalTrendIndexParams(q=q, r=r, s=s, u=u, ul=1))
        elif base == NonambiguousTrendFilterBase.TVI:
            self._base = TickVolumeIndicator(TickVolumeIndicatorParams(r=r, s=s, u=u))
        elif base == NonambiguousTrendFilterBase.MDI:
            self._base = MeanDeviationIndex(MeanDeviationIndexParams(
                r=r, s=s, u=u, ul=1,
                bar_component=bc, quote_component=qc, trade_component=tc))
        elif base == NonambiguousTrendFilterBase.CMI:
            self._base = CandlestickMomentumIndex(CandlestickMomentumIndexParams(r=r, s=s, u=u, ul=1))
        else:
            self._base = CandlestickStrengthIndex(CandlestickStrengthIndexParams(r=r, s=s, u=u, ul=1))

        # Filter state: the last finite base value, once one has been seen.
        self._previous = 0.0
        self._primed = False

        name = _BASE_MNEMONICS[base]
        triple = component_triple_mnemonic(
            bc if bc is not None else DEFAULT_BAR_COMPONENT,
            qc if qc is not None else DEFAULT_QUOTE_COMPONENT,
            tc if tc is not None else DEFAULT_TRADE_COMPONENT)
        if uses_q:
            self._mnemonic = f"ntf.{name}({q},{r},{s},{u}{triple})"
        else:
            self._mnemonic = f"ntf.{name}({r},{s},{u}{triple})"
        self._description = f"Nonambiguous Trend Filter {self._mnemonic}"

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.NONAMBIGUOUS_TREND_FILTER,
            self._mnemonic,
            self._description,
            [OutputText(self._mnemonic, self._description)],
        )

    def _filter(self, x: float) -> float:
        """Keep x when positive-and-rising or negative-and-falling, else 0."""
        if math.isnan(x):
            # The base is still warming up: do not touch the filter state.
            return math.nan

        if not self._primed:
            # First finite value: no prior slope, hence ambiguous.
            self._previous = x
            self._primed = True
            return 0.0

        delta = x - self._previous
        self._previous = x

        if x > 0.0 and delta > 0.0:
            return x          # positive and rising
        if x < 0.0 and delta < 0.0:
            return x          # negative and falling
        return 0.0            # ambiguous / flat / congestion

    def _wrap(self, time: datetime, output: List[Any]) -> List[Any]:
        """Filters the primary output of the base."""
        return [Scalar(time=time, value=self._filter(output[0].value))]

    def update(self, sample: float) -> float:
        """Update with a single sample value, fed to the base as a scalar."""
        return self._filter(self._base.update_scalar(Scalar(time=_EPOCH, value=sample))[0].value)

    def update_scalar(self, sample: Scalar) -> List[Any]:
        return self._wrap(sample.time, self._base.update_scalar(sample))

    def update_bar(self, sample: Bar) -> List[Any]:
        return self._wrap(sample.time, self._base.update_bar(sample))

    def update_quote(self, sample: Quote) -> List[Any]:
        return self._wrap(sample.time, self._base.update_quote(sample))

    def update_trade(self, sample: Trade) -> List[Any]:
        return self._wrap(sample.time, self._base.update_trade(sample))
